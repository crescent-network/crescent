package mint_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmosquad-labs/squad/app"
	simapp "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/mint"
	"github.com/cosmosquad-labs/squad/x/mint/keeper"
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	)
)

type ModuleTestSuite struct {
	suite.Suite

	app    *simapp.SquadApp
	ctx    sdk.Context
	keeper keeper.Keeper
	addrs  []sdk.AccAddress
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (suite *ModuleTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.keeper = suite.app.MintKeeper
	suite.addrs = simapp.AddTestAddrs(suite.app, suite.ctx, 6, sdk.ZeroInt())
	for _, addr := range suite.addrs {
		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
		suite.Require().NoError(err)
	}
}

func TestConstantInflation(t *testing.T) {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
			ChainId:       "test-chain-id",
		},
	)

	blockTime := 5 * time.Second

	feeCollector := app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	advanceHeight := func() sdk.Int {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(blockTime))
		beforeBalance := app.BankKeeper.GetBalance(ctx, feeCollector, sdk.DefaultBondDenom)
		mint.BeginBlocker(ctx, app.MintKeeper)
		afterBalance := app.BankKeeper.GetBalance(ctx, feeCollector, sdk.DefaultBondDenom)
		mintedAmt := afterBalance.Sub(beforeBalance)
		require.False(t, mintedAmt.IsNegative())
		return mintedAmt.Amount
	}

	//ctx = ctx.WithBlockHeight(0).WithBlockTime(liquidstakingtypes.MustParseRFC3339("2021-12-31T23:59:50Z"))
	ctx = ctx.WithBlockHeight(0).WithBlockTime(squadtypes.MustParseRFC3339("2022-01-01T00:00:00Z"))

	// skip first block inflation, not set LastBlockTime
	require.EqualValues(t, advanceHeight(), sdk.NewInt(0))

	// after 2022-01-01 00:00:00
	// 47564687 / 5 * (365 * 24 * 60 * 60) / 300000000000000 ~= 1
	// 47564687 ~= 300000000000000 / (365 * 24 * 60 * 60) * 5
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))

	ctx = ctx.WithBlockHeight(100).WithBlockTime(squadtypes.MustParseRFC3339("2022-12-31T23:59:50Z"))

	// applied 10sec(params.BlockTimeThreshold) block time due to block time diff is over params.BlockTimeThreshold
	require.EqualValues(t, advanceHeight(), sdk.NewInt(95129375))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))

	// 317097919 / 5 * (365 * 24 * 60 * 60) / 200000000000000 ~= 1
	// 317097919 ~= 200000000000000 / (365 * 24 * 60 * 60) * 5
	require.EqualValues(t, advanceHeight(), sdk.NewInt(31709791))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(31709791))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(31709791))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(31709791))

	blockTime = 10 * time.Second
	// 634195839 / 10 * (365 * 24 * 60 * 60) / 200000000000000 ~= 1
	// 634195839 ~= 200000000000000 / (365 * 24 * 60 * 60) * 10
	require.EqualValues(t, advanceHeight(), sdk.NewInt(63419583))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(63419583))

	// over BlockTimeThreshold 10sec
	blockTime = 20 * time.Second
	require.EqualValues(t, advanceHeight(), sdk.NewInt(63419583))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(63419583))

	// no inflation
	ctx = ctx.WithBlockHeight(300).WithBlockTime(squadtypes.MustParseRFC3339("2030-01-01T01:00:00Z"))
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
}
