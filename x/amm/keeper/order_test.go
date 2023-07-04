package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func BenchmarkPlaceMarketOrder(b *testing.B) {
	app := chain.Setup(false)
	ctx := app.NewContext(false, tmproto.Header{
		Height: 0,
		Time:   utils.ParseTime("2023-01-01T00:00:00Z"),
	})

	// Create market.
	creatorAddr := utils.TestAddress(1)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, creatorAddr, enoughCoins))
	market, err := app.ExchangeKeeper.CreateMarket(ctx, creatorAddr, "ucre", "uusd")
	require.NoError(b, err)
	// Create pool and add liquidity
	pool, err := app.AMMKeeper.CreatePool(ctx, creatorAddr, market.Id, utils.ParseDec("5"))
	require.NoError(b, err)
	_, _, _, err = app.AMMKeeper.AddLiquidity(
		ctx, creatorAddr, creatorAddr, pool.Id,
		exchangetypes.PriceAtTick(types.MinTick), exchangetypes.PriceAtTick(types.MaxTick),
		utils.ParseCoins("10_000000ucre,50_000000uusd"))
	require.NoError(b, err)

	// Prepare orderer.
	ordererAddr := utils.TestAddress(2)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, ordererAddr, enoughCoins))

	// Pre-run.
	isBuy := false
	for i := 0; i < 100; i++ {
		hdr := ctx.BlockHeader()
		hdr.Height++
		hdr.Time = hdr.Time.Add(5 * time.Second)
		app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
		ctx = app.NewContext(false, hdr)

		_, _, err := app.ExchangeKeeper.PlaceMarketOrder(
			ctx, market.Id, ordererAddr, isBuy, sdk.NewInt(10_000000))
		require.NoError(b, err)
		isBuy = !isBuy

		app.EndBlock(abci.RequestEndBlock{Height: ctx.BlockHeight()})
		app.Commit()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hdr := ctx.BlockHeader()
		hdr.Height++
		hdr.Time = hdr.Time.Add(5 * time.Second)
		app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
		ctx = app.NewContext(false, hdr)
		b.StartTimer()

		_, _, err := app.ExchangeKeeper.PlaceMarketOrder(
			ctx, market.Id, ordererAddr, isBuy, sdk.NewInt(10_000000))
		require.NoError(b, err)
		isBuy = !isBuy

		b.StopTimer()

		app.EndBlock(abci.RequestEndBlock{Height: ctx.BlockHeight()})
		app.Commit()
	}
}

func (s *KeeperTestSuite) TestOrderGas() {
	currentPrice := utils.ParseDec("67.855")
	market, pool := s.CreateMarketAndPool("ucre", "uusd", currentPrice)
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
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(50_000000))
	fmt.Println(s.Ctx.GasMeter().GasConsumed() - gasConsumedBefore)
}
