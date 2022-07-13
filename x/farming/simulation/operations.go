package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v2/app/params"
	utils "github.com/crescent-network/crescent/v2/types"
	farmingkeeper "github.com/crescent-network/crescent/v2/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
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
		if uint32(k.GetNumActivePrivatePlans(ctx)) >= params.MaxNumPrivatePlans {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateFixedAmountPlan, "maximum number of private plans reached"), nil, nil
		}

		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateFixedAmountPlan, "insufficient balance for plan creation fee"), nil, nil
		}

		rewardDenom := testCoinDenoms[r.Intn(len(testCoinDenoms))]
		if err := ensurePositiveSupply(bk, ctx, rewardDenom); err != nil {
			panic(fmt.Errorf("ensure positive supply of reward denom: %w", err))
		}

		msg := farmingtypes.NewMsgCreateFixedAmountPlan(
			"plan-"+simtypes.RandStringOfLength(r, 5),
			simAccount.Address,
			sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1)),
			ctx.BlockTime(),
			ctx.BlockTime().AddDate(0, 0, 1+r.Intn(5)),
			sdk.NewCoins(
				sdk.NewInt64Coin(
					rewardDenom,
					int64(simtypes.RandIntBetween(r, 1e8, 1e9)),
				),
			),
		)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			Fees,
			Gas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		planId := k.GetGlobalPlanId(ctx) // Newly created plan's id
		plan, found := k.GetPlan(ctx, planId)
		if !found {
			panic(fmt.Errorf("plan %d is not created", planId))
		}

		farmingPoolAcc := plan.GetFarmingPoolAddress()
		if _, err := fundBalances(ctx, r, bk, farmingPoolAcc, testCoinDenoms); err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "failed to fund farming pool"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgCreateRatioPlan generates a MsgCreateRatioPlan with random values
// nolint: interfacer
func SimulateMsgCreateRatioPlan(ak farmingtypes.AccountKeeper, bk farmingtypes.BankKeeper, k farmingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		farmingkeeper.EnableRatioPlan = true

		simAccount, _ := simtypes.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		params := k.GetParams(ctx)
		if uint32(k.GetNumActivePrivatePlans(ctx)) >= params.MaxNumPrivatePlans {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateRatioPlan, "maximum number of private plans reached"), nil, nil
		}

		_, hasNeg := spendable.SafeSub(params.PrivatePlanCreationFee)
		if hasNeg {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateRatioPlan, "insufficient balance for plan creation fee"), nil, nil
		}

		if err := ensurePositiveSupply(bk, ctx, testCoinDenoms...); err != nil {
			panic(fmt.Errorf("ensure positive supply of reward denoms: %w", err))
		}

		msg := farmingtypes.NewMsgCreateRatioPlan(
			"plan-"+simtypes.RandStringOfLength(r, 5),
			simAccount.Address,
			sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1)),
			ctx.BlockTime(),
			ctx.BlockTime().AddDate(0, 0, 1+r.Intn(5)),
			sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 3),
		)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			Fees,
			Gas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		planId := k.GetGlobalPlanId(ctx) // Newly created plan's id
		plan, found := k.GetPlan(ctx, planId)
		if !found {
			panic(fmt.Errorf("plan %d is not created", planId))
		}

		farmingPoolAcc := plan.GetFarmingPoolAddress()
		if _, err := fundBalances(ctx, r, bk, farmingPoolAcc, testCoinDenoms); err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, msg.Type(), "failed to fund farming pool"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
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
		accs = utils.ShuffleSimAccounts(r, accs)

		var simAccount simtypes.Account
		var totalCoins sdk.Coins
		skip := true
		for _, simAccount = range accs {
			stakedCoins := k.GetAllStakedCoinsByFarmer(ctx, simAccount.Address)
			queuedCoins := k.GetAllQueuedCoinsByFarmer(ctx, simAccount.Address)
			totalCoins = stakedCoins.Add(queuedCoins...)
			if !totalCoins.IsZero() {
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgUnstake, "no account to unstake"), nil, nil
		}

		unstakingCoins := simtypes.RandSubsetCoins(r, totalCoins)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := farmingtypes.NewMsgUnstake(simAccount.Address, unstakingCoins)

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
		var stakingCoinDenoms []string

		skip := true
		// find staking from the simulated accounts
		for _, acc := range accs {
			staked := k.GetAllStakedCoinsByFarmer(ctx, acc.Address)
			stakingCoinDenoms = nil
			for _, coin := range staked {
				rewards := k.Rewards(ctx, acc.Address, coin.Denom)
				if !rewards.IsZero() {
					stakingCoinDenoms = append(stakingCoinDenoms, coin.Denom)
				}
			}
			if len(stakingCoinDenoms) > 0 {
				simAccount = acc
				skip = false
				break
			}
		}
		if skip {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgHarvest, "no account to harvest rewards"), nil, nil
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
		accs = utils.ShuffleSimAccounts(r, accs)

		plans := k.GetPlans(ctx)
		r.Shuffle(len(plans), func(i, j int) {
			plans[i], plans[j] = plans[j], plans[i]
		})

		var simAccount simtypes.Account
		var plan farmingtypes.PlanI
		skip := true
	loop:
		for _, simAccount = range accs {
			for _, plan = range plans {
				// Only the plan creator can remove the plan.
				if plan.GetType() == farmingtypes.PlanTypePrivate &&
					plan.IsTerminated() &&
					plan.GetTerminationAddress().Equals(simAccount.Address) {
					skip = false
					break loop
				}
			}
		}
		if skip {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgRemovePlan, "no plan to remove"), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := farmingtypes.NewMsgRemovePlan(simAccount.Address, plan.GetId())
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

// ensurePositiveSupply mints coins for each denom, with amount of 1 to ensure
// the supply of the denom is positive.
func ensurePositiveSupply(bk farmingtypes.BankKeeper, ctx sdk.Context, denoms ...string) error {
	mintingCoins := sdk.Coins{}
	for _, denom := range denoms {
		mintingCoins = mintingCoins.Add(sdk.NewInt64Coin(denom, 1))
	}
	return bk.MintCoins(ctx, minttypes.ModuleName, mintingCoins)
}
