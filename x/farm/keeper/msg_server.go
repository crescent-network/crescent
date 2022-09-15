package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/farm/types"
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

// CreatePrivatePlan defines a method to create a new private plan.
func (k msgServer) CreatePrivatePlan(goCtx context.Context, msg *types.MsgCreatePrivatePlan) (*types.MsgCreatePrivatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	plan, err := k.Keeper.CreatePrivatePlan(
		ctx, creatorAddr, msg.Description, msg.RewardAllocations, msg.StartTime, msg.EndTime)
	if err != nil {
		return nil, err
	}

	// TODO: do we need below event?
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Creator),
		),
	)

	return &types.MsgCreatePrivatePlanResponse{
		Id:                 plan.Id,
		FarmingPoolAddress: plan.SourceAddress,
	}, nil
}

// Farm defines a method for farming coins.
func (k msgServer) Farm(goCtx context.Context, msg *types.MsgFarm) (*types.MsgFarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	farmerAddr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}

	withdrawnRewards, err := k.Keeper.Farm(ctx, farmerAddr, msg.Coin)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Farmer),
		),
	)

	return &types.MsgFarmResponse{
		WithdrawnRewards: withdrawnRewards,
	}, nil
}

// Unfarm defines a method for un-farming coins.
func (k msgServer) Unfarm(goCtx context.Context, msg *types.MsgUnfarm) (*types.MsgUnfarmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	farmerAddr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}

	withdrawnRewards, err := k.Keeper.Unfarm(ctx, farmerAddr, msg.Coin)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Farmer),
		),
	)

	return &types.MsgUnfarmResponse{
		WithdrawnRewards: withdrawnRewards,
	}, nil
}

// Harvest defines a method for harvesting farming rewards.
func (k msgServer) Harvest(goCtx context.Context, msg *types.MsgHarvest) (*types.MsgHarvestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	farmerAddr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}

	withdrawnRewards, err := k.Keeper.Harvest(ctx, farmerAddr, msg.Denom)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Farmer),
		),
	)

	return &types.MsgHarvestResponse{
		WithdrawnRewards: withdrawnRewards,
	}, nil
}
