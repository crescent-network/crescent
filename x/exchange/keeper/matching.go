package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) ConstructMemOrderBookSide(
	ctx sdk.Context, market types.Market, opts types.MemOrderBookSideOptions, escrow *types.Escrow) *types.MemOrderBookSide {
	accQty := utils.ZeroDec
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
			accQuote = accQuote.Add(types.QuoteAmount(!opts.IsBuy, order.Price, order.OpenQuantity))
		}
		numPriceLevels++
		return false
	})
	for _, name := range k.sourceNames {
		source := k.sources[name]
		if err := source.ConstructMemOrderBookSide(ctx, market, func(ordererAddr sdk.AccAddress, price, qty, openQty sdk.Dec) {
			deposit := types.DepositAmount(opts.IsBuy, price, openQty)
			if escrow != nil {
				payDenom, _ := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, opts.IsBuy)
				escrow.Lock(ordererAddr, sdk.NewDecCoinFromDec(payDenom, deposit))
			}
			obs.AddOrder(types.NewOrderSourceMemOrder(ordererAddr, opts.IsBuy, price, qty, openQty, source))
		}, opts); err != nil {
			panic(err)
		}
	}
	// You can optimize performance by only accepting up to opts.MaxNumPriceLevels for each user order and order source order, and then applying a limit.
	if opts.MaxNumPriceLevels > 0 {
		// TODO: can refund?
		obs.Limit(opts.MaxNumPriceLevels)
	}
	return obs
}

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	opts types.MemOrderBookSideOptions, halveFees, simulate bool) (res types.ExecuteOrderResult, err error) {
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	escrow := types.NewEscrow(market.MustGetEscrowAddress())
	mCtx := types.NewMatchingContext(market, halveFees)
	obs := k.ConstructMemOrderBookSide(ctx, market, opts, escrow)
	res = mCtx.ExecuteOrder(obs, opts.QuantityLimit, opts.QuoteLimit)
	if res.Executed() {
		res.Paid.Amount = res.Paid.Amount.Ceil()
		res.Received.Amount = res.Received.Amount.TruncateDec()
		// TODO fee?
		if !simulate {
			escrow.Lock(ordererAddr, res.Paid)
			escrow.Unlock(ordererAddr, res.Received)
			if err = k.finalizeMatching(ctx, market, obs.Orders(), escrow); err != nil {
				return
			}
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &res.LastPrice
			state.LastMatchingHeight = ctx.BlockHeight()
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	return
}

func (k Keeper) finalizeMatching(ctx sdk.Context, market types.Market, orders []*types.MemOrder, escrow *types.Escrow) error {
	if escrow == nil {
		escrow = types.NewEscrow(market.MustGetEscrowAddress())
	}
	var sourceNames []string
	ordersBySource := map[string][]*types.MemOrder{}
	for _, memOrder := range orders {
		if memOrder.IsMatched() && memOrder.Type() == types.OrderSourceMemOrder {
			sourceName := memOrder.Source().Name()
			results, ok := ordersBySource[sourceName]
			if !ok {
				sourceNames = append(sourceNames, sourceName)
			}
			ordersBySource[sourceName] = append(results, memOrder)
		}

		payDenom, receiveDenom := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, memOrder.IsBuy())
		ordererAddr := memOrder.OrdererAddress()
		if memOrder.IsMatched() {
			receivedCoin := sdk.NewDecCoinFromDec(receiveDenom, memOrder.Received())
			if memOrder.Type() == types.UserMemOrder {
				paid := memOrder.Paid().Ceil()
				receivedCoin.Amount = receivedCoin.Amount.TruncateDec()
				order := memOrder.Order()
				order.OpenQuantity = order.OpenQuantity.Sub(memOrder.ExecutedQuantity())
				order.RemainingDeposit = order.RemainingDeposit.Sub(paid)
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderFilled{
					MarketId:         market.Id,
					OrderId:          order.Id,
					Orderer:          order.Orderer,
					IsBuy:            order.IsBuy,
					Price:            order.Price,
					Quantity:         order.Quantity,
					OpenQuantity:     order.OpenQuantity,
					ExecutedQuantity: memOrder.ExecutedQuantity(),
					Paid:             sdk.NewDecCoinFromDec(payDenom, paid),
					Received:         receivedCoin,
				}); err != nil {
					return err
				}
				// Update user orders
				executableQty := order.ExecutableQuantity()
				if executableQty.TruncateDec().IsZero() || // How about adding order.IsBuy? And it works fine in the intended order, but adding parentheses seems like a good idea.
					!order.IsBuy && executableQty.MulTruncate(order.Price).TruncateDec().IsZero() {
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
			escrow.Unlock(ordererAddr, receivedCoin)
		}
		// Should refund deposit
		if memOrder.Type() == types.OrderSourceMemOrder && memOrder.RemainingDeposit().IsPositive() {
			escrow.Unlock(ordererAddr, sdk.NewDecCoinFromDec(payDenom, memOrder.RemainingDeposit()))
		}
	}
	if err := escrow.Transact(ctx, k.bankKeeper); err != nil {
		return err
	}
	for _, sourceName := range sourceNames {
		results := ordersBySource[sourceName]
		if len(results) > 0 {
			source := k.sources[sourceName]
			// TODO: pass grouped results
			totalExecQty := utils.ZeroDec
			ordererAddrs, m := types.GroupMemOrdersByOrderer(results) // TODO: bit redundant
			for _, ordererAddr := range ordererAddrs {
				if err := source.AfterOrdersExecuted(ctx, market, ordererAddr, results); err != nil {
					return err
				}
				var (
					isBuy         bool
					totalPaid     sdk.Dec
					totalReceived sdk.Dec
					totalFee      sdk.Dec
				)
				for _, order := range m[ordererAddr.String()] {
					totalExecQty = totalExecQty.Add(order.ExecutedQuantity())
					if totalPaid.IsNil() {
						isBuy = order.IsBuy()
						totalPaid = order.Paid()
						totalReceived = order.Received()
						totalFee = order.Fee()
					} else {
						if order.IsBuy() != isBuy { // sanity check
							panic("inconsistent isBuy")
						}
						totalPaid = totalPaid.Add(order.Paid())
						totalReceived = totalReceived.Add(order.Received())
						totalFee = totalFee.Add(order.Fee())
					}
				}
				payDenom, receiveDenom := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, isBuy)
				paid := sdk.NewDecCoinFromDec(payDenom, totalPaid)
				if totalFee.IsNegative() {
					paid.Amount = paid.Amount.Add(totalFee)
				}
				paid.Amount = paid.Amount.Ceil()
				received := sdk.NewDecCoinFromDec(receiveDenom, totalReceived.TruncateDec())
				if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderSourceOrdersFilled{
					MarketId:         market.Id,
					SourceName:       sourceName,
					Orderer:          ordererAddr.String(),
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

// getBestPrice returns the best(the highest for buy and the lowest for sell)
// price on the order book.
func (k Keeper) getBestPrice(ctx sdk.Context, market types.Market, isBuy bool) (bestPrice sdk.Dec, found bool) {
	obs := k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
		IsBuy:             isBuy,
		MaxNumPriceLevels: 1,
	}, nil)
	if len(obs.Levels()) > 0 {
		return obs.Levels()[0].Price(), true
	}
	return bestPrice, false
}
