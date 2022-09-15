package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v3/app/params"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgCreatePrivatePlan = "op_weight_msg_create_private_plan"
	OpWeightMsgFarm              = "op_weight_msg_farm"
	OpWeightMsgUnfarm            = "op_weight_msg_unfarm"
	OpWeightMsgHarvest           = "op_weight_msg_harvest"

	DefaultWeightCreatePrivatePlan = 10
	DefaultWeightFarm              = 40
	DefaultWeightUnfarm            = 50
	DefaultWeightHarvest           = 30
)

var (
	gas  = uint64(20000000)
	fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreatePrivatePlan int
		weightMsgFarm              int
		weightMsgUnfarm            int
		weightMsgHarvest           int
	)
	appParams.GetOrGenerate(cdc, OpWeightMsgCreatePrivatePlan, &weightMsgCreatePrivatePlan, nil, func(_ *rand.Rand) {
		weightMsgCreatePrivatePlan = DefaultWeightCreatePrivatePlan
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgFarm, &weightMsgFarm, nil, func(_ *rand.Rand) {
		weightMsgFarm = DefaultWeightFarm
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgUnfarm, &weightMsgUnfarm, nil, func(_ *rand.Rand) {
		weightMsgUnfarm = DefaultWeightUnfarm
	})
	appParams.GetOrGenerate(cdc, OpWeightMsgHarvest, &weightMsgHarvest, nil, func(_ *rand.Rand) {
		weightMsgHarvest = DefaultWeightHarvest
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreatePrivatePlan,
			SimulateMsgCreatePrivatePlan(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgFarm,
			SimulateMsgFarm(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgUnfarm,
			SimulateMsgUnfarm(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgHarvest,
			SimulateMsgHarvest(ak, bk),
		),
	}
}

func SimulateMsgCreatePrivatePlan(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[0]
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		msg := &types.MsgCreatePrivatePlan{}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           appparams.MakeTestEncodingConfig().TxConfig,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgFarm(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[0]
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		msg := &types.MsgFarm{}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           appparams.MakeTestEncodingConfig().TxConfig,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgUnfarm(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[0]
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		msg := &types.MsgUnfarm{}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           appparams.MakeTestEncodingConfig().TxConfig,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}

func SimulateMsgHarvest(ak types.AccountKeeper, bk types.BankKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[0]
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		msg := &types.MsgHarvest{}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           appparams.MakeTestEncodingConfig().TxConfig,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return utils.GenAndDeliverTxWithFees(txCtx, gas, fees)
	}
}
