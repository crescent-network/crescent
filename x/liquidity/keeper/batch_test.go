package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crescentapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestDepositWithdraw() {
	k, ctx := s.keeper, s.ctx

	params := k.GetParams(ctx)

	// Create a normal pool
	creator := s.addr(0)
	xCoin, yCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	s.createPool(creator, xCoin, yCoin, true)

	pool, found := k.GetPool(ctx, 1)
	s.Require().True(found)
	s.Require().Equal(params.InitialPoolCoinSupply, s.getBalance(creator, pool.PoolCoinDenom).Amount)

	// A depositor makes a deposit
	depositor := s.addr(1)
	s.depositBatch(depositor, pool.Id, parseCoin("500000denom1"), parseCoin("500000denom2"), true)
	s.nextBlock()

	// The depositor withdraws pool coin
	poolCoin := s.getBalance(depositor, pool.PoolCoinDenom)
	s.withdrawBatch(depositor, pool.Id, poolCoin)
	s.nextBlock()
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

	_, err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user1, "denom1", "denom2", sdk.NewInt64Coin("denom1", 500), "denom2", sdk.MustNewDecFromStr("1.5"), 0))
	require.NoError(t, err)
	_, err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user2, "denom1", "denom2", sdk.NewInt64Coin("denom2", 1100), "denom1", sdk.MustNewDecFromStr("0.5"), 0))
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
	_, err = app.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, depositXCoin, depositYCoin))
	require.NoError(t, err)
	pool, found := app.LiquidityKeeper.GetPool(ctx, 1)
	require.True(t, found)

	err = crescentapp.FundAccount(app.BankKeeper, ctx, user, sdk.NewCoins(sdk.NewInt64Coin("denom1", 1000000)))
	require.NoError(t, err)

	_, err = app.LiquidityKeeper.SwapBatch(ctx, types.NewMsgSwapBatch(user, "denom1", "denom2", sdk.NewInt64Coin("denom1", 1000000), "denom2", sdk.MustNewDecFromStr("1.1"), 0))
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
