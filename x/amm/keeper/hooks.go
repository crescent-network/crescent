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

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotOrder, execQty sdk.Int) {
	ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
	// TODO: optimize
	pool, found := h.k.GetPoolByReserveAddress(ctx, ordererAddr)
	if found {
		poolState := h.k.MustGetPoolState(ctx, pool.Id)

		var nextSqrtPrice sdk.Dec
		if order.OpenQuantity.IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(order.Price)
		} else { // Partially executed
			// TODO: fix nextSqrtPrice?
			if order.IsBuy {
				quote := exchangetypes.QuoteAmount(true, order.Price, execQty)
				nextSqrtPrice = types.NextSqrtPriceFromOutput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, quote, true)
			} else {
				nextSqrtPrice = types.NextSqrtPriceFromOutput(
					poolState.CurrentSqrtPrice, poolState.CurrentLiquidity, execQty, false)
			}
		}

		var expectedAmtIn, amtIn, amtInDiff sdk.Int
		if order.IsBuy {
			expectedAmtIn = types.Amount0DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn = execQty
			amtInDiff = amtIn.Sub(expectedAmtIn)
		} else {
			expectedAmtIn = types.Amount1DeltaRounding(
				poolState.CurrentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, false)
			amtIn = exchangetypes.QuoteAmount(false, order.Price, execQty)
			amtInDiff = amtIn.Sub(expectedAmtIn)
		}
		if amtInDiff.IsPositive() {
			coinDiff := sdk.NewCoin(pool.DenomIn(order.IsBuy), amtInDiff)
			reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
			if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
				ctx, reserveAddr, types.ModuleName, sdk.NewCoins(coinDiff)); err != nil {
				panic(err)
			}
			if order.IsBuy {
				poolState.FeeGrowthGlobal0 = poolState.FeeGrowthGlobal0.Add(amtInDiff.ToDec().QuoTruncate(poolState.CurrentLiquidity))
			} else {
				poolState.FeeGrowthGlobal1 = poolState.FeeGrowthGlobal1.Add(amtInDiff.ToDec().QuoTruncate(poolState.CurrentLiquidity))
			}
		} else if amtInDiff.IsNegative() { // sanity check
			//panic(amtInDiff)
		}

		poolState.CurrentSqrtPrice = nextSqrtPrice
		nextTick := types.TickAtSqrtPrice(nextSqrtPrice, TickPrecision)
		poolState.CurrentTick = nextTick

		if order.OpenQuantity.IsZero() {
			// TODO: handle liquidity = 0 case
			tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, poolState.CurrentTick)
			if found { // TODO: handle tick crossing properly!
				tickInfo.FeeGrowthOutside0 = poolState.FeeGrowthGlobal0.Sub(tickInfo.FeeGrowthOutside0)
				tickInfo.FeeGrowthOutside1 = poolState.FeeGrowthGlobal1.Sub(tickInfo.FeeGrowthOutside1)
				h.k.SetTickInfo(ctx, pool.Id, poolState.CurrentTick, tickInfo)
				var netLiquidity sdk.Dec
				if order.IsBuy {
					netLiquidity = tickInfo.NetLiquidity.Neg()
				} else {
					netLiquidity = tickInfo.NetLiquidity
				}
				// TODO: fix liquidity calculation
				poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(netLiquidity)
			}
		}
		h.k.SetPoolState(ctx, pool.Id, poolState)
	}
}
