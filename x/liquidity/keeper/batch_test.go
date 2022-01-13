package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crescentapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func TestDepositWithdraw(t *testing.T) {
	app := crescentapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	params := app.LiquidityKeeper.GetParams(ctx)

	addrs := crescentapp.AddTestAddrs(app, ctx, 2, sdk.ZeroInt())

	poolCreator := addrs[0]
	depositor := addrs[1]

	xCoin := sdk.NewCoin("denom1", sdk.NewInt(1000000))
	yCoin := sdk.NewCoin("denom2", sdk.NewInt(1000000))
	depositCoins := sdk.NewCoins(xCoin, yCoin)
	err := crescentapp.FundAccount(app.BankKeeper, ctx, poolCreator, depositCoins.Add(params.PoolCreationFee...))
	require.NoError(t, err)

	err = app.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, xCoin, yCoin))
	require.NoError(t, err)
	pool, found := app.LiquidityKeeper.GetPool(ctx, 1)
	require.True(t, found)

	xCoin = sdk.NewCoin("denom1", sdk.NewInt(1000000))
	yCoin = sdk.NewCoin("denom2", sdk.NewInt(1000000))
	depositCoins = sdk.NewCoins(xCoin, yCoin)
	err = crescentapp.FundAccount(app.BankKeeper, ctx, depositor, depositCoins)
	require.NoError(t, err)

	liquidity.BeginBlocker(ctx, app.LiquidityKeeper)
	err = app.LiquidityKeeper.DepositBatch(ctx, types.NewMsgDepositBatch(depositor, pool.Id, xCoin, yCoin))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, app.LiquidityKeeper)

	poolCoin := app.BankKeeper.GetBalance(ctx, depositor, pool.PoolCoinDenom)

	liquidity.BeginBlocker(ctx, app.LiquidityKeeper)
	err = app.LiquidityKeeper.WithdrawBatch(ctx, types.NewMsgWithdrawBatch(depositor, pool.Id, poolCoin))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, app.LiquidityKeeper)
}
