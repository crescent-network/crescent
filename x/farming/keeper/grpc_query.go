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

	"github.com/tendermint/farming/x/farming/types"
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
	k.paramSpace.GetParamSet(ctx, &params)
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
		plan, err := k.UnmarshalPlan(value)
		if err != nil {
			return false, err
		}
		any, err := codectypes.NewAnyWithValue(plan)
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
			plans = append(plans, any)
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
		k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
			resp.StakedCoins = resp.StakedCoins.Add(sdk.NewCoin(stakingCoinDenom, staking.Amount))
			return false
		})
		k.IterateQueuedStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, queuedStaking types.QueuedStaking) (stop bool) {
			resp.QueuedCoins = resp.QueuedCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
			return false
		})
	} else {
		staking, found := k.GetStaking(ctx, req.StakingCoinDenom, farmerAcc)
		if found {
			resp.StakedCoins = resp.StakedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, staking.Amount))
		}
		queuedStaking, found := k.GetQueuedStaking(ctx, req.StakingCoinDenom, farmerAcc)
		if found {
			resp.QueuedCoins = resp.QueuedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, queuedStaking.Amount))
		}
	}

	return resp, nil
}

func (k Querier) TotalStakings(c context.Context, req *types.QueryTotalStakingsRequest) (*types.QueryTotalStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	totalStakings, found := k.GetTotalStakings(ctx, req.StakingCoinDenom)
	if !found {
		totalStakings.Amount = sdk.ZeroInt()
	}

	return &types.QueryTotalStakingsResponse{
		Amount: totalStakings.Amount,
	}, nil
}

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
		var err error
		rewards, err = k.WithdrawAllRewards(ctx, farmerAcc)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		rewards, err = k.WithdrawRewards(ctx, farmerAcc, req.StakingCoinDenom)
		if err != nil {
			return nil, err
		}
	}
	resp.Rewards = rewards

	return resp, nil
}
