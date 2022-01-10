package keeper

import (
	"context"

	"github.com/tendermint/farming/x/liquidity/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func (m msgServer) CreatePool(ctx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	panic("implement me")
}

func (m msgServer) DepositBatch(ctx context.Context, msg *types.MsgDepositBatch) (*types.MsgDepositBatchResponse, error) {
	panic("implement me")
}

func (m msgServer) WithdrawBatch(ctx context.Context, msg *types.MsgWithdrawBatch) (*types.MsgWithdrawBatchResponse, error) {
	panic("implement me")
}

func (m msgServer) SwapBatch(ctx context.Context, msg *types.MsgSwapBatch) (*types.MsgSwapBatchResponse, error) {
	panic("implement me")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}
