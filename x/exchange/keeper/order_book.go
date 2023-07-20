package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	opts types.ConstructMemOrderBookOptions, halveFees, simulate bool) (res types.ExecuteOrderResult, err error) {
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	payDenom, receiveDenom := market.PayReceiveDenoms(opts.IsBuy)
	res = types.NewExecuteOrderResult(payDenom, receiveDenom)
	var lastPrice sdk.Dec
	obs := k.ConstructTempOrderBookSide(ctx, market, types.ConstructMemOrderBookOptions{
		IsBuy:             !opts.IsBuy,
		PriceLimit:        opts.PriceLimit,
		QuantityLimit:     opts.QuantityLimit,
		QuoteLimit:        opts.QuoteLimit,
		MaxNumPriceLevels: 0,
	})
	for _, level := range obs.Levels {
		if opts.QuantityLimit != nil && !opts.QuantityLimit.Sub(res.ExecutedQuantity).IsPositive() {
			break
		}
		if opts.QuoteLimit != nil && !opts.QuoteLimit.Sub(res.ExecutedQuote).IsPositive() {
			break
		}

		executableQty := types.TotalExecutableQuantity(level.Orders, level.Price)
		var remainingQty sdk.Dec
		if opts.QuantityLimit != nil {
			remainingQty = opts.QuantityLimit.Sub(res.ExecutedQuantity)
		}
		if opts.QuoteLimit != nil {
			qty := opts.QuoteLimit.Sub(res.ExecutedQuote).QuoTruncate(level.Price)
			if remainingQty.IsNil() {
				remainingQty = qty
			} else {
				remainingQty = sdk.MinDec(remainingQty, qty)
			}
		}
		execQty := executableQty
		if !remainingQty.IsNil() {
			execQty = sdk.MinDec(execQty, remainingQty)
			if execQty.Equal(remainingQty) {
				res.FullyExecuted = true
			}
		}

		market.FillTempOrderBookLevel(level, execQty, level.Price, true, halveFees)
		execQuote := types.QuoteAmount(opts.IsBuy, level.Price, execQty)
		var paid, received, fee sdk.DecCoin
		if opts.IsBuy {
			paid = sdk.NewDecCoinFromDec(market.QuoteDenom, execQuote)
			deductedQty, feeQty := market.DeductTakerFee(execQty, halveFees)
			received = sdk.NewDecCoinFromDec(market.BaseDenom, deductedQty)
			fee = sdk.NewDecCoinFromDec(market.BaseDenom, feeQty)
		} else {
			paid = sdk.NewDecCoinFromDec(market.BaseDenom, execQty)
			deductedQuote, feeQuote := market.DeductTakerFee(execQuote, halveFees)
			received = sdk.NewDecCoinFromDec(market.QuoteDenom, deductedQuote)
			fee = sdk.NewDecCoinFromDec(market.QuoteDenom, feeQuote)
		}
		res.ExecutedQuantity = res.ExecutedQuantity.Add(execQty)
		res.ExecutedQuote = res.ExecutedQuote.Add(execQuote)
		res.Paid = res.Paid.Add(paid)
		res.Received = res.Received.Add(received)
		res.Fee = res.Fee.Add(fee)
		lastPrice = level.Price
	}
	if !simulate {
		var tempOrders []*types.TempOrder
		for _, level := range obs.Levels {
			tempOrders = append(tempOrders, level.Orders...)
		}
		if err = k.FinalizeMatching(ctx, market, tempOrders); err != nil {
			return
		}
		if !lastPrice.IsNil() {
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &lastPrice
			state.LastMatchingHeight = ctx.BlockHeight()
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	return
}

func (k Keeper) ConstructTempOrderBookSide(
	ctx sdk.Context, market types.Market, opts types.ConstructMemOrderBookOptions) *types.TempOrderBookSide {
	accQty := utils.ZeroDec
	accQuote := utils.ZeroDec
	// TODO: adjust price limit
	obs := types.NewTempOrderBookSide(opts.IsBuy)
	numPriceLevels := 0
	k.IterateOrderBookSideByMarket(ctx, market.Id, opts.IsBuy, false, func(order types.Order) (stop bool) {
		if opts.PriceLimit != nil &&
			((opts.IsBuy && order.Price.LT(*opts.PriceLimit)) ||
				(!opts.IsBuy && order.Price.GT(*opts.PriceLimit))) {
			return true
		}
		if opts.QuantityLimit != nil && !opts.QuantityLimit.Sub(accQty).IsPositive() {
			return true
		}
		if opts.QuoteLimit != nil && !opts.QuoteLimit.Sub(accQuote).IsPositive() {
			return true
		}
		obs.AddOrder(types.NewTempOrder(order, market, nil))
		accQty = accQty.Add(order.OpenQuantity)
		accQuote = accQuote.Add(types.QuoteAmount(!opts.IsBuy, order.Price, order.OpenQuantity))
		if opts.MaxNumPriceLevels > 0 {
			numPriceLevels++
			if numPriceLevels >= opts.MaxNumPriceLevels {
				return true
			}
		}
		return false
	})
	for _, name := range k.sourceNames {
		source := k.sources[name]
		source.ConstructMemOrderBook(ctx, market, func(ordererAddr sdk.AccAddress, price, qty sdk.Dec) error {
			// orders from OrderSource don't have id - priority among them will
			// be determined the source's name.
			order := types.NewOrder(
				0, types.OrderTypeLimit, ordererAddr, market.Id, opts.IsBuy, price, qty, 0, qty, deposit.ToDec(), ctx.BlockTime())
			order, err := k.newOrder(
				ctx, 0, types.OrderTypeLimit, market, ordererAddr, isBuy, price,
				qty, qty, ctx.BlockTime(), true)
			if err != nil {
				return err
			}
			obs.AddOrder(types.NewTempOrder(order, market, source))
			return nil
		}, opts)
	}
	if opts.MaxNumPriceLevels > 0 {
		bound := len(obs.Levels)
		if opts.MaxNumPriceLevels < bound {
			bound = opts.MaxNumPriceLevels
		}
		obs.Levels = obs.Levels[:bound]
	}
	return obs
}

func (k Keeper) FinalizeMatching(ctx sdk.Context, market types.Market, orders []*types.TempOrder) error {
	var sourceNames []string
	resultsBySourceName := map[string][]types.TempOrder{}
	for _, order := range orders {
		if order.IsUpdated && order.Source != nil {
			sourceName := order.Source.Name()
			results, ok := resultsBySourceName[sourceName]
			if !ok {
				sourceNames = append(sourceNames, sourceName)
			}
			resultsBySourceName[sourceName] = append(results, *order)
		}

		ordererAddr := order.MustGetOrdererAddress()
		if order.IsUpdated {
			if err := k.ReleaseCoins(ctx, market, ordererAddr, order.Received, true); err != nil {
				return err
			}
			if order.Source == nil {
				payDenom, receiveDenom := market.PayReceiveDenoms(order.IsBuy)
				paid := order.Paid.SubAmount(order.Received.AmountOf(payDenom))
				received := sdk.NewCoin(receiveDenom, order.Received.AmountOf(receiveDenom))
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderFilled{
					MarketId:         market.Id,
					OrderId:          order.Id,
					Orderer:          order.Orderer,
					IsBuy:            order.IsBuy,
					Price:            order.Price,
					Quantity:         order.Quantity,
					OpenQuantity:     order.OpenQuantity,
					ExecutedQuantity: order.ExecutedQuantity,
					Paid:             paid,
					Received:         received,
				}); err != nil {
					return err
				}
				// Update user orders
				if order.ExecutableQuantity(order.Price).IsZero() {
					if err := k.cancelOrder(ctx, market, order.Order, true); err != nil {
						return err
					}
				} else {
					k.SetOrder(ctx, order.Order)
				}
			}
		}
		// Should refund deposit
		if order.Source != nil && order.RemainingDeposit.IsPositive() {
			if err := k.ReleaseCoin(
				ctx, market, ordererAddr,
				market.DepositCoin(order.IsBuy, order.RemainingDeposit), true); err != nil {
				return err
			}
		}
	}
	if err := k.ExecuteSendCoins(ctx); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := resultsBySourceName[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			// TODO: pass grouped results
			source.AfterOrdersExecuted(ctx, market, results)
			totalExecQty := utils.ZeroInt
			orderers, m := types.GroupTempOrderResultsByOrderer(results) // TODO: bit redundant
			for _, orderer := range orderers {
				var (
					isBuy                  bool
					payDenom, receiveDenom string
					totalPaid              sdk.Coin
					totalReceived          sdk.Coins
				)
				for _, res := range m[orderer] {
					totalExecQty = totalExecQty.Add(res.ExecutedQuantity)
					if totalPaid.IsNil() {
						isBuy = res.IsBuy
						totalPaid = res.Paid
						totalReceived = res.Received
					} else {
						if res.IsBuy != isBuy { // sanity check
							panic("inconsistent isBuy")
						}
						totalPaid = totalPaid.Add(res.Paid)
						totalReceived = totalReceived.Add(res.Received...)
					}
				}
				payDenom, receiveDenom = market.PayReceiveDenoms(isBuy)
				paid := totalPaid.SubAmount(totalReceived.AmountOf(payDenom))
				received := sdk.NewCoin(receiveDenom, totalReceived.AmountOf(receiveDenom))
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderSourceOrdersFilled{
					MarketId:         market.Id,
					SourceName:       sourceName,
					Orderer:          orderer,
					IsBuy:            isBuy,
					ExecutedQuantity: totalExecQty,
					Paid:             paid,
					Received:         received,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
