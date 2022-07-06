package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func BenchmarkMatching(b *testing.B) {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	keeper := app.LiquidityKeeper

	for i := 0; i < 2; i++ {
		require.NoError(b, chain.FundAccount(
			app.BankKeeper, ctx, utils.TestAddress(i),
			utils.ParseCoins("9999999999999999denom1,9999999999999999denom2,9999999999999999stake")))
	}

	pair, err := keeper.CreatePair(ctx, types.NewMsgCreatePair(utils.TestAddress(0), "denom1", "denom2"))
	require.NoError(b, err)
	pair.LastPrice = utils.ParseDecP("0.99999")
	keeper.SetPair(ctx, pair)

	_, err = keeper.CreatePool(ctx, types.NewMsgCreatePool(
		utils.TestAddress(0), pair.Id, utils.ParseCoins("1000_000000denom1,1000_00000denom2")))
	require.NoError(b, err)

	_, err = keeper.CreateRangedPool(ctx, types.NewMsgCreateRangedPool(
		utils.TestAddress(0), pair.Id, utils.ParseCoins("1000_000000denom1,1000_000000denom2"),
		utils.ParseDec("0.95"), utils.ParseDec("1.05"), utils.ParseDec("1.02")))
	require.NoError(b, err)

	_, err = keeper.CreateRangedPool(ctx, types.NewMsgCreateRangedPool(
		utils.TestAddress(0), pair.Id, utils.ParseCoins("1000_000000denom1,1000_000000denom2"),
		utils.ParseDec("0.9"), utils.ParseDec("1.2"), utils.ParseDec("0.98")))
	require.NoError(b, err)

	amt := sdk.NewInt(50_000000)
	price := utils.ParseDec("1.05")
	_, err = keeper.LimitOrder(ctx, types.NewMsgLimitOrder(
		utils.TestAddress(1), pair.Id, types.OrderDirectionSell,
		sdk.NewCoin("denom1", amt), "denom2", price, amt, 0))
	require.NoError(b, err)

	amt = sdk.NewInt(100_000000)
	price = utils.ParseDec("0.97")
	_, err = keeper.LimitOrder(ctx, types.NewMsgLimitOrder(
		utils.TestAddress(1), pair.Id, types.OrderDirectionBuy,
		sdk.NewCoin("denom2", amm.OfferCoinAmount(amm.Buy, price, amt)), "denom1", price, amt, 0))
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheCtx, _ := ctx.CacheContext()
		liquidity.EndBlocker(cacheCtx, keeper)
	}
}
