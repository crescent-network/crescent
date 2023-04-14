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
		poolState := h.k.MustGetPoolState(ctx, pool.Id)

		var nextSqrtPrice sdk.Dec
		if order.OpenQuantity.IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(*order.Price)
			h.k.DeletePoolOrder(ctx, pool.Id, order.MarketId, exchangetypes.TickAtPrice(*order.Price, TickPrecision))
		} else { // Partially executed
			// TODO: fix nextSqrtPrice?
			if order.IsBuy {
				quote := exchangetypes.QuoteAmount(true, *order.Price, execQty)
				nextSqrtPrice = types.NextSqrtPriceFromOutput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, quote, true)
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromOutput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, execQty, false)
			}
		}

		if order.IsBuy {
			expectedAmtIn := types.Amount0DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn := execQty
			amtInDiff := amtIn.Sub(expectedAmtIn)
			if amtInDiff.IsPositive() {
				reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
				if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
					ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom0, amtInDiff))); err != nil {
					panic(err)
				}
			} else if amtInDiff.IsNegative() { // sanity check
				//panic(amtInDiff)
			}
		} else {
			expectedAmtIn := types.Amount1DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn := exchangetypes.QuoteAmount(false, *order.Price, execQty)
			amtInDiff := amtIn.Sub(expectedAmtIn)
			if amtInDiff.IsPositive() {
				reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
				if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
					ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom1, amtInDiff))); err != nil {
					panic(err)
				}
			} else if amtInDiff.IsNegative() { // sanity check
				//panic(amtInDiff)
			}
		}

		tick := exchangetypes.TickAtPrice(*order.Price, TickPrecision)
		var prevTick int32
		if order.IsBuy {
			prevTick = tick + int32(pool.TickSpacing)
		} else {
			prevTick = tick - int32(pool.TickSpacing)
		}

		poolState.CurrentSqrtPrice = nextSqrtPrice
		nextTick := types.TickAtSqrtPrice(nextSqrtPrice, TickPrecision)
		poolState.CurrentTick = nextTick

		// TODO: use previous liquidity?
		if err := h.k.PlacePoolOrder(
			ctx, pool, poolState, order.MarketId, !order.IsBuy, prevTick); err != nil {
			panic(err)
		}

		if order.OpenQuantity.IsZero() {
			// TODO: handle liquidity = 0 case
			tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, poolState.CurrentTick)
			if found { // TODO: handle tick crossing properly!
				var netLiquidity sdk.Dec
				if order.IsBuy {
					netLiquidity = tickInfo.NetLiquidity.Neg()
				} else {
					netLiquidity = tickInfo.NetLiquidity
				}
				_ = netLiquidity
				// TODO: fix liquidity calculation
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
			}
		}
		h.k.SetPoolState(ctx, pool.Id, poolState)
	}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, firstPrice, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
}
