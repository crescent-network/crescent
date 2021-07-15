package keeper

import (
	"context"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/tendermint/farming/x/farming/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Querier) Plans(c context.Context, req *types.QueryPlansRequest) (*types.QueryPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Type != "" && !(req.Type == types.PlanTypePublic.String() || req.Type == types.PlanTypePrivate.String()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid plan type %s", req.Type)
	}

	if req.FarmingPoolAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.FarmingPoolAddress); err != nil {
			return nil, err
		}
	}

	if req.RewardPoolAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.RewardPoolAddress); err != nil {
			return nil, err
		}
	}

	if req.TerminationAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.TerminationAddress); err != nil {
			return nil, err
		}
	}

	if req.StakingReserveAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.StakingReserveAddress); err != nil {
			return nil, err
		}
	}

	if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	planStore := prefix.NewStore(store, types.PlanKeyPrefix)

	var plans []*codectypes.Any
	pageRes, err := query.FilteredPaginate(planStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		plan, err := k.UnmarshalPlan(value)
		if err != nil {
			return false, err
		}
		any, err := codectypes.NewAnyWithValue(plan)
		if err != nil {
			return false, err
		}

		if req.FarmingPoolAddress != "" && plan.GetFarmingPoolAddress().String() != req.FarmingPoolAddress {
			return false, nil
		}

		if req.RewardPoolAddress != "" && plan.GetRewardPoolAddress().String() != req.RewardPoolAddress {
			return false, nil
		}

		if req.TerminationAddress != "" && plan.GetTerminationAddress().String() != req.TerminationAddress {
			return false, nil
		}

		if req.StakingReserveAddress != "" && plan.GetStakingReserveAddress().String() != req.StakingReserveAddress {
			return false, nil
		}

		if req.StakingCoinDenom != "" {
			found := false
			for _, coin := range plan.GetStakingCoinWeights() {
				if coin.Denom == req.StakingCoinDenom {
					found = true
					break
				}
			}
			if !found {
				return false, nil
			}
		}

		if accumulate {
			plans = append(plans, any)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlansResponse{Plans: plans, Pagination: pageRes}, nil
}

func (k Querier) Plan(c context.Context, req *types.QueryPlanRequest) (*types.QueryPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "plan %d not found", req.PlanId)
	}

	any, err := codectypes.NewAnyWithValue(plan)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlanResponse{Plan: any}, nil
}

func (k Querier) PlanStakings(c context.Context, req *types.QueryPlanStakingsRequest) (*types.QueryPlanStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	stakingStore := prefix.NewStore(store, types.GetStakingPrefix(req.PlanId))

	var stakings []*types.Staking
	pageRes, err := query.Paginate(stakingStore, req.Pagination, func(key []byte, value []byte) error {
		staking, err := k.UnmarshalStaking(value)
		if err != nil {
			return err
		}
		stakings = append(stakings, &staking)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlanStakingsResponse{Stakings: stakings, Pagination: pageRes}, nil
}

func (k Querier) FarmerStakings(c context.Context, req *types.QueryFarmerStakingsRequest) (*types.QueryFarmerStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	farmerAddr, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	planStore := prefix.NewStore(store, types.GetPlansByFarmerAddrIndexKey(farmerAddr))

	var stakings []*types.Staking
	pageRes, err := query.Paginate(planStore, req.Pagination, func(key []byte, value []byte) error {
		var val gogotypes.UInt64Value
		if err := k.cdc.Unmarshal(value, &val); err != nil {
			return err
		}
		planID := val.GetValue()
		staking, found := k.GetStaking(ctx, planID, farmerAddr)
		if !found { // TODO: Remove this check if we can sure that we're cleaning the store correctly.
			return fmt.Errorf("staking not found")
		}
		stakings = append(stakings, &staking)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmerStakingsResponse{Stakings: stakings, Pagination: pageRes}, nil
}

func (k Querier) FarmerStaking(c context.Context, req *types.QueryFarmerStakingRequest) (*types.QueryFarmerStakingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	farmerAddr, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	staking, found := k.GetStaking(ctx, req.PlanId, farmerAddr)
	if !found {
		return nil, status.Error(codes.NotFound, "staking not found")
	}

	return &types.QueryFarmerStakingResponse{Staking: &staking}, nil
}

func (k Querier) PlanRewards(c context.Context, req *types.QueryPlanRewardsRequest) (*types.QueryPlanRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	rewardStore := prefix.NewStore(store, types.GetRewardPrefix(req.PlanId))

	var rewards []*types.Reward
	pageRes, err := query.Paginate(rewardStore, req.Pagination, func(key []byte, value []byte) error {
		reward, err := k.UnmarshalReward(value)
		if err != nil {
			return err
		}
		rewards = append(rewards, &reward)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlanRewardsResponse{Rewards: rewards, Pagination: pageRes}, nil
}

func (k Querier) FarmerRewards(c context.Context, req *types.QueryFarmerRewardsRequest) (*types.QueryFarmerRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	farmerAddr, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	planStore := prefix.NewStore(store, types.GetPlansByFarmerAddrIndexKey(farmerAddr))

	var rewards []*types.Reward
	pageRes, err := query.Paginate(planStore, req.Pagination, func(key []byte, value []byte) error {
		var val gogotypes.UInt64Value
		if err := k.cdc.Unmarshal(value, &val); err != nil {
			return err
		}
		planID := val.GetValue()
		reward, found := k.GetReward(ctx, planID, farmerAddr)
		if !found { // TODO: Remove this check if we can sure that we're cleaning the store correctly.
			return fmt.Errorf("reward not found")
		}
		rewards = append(rewards, &reward)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmerRewardsResponse{Rewards: rewards, Pagination: pageRes}, nil
}

func (k Querier) FarmerReward(c context.Context, req *types.QueryFarmerRewardRequest) (*types.QueryFarmerRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	farmerAddr, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	reward, found := k.GetReward(ctx, req.PlanId, farmerAddr)
	if !found {
		return nil, status.Error(codes.NotFound, "reward not found")
	}

	return &types.QueryFarmerRewardResponse{Reward: &reward}, nil
}
