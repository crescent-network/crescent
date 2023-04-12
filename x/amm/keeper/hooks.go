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
		if qty.Equal(order.Quantity) {
			sqrtPrice, err := order.Price.ApproxSqrt()
			if err != nil {
				panic(err)
			}
			if order.IsBuy {
				expectedAmt0 := types.Amount0Delta(sqrtPrice, pool.CurrentSqrtPrice, pool.CurrentLiquidity)
				amt0 := qty
				diff := amt0.Sub(expectedAmt0)
				if diff.IsPositive() {
					reserveAddr:= sdk.MustAccAddressFromBech32(pool.ReserveAddress)
					if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
						ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom0, diff))); err != nil {
						panic(err)
					}
				}
			} else {
				expectedAmt1 := types.Amount1Delta(pool.CurrentSqrtPrice, sqrtPrice, pool.CurrentLiquidity)
				amt1 := order.Price.MulInt(qty).TruncateInt()
				diff := amt1.Sub(expectedAmt1)
				if diff.IsPositive() {
					reserveAddr:= sdk.MustAccAddressFromBech32(pool.ReserveAddress)
					if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
						ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom1, diff))); err != nil {
						panic(err)
					}
				}
			}
			pool.CurrentSqrtPrice = sqrtPrice
			h.k.DeletePoolOrder(ctx, pool.Id, order.MarketId, exchangetypes.TickAtPrice(*order.Price, TickPrecision))
		} else {
			var sqrtPrice sdk.Dec
			if order.IsBuy {
				sqrtPrice = types.NextSqrtPriceFromAmount1OutRoundingDown(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty)
				expectedAmt0 := types.Amount0Delta(sqrtPrice, pool.CurrentSqrtPrice, pool.CurrentLiquidity)
				amt0 := qty
				diff := amt0.Sub(expectedAmt0)
				if diff.IsPositive() {
					reserveAddr:= sdk.MustAccAddressFromBech32(pool.ReserveAddress)
					if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
						ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom0, diff))); err != nil {
						panic(err)
					}
				}
			} else {
				sqrtPrice = types.NextSqrtPriceFromAmount0OutRoundingUp(pool.CurrentSqrtPrice, pool.CurrentLiquidity, qty)
				expectedAmt1 := types.Amount1Delta(pool.CurrentSqrtPrice, sqrtPrice, pool.CurrentLiquidity)
				amt1 := order.Price.MulInt(qty).TruncateInt()
				diff := amt1.Sub(expectedAmt1)
				if diff.IsPositive() {
					reserveAddr:= sdk.MustAccAddressFromBech32(pool.ReserveAddress)
					if err := h.k.bankKeeper.SendCoinsFromAccountToModule(
						ctx, reserveAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(pool.Denom1, diff))); err != nil {
						panic(err)
					}
				}
			}
			pool.CurrentSqrtPrice = sqrtPrice
		}
		nextTick := types.TickAtSqrtPrice(pool.CurrentSqrtPrice, TickPrecision)
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
