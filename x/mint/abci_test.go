package mint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/mint"
	"github.com/crescent-network/crescent/v2/x/mint/keeper"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	)
)

type ModuleTestSuite struct {
	suite.Suite

	app    *chain.App
	ctx    sdk.Context
	keeper keeper.Keeper
	addrs  []sdk.AccAddress
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (suite *ModuleTestSuite) SetupTest() {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.keeper = suite.app.MintKeeper
	suite.addrs = chain.AddTestAddrs(suite.app, suite.ctx, 6, sdk.ZeroInt())
	for _, addr := range suite.addrs {
		err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
		suite.Require().NoError(err)
	}
}

func (s *ModuleTestSuite) TestInitGenesis() {
	// default gent state case
	genState := types.DefaultGenesisState()
	mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, genState)
	got := mint.ExportGenesis(s.ctx, s.app.MintKeeper)
	s.Require().Equal(*genState, *got)

	// not nil last block time case
	testTime := utils.ParseTime("2023-01-01T00:00:00Z")
	genState.LastBlockTime = &testTime
	mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, genState)
	got = mint.ExportGenesis(s.ctx, s.app.MintKeeper)
	s.Require().Equal(*genState, *got)

	// invalid last block time case
	testTime2 := time.Unix(-62136697901, 0)
	genState.LastBlockTime = &testTime2
	s.Require().Panics(func() {
		mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, genState)
	})
	got = mint.ExportGenesis(s.ctx, s.app.MintKeeper)
	s.Require().NotEqual(*genState, *got)
}

func (s *ModuleTestSuite) TestImportExportGenesis() {
	k, ctx := s.keeper, s.ctx
	genState := mint.ExportGenesis(ctx, k)
	bz := s.app.AppCodec().MustMarshalJSON(genState)

	var genState2, genState5 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	mint.InitGenesis(ctx, s.app.MintKeeper, s.app.AccountKeeper, &genState2)

	genState3 := mint.ExportGenesis(ctx, k)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)

	ctx = ctx.WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))
	mint.BeginBlocker(ctx, k)
	genState4 := mint.ExportGenesis(ctx, k)
	bz = s.app.AppCodec().MustMarshalJSON(genState4)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState5)
	s.Require().Equal(*genState5.LastBlockTime, utils.ParseTime("2022-01-01T00:00:00Z"))
	mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, &genState5)
	genState6 := mint.ExportGenesis(ctx, k)
	s.Require().Equal(*genState4, genState5, genState6)
}

func TestConstantInflation(t *testing.T) {
	app := chain.Setup(false)
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

	ctx = ctx.WithBlockHeight(0).WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	// skip first block inflation, not set LastBlockTime
	require.EqualValues(t, advanceHeight(), sdk.NewInt(0))

	// after 2022-01-01 00:00:00
	// 47564687 / 5 * (365 * 24 * 60 * 60) / 300000000000000 ~= 1
	// 47564687 ~= 300000000000000 / (365 * 24 * 60 * 60) * 5
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(47564687))

	ctx = ctx.WithBlockHeight(100).WithBlockTime(utils.ParseTime("2023-01-01T00:00:00Z"))

	// applied 10sec(params.BlockTimeThreshold) block time due to block time diff is over params.BlockTimeThreshold
	require.EqualValues(t, advanceHeight(), sdk.NewInt(63419583))
	require.EqualValues(t, advanceHeight(), sdk.NewInt(31709791))

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
	ctx = ctx.WithBlockHeight(300).WithBlockTime(utils.ParseTime("2030-01-01T01:00:00Z"))
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
	require.True(t, advanceHeight().IsZero())
}

func TestChangeMintPool(t *testing.T) {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
			ChainId:       "test-chain-id",
		},
	)

	blockTime := 5 * time.Second
	params := app.MintKeeper.GetParams(ctx)
	require.EqualValues(t, params.MintPoolAddress, types.DefaultMintPoolAddress.String())
	require.EqualValues(t, params.MintPoolAddress, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName).String())

	advanceHeight := func(mintPool sdk.AccAddress) sdk.Int {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(ctx.BlockTime().Add(blockTime))
		beforeBalance := app.BankKeeper.GetBalance(ctx, mintPool, sdk.DefaultBondDenom)
		mint.BeginBlocker(ctx, app.MintKeeper)
		afterBalance := app.BankKeeper.GetBalance(ctx, mintPool, sdk.DefaultBondDenom)
		mintedAmt := afterBalance.Sub(beforeBalance)
		require.False(t, mintedAmt.IsNegative())
		return mintedAmt.Amount
	}

	ctx = ctx.WithBlockHeight(0).WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	// skip first block inflation, not set LastBlockTime
	require.EqualValues(t, advanceHeight(types.DefaultMintPoolAddress), sdk.NewInt(0))

	require.EqualValues(t, advanceHeight(types.DefaultMintPoolAddress), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(types.DefaultMintPoolAddress), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(types.MintModuleAcc).String(), sdk.NewInt(0).String())
	require.EqualValues(t, advanceHeight(types.MintModuleAcc).String(), sdk.NewInt(0).String())

	// change mint pool address to mint module account from fee collector
	params.MintPoolAddress = types.MintModuleAcc.String()
	app.MintKeeper.SetParams(ctx, params)

	require.EqualValues(t, advanceHeight(types.DefaultMintPoolAddress).String(), sdk.NewInt(0).String())
	require.EqualValues(t, advanceHeight(types.DefaultMintPoolAddress).String(), sdk.NewInt(0).String())
	require.EqualValues(t, advanceHeight(types.MintModuleAcc), sdk.NewInt(47564687))
	require.EqualValues(t, advanceHeight(types.MintModuleAcc), sdk.NewInt(47564687))
}

func (s *ModuleTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesisState()

	mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, &genState)
	got := mint.ExportGenesis(s.ctx, s.app.MintKeeper)
	s.Require().Equal(genState, *got)
}

func (s *ModuleTestSuite) TestImportExportGenesisEmpty() {
	emptyParams := types.DefaultParams()
	emptyParams.InflationSchedules = []types.InflationSchedule{}
	s.app.MintKeeper.SetParams(s.ctx, emptyParams)
	genState := mint.ExportGenesis(s.ctx, s.app.MintKeeper)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	mint.InitGenesis(s.ctx, s.app.MintKeeper, s.app.AccountKeeper, &genState2)

	genState3 := mint.ExportGenesis(s.ctx, s.app.MintKeeper)
	s.Require().Equal(*genState, genState2)
	s.Require().EqualValues(genState2, *genState3)
}
