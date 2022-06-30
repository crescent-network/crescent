package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/app/params"
	"github.com/crescent-network/crescent/v2/x/farming"
	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/simulation"
	"github.com/crescent-network/crescent/v2/x/farming/types"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(false)

	ctx.WithChainID("test-chain")

	cdc := types.ModuleCdc
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 1)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{params.DefaultWeightMsgCreateFixedAmountPlan, types.ModuleName, types.TypeMsgCreateFixedAmountPlan},
		{params.DefaultWeightMsgCreateRatioPlan, types.ModuleName, types.TypeMsgCreateRatioPlan},
		{params.DefaultWeightMsgStake, types.ModuleName, types.TypeMsgStake},
		{params.DefaultWeightMsgUnstake, types.ModuleName, types.TypeMsgUnstake},
		{params.DefaultWeightMsgHarvest, types.ModuleName, types.TypeMsgHarvest},
		{params.DefaultWeightMsgRemovePlan, types.ModuleName, types.TypeMsgRemovePlan},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgCreateFixedAmountPlan tests the normal scenario of a valid message of type TypeMsgCreateFixedAmountPlan.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgCreateFixedAmountPlan(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated private plan creation fees
	feeCoins := simulation.GenPrivatePlanCreationFee(r)
	params := app.FarmingKeeper.GetParams(ctx)
	params.PrivatePlanCreationFee = feeCoins
	app.FarmingKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgCreateFixedAmountPlan(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgCreateFixedAmountPlan
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgCreateFixedAmountPlan, msg.Type())
	require.Equal(t, "plan-LfGaE", msg.Name)
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Creator)
	require.Equal(t, "1.000000000000000000stake", msg.StakingCoinWeights.String())
	require.Equal(t, "308240456testa", msg.EpochAmount.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgCreateRatioPlan tests the normal scenario of a valid message of type TypeMsgCreateRatioPlan.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgCreateRatioPlan(t *testing.T) {
	app, ctx := createTestApp(false)
	keeper.EnableRatioPlan = true

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated private plan creation fees
	feeCoins := simulation.GenPrivatePlanCreationFee(r)
	params := app.FarmingKeeper.GetParams(ctx)
	params.PrivatePlanCreationFee = feeCoins
	app.FarmingKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgCreateRatioPlan(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgCreateRatioPlan
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgCreateRatioPlan, msg.Type())
	require.Equal(t, "plan-nhwJy", msg.Name)
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Creator)
	require.Equal(t, "1.000000000000000000stake", msg.StakingCoinWeights.String())
	require.Equal(t, "0.009000000000000000", msg.EpochRatio.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgStake tests the normal scenario of a valid message of type TypeMsgStake.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgStake(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated staking creation fees
	params := app.FarmingKeeper.GetParams(ctx)
	app.FarmingKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgStake(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgStake
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgStake, msg.Type())
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Farmer)
	require.Equal(t, "912902081stake", msg.StakingCoins.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgUnstake tests the normal scenario of a valid message of type TypeMsgUnstake.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgUnstake(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated staking creation fees
	params := app.FarmingKeeper.GetParams(ctx)
	app.FarmingKeeper.SetParams(ctx, params)

	// staking must exist in order to simulate unstake
	stakingCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000))
	err := app.FarmingKeeper.Stake(ctx, accounts[0].Address, stakingCoins)
	require.NoError(t, err)

	// begin a new block and advance epoch
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})
	err = app.FarmingKeeper.AdvanceEpoch(ctx)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgUnstake(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgUnstake
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgUnstake, msg.Type())
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Farmer)
	require.Equal(t, "78973699stake", msg.UnstakingCoins.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgHarvest tests the normal scenario of a valid message of type TypeMsgHarvest.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgHarvest(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 2)

	// setup epoch days to 1 to ease the test
	params := app.FarmingKeeper.GetParams(ctx)
	params.NextEpochDays = 1
	app.FarmingKeeper.SetParams(ctx, params)

	// setup a fixed amount plan
	_, err := app.FarmingKeeper.CreateFixedAmountPlan(
		ctx,
		&types.MsgCreateFixedAmountPlan{
			Name:    "simulation-test",
			Creator: accounts[0].Address.String(),
			StakingCoinWeights: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10, 1)), // 100%
			),
			StartTime: types.ParseTime("0001-01-01T00:00:00Z"),
			EndTime:   types.ParseTime("9999-01-01T00:00:00Z"),
			EpochAmount: sdk.NewCoins(
				sdk.NewInt64Coin("pool1", 300_000_000),
			),
		},
		accounts[0].Address,
		accounts[0].Address,
		types.PlanTypePrivate,
	)
	require.NoError(t, err)

	// stake
	err = app.FarmingKeeper.Stake(ctx, accounts[1].Address, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)))
	require.NoError(t, err)

	queuedStakingAmt := app.FarmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(ctx, accounts[1].Address, sdk.DefaultBondDenom)
	require.True(t, queuedStakingAmt.IsPositive())

	// begin a new block and advance epoch
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.Day))
	farming.EndBlocker(ctx, app.FarmingKeeper)

	// check that queue coins are moved to staked coins
	_, sf := app.FarmingKeeper.GetStaking(ctx, sdk.DefaultBondDenom, accounts[1].Address)
	require.True(t, sf)

	err = app.FarmingKeeper.AdvanceEpoch(ctx)
	require.NoError(t, err)

	// execute operation
	op := simulation.SimulateMsgHarvest(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgHarvest
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgHarvest, msg.Type())
	require.Equal(t, "cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.Farmer)
	require.Equal(t, []string{"stake"}, msg.StakingCoinDenoms)
	require.Len(t, futureOperations, 0)

	balances := app.BankKeeper.GetBalance(ctx, accounts[1].Address, "pool1")
	require.Equal(t, sdk.NewInt64Coin("pool1", 100300000000), balances)
}

