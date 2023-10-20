package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) ConstructMemOrderBookSide(
	ctx sdk.Context, market types.Market, opts types.MemOrderBookSideOptions) *types.MemOrderBookSide {
	accQty := utils.ZeroInt
	accQuote := utils.ZeroDec
	obs := types.NewMemOrderBookSide(opts.IsBuy)
	numPriceLevels := 0
	k.IterateOrderBookSide(ctx, market.Id, opts.IsBuy, opts.PriceLimit, func(price sdk.Dec, orders []types.Order) (stop bool) {
		if opts.ReachedLimit(price, accQty, accQuote, numPriceLevels) {
			return true
		}
		for _, order := range orders {
			obs.AddOrder(types.NewUserMemOrder(order))
			accQty = accQty.Add(order.OpenQuantity)
			accQuote = accQuote.Add(order.Price.MulInt(order.OpenQuantity))
		}
		numPriceLevels++
		return false
	})
	for _, name := range k.sourceNames {
		source := k.sources[name]
		if err := source.ConstructMemOrderBookSide(ctx, market, func(ordererAddr sdk.AccAddress, price sdk.Dec, qty sdk.Int) {
			deposit := types.DepositAmount(opts.IsBuy, price, qty)
			obs.AddOrder(types.NewOrderSourceMemOrder(
				ordererAddr, opts.IsBuy, price, qty, qty, deposit, source))
		}, opts); err != nil {
			panic(err)
		}
	}
	if opts.MaxNumPriceLevels > 0 {
		obs.Limit(opts.MaxNumPriceLevels)
	}
	return obs
}

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress, isBuy bool,
	opts types.MemOrderBookSideOptions, halveFees, simulate bool) (res types.ExecuteOrderResult, full bool, err error) {
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	mCtx := types.NewMatchingContext(market, halveFees)
	obs := k.ConstructMemOrderBookSide(ctx, market, opts)
	mr, full, lastPrice := mCtx.ExecuteOrder(isBuy, obs, opts.QuantityLimit, opts.QuoteLimit)
	if mr.IsMatched() {
		if !simulate {
			ledger := types.NewLedger(market.BaseDenom, market.QuoteDenom)
			payDenom, receiveDenom := types.PayReceiveDenoms(
				market.BaseDenom, market.QuoteDenom, isBuy)
			ledger.FeedMatchResult(isBuy, mr)
			ledger.Pay(ordererAddr, sdk.NewCoin(payDenom, mr.Paid))
			ledger.Receive(ordererAddr, sdk.NewCoin(receiveDenom, mr.Received))
			if err = k.finalizeMatching(ctx, market, obs.Orders(), ledger); err != nil {
				return
			}
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &lastPrice
			state.LastMatchingHeight = ctx.BlockHeight()
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	payDenom, receiveDenom := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, isBuy)
	return types.ExecuteOrderResult{
		LastPrice:        lastPrice,
		ExecutedQuantity: mr.ExecutedQuantity,
		Paid:             sdk.NewCoin(payDenom, mr.Paid),
		Received:         sdk.NewCoin(receiveDenom, mr.Received),
		FeePaid:          sdk.NewCoin(receiveDenom, mr.FeePaid),
		FeeReceived:      sdk.NewCoin(payDenom, mr.FeeReceived),
	}, full, nil
}

func (k Keeper) finalizeMatching(
	ctx sdk.Context, market types.Market, orders []*types.MemOrder,
	ledger *types.Ledger) error {
	if ledger == nil {
		ledger = types.NewLedger(market.BaseDenom, market.QuoteDenom)
	}
	var sourceNames []string
	ordersBySource := map[string][]*types.MemOrder{}
	for _, memOrder := range orders {
		if !memOrder.IsMatched() {
			continue
		}
		res := memOrder.Result()
		ledger.FeedMatchResult(memOrder.IsBuy, res)
		payDenom, receiveDenom := types.PayReceiveDenoms(
			market.BaseDenom, market.QuoteDenom, memOrder.IsBuy)
		ledger.Receive(memOrder.OrdererAddress, sdk.NewCoin(receiveDenom, res.Received))

		if memOrder.Type == types.OrderSourceMemOrder {
			ledger.Pay(memOrder.OrdererAddress, sdk.NewCoin(payDenom, res.Paid))

			sourceName := memOrder.Source.Name()
			results, ok := ordersBySource[sourceName]
			if !ok {
				sourceNames = append(sourceNames, sourceName)
			}
			ordersBySource[sourceName] = append(results, memOrder)
		}

		if memOrder.Type == types.UserMemOrder {
			order := *memOrder.Order
			order.OpenQuantity = order.OpenQuantity.Sub(res.ExecutedQuantity)
			order.RemainingDeposit = order.RemainingDeposit.Sub(res.Paid)
			if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderFilled{
				OrderId:          order.Id,
				Quantity:         order.Quantity,
				OpenQuantity:     order.OpenQuantity,
				ExecutedQuantity: res.ExecutedQuantity,
				Paid:             sdk.NewCoin(payDenom, res.Paid),
				Received:         sdk.NewCoin(receiveDenom, res.Received),
				FeePaid:          sdk.NewCoin(receiveDenom, res.FeePaid),
				FeeReceived:      sdk.NewCoin(payDenom, res.FeePaid),
			}); err != nil {
				return err
			}
			// Update user orders
			executableQty := order.ExecutableQuantity()
			if executableQty.IsZero() ||
				(!order.IsBuy && order.Price.MulInt(executableQty).TruncateDec().IsZero()) {
				if err := k.cancelOrder(ctx, market, order); err != nil {
					return err
				}
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderCompleted{
					OrderId: order.Id,
				}); err != nil {
					return err
				}
			} else {
				k.SetOrder(ctx, order)
			}
		}
	}
	if err := ledger.Transact(ctx, k.bankKeeper, market.MustGetEscrowAddress()); err != nil {
		return err
	}
	if err := k.CollectFees(ctx, market); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := ordersBySource[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			totalExecQty := utils.ZeroInt
			ordererAddrs, m := types.GroupMemOrdersByOrderer(results)
			for _, ordererAddr := range ordererAddrs {
				if err := source.AfterOrdersExecuted(ctx, market, ordererAddr, results); err != nil {
					return err
				}
				var (
					isBuy            bool
					totalPaid        sdk.Int
					totalReceived    sdk.Int
					totalFeePaid     sdk.Int // NOTE: this will always be 0
					totalFeeReceived sdk.Int
				)
				for _, order := range m[ordererAddr.String()] {
					res := order.Result()
					totalExecQty = totalExecQty.Add(res.ExecutedQuantity)
					if totalPaid.IsNil() {
						isBuy = order.IsBuy
						totalPaid = res.Paid
						totalReceived = res.Received
						totalFeePaid = res.FeePaid
						totalFeeReceived = res.FeeReceived
					} else {
						if order.IsBuy != isBuy { // sanity check
							panic("inconsistent isBuy")
						}
						totalPaid = totalPaid.Add(res.Paid)
						totalReceived = totalReceived.Add(res.Received)
						totalFeePaid = totalFeePaid.Add(res.FeePaid)
						totalFeeReceived = totalFeeReceived.Add(res.FeeReceived)
					}
				}
				payDenom, receiveDenom := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, isBuy)
				paidCoin := sdk.NewCoin(payDenom, totalPaid)
				receivedCoin := sdk.NewCoin(receiveDenom, totalReceived)
				feePaidCoin := sdk.NewCoin(receiveDenom, totalFeePaid)
				feeReceivedCoin := sdk.NewCoin(payDenom, totalFeeReceived)
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderSourceOrdersFilled{
					MarketId:         market.Id,
					SourceName:       sourceName,
					Orderer:          ordererAddr.String(),
					IsBuy:            isBuy,
					ExecutedQuantity: totalExecQty,
					Paid:             paidCoin,
					Received:         receivedCoin,
					FeePaid:          feePaidCoin,
					FeeReceived:      feeReceivedCoin,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// getBestPrice returns the best(the highest for buy and the lowest for sell)
// price on the order book.
func (k Keeper) getBestPrice(ctx sdk.Context, market types.Market, isBuy bool) (bestPrice sdk.Dec, found bool) {
	obs := k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
		IsBuy:             isBuy,
		MaxNumPriceLevels: 1,
	})
	if len(obs.Levels) > 0 {
		return obs.Levels[0].Price, true
	}
	return bestPrice, false
}
