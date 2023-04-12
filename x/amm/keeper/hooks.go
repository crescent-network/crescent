package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (h Hooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotOrder, qty sdk.Int) {
	ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
	// TODO: optimize
	pool, found := h.k.GetPoolByReserveAddress(ctx, ordererAddr)
	if found {
		var nextSqrtPrice sdk.Dec
		if order.OpenQuantity.IsZero() {
			var err error
			nextSqrtPrice, err = order.Price.ApproxSqrt()
			if err != nil {
				panic(err)
			}
			// TODO: use sdk.Dec for key
			h.k.DeletePoolOrder(ctx, pool.Id, order.MarketId, exchangetypes.TickAtPrice(*order.Price, TickPrecision))
		} else {
			if order.IsBuy {
				nextSqrtPrice = types.NextSqrtPriceFromInput(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty, order.IsBuy)
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromOutput(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty, order.IsBuy)
			}
		}
		expectedQuote := types.Amount1Delta(pool.CurrentSqrtPrice, nextSqrtPrice, pool.CurrentLiquidity)
		quote := exchangetypes.QuoteAmount(!order.IsBuy, *order.Price, qty)
		var diff sdk.Int
		if order.IsBuy {
			diff = expectedQuote.Sub(quote)
		} else {
			diff = quote.Sub(expectedQuote)
		}
		if diff.IsPositive() {
			reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
			if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
				ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom1, diff))); err != nil {
				panic(err)
			}
		} else if diff.IsNegative() { // sanity check
			panic("insufficient amount in")
		}

		nextTick := types.TickAtSqrtPrice(nextSqrtPrice, TickPrecision)
		pool.CurrentSqrtPrice = nextSqrtPrice
		pool.CurrentTick = nextTick
		tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, nextTick)
		if found {
			var netLiquidity sdk.Int
			if order.IsBuy {
				netLiquidity = tickInfo.NetLiquidity
			} else {
				netLiquidity = tickInfo.NetLiquidity.Neg()
			}
			pool.CurrentLiquidity = pool.CurrentLiquidity.Add(netLiquidity)
		}
		h.k.SetPool(ctx, pool)
	}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, firstPrice, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	tickA := exchangetypes.TickAtPrice(firstPrice, TickPrecision)
	tickB := exchangetypes.TickAtPrice(lastPrice, TickPrecision)
	if tickA > tickB {
		tickA, tickB = tickB, tickA
	}
	h.k.UpdateSpotMarketOrders(ctx, market.Id, tickA, tickB)
}
