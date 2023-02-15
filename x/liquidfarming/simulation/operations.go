package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v5/app/params"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgLiquidFarm   = "op_weight_msg_liquid_farm"
	OpWeightMsgLiquidUnfarm = "op_weight_msg_liquid_unfarm"
	OpWeightMsgPlaceBid     = "op_weight_msg_place_bid"
	OpWeightMsgRefundBid    = "op_weight_msg_refund_bid"
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

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgLiquidFarm int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidFarm, &weightMsgLiquidFarm, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidFarm = appparams.DefaultWeightMsgLiquidFarm
		},
	)

	var weightMsgLiquidUnfarm int
	appParams.GetOrGenerate(cdc, OpWeightMsgLiquidUnfarm, &weightMsgLiquidUnfarm, nil,
		func(_ *rand.Rand) {
			weightMsgLiquidUnfarm = appparams.DefaultWeightMsgLiquidUnfarm
		},
	)

	var weightMsgPlaceBid int
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceBid, &weightMsgPlaceBid, nil,
		func(_ *rand.Rand) {
			weightMsgPlaceBid = appparams.DefaultWeightMsgPlaceBid
		},
	)

	var weightMsgRefundBid int
	appParams.GetOrGenerate(cdc, OpWeightMsgRefundBid, &weightMsgRefundBid, nil,
		func(_ *rand.Rand) {
			weightMsgRefundBid = appparams.DefaultWeightMsgRefundBid
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgLiquidFarm,
			SimulateMsgLiquidFarm(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgLiquidUnfarm,
			SimulateMsgLiquidUnfarm(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceBid,
			SimulateMsgPlaceBid(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRefundBid,
			SimulateMsgRefundBid(ak, bk, k),
		),
	}
}

// SimulateMsgLiquidFarm generates a MsgLiquidFarm with random values
// nolint: interfacer
func SimulateMsgLiquidFarm(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgLiquidUnfarm generates a MsgLiquidUnfarm with random values
// nolint: interfacer
func SimulateMsgLiquidUnfarm(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgPlaceBid generates a MsgPlaceBid with random values
// nolint: interfacer
func SimulateMsgPlaceBid(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgRefundBid generates a MsgRefundBid with random values
// nolint: interfacer
func SimulateMsgRefundBid(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}
