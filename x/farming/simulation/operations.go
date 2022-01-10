package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	crescentappparams "github.com/crescent-network/crescent/app/params"
	farmingkeeper "github.com/crescent-network/crescent/x/farming/keeper"
	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreateFixedAmountPlan = "op_weight_msg_create_fixed_amount_plan"
	OpWeightMsgCreateRatioPlan       = "op_weight_msg_create_ratio_plan"
	OpWeightMsgStake                 = "op_weight_msg_stake"
	OpWeightMsgUnstake               = "op_weight_msg_unstake"
	OpWeightMsgHarvest               = "op_weight_msg_harvest"
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak farmingtypes.AccountKeeper,
	bk farmingtypes.BankKeeper, k farmingkeeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgCreateFixedAmountPlan int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateFixedAmountPlan, &weightMsgCreateFixedAmountPlan, nil,
		func(_ *rand.Rand) {
			weightMsgCreateFixedAmountPlan = crescentappparams.DefaultWeightMsgCreateFixedAmountPlan
		},
	)

	var weightMsgCreateRatioPlan int
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateRatioPlan, &weightMsgCreateRatioPlan, nil,
		func(_ *rand.Rand) {
			weightMsgCreateRatioPlan = crescentappparams.DefaultWeightMsgCreateRatioPlan
		},
	)

	var weightMsgStake int
	appParams.GetOrGenerate(cdc, OpWeightMsgStake, &weightMsgStake, nil,
		func(_ *rand.Rand) {
			weightMsgStake = crescentappparams.DefaultWeightMsgStake
		},
	)

	var weightMsgUnstake int
	appParams.GetOrGenerate(cdc, OpWeightMsgUnstake, &weightMsgUnstake, nil,
		func(_ *rand.Rand) {
			weightMsgUnstake = crescentappparams.DefaultWeightMsgUnstake
		},
	)

	var weightMsgHarvest int
	appParams.GetOrGenerate(cdc, OpWeightMsgHarvest, &weightMsgHarvest, nil,
		func(_ *rand.Rand) {
			weightMsgHarvest = crescentappparams.DefaultWeightMsgHarvest
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

		// mint pool coins to simulate the real-world cases
		poolCoins, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateFixedAmountPlan, "unable to mint pool coins"), nil, nil
		}
		name := "simulation-test-" + simtypes.RandStringOfLength(r, 5) // name must be unique
		creatorAcc := account.GetAddress()
		stakingCoinWeights := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
		startTime := ctx.BlockTime()
		endTime := startTime.AddDate(0, 1, 0)
		epochAmount := sdk.NewCoins(
			sdk.NewInt64Coin(poolCoins[r.Intn(3)].Denom, int64(simtypes.RandIntBetween(r, 10_000_000, 1_000_000_000))),
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

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
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

		// mint pool coins to simulate the real-world cases
		_, err := mintPoolCoins(ctx, r, bk, simAccount)
		if err != nil {
			return simtypes.NoOpMsg(farmingtypes.ModuleName, farmingtypes.TypeMsgCreateRatioPlan, "unable to mint pool coins"), nil, nil
		}

		name := "simulation-test-" + simtypes.RandStringOfLength(r, 5) // name must be unique
		creatorAcc := account.GetAddress()
		stakingCoinWeights := sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1))
		startTime := ctx.BlockTime()
		endTime := startTime.AddDate(0, 1, 0)
		epochRatio := sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 1)

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

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
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

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
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
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
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

// mintPoolCoins mints random amount of coins with the provided pool coin denoms and
// send them to the simulated account.
func mintPoolCoins(ctx sdk.Context, r *rand.Rand, bk farmingtypes.BankKeeper, acc simtypes.Account) (mintCoins sdk.Coins, err error) {
	for _, denom := range []string{
		"pool93E069B333B5ECEBFE24C6E1437E814003248E0DD7FF8B9F82119F4587449BA5",
		"pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C",
		"poolE4D2617BFE03E1146F6BBA1D9893F2B3D77BA29E7ED532BB721A39FF1ECC1B07",
	} {
		mintCoins = mintCoins.Add(sdk.NewInt64Coin(denom, int64(simtypes.RandIntBetween(r, 1e14, 1e15))))
	}

	if err := bk.MintCoins(ctx, liquiditytypes.ModuleName, mintCoins); err != nil {
		return nil, err
	}

	if err := bk.SendCoinsFromModuleToAccount(ctx, liquiditytypes.ModuleName, acc.Address, mintCoins); err != nil {
		return nil, err
	}
	return mintCoins, nil
}
