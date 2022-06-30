package keeper

// DONTCOVER

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
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
	poolAcc, err := k.DerivePrivatePlanFarmingPoolAcc(ctx, msg.Name)
	if err != nil {
		return nil, err
	}

	if _, err := k.Keeper.CreateFixedAmountPlan(ctx, msg, poolAcc, msg.GetCreator(), types.PlanTypePrivate); err != nil {
		return nil, err
	}

	return &types.MsgCreateFixedAmountPlanResponse{}, nil
}

// CreateRatioPlan defines a method for creating ratio farming plan.
func (k msgServer) CreateRatioPlan(goCtx context.Context, msg *types.MsgCreateRatioPlan) (*types.MsgCreateRatioPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !EnableRatioPlan {
		return nil, types.ErrRatioPlanDisabled
	}

	poolAcc, err := k.DerivePrivatePlanFarmingPoolAcc(ctx, msg.Name)
	if err != nil {
		return nil, err
	}

	if _, err := k.Keeper.CreateRatioPlan(ctx, msg, poolAcc, msg.GetCreator(), types.PlanTypePrivate); err != nil {
		return nil, err
	}

	plans := k.GetPlans(ctx)
	if err := types.ValidateTotalEpochRatio(plans); err != nil {
		return nil, err
	}

	return &types.MsgCreateRatioPlanResponse{}, nil
}

// Stake defines a method for staking coins to the farming plan.
func (k msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.Stake(ctx, msg.GetFarmer(), msg.StakingCoins); err != nil {
		return nil, err
	}

	return &types.MsgStakeResponse{}, nil
}

// Unstake defines a method for unstaking coins from the farming plan.
func (k msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.Unstake(ctx, msg.GetFarmer(), msg.UnstakingCoins); err != nil {
		return nil, err
	}

	return &types.MsgUnstakeResponse{}, nil
}

// Harvest defines a method for claiming farming rewards from the farming plan.
func (k msgServer) Harvest(goCtx context.Context, msg *types.MsgHarvest) (*types.MsgHarvestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.Harvest(ctx, msg.GetFarmer(), msg.StakingCoinDenoms); err != nil {
		return nil, err
	}

	return &types.MsgHarvestResponse{}, nil
}

func (k msgServer) RemovePlan(goCtx context.Context, msg *types.MsgRemovePlan) (*types.MsgRemovePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.RemovePlan(ctx, msg.GetCreator(), msg.PlanId); err != nil {
		return nil, err
	}

	return &types.MsgRemovePlanResponse{}, nil
}

// AdvanceEpoch defines a method for advancing epoch by one, just for testing purpose
// and shouldn't be used in real world.
func (k msgServer) AdvanceEpoch(goCtx context.Context, msg *types.MsgAdvanceEpoch) (*types.MsgAdvanceEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if EnableAdvanceEpoch {
		if err := k.Keeper.AdvanceEpoch(ctx); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("AdvanceEpoch is disabled")
	}

	return &types.MsgAdvanceEpochResponse{}, nil
}
