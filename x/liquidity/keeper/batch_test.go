package keeper_test

import (
	"fmt"
	"testing"
	"time"

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

func TestSwap(t *testing.T) {
	app := crescentapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := crescentapp.AddTestAddrs(app, ctx, 2, sdk.ZeroInt())

	user1, user2 := addrs[0], addrs[1]

	err := crescentapp.FundAccount(app.BankKeeper, ctx, user1, sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000)))
	require.NoError(t, err)
	err = crescentapp.FundAccount(app.BankKeeper, ctx, user2, sdk.NewCoins(sdk.NewInt64Coin("denom2", 1100)))
	require.NoError(t, err)

	err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user1, "denom1", "denom2", sdk.NewInt64Coin("denom1", 500), "denom2", sdk.MustNewDecFromStr("1.5"), 0))
	require.NoError(t, err)
	err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user2, "denom1", "denom2", sdk.NewInt64Coin("denom2", 1100), "denom1", sdk.MustNewDecFromStr("0.5"), 0))
	require.NoError(t, err)
	pair, found := app.LiquidityKeeper.GetPairByDenoms(ctx, "denom1", "denom2")
	require.True(t, found)

	liquidity.EndBlocker(ctx, app.LiquidityKeeper)

	fmt.Println(app.BankKeeper.GetAllBalances(ctx, user1))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, user2))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, pair.GetEscrowAddress()))

	liquidity.BeginBlocker(ctx, app.LiquidityKeeper)

	fmt.Println(app.BankKeeper.GetAllBalances(ctx, user1))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, user2))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, pair.GetEscrowAddress()))
}

func TestSwapWithPool(t *testing.T) {
	app := crescentapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	params := app.LiquidityKeeper.GetParams(ctx)

	addrs := crescentapp.AddTestAddrs(app, ctx, 2, sdk.ZeroInt())

	poolCreator, user := addrs[0], addrs[1]

	depositXCoin := sdk.NewInt64Coin("denom1", 1000000)
	depositYCoin := sdk.NewInt64Coin("denom2", 1000000)
	depositCoins := sdk.NewCoins(depositXCoin, depositYCoin)
	err := crescentapp.FundAccount(app.BankKeeper, ctx, poolCreator, depositCoins.Add(params.PoolCreationFee...))
	require.NoError(t, err)
	err = app.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, depositXCoin, depositYCoin))
	require.NoError(t, err)
	pool, found := app.LiquidityKeeper.GetPool(ctx, 1)
	require.True(t, found)

	err = crescentapp.FundAccount(app.BankKeeper, ctx, user, sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000000)))
	require.NoError(t, err)

	err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user, "denom1", "denom2", sdk.NewInt64Coin("denom1", 1000000), "denom2", sdk.MustNewDecFromStr("1.1"), 0))
	require.NoError(t, err)
	pair, found := app.LiquidityKeeper.GetPairByDenoms(ctx, "denom1", "denom2")
	require.True(t, found)

	st := time.Now()
	liquidity.EndBlocker(ctx, app.LiquidityKeeper)
	fmt.Println(time.Since(st))
	liquidity.BeginBlocker(ctx, app.LiquidityKeeper)

	fmt.Println(app.BankKeeper.GetAllBalances(ctx, user))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, pool.GetReserveAddress()))
	fmt.Println(app.BankKeeper.GetAllBalances(ctx, pair.GetEscrowAddress()))
}
