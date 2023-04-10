package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pool, err := k.Keeper.CreatePool(ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.Denom0, msg.Denom1, msg.TickSpacing, msg.Price)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePoolResponse{PoolId: pool.Id}, nil
}

func (k msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, liquidity, amt0, amt1, err := k.Keeper.AddLiquidity(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.PoolId, msg.LowerTick, msg.UpperTick,
		msg.DesiredAmount0, msg.DesiredAmount1, msg.MinAmount0, msg.MinAmount1)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddLiquidityResponse{
		Liquidity: liquidity,
		Amount0:   amt0,
		Amount1:   amt1,
	}, nil
}

func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, amt0, amt1, err := k.Keeper.RemoveLiquidity(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.PositionId,
		msg.Liquidity, msg.MinAmount0, msg.MinAmount1)
	if err != nil {
		return nil, err
	}

	return &types.MsgRemoveLiquidityResponse{
		Amount0: amt0,
		Amount1: amt1,
	}, nil
}
