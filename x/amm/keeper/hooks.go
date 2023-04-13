package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.ExchangeHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotOrder, execQty sdk.Int) {
	ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
	// TODO: optimize
	pool, found := h.k.GetPoolByReserveAddress(ctx, ordererAddr)
	if found {
		reserveAddr := ordererAddr
		poolState := h.k.MustGetPoolState(ctx, pool.Id)

		var nextSqrtPrice sdk.Dec
		if order.OpenQuantity.IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(*order.Price)
			h.k.DeletePoolOrder(ctx, pool.Id, order.MarketId, exchangetypes.TickAtPrice(*order.Price, TickPrecision))
		} else { // Partially executed
			if order.IsBuy {
				nextSqrtPrice = types.NextSqrtPriceFromInput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, execQty, true)
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromOutput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, execQty, false)
			}
		}

		var (
			expectedAmtIn, amtIn sdk.Int
			amtInDenom           string
		)
		if order.IsBuy {
			expectedAmtIn = types.Amount0DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn = execQty
			amtInDenom = pool.Denom0
		} else {
			expectedAmtIn = types.Amount1DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn = exchangetypes.QuoteAmount(false, *order.Price, execQty)
			amtInDenom = pool.Denom1
		}
		amtInDiff := amtIn.Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
			if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
				ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(amtInDenom, amtInDiff))); err != nil {
				panic(err)
			}
		} else if amtInDiff.IsNegative() { // sanity check
			panic("insufficient amount in")
		}

		// Place a new order
		tick := exchangetypes.TickAtPrice(*order.Price, TickPrecision)
		var (
			prevTick      int32
			prevSqrtPrice sdk.Dec
		)
		if order.IsBuy {
			prevTick = tick + int32(pool.TickSpacing)
			prevSqrtPrice = types.SqrtPriceAtTick(prevTick, TickPrecision)
		} else {
			prevTick = tick - int32(pool.TickSpacing)
			prevSqrtPrice = types.SqrtPriceAtTick(prevTick, TickPrecision)
		}
		prevPrice := exchangetypes.PriceAtTick(prevTick, TickPrecision)

		// Cancel order at previous tick
		prevOrderId, found := h.k.GetPoolOrder(ctx, pool.Id, order.MarketId, prevTick)
		if found {
			if _, err := h.k.exchangeKeeper.CancelSpotOrder(ctx, reserveAddr, order.MarketId, prevOrderId); err != nil {
				panic(err)
			}
			h.k.DeletePoolOrder(ctx, pool.Id, order.MarketId, prevTick) // TODO: use cancel hook to delete pool order
		}

		var qty sdk.Int
		if order.IsBuy { // New order is a sell order
			// TODO: use previous liquidity?
			qty = utils.MinInt(
				h.k.bankKeeper.GetBalance(ctx, reserveAddr, pool.Denom0).Amount,
				types.Amount0DeltaRounding(prevSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false))
		} else { // New order is a buy order
			quote := utils.MinInt(
				h.k.bankKeeper.GetBalance(ctx, reserveAddr, pool.Denom1).Amount,
				types.Amount1DeltaRounding(prevSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false))
			qty = quote.ToDec().QuoTruncate(prevPrice).TruncateInt()
		}
		if qty.IsPositive() {
			order, execQuote, err := h.k.exchangeKeeper.PlaceSpotLimitOrder(
				ctx, reserveAddr, order.MarketId, !order.IsBuy, prevPrice, qty)
			if err != nil {
				panic(err)
			}
			if !execQuote.IsZero() { // sanity check
				panic("pool order matched with another order")
			}
			h.k.SetPoolOrder(ctx, pool.Id, order.MarketId, prevTick, order.Id)
		}

		nextTick := types.TickAtSqrtPrice(nextSqrtPrice, TickPrecision)
		poolState.CurrentSqrtPrice = nextSqrtPrice
		if nextTick != poolState.CurrentTick { // Cross the tick
			tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, nextTick)
			if found {
				var netLiquidity sdk.Int
				if order.IsBuy {
					netLiquidity = tickInfo.NetLiquidity
				} else {
					netLiquidity = tickInfo.NetLiquidity.Neg()
				}
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
			}
			poolState.CurrentTick = nextTick
		}
		h.k.SetPoolState(ctx, pool.Id, poolState)
	}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, firstPrice, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
}
