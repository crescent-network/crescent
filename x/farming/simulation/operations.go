package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/app/params"
	utils "github.com/crescent-network/crescent/types"
	farmingkeeper "github.com/crescent-network/crescent/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
	minttypes "github.com/crescent-network/crescent/x/mint/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreateFixedAmountPlan = "op_weight_msg_create_fixed_amount_plan"
	OpWeightMsgCreateRatioPlan       = "op_weight_msg_create_ratio_plan"
	OpWeightMsgStake                 = "op_weight_msg_stake"
	OpWeightMsgUnstake               = "op_weight_msg_unstake"
	OpWeightMsgHarvest               = "op_weight_msg_harvest"
	OpWeightMsgRemovePlan            = "op_weight_msg_remove_plan"
)

var (
	Gas  = uint64(20000000)
	Fees = sdk.Coins{
		{
			Denom:  "stake",
			Amount: sdk.NewInt(0),
		},
	}
)

var (
	poolCoinDenoms = []string{
		"pool1",
		"pool2",
		"pool3",
	}

	testCoinDenoms = []string{
		"testa",
		"testb",
		"testc",
	}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak farmingtypes.AccountKeeper,
	bk farmingtypes.BankKeeper, k farmingkeeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgCreateFixedAmountPlan int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateFixedAmountPlan, &weightMsgCreateFixedAmountPlan, nil,
		func(_ *rand.Rand) {
			weightMsgCreateFixedAmountPlan = appparams.DefaultWeightMsgCreateFixedAmountPlan
		},
	)

	var weightMsgCreateRatioPlan int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateRatioPlan, &weightMsgCreateRatioPlan, nil,
		func(_ *rand.Rand) {
			weightMsgCreateRatioPlan = appparams.DefaultWeightMsgCreateRatioPlan
		},
	)

	var weightMsgStake int
	appParams.GetOrGenerate(cdc, OpWeightMsgStake, &weightMsgStake, nil,
		func(_ *rand.Rand) {
			weightMsgStake = appparams.DefaultWeightMsgStake
		},
	)

	var weightMsgUnstake int
	appParams.GetOrGenerate(cdc, OpWeightMsgUnstake, &weightMsgUnstake, nil,
		func(_ *rand.Rand) {
			weightMsgUnstake = appparams.DefaultWeightMsgUnstake
		},
	)

	var weightMsgHarvest int
	appParams.GetOrGenerate(cdc, OpWeightMsgHarvest, &weightMsgHarvest, nil,
		func(_ *rand.Rand) {
			weightMsgHarvest = appparams.DefaultWeightMsgHarvest
		},
	)

	var weightMsgRemovePlan int
	appParams.GetOrGenerate(cdc, OpWeightMsgRemovePlan, &weightMsgRemovePlan, nil,
		func(r *rand.Rand) {
			weightMsgRemovePlan = appparams.DefaultWeightMsgRemovePlan
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateFixedAmountPlan,
			SimulateMsgCreateFixedAmountPlan(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateRatioPlan,
			SimulateMsgCreateRatioPlan(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgStake,
			SimulateMsgStake(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUnstake,
			SimulateMsgUnstake(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgHarvest,
			SimulateMsgHarvest(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRemovePlan,
			SimulateMsgRemovePlan(ak, bk, k),
		),
	}
}

// SimulateMsgCreateFixedAmountPlan generates a MsgCreateFixedAmountPlan with random values
// nolint: interfacer
func SimulateMsgCreateFixedAmountPlan(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateFixedAmountPlan, "insufficient balance for plan creation fee"), nil, nil
		}

		name := "simulation-test-" + simtypes.RandStringOfLength(r, 5) // name must be unique
		creatorAcc := account.GetAddress()
		// mint pool coins to simulate the real-world cases
		funds, err := fundBalances(ctx, r, bk, creatorAcc, testCoinDenoms)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateFixedAmountPlan, "unable to mint pool coins"), nil, nil
		}
		stakingCoinWeights := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
		startTime := ctx.BlockTime()
		endTime := startTime.AddDate(1, 0, 0)
		epochAmount := sdk.NewCoins(
			sdk.NewInt64Coin(funds[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))),
		)

		msg := farmingtypes.NewMsgCreateFixedAmountPlan(
			name,
			creatorAcc,
			stakingCoinWeights,
			startTime,
			endTime,
			epochAmount,
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgCreateRatioPlan generates a MsgCreateRatioPlan with random values
// nolint: interfacer
func SimulateMsgCreateRatioPlan(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateRatioPlan, "insufficient balance for plan creation fee"), nil, nil
		}

		name := "simulation-test-" + simtypes.RandStringOfLength(r, 5) // name must be unique
		creatorAcc := account.GetAddress()
		// mint pool coins to simulate the real-world cases
		_, err := fundBalances(ctx, r, bk, account.GetAddress(), testCoinDenoms)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateRatioPlan, "unable to mint pool coins"), nil, nil
		}
		stakingCoinWeights := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
		startTime := ctx.BlockTime()
		endTime := startTime.AddDate(1, 0, 0)
		epochRatio := sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 3)

		msg := farmingtypes.NewMsgCreateRatioPlan(
			name,
			creatorAcc,
			stakingCoinWeights,
			startTime,
			endTime,
			epochRatio,
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgStake generates a MsgStake with random values
// nolint: interfacer
func SimulateMsgStake(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		farmer := account.GetAddress()
		stakingCoins := sdk.NewCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 1_000_000, 1_000_000_000))),
		)
		if !spendable.IsAllGTE(stakingCoins) {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgUnstake, "insufficient funds"), nil, nil
		}

		msg := farmingtypes.NewMsgStake(farmer, stakingCoins)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgUnstake generates a SimulateMsgUnstake with random values
