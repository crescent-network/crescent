package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/x/liquidity/types"
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
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(req.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapDirection, req.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyRemainingAmount, req.RemainingCoin.String()),
			sdk.NewAttribute(types.AttributeKeyReceivedAmount, req.ReceivedCoin.String()),
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
		if req.Status.IsMatchable() {
			ob.AddOrder(types.NewUserOrder(req))
			req.Status = types.SwapRequestStatusExecuted
			k.SetSwapRequest(ctx, req)
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
		poolBuySources = append(poolBuySources, types.NewPoolOrderSource(poolInfo, poolReserveAddr, types.SwapDirectionBuy, tickPrec))
		poolSellSources = append(poolSellSources, types.NewPoolOrderSource(poolInfo, poolReserveAddr, types.SwapDirectionSell, tickPrec))
		return false
	})

	buySource := types.MergeOrderSources(append(poolBuySources, ob.OrderSource(types.SwapDirectionBuy))...)
	sellSource := types.MergeOrderSources(append(poolSellSources, ob.OrderSource(types.SwapDirectionSell))...)

	var lastPrice sdk.Dec
	if pair.LastPrice != nil {
		lastPrice = *pair.LastPrice
	} else {
		// If there is a pool, then the last price is the pool's price.
		// TODO: assuming there is only one active(not disabled) pool right now
		//   Later, the algorithm to determine the initial last price should be changed
		if len(pools) > 0 {
			lastPrice = pools[0].Price()
		} else {
			highestBuyPrice, found := buySource.HighestTick()
			if !found {
				// There is no buy order.
				return nil
			}
			lowestSellPrice, found := sellSource.LowestTick()
			if !found {
				// There is no sell order.
				return nil
			}
			lastPrice = highestBuyPrice.Add(lowestSellPrice).QuoInt64(2)
		}
	}
	lastPrice = types.PriceToTick(lastPrice, tickPrec) // TODO: remove this and make Match to handle this

	engine := types.NewMatchEngine(buySource, sellSource, tickPrec)
	ob, swapPrice, matched := engine.Match(lastPrice)

	if matched {
		orders := ob.AllOrders()
		bulkOp := types.NewBulkSendCoinsOperation()
		for _, order := range orders {
			if order.IsMatched() {
				if order, ok := order.(*types.PoolOrder); ok {
					var offerCoinDenom string
					switch order.Direction {
					case types.SwapDirectionBuy:
						offerCoinDenom = pair.QuoteCoinDenom
					case types.SwapDirectionSell:
						offerCoinDenom = pair.BaseCoinDenom
					}
					offerCoin := sdk.NewCoin(offerCoinDenom, order.Amount.Sub(order.RemainingAmount))
					bulkOp.SendCoins(order.ReserveAddress, pair.GetEscrowAddress(), sdk.NewCoins(offerCoin))
				}
			}
		}
		if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
			return err
		}
		bulkOp = types.NewBulkSendCoinsOperation()
		for _, order := range orders {
			if order.IsMatched() {
				switch order := order.(type) {
				case *types.UserOrder:
					// TODO: optimize read/write (can there be only one write?)
					req, _ := k.GetSwapRequest(ctx, pair.Id, order.RequestId)
					req.RemainingCoin.Amount = order.RemainingAmount
					req.ReceivedCoin.Amount = req.ReceivedCoin.Amount.Add(order.ReceivedAmount)
					req.Matched = true
					k.SetSwapRequest(ctx, req)

					var demandCoinDenom string
					switch order.Direction {
					case types.SwapDirectionBuy:
						demandCoinDenom = pair.BaseCoinDenom
					case types.SwapDirectionSell:
						demandCoinDenom = pair.QuoteCoinDenom
					}
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
		}
		if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
			return err
		}

		pair.LastPrice = &swapPrice
	}

	pair.CurrentBatchId++
	k.SetPair(ctx, pair)

	// TODO: emit an event?
	_ = swapPrice
	return nil
}

func (k Keeper) RefundSwapRequestAndSetStatus(ctx sdk.Context, req types.SwapRequest, status types.SwapRequestStatus) error {
	if req.Status.IsCanceledOrExpired() { // sanity check
		return nil
	}
	if req.RemainingCoin.IsPositive() {
		pair, _ := k.GetPair(ctx, req.PairId)
		if err := k.bankKeeper.SendCoins(ctx, pair.GetEscrowAddress(), req.GetOrderer(), sdk.NewCoins(req.RemainingCoin)); err != nil {
			return err
		}
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
