package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

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

// ExecuteDepositRequest executes a deposit request.
func (k Keeper) ExecuteDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	// TODO: check if pool is disabled

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)

	poolInfo := types.NewPoolInfo(rx, ry, ps)
	ax, ay, pc := types.DepositToPool(poolInfo, req.XCoin.Amount, req.YCoin.Amount)

	if pc.IsZero() {
		req.Succeeded = false
		req.ToBeDeleted = true
		k.SetDepositRequest(ctx, req)
		return nil
	}

	req.AcceptedXCoin = sdk.NewCoin(req.XCoin.Denom, ax)
	req.AcceptedYCoin = sdk.NewCoin(req.YCoin.Denom, ay)
	acceptedCoins := sdk.NewCoins(req.AcceptedXCoin, req.AcceptedYCoin)
	mintingCoins := sdk.NewCoins(sdk.NewCoin(pool.PoolCoinDenom, pc))

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintingCoins); err != nil {
		return err
	}

	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, pool.GetReserveAddress(), acceptedCoins)
	bulkOp.SendCoins(k.accountKeeper.GetModuleAddress(types.ModuleName), req.GetDepositor(), mintingCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	req.Succeeded = true
	req.ToBeDeleted = true
	k.SetDepositRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) RefundDepositRequest(ctx sdk.Context, req types.DepositRequest) error {
	refundingCoins := sdk.NewCoins(req.XCoin.Sub(req.AcceptedXCoin), req.YCoin.Sub(req.AcceptedYCoin))
	if !refundingCoins.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, types.GlobalEscrowAddr, req.GetDepositor(), refundingCoins); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) RefundAndDeleteDepositRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateDepositRequestsToBeDeleted(ctx, func(req types.DepositRequest) (stop bool) {
		if err := k.RefundDepositRequest(ctx, req); err != nil {
			panic(err)
		}
		k.DeleteDepositRequest(ctx, req.PoolId, req.Id)
		return false
	})
}

// ExecuteWithdrawRequest executes a withdraw request.
func (k Keeper) ExecuteWithdrawRequest(ctx sdk.Context, req types.WithdrawRequest) error {
	pool, _ := k.GetPool(ctx, req.PoolId)
	// TODO: check if pool is disabled

	rx, ry := k.GetPoolBalance(ctx, pool)
	ps := k.GetPoolCoinSupply(ctx, pool)

	poolInfo := types.NewPoolInfo(rx, ry, ps)
	params := k.GetParams(ctx)
	x, y := types.WithdrawFromPool(poolInfo, req.PoolCoin.Amount, params.WithdrawFeeRate)

	req.WithdrawnXCoin = sdk.NewCoin(pool.XCoinDenom, x)
	req.WithdrawnYCoin = sdk.NewCoin(pool.YCoinDenom, y)
	withdrawnCoins := sdk.NewCoins(req.WithdrawnXCoin, req.WithdrawnYCoin)
	burningCoins := sdk.NewCoins(req.PoolCoin)

	bulkOp := types.NewBulkSendCoinsOperation()
	bulkOp.SendCoins(types.GlobalEscrowAddr, k.accountKeeper.GetModuleAddress(types.ModuleName), burningCoins)
	bulkOp.SendCoins(pool.GetReserveAddress(), req.GetWithdrawer(), withdrawnCoins)
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burningCoins); err != nil {
		return err
	}

	req.Succeeded = true
	req.ToBeDeleted = true
	k.SetWithdrawRequest(ctx, req)
	// TODO: emit an event?
	return nil
}

func (k Keeper) RefundAndDeleteWithdrawRequestsToBeDeleted(ctx sdk.Context) {
	k.IterateWithdrawRequestsToBeDeleted(ctx, func(req types.WithdrawRequest) (stop bool) {
		// TODO: need a refund? maybe not
		k.DeleteWithdrawRequest(ctx, req.PoolId, req.Id)
		return false
	})
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
		rx, ry := k.GetPoolBalance(ctx, pool)
		poolInfo := types.NewPoolInfo(rx, ry, sdk.ZeroInt()) // Pool coin supply is not used when matching
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
						offerCoinDenom = pair.XCoinDenom
					case types.SwapDirectionSell:
						offerCoinDenom = pair.YCoinDenom
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
						demandCoinDenom = pair.YCoinDenom
					case types.SwapDirectionSell:
						demandCoinDenom = pair.XCoinDenom
					}
					demandCoin := sdk.NewCoin(demandCoinDenom, order.ReceivedAmount)
					bulkOp.SendCoins(pair.GetEscrowAddress(), order.Orderer, sdk.NewCoins(demandCoin))
				case *types.PoolOrder:
					var demandCoinDenom string
					switch order.Direction {
					case types.SwapDirectionBuy:
						demandCoinDenom = pair.YCoinDenom
					case types.SwapDirectionSell:
						demandCoinDenom = pair.XCoinDenom
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
