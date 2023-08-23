package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func BenchmarkPoolOrders(b *testing.B) {
	app := chain.Setup(false)
	ctx := app.NewContext(false, tmproto.Header{
		Height: 0,
		Time:   utils.ParseTime("2023-01-01T00:00:00Z"),
	})

	creatorAddr := utils.TestAddress(0)
	require.NoError(
		b, chain.FundAccount(app.BankKeeper, ctx, creatorAddr, enoughCoins))

	market, err := app.ExchangeKeeper.CreateMarket(ctx, creatorAddr, "ucre", "uusd")
	require.NoError(b, err)

	pool, err := app.AMMKeeper.CreatePool(ctx, creatorAddr, market.Id, utils.ParseDec("5"))
	require.NoError(b, err)

	lpAddr := utils.TestAddress(1)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, lpAddr, enoughCoins))

	_, _, _, err = app.AMMKeeper.AddLiquidity(
		ctx, lpAddr, lpAddr, pool.Id, types.MinPrice, types.MaxPrice,
		utils.ParseCoins("10_000000ucre,50_000000uusd"))
	require.NoError(b, err)

	ordererAddr := utils.TestAddress(2)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, ordererAddr, enoughCoins))
	b.ResetTimer()

	b.Run("buy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cacheCtx, _ := ctx.CacheContext()
			_, _, err := app.ExchangeKeeper.PlaceMarketOrder(cacheCtx, market.Id, ordererAddr, true, sdk.NewDec(5_000000))
			require.NoError(b, err)
		}
	})
	b.Run("sell", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cacheCtx, _ := ctx.CacheContext()
			_, _, err := app.ExchangeKeeper.PlaceMarketOrder(cacheCtx, market.Id, ordererAddr, false, sdk.NewDec(5_000000))
			require.NoError(b, err)
		}
	})
}

// XXX
func BenchmarkPlaceBuyMarketOrder(b *testing.B) {
	app := chain.Setup(false)
	ctx := app.NewContext(false, tmproto.Header{
		Height: 0,
		Time:   utils.ParseTime("2023-01-01T00:00:00Z"),
	})

	creatorAddr := utils.TestAddress(0)
	require.NoError(
		b, chain.FundAccount(app.BankKeeper, ctx, creatorAddr, enoughCoins))

	market, err := app.ExchangeKeeper.CreateMarket(ctx, creatorAddr, "ucre", "uusd")
	require.NoError(b, err)

	pool, err := app.AMMKeeper.CreatePool(ctx, creatorAddr, market.Id, utils.ParseDec("5"))
	require.NoError(b, err)

	lpAddr := utils.TestAddress(1)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, lpAddr, enoughCoins))

	_, _, _, err = app.AMMKeeper.AddLiquidity(
		ctx, lpAddr, lpAddr, pool.Id, types.MinPrice, types.MaxPrice,
		utils.ParseCoins("10_000000ucre,50_000000uusd"))
	require.NoError(b, err)

	ordererAddr := utils.TestAddress(2)
	require.NoError(b, chain.FundAccount(app.BankKeeper, ctx, ordererAddr, enoughCoins))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheCtx, _ := ctx.CacheContext()
		_, _, err := app.ExchangeKeeper.PlaceMarketOrder(cacheCtx, market.Id, ordererAddr, true, sdk.NewDec(100_000000))
		require.NoError(b, err)
	}
}
