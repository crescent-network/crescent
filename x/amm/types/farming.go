package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	MaxPlanDescriptionLen = 200 // Maximum length of a plan's description

	day = 24 * time.Hour
)

var RewardsPoolAddress = address.Module(ModuleName, []byte("RewardsPool"))

func DeriveFarmingPoolAddress(planId uint64) sdk.AccAddress {
	return address.Module(ModuleName, []byte(fmt.Sprintf("FarmingPool/%d", planId)))
}

func NewFarmingPlan(
	id uint64, description string, farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []FarmingRewardAllocation, startTime, endTime time.Time,
	isPrivate bool) FarmingPlan {
	return FarmingPlan{
		Id:                 id,
		Description:        description,
		FarmingPoolAddress: farmingPoolAddr.String(),
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
		IsPrivate:          isPrivate,
		IsTerminated:       false,
	}
}

func (plan FarmingPlan) IsActiveAt(t time.Time) bool {
	return !plan.StartTime.After(t) && plan.EndTime.After(t)
}

func (plan FarmingPlan) Validate() error {
	if plan.Id == 0 {
		return fmt.Errorf("plan id must be positive")
	}
	if len(plan.Description) > MaxPlanDescriptionLen {
		return fmt.Errorf("too long plan description, maximum %d", MaxPlanDescriptionLen)
	}
	if _, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress); err != nil {
		return fmt.Errorf("invalid farming pool address: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(plan.TerminationAddress); err != nil {
		return fmt.Errorf("invalid termination address: %w", err)
	}
	if err := ValidateFarmingRewardAllocations(plan.RewardAllocations); err != nil {
		return fmt.Errorf("invalid reward allocations: %w", err)
	}
	if !plan.StartTime.Before(plan.EndTime) {
		return fmt.Errorf("end time must be after start time")
	}
	return nil
}

func NewFarmingRewardAllocation(poolId uint64, rewardsPerDay sdk.Coins) FarmingRewardAllocation {
	return FarmingRewardAllocation{
		PoolId:        poolId,
		RewardsPerDay: rewardsPerDay,
	}
}

func ValidateFarmingRewardAllocations(rewardAllocs []FarmingRewardAllocation) error {
	if len(rewardAllocs) == 0 {
		return fmt.Errorf("empty reward allocations")
	}
	poolIdSet := map[uint64]struct{}{}
	for _, rewardAlloc := range rewardAllocs {
		if rewardAlloc.PoolId == 0 {
			return fmt.Errorf("pool id must not be 0")
		}
		if _, ok := poolIdSet[rewardAlloc.PoolId]; ok {
			return fmt.Errorf("duplicate pool id: %d", rewardAlloc.PoolId)
		}
		poolIdSet[rewardAlloc.PoolId] = struct{}{}
		if err := rewardAlloc.RewardsPerDay.Validate(); err != nil {
			return fmt.Errorf("invalid rewards per day: %w", err)
		}
		overflow := false
		utils.SafeMath(func() {
			RewardsForBlock(rewardAlloc.RewardsPerDay, day)
		}, func() {
			overflow = true
		})
		if overflow {
			return fmt.Errorf("too much rewards per day")
		}
	}
	return nil
}

func RewardsForBlock(rewardsPerDay sdk.Coins, blockDuration time.Duration) sdk.DecCoins {
	return sdk.NewDecCoinsFromCoins(rewardsPerDay...).
		MulDecTruncate(sdk.NewDec(blockDuration.Milliseconds())).
		QuoDecTruncate(sdk.NewDec(day.Milliseconds()))
}
