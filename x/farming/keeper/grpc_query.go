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

	"github.com/crescent-network/crescent/v2/x/farming/types"
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
			if plan.IsTerminated() != terminated {
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

// Position queries farming position for a farmer.
func (k Querier) Position(c context.Context, req *types.QueryPositionRequest) (*types.QueryPositionResponse, error) {
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

	resp := &types.QueryPositionResponse{
		StakedCoins: sdk.Coins{},
		QueuedCoins: sdk.Coins{},
		Rewards:     sdk.Coins{},
	}
	if req.StakingCoinDenom == "" {
		resp.StakedCoins = k.Keeper.GetAllStakedCoinsByFarmer(ctx, farmerAcc)
		resp.QueuedCoins = k.Keeper.GetAllQueuedCoinsByFarmer(ctx, farmerAcc)
		resp.Rewards = k.Keeper.AllRewards(ctx, farmerAcc).Add(k.Keeper.AllUnharvestedRewards(ctx, farmerAcc)...)
	} else {
		staking, found := k.Keeper.GetStaking(ctx, req.StakingCoinDenom, farmerAcc)
		if found {
			resp.StakedCoins = resp.StakedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, staking.Amount))
		}
		queuedStakingAmt := k.Keeper.GetAllQueuedStakingAmountByFarmerAndDenom(ctx, farmerAcc, req.StakingCoinDenom)
		if queuedStakingAmt.IsPositive() {
			resp.QueuedCoins = resp.QueuedCoins.Add(sdk.NewCoin(req.StakingCoinDenom, queuedStakingAmt))
		}
		unharvested, _ := k.Keeper.GetUnharvestedRewards(ctx, farmerAcc, req.StakingCoinDenom)
		resp.Rewards = k.Keeper.Rewards(ctx, farmerAcc, req.StakingCoinDenom).Add(unharvested.Rewards...)
	}

	return resp, nil
}

// Stakings queries all stakings of the farmer.
func (k Querier) Stakings(c context.Context, req *types.QueryStakingsRequest) (*types.QueryStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.GetStakingsByFarmerPrefix(farmerAcc)
	stakingStore := prefix.NewStore(store, keyPrefix)
	var stakings []types.StakingResponse

	pageRes, _ := query.FilteredPaginate(stakingStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, stakingCoinDenom := types.ParseStakingIndexKey(append(keyPrefix, key...))

		if req.StakingCoinDenom != "" && stakingCoinDenom != req.StakingCoinDenom {
			return false, nil
		}

		staking, _ := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)

		if accumulate {
			stakings = append(stakings, types.StakingResponse{
				StakingCoinDenom: stakingCoinDenom,
				Amount:           staking.Amount,
				StartingEpoch:    staking.StartingEpoch,
			})
		}

		return true, nil
	})

	return &types.QueryStakingsResponse{Stakings: stakings, Pagination: pageRes}, nil
}

// QueuedStakings queries all queued stakings of the farmer.
func (k Querier) QueuedStakings(c context.Context, req *types.QueryQueuedStakingsRequest) (*types.QueryQueuedStakingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	var keyPrefix []byte
	if req.StakingCoinDenom != "" {
		keyPrefix = types.GetQueuedStakingsByFarmerAndDenomPrefix(farmerAcc, req.StakingCoinDenom)
	} else {
		keyPrefix = types.GetQueuedStakingsByFarmerPrefix(farmerAcc)
	}
	queuedStakingStore := prefix.NewStore(store, keyPrefix)
	var queuedStakings []types.QueuedStakingResponse

	pageRes, _ := query.FilteredPaginate(queuedStakingStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, stakingCoinDenom, endTime := types.ParseQueuedStakingIndexKey(append(keyPrefix, key...))

		queuedStaking, _ := k.GetQueuedStaking(ctx, endTime, stakingCoinDenom, farmerAcc)

		if accumulate {
			queuedStakings = append(queuedStakings, types.QueuedStakingResponse{
				StakingCoinDenom: stakingCoinDenom,
				Amount:           queuedStaking.Amount,
				EndTime:          endTime,
			})
		}

		return true, nil
	})

	return &types.QueryQueuedStakingsResponse{QueuedStakings: queuedStakings, Pagination: pageRes}, nil
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

