package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
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

// CreateFixedAmountPlan defines a method for creating fixed amount farming plan.
func (k msgServer) CreateFixedAmountPlan(goCtx context.Context, msg *types.MsgCreateFixedAmountPlan) (*types.MsgCreateFixedAmountPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fixedPlan := k.Keeper.CreateFixedAmountPlan(ctx, msg, types.PlanTypePrivate)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateFixedAmountPlan,
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, msg.FarmingPoolAddress),
			sdk.NewAttribute(types.AttributeKeyRewardPoolAddress, fixedPlan.RewardPoolAddress),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochAmount, msg.EpochAmount.String()),
		),
	})

	return &types.MsgCreateFixedAmountPlanResponse{}, nil
}

// CreateRatioPlan defines a method for creating ratio farming plan.
func (k msgServer) CreateRatioPlan(goCtx context.Context, msg *types.MsgCreateRatioPlan) (*types.MsgCreateRatioPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ratioPlan := k.Keeper.CreateRatioPlan(ctx, msg, types.PlanTypePrivate)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateRatioPlan,
			sdk.NewAttribute(types.AttributeKeyFarmingPoolAddress, msg.FarmingPoolAddress),
			sdk.NewAttribute(types.AttributeKeyRewardPoolAddress, ratioPlan.RewardPoolAddress),
			sdk.NewAttribute(types.AttributeKeyStartTime, msg.StartTime.String()),
			sdk.NewAttribute(types.AttributeKeyEndTime, msg.EndTime.String()),
			sdk.NewAttribute(types.AttributeKeyEpochRatio, fmt.Sprint(msg.EpochRatio)),
		),
	})

	return &types.MsgCreateRatioPlanResponse{}, nil
}

// Stake defines a method for staking coins to the farming plan.
func (k msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.Keeper.Stake(ctx, msg.GetFarmer(), msg.StakingCoins); err != nil {
		return &types.MsgStakeResponse{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyFarmer, msg.Farmer),
			sdk.NewAttribute(types.AttributeKeyStakingCoins, msg.StakingCoins.String()),
		),
	})

	return &types.MsgStakeResponse{}, nil
}

// Unstake defines a method for unstaking coins from the farming plan.
func (k msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := k.Keeper.Unstake(ctx, msg.GetFarmer(), msg.UnstakingCoins); err != nil {
		return &types.MsgUnstakeResponse{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyFarmer, msg.Farmer),
			sdk.NewAttribute(types.AttributeKeyUnstakingCoins, msg.UnstakingCoins.String()),
		),
	})

	return &types.MsgUnstakeResponse{}, nil
}

// Harvest defines a method for claiming farming rewards from the farming plan.
func (k msgServer) Harvest(goCtx context.Context, msg *types.MsgHarvest) (*types.MsgHarvestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rewardCoins, err := k.Keeper.Harvest(ctx, msg.GetFarmer(), msg.StakingCoinDenoms)
	if err != nil {
		return &types.MsgHarvestResponse{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeHarvest,
			sdk.NewAttribute(types.AttributeKeyFarmer, msg.Farmer),
			sdk.NewAttribute(types.AttributeKeyRewardCoins, rewardCoins.String()),
		),
	})

	return &types.MsgHarvestResponse{}, nil
}
