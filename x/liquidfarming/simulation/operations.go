package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/crescent-network/crescent/v2/app/params"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgFarm      = "op_weight_msg_farm"
	OpWeightMsgUnfarm    = "op_weight_msg_unfarm"
	OpWeightMsgPlaceBid  = "op_weight_msg_place_bid"
	OpWeightMsgRefundBid = "op_weight_msg_refund_bid"
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

	var weightMsgFarm int
	appParams.GetOrGenerate(cdc, OpWeightMsgFarm, &weightMsgFarm, nil,
		func(_ *rand.Rand) {
			weightMsgFarm = appparams.DefaultWeightMsgFarm
		},
	)

	var weightMsgUnfarm int
	appParams.GetOrGenerate(cdc, OpWeightMsgUnfarm, &weightMsgUnfarm, nil,
		func(_ *rand.Rand) {
			weightMsgUnfarm = appparams.DefaultWeightMsgUnfarm
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
			weightMsgFarm,
			SimulateMsgFarm(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUnfarm,
			SimulateMsgUnfarm(ak, bk, k),
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

// SimulateMsgFarm generates a MsgFarm with random values
// nolint: interfacer
func SimulateMsgFarm(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgUnfarm generates a MsgUnfarm with random values
// nolint: interfacer
func SimulateMsgUnfarm(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
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
