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

func (h Hooks) AfterOrderExecuted(ctx sdk.Context, order exchangetypes.Order, execQty sdk.Int, paid, received, fee sdk.Coin) {
	ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
	// TODO: optimize
	pool, found := h.k.GetPoolByReserveAddress(ctx, ordererAddr)
	if found {
		reserveAddr := ordererAddr
		poolState := h.k.MustGetPoolState(ctx, pool.Id)
		currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)

		if order.IsBuy {
			if _, valid := exchangetypes.ValidateTickPrice(poolState.CurrentPrice, TickPrecision); valid {
				tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, poolState.CurrentTick)
				if found {
					// Cross the tick
					poolState.CurrentLiquidity = poolState.CurrentLiquidity.Sub(tickInfo.NetLiquidity)
					tickInfo.FeeGrowthOutside0 = poolState.FeeGrowthGlobal0.Sub(tickInfo.FeeGrowthOutside0)
					tickInfo.FeeGrowthOutside1 = poolState.FeeGrowthGlobal1.Sub(tickInfo.FeeGrowthOutside1)
					h.k.SetTickInfo(ctx, pool.Id, poolState.CurrentTick, tickInfo)
				}
			}
		}

		var nextSqrtPrice, nextPrice sdk.Dec
		if order.OpenQuantity.IsZero() { // Fully executed
			nextSqrtPrice = utils.DecApproxSqrt(order.Price)
			nextPrice = order.Price
		} else { // Partially executed
			// TODO: fix nextSqrtPrice?
			nextSqrtPrice = types.NextSqrtPriceFromOutput(
				currentSqrtPrice, poolState.CurrentLiquidity, paid.Amount, order.IsBuy)
			nextPrice = nextSqrtPrice.Power(2)
		}

		var expectedAmtIn sdk.Int
		if order.IsBuy {
			expectedAmtIn = types.Amount0DeltaRounding(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		} else {
			expectedAmtIn = types.Amount1DeltaRounding(
				currentSqrtPrice, nextSqrtPrice, poolState.CurrentLiquidity, true)
		}
		amtInDiff := received.Amount.Sub(expectedAmtIn)
		if amtInDiff.IsPositive() {
			if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
				ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(received.Denom, amtInDiff))); err != nil {
				panic(err)
			}
			feeGrowth := amtInDiff.ToDec().QuoTruncate(poolState.CurrentLiquidity)
			if order.IsBuy {
				poolState.FeeGrowthGlobal0 = poolState.FeeGrowthGlobal0.Add(feeGrowth)
			} else {
				poolState.FeeGrowthGlobal1 = poolState.FeeGrowthGlobal1.Add(feeGrowth)
			}
		} else if amtInDiff.IsNegative() { // sanity check
			//panic(amtInDiff)
		}

		if fee.IsNegative() {
			if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
				ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.Neg()))); err != nil {
				panic(err)
			}
			feeGrowth := fee.Amount.Neg().ToDec().QuoTruncate(poolState.CurrentLiquidity)
			if fee.Denom == pool.Denom0 {
				poolState.FeeGrowthGlobal0 = poolState.FeeGrowthGlobal0.Add(feeGrowth)
			} else {
				poolState.FeeGrowthGlobal1 = poolState.FeeGrowthGlobal1.Add(feeGrowth)
			}
		}

		nextTick := exchangetypes.TickAtPrice(nextPrice, TickPrecision)
		if nextTick != poolState.CurrentTick {
			poolState.CurrentTick = nextTick
			if !order.IsBuy {
				tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, poolState.CurrentTick)
				if found {
					// Cross
					poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(tickInfo.NetLiquidity)
					tickInfo.FeeGrowthOutside0 = poolState.FeeGrowthGlobal0.Sub(tickInfo.FeeGrowthOutside0)
					tickInfo.FeeGrowthOutside1 = poolState.FeeGrowthGlobal1.Sub(tickInfo.FeeGrowthOutside1)
					h.k.SetTickInfo(ctx, pool.Id, poolState.CurrentTick, tickInfo)
				}
			}
		}
		poolState.CurrentPrice = nextPrice
		h.k.SetPoolState(ctx, pool.Id, poolState)
	}
}