func TestSimulateMsgRemovePlan(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  app.LastBlockHeight() + 1,
			Time:    types.ParseTime("2022-01-01T00:00:00Z"),
			AppHash: app.LastCommitID().Hash,
		},
	})

	// Create a new terminated plan.
	_, err := app.FarmingKeeper.CreateFixedAmountPlan(
		ctx,
		&types.MsgCreateFixedAmountPlan{
			Name:    "simulation-test",
			Creator: accounts[0].Address.String(),
			StakingCoinWeights: sdk.NewDecCoins(
				sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10, 1)), // 100%
			),
			StartTime: types.ParseTime("0001-01-01T00:00:00Z"),
			EndTime:   types.ParseTime("0001-01-02T00:00:00Z"),
			EpochAmount: sdk.NewCoins(
				sdk.NewInt64Coin("pool1", 300_000_000),
			),
		},
		accounts[0].Address,
		accounts[0].Address,
		types.PlanTypePrivate,
	)
	require.NoError(t, err)
	app.EndBlock(abci.RequestEndBlock{Height: app.LastBlockHeight()})
	app.BeginBlock(abci.RequestBeginBlock{
		Header: tmproto.Header{
			Height:  app.LastBlockHeight() + 1,
			Time:    types.ParseTime("2022-01-01T00:00:00Z"),
			AppHash: app.LastCommitID().Hash,
		},
	})

	// execute operation
	op := simulation.SimulateMsgRemovePlan(app.AccountKeeper, app.BankKeeper, app.FarmingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgRemovePlan
	err = types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgRemovePlan, msg.Type())
	require.Equal(t, "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Creator)
	require.Equal(t, uint64(1), msg.PlanId)
	require.Len(t, futureOperations, 0)
}

func createTestApp(isCheckTx bool) (*chain.App, sdk.Context) {
	app := chain.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *chain.App, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 100_000_000_000)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewInt64Coin("pool1", 100_000_000_000),
	)

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		err := simapp.FundAccount(app.BankKeeper, ctx, account.Address, initCoins)
		require.NoError(t, err)
	}

	return accounts
}