// nolint: interfacer
func SimulateMsgUnstake(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		farmer := account.GetAddress()
		unstakingCoins := sdk.NewCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 1_000_000, 100_000_000))),
		)

		// staking must exist in order to unstake
		staking, sf := k.GetStaking(ctx, sdk.DefaultBondDenom, farmer)
		if !sf {
			staking = farmingtypes.Staking{
				Amount: sdk.ZeroInt(),
			}
		}
		queuedStaking, qsf := k.GetQueuedStaking(ctx, sdk.DefaultBondDenom, farmer)
		if !qsf {
			if !qsf {
				queuedStaking = farmingtypes.QueuedStaking{
					Amount: sdk.ZeroInt(),
				}
			}
		}
		if !sf && !qsf {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgUnstake, "unable to find staking and queued staking"), nil, nil
		}
		// sum of staked and queued coins must be greater than unstaking coins
		if !staking.Amount.Add(queuedStaking.Amount).GTE(unstakingCoins[0].Amount) {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgUnstake, "insufficient funds"), nil, nil
		}

		// spendable must be greater than unstaking coins
		if !spendable.IsAllGT(unstakingCoins) {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgUnstake, "insufficient funds"), nil, nil
		}

		msg := farmingtypes.NewMsgUnstake(farmer, unstakingCoins)
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}
		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// SimulateMsgHarvest generates a MsgHarvest with random values
// nolint: interfacer
func SimulateMsgHarvest(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// find staking from the simulated accounts
		var ranStaking sdk.Coins
		for _, acc := range accs {
			staking := k.GetAllStakedCoinsByFarmer(ctx, acc.Address)
			if !staking.IsZero() {
				simAccount = acc
				ranStaking = staking
				break
			}
		}

		var stakingCoinDenoms []string
		for _, coin := range ranStaking {
			stakingCoinDenoms = append(stakingCoinDenoms, coin.Denom)
		}

		var totalRewards sdk.Coins
		for _, denom := range stakingCoinDenoms {
			rewards := k.Rewards(ctx, simAccount.Address, denom)
			totalRewards = totalRewards.Add(rewards...)
		}

		if totalRewards.IsZero() {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgHarvest, "no rewards to harvest"), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := farmingtypes.NewMsgHarvest(simAccount.Address, stakingCoinDenoms)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRemovePlan generates a MsgRemovePlan with random values
func SimulateMsgRemovePlan(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		creator := account.GetAddress()

		var terminatedPlans []farmingtypes.PlanI
		for _, plan := range k.GetPlans(ctx) {
			if plan.IsTerminated() {
				terminatedPlans = append(terminatedPlans, plan)
			}
		}
		if len(terminatedPlans) == 0 {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgRemovePlan, "no terminated plans to remove"), nil, nil
		}

		// Select a random terminated plan.
		plan := terminatedPlans[simtypes.RandIntBetween(r, 0, len(terminatedPlans))]

		msg := farmingtypes.NewMsgRemovePlan(creator, plan.GetId())
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      farmingtypes.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, Gas, Fees)
	}
}

// fundBalances mints random amount of coins with the provided coin denoms and
// send them to the simulated account.
func fundBalances(ctx sdk.Context, r *rand.Rand, bk farmingtypes.BankKeeper, acc sdk.AccAddress, denoms []string) (mintCoins sdk.Coins, err error) {
	for _, denom := range denoms {
		mintCoins = mintCoins.Add(sdk.NewInt64Coin(denom, int64(simtypes.RandIntBetween(r, 1e14, 1e15))))
	}

	if err := bk.MintCoins(ctx, minttypes.ModuleName, mintCoins); err != nil {
		return nil, err
	}

	if err := bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, acc, mintCoins); err != nil {
		return nil, err
	}
	return mintCoins, nil
}
