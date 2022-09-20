package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const day = 24 * time.Hour

// NewPlan returns a new Plan.
func NewPlan(
	id uint64, description string, farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []RewardAllocation, startTime, endTime time.Time,
	isPrivate bool) Plan {
	return Plan{
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

// IsActiveAt returns whether the plan is active(being able to distribute rewards)
// at given time t.
func (plan Plan) IsActiveAt(t time.Time) bool {
	return !plan.StartTime.After(t) && plan.EndTime.After(t)
}

func (plan Plan) Validate() error {
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
	if err := ValidateRewardAllocations(plan.RewardAllocations); err != nil {
		return fmt.Errorf("invalid reward allocations: %w", err)
	}
	if !plan.StartTime.Before(plan.EndTime) {
		return fmt.Errorf("end time must be after start time")
	}
	return nil
}

func (plan Plan) GetFarmingPoolAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (plan Plan) GetTerminationAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(plan.TerminationAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewRewardAllocation returns a new RewardAllocation.
func NewRewardAllocation(pairId uint64, rewardsPerDay sdk.Coins) RewardAllocation {
	return RewardAllocation{
		PairId:        pairId,
		RewardsPerDay: rewardsPerDay,
	}
}

func ValidateRewardAllocations(rewardAllocs []RewardAllocation) error {
	if len(rewardAllocs) == 0 {
		return fmt.Errorf("empty reward allocations")
	}
	pairIdSet := map[uint64]struct{}{}
	for _, rewardAlloc := range rewardAllocs {
		if rewardAlloc.PairId == 0 {
			return fmt.Errorf("pair id must not be zero")
		}
		if _, ok := pairIdSet[rewardAlloc.PairId]; ok {
			return fmt.Errorf("duplicate pair id: %d", rewardAlloc.PairId)
		}
		pairIdSet[rewardAlloc.PairId] = struct{}{}
		if err := rewardAlloc.RewardsPerDay.Validate(); err != nil {
			return fmt.Errorf("invalid rewards per day: %w", err)
		}
	}
	return nil
}
