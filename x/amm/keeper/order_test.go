package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestOrderGas() {
	currentPrice := utils.ParseDec("67.855")
	market, pool := s.CreateMarketAndPool("ucre", "uusd", currentPrice)
	pool.TickSpacing = 50
	s.keeper.SetPool(s.Ctx, pool)
	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	lpAddr := s.FundedAccount(1, enoughCoins)
	for _, info := range []struct {
		lowerPrice, upperPrice sdk.Dec
		liquidity              sdk.Int
	}{
		{utils.ParseDec("43.95"), utils.ParseDec("150.5"), sdk.NewInt(34708676)},
		{utils.ParseDec("78.05"), utils.ParseDec("125.5"), sdk.NewInt(2572344642)},
		{utils.ParseDec("64.60"), utils.ParseDec("164"), sdk.NewInt(96518823)},
	} {
		lowerTick := exchangetypes.TickAtPrice(info.lowerPrice)
		upperTick := exchangetypes.TickAtPrice(info.upperPrice)
		sqrtPriceA := types.SqrtPriceAtTick(lowerTick)
		sqrtPriceB := types.SqrtPriceAtTick(upperTick)
		amt0 := utils.ZeroInt
		amt1 := utils.ZeroInt
		if poolState.CurrentTick < lowerTick {
			amt0 = types.Amount0Delta(sqrtPriceA, sqrtPriceB, info.liquidity)
		} else if poolState.CurrentTick < upperTick {
			currentSqrtPrice := utils.DecApproxSqrt(poolState.CurrentPrice)
			amt0 = types.Amount0Delta(currentSqrtPrice, sqrtPriceB, info.liquidity)
			amt1 = types.Amount1Delta(sqrtPriceA, currentSqrtPrice, info.liquidity)
		} else {
			amt1 = types.Amount1Delta(sqrtPriceA, sqrtPriceB, info.liquidity)
		}
		desiredAmt := sdk.NewCoins(sdk.NewCoin(pool.Denom0, amt0), sdk.NewCoin(pool.Denom1, amt1))
		s.AddLiquidity(
			lpAddr, lpAddr, pool.Id, info.lowerPrice, info.upperPrice, desiredAmt)
	}
	ordererAddr := s.FundedAccount(2, enoughCoins)
	gasConsumedBefore := s.Ctx.GasMeter().GasConsumed()
	qtyLimit := sdk.NewInt(50_000000)
	keeper.NewOrderSource(s.keeper).GenerateOrders(s.Ctx, market, func(ordererAddr sdk.AccAddress, price sdk.Dec, qty sdk.Int) error {
		fmt.Println("createOrder", price, qty)
		return nil
	}, exchangetypes.GenerateOrdersOptions{
		IsBuy:         false,
		PriceLimit:    nil,
		QuantityLimit: &qtyLimit,
		QuoteLimit:    nil,
	})
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(50_000000))
	fmt.Println(s.Ctx.GasMeter().GasConsumed() - gasConsumedBefore)
}
