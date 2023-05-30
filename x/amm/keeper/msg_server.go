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
	pool, err := k.Keeper.CreatePool(ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.MarketId, msg.Price)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreatePoolResponse{PoolId: pool.Id}, nil
}

func (k msgServer) AddLiquidity(goCtx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr := sdk.MustAccAddressFromBech32(msg.Sender)
	_, liquidity, amt, err := k.Keeper.AddLiquidity(
		ctx, senderAddr, senderAddr, msg.PoolId,
		msg.LowerPrice, msg.UpperPrice, msg.DesiredAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgAddLiquidityResponse{
		Liquidity: liquidity,
		Amount:    amt,
	}, nil
}

func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr := sdk.MustAccAddressFromBech32(msg.Sender)
	_, amt, err := k.Keeper.RemoveLiquidity(ctx, senderAddr, senderAddr, msg.PositionId, msg.Liquidity)
	if err != nil {
		return nil, err
	}
	return &types.MsgRemoveLiquidityResponse{
		Amount: amt,
	}, nil
}

func (k msgServer) Collect(goCtx context.Context, msg *types.MsgCollect) (*types.MsgCollectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	senderAddr := sdk.MustAccAddressFromBech32(msg.Sender)
	if err := k.Keeper.Collect(
		ctx, senderAddr, senderAddr, msg.PositionId, msg.Amount); err != nil {
		return nil, err
	}
	return &types.MsgCollectResponse{}, nil
}

func (k msgServer) CreatePrivateFarmingPlan(goCtx context.Context, msg *types.MsgCreatePrivateFarmingPlan) (*types.MsgCreatePrivateFarmingPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	plan, err := k.Keeper.CreatePrivateFarmingPlan(
		ctx, sdk.MustAccAddressFromBech32(msg.Sender), msg.Description,
		sdk.MustAccAddressFromBech32(msg.TerminationAddress), msg.RewardAllocations, msg.StartTime, msg.EndTime)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreatePrivateFarmingPlanResponse{
		FarmingPlanId:      plan.Id,
		FarmingPoolAddress: plan.FarmingPoolAddress,
	}, nil
}
