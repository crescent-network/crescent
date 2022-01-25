package keeper

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// SwapBatch handles types.MsgSwapBatch and stores it.
func (k Keeper) SwapBatch(ctx sdk.Context, msg *types.MsgSwapBatch) (types.SwapRequest, error) {
	params := k.GetParams(ctx)

	if price := types.PriceToTick(msg.Price, int(params.TickPrecision)); !msg.Price.Equal(price) {
		return types.SwapRequest{}, types.ErrInvalidPriceTick
	}

	if msg.OrderLifespan > params.MaxOrderLifespan {
		return types.SwapRequest{}, types.ErrTooLongOrderLifespan
	}
	canceledAt := ctx.BlockTime().Add(msg.OrderLifespan)

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found {
		return types.SwapRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair not found")
	}

	switch msg.Direction {
	case types.SwapDirectionBuy:
		if msg.OfferCoin.Denom != pair.QuoteCoinDenom || msg.DemandCoinDenom != pair.BaseCoinDenom {
			return types.SwapRequest{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
	case types.SwapDirectionSell:
		if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
			return types.SwapRequest{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
	}

	if pair.LastPrice != nil {
		lastPrice := *pair.LastPrice
		switch {
		case msg.Price.GT(lastPrice):
			priceLimit := msg.Price.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio))
			if msg.Price.GT(priceLimit) {
				return types.SwapRequest{}, types.ErrPriceOutOfRange
			}
		case msg.Price.LT(lastPrice):
			priceLimit := msg.Price.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio))
			if msg.Price.LT(priceLimit) {
				return types.SwapRequest{}, types.ErrPriceOutOfRange
			}
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(msg.OfferCoin)); err != nil {
		return types.SwapRequest{}, err
	}

	requestId := k.GetNextSwapRequestIdWithUpdate(ctx, pair)
	req := types.NewSwapRequest(msg, requestId, pair, canceledAt, ctx.BlockHeight())
	k.SetSwapRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSwapBatch,
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, msg.OfferCoin.String()),
			sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, msg.Price.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(req.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyExpireAt, req.ExpireAt.Format(time.RFC3339)),
		),
	})

	return req, nil
}

// CancelSwapBatch handles types.MsgCancelSwapBatch and stores it.
func (k Keeper) CancelSwapBatch(ctx sdk.Context, msg *types.MsgCancelSwapBatch) (types.CancelSwapRequest, error) {
	swapReq, found := k.GetSwapRequest(ctx, msg.PairId, msg.SwapRequestId)
	if !found {
		return types.CancelSwapRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "swap request with id %d in pair %d not found", msg.SwapRequestId, msg.PairId)
	}

	if msg.Orderer != swapReq.Orderer {
		return types.CancelSwapRequest{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "mismatching orderer")
	}

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found { // TODO: will it ever happen?
		return types.CancelSwapRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair with id %d not found", msg.PairId)
	}

	requestId := k.GetNextCancelSwapRequestIdWithUpdate(ctx, pair)
	req := types.NewCancelSwapRequest(msg, requestId, pair, ctx.BlockHeight())
	k.SetCancelSwapRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelSwapBatch,
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(req.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapRequestId, strconv.FormatUint(req.SwapRequestId, 10)),
		),
	})

	return req, nil
}

