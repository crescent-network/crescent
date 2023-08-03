package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// Simulation operation weights constants.
const (
	OpWeightMsgMintShare = "op_weight_msg_mint_share"
	OpWeightMsgBurnShare = "op_weight_msg_burn_share"
	OpWeightMsgPlaceBid  = "op_weight_msg_place_bid"

	DefaultWeightMsgMintShare int = 50
	DefaultWeightMsgBurnShare int = 10
	DefaultWeightMsgPlaceBid  int = 20
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

	var weightMsgMintShare int
	appParams.GetOrGenerate(cdc, OpWeightMsgMintShare, &weightMsgMintShare, nil,
		func(_ *rand.Rand) {
			weightMsgMintShare = DefaultWeightMsgMintShare
		},
	)

	var weightMsgBurnShare int
	appParams.GetOrGenerate(cdc, OpWeightMsgBurnShare, &weightMsgBurnShare, nil,
		func(_ *rand.Rand) {
			weightMsgBurnShare = DefaultWeightMsgBurnShare
		},
	)

	var weightMsgPlaceBid int
	appParams.GetOrGenerate(cdc, OpWeightMsgPlaceBid, &weightMsgPlaceBid, nil,
		func(_ *rand.Rand) {
			weightMsgPlaceBid = DefaultWeightMsgPlaceBid
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgMintShare,
			SimulateMsgMintShare(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBurnShare,
			SimulateMsgBurnShare(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPlaceBid,
			SimulateMsgPlaceBid(ak, bk, k),
		),
	}
}

// SimulateMsgMintShare generates a MsgMintShare with random values
// nolint: interfacer
func SimulateMsgMintShare(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// TODO: not implemented yet

		return simtypes.OperationMsg{}, nil, nil
		// return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgBurnShare generates a MsgBurnShare with random values
// nolint: interfacer
func SimulateMsgBurnShare(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
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
