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

	var pair types.Pair
	pair, found := k.GetPairByDenoms(ctx, msg.XCoinDenom, msg.YCoinDenom)
	if !found {
		return types.SwapRequest{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair not found")
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
		ob.AddOrder(types.NewUserOrder(req))
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

func (k Keeper) RefundSwapRequest(ctx sdk.Context, pair types.Pair, req types.SwapRequest) error {
	if req.RemainingCoin.IsPositive() {
		if err := k.bankKeeper.SendCoins(ctx, pair.GetEscrowAddress(), req.GetOrderer(), sdk.NewCoins(req.RemainingCoin)); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) RefundAndDeleteSwapRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateAllPairs(ctx, func(pair types.Pair) (stop bool) {
		k.IterateSwapRequestsToBeDeletedByPair(ctx, pair.Id, func(req types.SwapRequest) (stop bool) {
			if err := k.RefundSwapRequest(ctx, pair, req); err != nil {
				panic(err)
			}
			k.DeleteSwapRequest(ctx, req.PairId, req.Id)
			return false
		})
		return false
	})
}

func (k Keeper) DeleteCancelSwapRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateCancelSwapRequestsToBeDeleted(ctx, func(req types.CancelSwapRequest) (stop bool) {
		k.DeleteCancelSwapRequest(ctx, req.PairId, req.Id)
		return false
	})
}

func (k Keeper) CancelSwapRequest(ctx sdk.Context, req types.SwapRequest) {
	req.Canceled = true
	req.ToBeDeleted = true
	k.SetSwapRequest(ctx, req)
}

func (k Keeper) MarkCancelSwapRequestToBeDeleted(ctx sdk.Context, req types.CancelSwapRequest, succeeded bool) {
	req.Succeeded = succeeded
	req.ToBeDeleted = true
	k.SetCancelSwapRequest(ctx, req)
}

// ExecuteCancelSwapRequest cancels swap requests and deletes cancel swap requests.
func (k Keeper) ExecuteCancelSwapRequest(ctx sdk.Context, req types.CancelSwapRequest) {
	swapReq, found := k.GetSwapRequest(ctx, req.PairId, req.SwapRequestId)
	if !found {
		k.MarkCancelSwapRequestToBeDeleted(ctx, req, false)
		return
	}

	if swapReq.BatchId < req.BatchId {
		if !swapReq.Canceled {
			k.CancelSwapRequest(ctx, swapReq)
		}
		k.MarkCancelSwapRequestToBeDeleted(ctx, req, true)
	}
}