// Rewards queries all accumulated rewards for a farmer.
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
	store := ctx.KVStore(k.storeKey)
	var keyPrefix []byte
	if req.StakingCoinDenom != "" {
		keyPrefix = types.GetStakingIndexKey(farmerAcc, req.StakingCoinDenom)
	} else {
		keyPrefix = types.GetStakingsByFarmerPrefix(farmerAcc)
	}
	stakingStore := prefix.NewStore(store, keyPrefix)
	var rewards []types.RewardsResponse

	pageRes, _ := query.FilteredPaginate(stakingStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, stakingCoinDenom := types.ParseStakingIndexKey(append(keyPrefix, key...))

		r := k.Keeper.Rewards(ctx, farmerAcc, stakingCoinDenom)

		if accumulate {
			rewards = append(rewards, types.RewardsResponse{
				StakingCoinDenom: stakingCoinDenom,
				Rewards:          r,
			})
		}

		return true, nil
	})

	return &types.QueryRewardsResponse{Rewards: rewards, Pagination: pageRes}, nil
}

// UnharvestedRewards queries all unharvested rewards for the farmer.
func (k Querier) UnharvestedRewards(c context.Context, req *types.QueryUnharvestedRewardsRequest) (*types.QueryUnharvestedRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	farmerAcc, err := sdk.AccAddressFromBech32(req.Farmer)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	var keyPrefix []byte
	if req.StakingCoinDenom != "" {
		keyPrefix = types.GetUnharvestedRewardsKey(farmerAcc, req.StakingCoinDenom)
	} else {
		keyPrefix = types.GetUnharvestedRewardsPrefix(farmerAcc)
	}
	unharvestedRewardsStore := prefix.NewStore(store, keyPrefix)
	var unharvestedRewards []types.UnharvestedRewardsResponse

	pageRes, _ := query.FilteredPaginate(unharvestedRewardsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, stakingCoinDenom := types.ParseUnharvestedRewardsKey(append(keyPrefix, key...))

		unharvested, _ := k.GetUnharvestedRewards(ctx, farmerAcc, stakingCoinDenom)

		if accumulate {
			unharvestedRewards = append(unharvestedRewards, types.UnharvestedRewardsResponse{
				StakingCoinDenom: stakingCoinDenom,
				Rewards:          unharvested.Rewards,
			})
		}

		return true, nil
	})

	return &types.QueryUnharvestedRewardsResponse{UnharvestedRewards: unharvestedRewards, Pagination: pageRes}, nil
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

// HistoricalRewards queries HistoricalRewards records for a staking coin denom.
func (k Querier) HistoricalRewards(c context.Context, req *types.QueryHistoricalRewardsRequest) (*types.QueryHistoricalRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := sdk.ValidateDenom(req.StakingCoinDenom); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid staking coin denom: %v", err)
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	keyPrefix := types.GetHistoricalRewardsPrefix(req.StakingCoinDenom)
	historicalRewardsStore := prefix.NewStore(store, keyPrefix)
	var historicalRewards []types.HistoricalRewardsResponse

	pageRes, _ := query.FilteredPaginate(historicalRewardsStore, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		_, epoch := types.ParseHistoricalRewardsKey(append(keyPrefix, key...))

		var rewards types.HistoricalRewards
		k.cdc.MustUnmarshal(value, &rewards)

		if accumulate {
			historicalRewards = append(historicalRewards, types.HistoricalRewardsResponse{
				Epoch:                 epoch,
				CumulativeUnitRewards: rewards.CumulativeUnitRewards,
			})
		}

		return true, nil
	})

	return &types.QueryHistoricalRewardsResponse{HistoricalRewards: historicalRewards, Pagination: pageRes}, nil

}
