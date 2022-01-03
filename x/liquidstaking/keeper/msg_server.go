package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"

	"github.com/tendermint/farming/x/liquidstaking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the farming MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: unimplemented
	//k.LiquidStaking()

	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) LiquidUnstake(goCtx context.Context, msg *types.MsgLiquidUnstake) (*types.MsgLiquidUnstakeResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: unimplemented
	//k.LiquidUnstaking()

	return &types.MsgLiquidUnstakeResponse{}, nil
}
