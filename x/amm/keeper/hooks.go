package keeper

import (
	"fmt"

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

func (h Hooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotLimitOrder, qty sdk.Int) {
	ordererAddr := sdk.MustAccAddressFromBech32(order.Orderer)
	// TODO: optimize
	pool, found := h.k.GetPoolByReserveAddress(ctx, ordererAddr)
	if found {
		fmt.Println("found pool", pool.Id)
		var tick int32
		if qty.Equal(order.Quantity) {
			sqrtPrice, err := order.Price.ApproxSqrt()
			if err != nil {
				panic(err)
			}
			pool.CurrentSqrtPrice = sqrtPrice
			tick = exchangetypes.TickAtPrice(order.Price, TickPrecision)
			fmt.Println(" fully matched", sqrtPrice, tick)
		} else {
			var sqrtPrice sdk.Dec
			if order.IsBuy {
				sqrtPrice = types.NextSqrtPriceFromAmount1OutRoundingDown(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty)
			} else {
				sqrtPrice = types.NextSqrtPriceFromAmount0OutRoundingUp(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty)
			}
			pool.CurrentSqrtPrice = sqrtPrice
			tick = types.TickAtSqrtPrice(pool.CurrentSqrtPrice, TickPrecision)
			fmt.Println(" partially matched", sqrtPrice, tick)
		}
		tickInfo, found := h.k.GetTickInfo(ctx, pool.Id, tick)
		if found {
			pool.CurrentTick = tick
			var netLiquidity sdk.Int
			if order.IsBuy {
				netLiquidity = tickInfo.NetLiquidity
			} else {
				netLiquidity = tickInfo.NetLiquidity.Neg()
			}
			pool.CurrentLiquidity = pool.CurrentLiquidity.Add(netLiquidity)
			fmt.Println("  pool current tick ->", tick, "current L ->", pool.CurrentLiquidity)
		}
		h.k.SetPool(ctx, pool)
	}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, ordererAddr sdk.AccAddress, isBuy bool, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	// Update pool orders from the market's last price to the new last price
}
