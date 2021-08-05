package keeper

import (
	"context"
	"fmt"

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

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
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

		if req.RewardPoolAddress != "" && plan.GetRewardPoolAddress().String() != req.RewardPoolAddress {
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

// Stakings queries all stakings.
func (k Querier) Stakings(c context.Context, req *types.QueryStakingsRequest) (*types.QueryStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	var stakings []types.Staking
	var pageRes *query.PageResponse
	var err error

	if req.Farmer != "" {
		farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
		if err != nil {
			return nil, err
		}

		staking, found := k.GetStakingByFarmer(ctx, farmerAcc)
		if found {
			found := false

			if req.StakingCoinDenom != "" {
				for _, denom := range staking.StakingCoinDenoms() {
					if denom == req.StakingCoinDenom {
						found = true
						break
					}
				}
			} else {
				found = true
			}

			if found {
				stakings = append(stakings, staking)
			}
		}

		pageRes = &query.PageResponse{
			NextKey: nil,
			Total:   uint64(len(stakings)),
		}
	} else {
		if req.StakingCoinDenom != "" {
			storePrefix := types.GetStakingsByStakingCoinDenomIndexKey(req.StakingCoinDenom)
			indexStore := prefix.NewStore(store, storePrefix)
			pageRes, err = query.Paginate(indexStore, req.Pagination, func(key, value []byte) error {
				_, stakingID := types.ParseStakingsByStakingCoinDenomIndexKey(append(storePrefix, key...))
				staking, _ := k.GetStaking(ctx, stakingID)
				stakings = append(stakings, staking)
				return nil
			})
		} else {
			stakingStore := prefix.NewStore(store, types.StakingKeyPrefix)
			pageRes, err = query.Paginate(stakingStore, req.Pagination, func(key, value []byte) error {
				var staking types.Staking
				k.cdc.MustUnmarshal(value, &staking)
				stakings = append(stakings, staking)
				return nil
			})
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &types.QueryStakingsResponse{Stakings: stakings, Pagination: pageRes}, nil
}

// Staking queries a specific staking.
func (k Querier) Staking(c context.Context, req *types.QueryStakingRequest) (*types.QueryStakingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	staking, found := k.GetStaking(ctx, req.StakingId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "staking %d not found", req.StakingId)
	}

	return &types.QueryStakingResponse{Staking: staking}, nil
}

// Rewards queries all rewards.
func (k Querier) Rewards(c context.Context, req *types.QueryRewardsRequest) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.StakingCoinDenom != "" {
		if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	var rewards []types.Reward
	var pageRes *query.PageResponse
	var err error

	if req.Farmer != "" {
		var farmerAcc sdk.AccAddress
		farmerAcc, err = sdk.AccAddressFromBech32(req.Farmer)
		if err != nil {
			return nil, err
		}

		storePrefix := types.GetRewardsByFarmerIndexKey(farmerAcc)
		indexStore := prefix.NewStore(store, storePrefix)
		pageRes, err = query.FilteredPaginate(indexStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
			_, stakingCoinDenom := types.ParseRewardsByFarmerIndexKey(append(storePrefix, key...))
			if req.StakingCoinDenom != "" {
				if stakingCoinDenom != req.StakingCoinDenom {
					return false, nil
				}
			}
			reward, found := k.GetReward(ctx, stakingCoinDenom, farmerAcc)
			if !found { // TODO: remove this check
				return false, fmt.Errorf("reward not found")
			}
			if accumulate {
				rewards = append(rewards, reward)
			}
			return true, nil
		})
	} else {
		var storePrefix []byte
		if req.StakingCoinDenom != "" {
			storePrefix = types.GetRewardsByStakingCoinDenomKey(req.StakingCoinDenom)
		} else {
			storePrefix = types.RewardKeyPrefix
		}
		rewardStore := prefix.NewStore(store, storePrefix)

		pageRes, err = query.Paginate(rewardStore, req.Pagination, func(key, value []byte) error {
			stakingCoinDenom, farmerAcc := types.ParseRewardKey(append(storePrefix, key...))
			rewardCoins, err := k.UnmarshalRewardCoins(value)
			if err != nil {
				return err
			}
			rewards = append(rewards, types.Reward{
				Farmer:           farmerAcc.String(),
				StakingCoinDenom: stakingCoinDenom,
				RewardCoins:      rewardCoins.RewardCoins,
			})
			return nil
		})
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryRewardsResponse{Rewards: rewards, Pagination: pageRes}, nil
}
