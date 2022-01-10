package keeper

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/x/farming/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the farming module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.Keeper.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

// Plans queries all plans.
func (k Querier) Plans(c context.Context, req *types.QueryPlansRequest) (*types.QueryPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Type != "" && !(req.Type == types.PlanTypePublic.String() || req.Type == types.PlanTypePrivate.String()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid plan type %s", req.Type)
	}

	if req.FarmingPoolAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.FarmingPoolAddress); err != nil {
			return nil, err
		}
	}

	if req.TerminationAddress != "" {
		if _, err := sdk.AccAddressFromBech32(req.TerminationAddress); err != nil {
			return nil, err
		}
	}

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
			return nil, err
		}
	}

	var terminated bool
	if req.Terminated != "" {
		var err error
		terminated, err = strconv.ParseBool(req.Terminated)
		if err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	planStore := prefix.NewStore(store, types.PlanKeyPrefix)

	var plans []*codectypes.Any
	pageRes, err := query.FilteredPaginate(planStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		plan, err := k.Keeper.UnmarshalPlan(value)
		if err != nil {
			return false, err
		}
		planAny, err := types.PackPlan(plan)
		if err != nil {
			return false, err
		}

		if req.Type != "" && plan.GetType().String() != req.Type {
			return false, nil
		}

		if req.FarmingPoolAddress != "" && plan.GetFarmingPoolAddress().String() != req.FarmingPoolAddress {
			return false, nil
		}

		if req.TerminationAddress != "" && plan.GetTerminationAddress().String() != req.TerminationAddress {
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

		if req.Terminated != "" {
			if plan.GetTerminated() != terminated {
				return false, nil
			}
		}

		if accumulate {
			plans = append(plans, planAny)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlansResponse{Plans: plans, Pagination: pageRes}, nil
}

// Plan queries a specific plan.
func (k Querier) Plan(c context.Context, req *types.QueryPlanRequest) (*types.QueryPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	plan, found := k.Keeper.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "plan %d not found", req.PlanId)
	}

	planAny, err := types.PackPlan(plan)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlanResponse{Plan: planAny}, nil
}

// Stakings queries stakings for a farmer.
func (k Querier) Stakings(c context.Context, req *types.QueryStakingsRequest) (*types.QueryStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)

	resp := &types.QueryStakingsResponse{
		StakedCoins: sdk.NewCoins(),
		QueuedCoins: sdk.NewCoins(),
	}
	if req.StakingCoinDenom == "" {
		resp.StakedCoins = k.Keeper.GetAllStakedCoinsByFarmer(ctx, farmerAcc)
		resp.QueuedCoins = k.Keeper.GetAllQueuedCoinsByFarmer(ctx, farmerAcc)
	} else {
		staking, found := k.Keeper.GetStaking(ctx, req.StakingCoinDenom, farmerAcc)
		if found {
			resp.StakedCoins = resp.StakedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, staking.Amount))
		}
		queuedStaking, found := k.Keeper.GetQueuedStaking(ctx, req.StakingCoinDenom, farmerAcc)
		if found {
			resp.QueuedCoins = resp.QueuedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, queuedStaking.Amount))
		}
	}

	return resp, nil
}

// TotalStakings queries total staking coin amount for a specific staking coin denom.
func (k Querier) TotalStakings(c context.Context, req *types.QueryTotalStakingsRequest) (*types.QueryTotalStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	totalStakings, found := k.Keeper.GetTotalStakings(ctx, req.StakingCoinDenom)
	if !found {
		totalStakings.Amount = sdk.ZeroInt()
	}

	return &types.QueryTotalStakingsResponse{
		Amount: totalStakings.Amount,
	}, nil
}

// Rewards queries accumulated rewards for a farmer.
func (k Querier) Rewards(c context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)

	resp := &types.QueryRewardsResponse{
		Rewards: sdk.NewCoins(),
	}

	var rewards sdk.Coins
	if req.StakingCoinDenom == "" {
		rewards = k.Keeper.AllRewards(ctx, farmerAcc)
	} else {
		rewards = k.Keeper.Rewards(ctx, farmerAcc, req.StakingCoinDenom)
	}
	resp.Rewards = rewards

	return resp, nil
}

// CurrentEpochDays queries current epoch days.
func (k Querier) CurrentEpochDays(c context.Context, req *types.QueryCurrentEpochDaysRequest) (*types.QueryCurrentEpochDaysResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	currentEpochDays := k.Keeper.GetCurrentEpochDays(ctx)

	return &types.QueryCurrentEpochDaysResponse{CurrentEpochDays: currentEpochDays}, nil
}