func (k Keeper) ExecuteMatching(ctx sdk.Context, pair types.Pair) error {
	params := k.GetParams(ctx)
	tickPrec := int(params.TickPrecision)

	ob := types.NewOrderBook(tickPrec)
	k.IterateSwapRequestsByPair(ctx, pair.Id, func(req types.SwapRequest) (stop bool) {
		switch req.Status {
		case types.SwapRequestStatusNotExecuted,
			types.SwapRequestStatusNotMatched,
			types.SwapRequestStatusPartiallyMatched:
			if !ctx.BlockTime().Before(req.ExpireAt) {
				if err := k.RefundSwapRequestAndSetStatus(ctx, req, types.SwapRequestStatusExpired); err != nil {
					panic(err)
				}
				return false
			}
			ob.AddOrder(types.NewUserOrder(req))
			if req.Status == types.SwapRequestStatusNotExecuted {
				req.Status = types.SwapRequestStatusNotMatched
				k.SetSwapRequest(ctx, req)
			}
		default:
			panic(fmt.Sprintf("invalid swap request status: %s", req.Status))
		}
		return false
	})

	var pools []types.PoolI
	var poolBuySources, poolSellSources []types.OrderSource
	k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool) {
		rx, ry := k.GetPoolBalance(ctx, pool, pair)
		ps := k.GetPoolCoinSupply(ctx, pool)
		poolInfo := types.NewPoolInfo(rx, ry, ps) // Pool coin supply is not used when matching
		if types.IsDepletedPool(poolInfo) {
			k.MarkPoolAsDisabled(ctx, pool)
			return false
		}
		pools = append(pools, poolInfo)

		poolReserveAddr := pool.GetReserveAddress()
		poolBuySources = append(poolBuySources, types.NewPoolOrderSource(poolInfo, pool.Id, poolReserveAddr, types.SwapDirectionBuy, tickPrec))
		poolSellSources = append(poolSellSources, types.NewPoolOrderSource(poolInfo, pool.Id, poolReserveAddr, types.SwapDirectionSell, tickPrec))
		return false
	})

	buySource := types.MergeOrderSources(append(poolBuySources, ob.OrderSource(types.SwapDirectionBuy))...)
	sellSource := types.MergeOrderSources(append(poolSellSources, ob.OrderSource(types.SwapDirectionSell))...)

	engine := types.NewMatchEngine(buySource, sellSource, tickPrec)
	ob, matchPrice, quoteCoinDustAmt, matched := engine.Match()

	if matched {
		orders := ob.AllOrders()
		bulkOp := types.NewBulkSendCoinsOperation()
		for _, order := range orders {
			if order, ok := order.(*types.PoolOrder); ok {
				var offerCoinDenom string
				switch order.Direction {
				case types.SwapDirectionBuy:
					offerCoinDenom = pair.QuoteCoinDenom
				case types.SwapDirectionSell:
					offerCoinDenom = pair.BaseCoinDenom
				}
				paidCoin := sdk.NewCoin(offerCoinDenom, order.OfferCoinAmount.Sub(order.RemainingOfferCoinAmount))
				bulkOp.SendCoins(order.ReserveAddress, pair.GetEscrowAddress(), sdk.NewCoins(paidCoin))
			}
		}
		if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
			return err
		}
		bulkOp = types.NewBulkSendCoinsOperation()
		for _, order := range orders {
			switch order := order.(type) {
			case *types.UserOrder:
				var offerCoinDenom, demandCoinDenom string
				switch order.Direction {
				case types.SwapDirectionBuy:
					offerCoinDenom = pair.QuoteCoinDenom
					demandCoinDenom = pair.BaseCoinDenom
				case types.SwapDirectionSell:
					offerCoinDenom = pair.BaseCoinDenom
					demandCoinDenom = pair.QuoteCoinDenom
				}

				// TODO: optimize read/write (can there be only one write?)
				req, _ := k.GetSwapRequest(ctx, pair.Id, order.RequestId)
				req.OpenAmount = order.OpenAmount
				req.RemainingOfferCoin = sdk.NewCoin(offerCoinDenom, order.RemainingOfferCoinAmount)
				req.ReceivedCoin.Amount = req.ReceivedCoin.Amount.Add(order.ReceivedAmount)
				if order.OpenAmount.IsZero() {
					req.Status = types.SwapRequestStatusCompleted
				} else {
					req.Status = types.SwapRequestStatusPartiallyMatched
				}
				k.SetSwapRequest(ctx, req)

				demandCoin := sdk.NewCoin(demandCoinDenom, order.ReceivedAmount)
				bulkOp.SendCoins(pair.GetEscrowAddress(), order.Orderer, sdk.NewCoins(demandCoin))
			case *types.PoolOrder:
				var demandCoinDenom string
				switch order.Direction {
				case types.SwapDirectionBuy:
					demandCoinDenom = pair.BaseCoinDenom
				case types.SwapDirectionSell:
					demandCoinDenom = pair.QuoteCoinDenom
				}
				demandCoin := sdk.NewCoin(demandCoinDenom, order.ReceivedAmount)
				bulkOp.SendCoins(pair.GetEscrowAddress(), order.ReserveAddress, sdk.NewCoins(demandCoin))
			}
		}
		if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
			return err
		}

		pair.LastPrice = &matchPrice
	}

	pair.CurrentBatchId++
	k.SetPair(ctx, pair)

	// TODO: emit an event?
	_ = matchPrice
	_ = quoteCoinDustAmt
	return nil
}

func (k Keeper) RefundSwapRequest(ctx sdk.Context, req types.SwapRequest) error {
	if req.Status.IsCanceledOrExpired() { // sanity check
		return nil
	}
	if req.RemainingOfferCoin.IsPositive() {
		pair, _ := k.GetPair(ctx, req.PairId)
		if err := k.bankKeeper.SendCoins(ctx, pair.GetEscrowAddress(), req.GetOrderer(), sdk.NewCoins(req.RemainingOfferCoin)); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) RefundSwapRequestAndSetStatus(ctx sdk.Context, req types.SwapRequest, status types.SwapRequestStatus) error {
	if err := k.RefundSwapRequest(ctx, req); err != nil {
		return err
	}
	req.Status = status
	k.SetSwapRequest(ctx, req)
	return nil
}

// ExecuteCancelSwapRequest cancels and refunds swap request.
func (k Keeper) ExecuteCancelSwapRequest(ctx sdk.Context, req types.CancelSwapRequest) error {
	swapReq, found := k.GetSwapRequest(ctx, req.PairId, req.SwapRequestId)
	if !found {
		req.Status = types.RequestStatusFailed
		k.SetCancelSwapRequest(ctx, req)
		return nil
	}

	if swapReq.BatchId < req.BatchId {
		if !swapReq.Status.IsCanceledOrExpired() {
			if err := k.RefundSwapRequestAndSetStatus(ctx, swapReq, types.SwapRequestStatusCanceled); err != nil {
				return err
			}
			req.Status = types.RequestStatusSucceeded
		} else {
			req.Status = types.RequestStatusFailed
		}
		k.SetCancelSwapRequest(ctx, req)
	}

	// if swapReq.BatchId == req.BatchId, then do not change the cancel swap
	// request's status and just wait for next batch.
	return nil
}
