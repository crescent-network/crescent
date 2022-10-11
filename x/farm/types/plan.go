package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const day = 24 * time.Hour

// NewPlan creates a new Plan.
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

// NewPairRewardAllocation creates a new RewardAllocation for a pair.
func NewPairRewardAllocation(pairId uint64, rewardsPerDay sdk.Coins) RewardAllocation {
	return RewardAllocation{
		PairId:        pairId,
		RewardsPerDay: rewardsPerDay,
	}
}

// NewDenomRewardAllocation creates a new RewardAllocation for a target denom.
func NewDenomRewardAllocation(denom string, rewardsPerDay sdk.Coins) RewardAllocation {
	return RewardAllocation{
		Denom:         denom,
		RewardsPerDay: rewardsPerDay,
	}
}

// ValidateRewardAllocations validates a slice of RewardAllocation.
// It also checks whether there's any duplication of pair id among the
// reward allocations.
func ValidateRewardAllocations(rewardAllocs []RewardAllocation) error {
	if len(rewardAllocs) == 0 {
		return fmt.Errorf("empty reward allocations")
	}
	denomSet := map[string]struct{}{}
	pairIdSet := map[uint64]struct{}{}
	for _, rewardAlloc := range rewardAllocs {
		if rewardAlloc.Denom == "" && rewardAlloc.PairId == 0 {
			return fmt.Errorf("target denom or pair id must be specified")
		} else if rewardAlloc.Denom != "" && rewardAlloc.PairId != 0 {
			return fmt.Errorf("target denom and pair id cannot be specified together")
		}
		if rewardAlloc.Denom != "" {
			if err := sdk.ValidateDenom(rewardAlloc.Denom); err != nil {
				return fmt.Errorf("invalid target denom: %w", err)
			}
			if _, ok := denomSet[rewardAlloc.Denom]; ok {
				return fmt.Errorf("duplicate target denom: %s", rewardAlloc.Denom)
			}
			denomSet[rewardAlloc.Denom] = struct{}{}
		} else if rewardAlloc.PairId > 0 {
			if _, ok := pairIdSet[rewardAlloc.PairId]; ok {
				return fmt.Errorf("duplicate pair id: %d", rewardAlloc.PairId)
			}
			pairIdSet[rewardAlloc.PairId] = struct{}{}
		}
		if err := rewardAlloc.RewardsPerDay.Validate(); err != nil {
			return fmt.Errorf("invalid rewards per day: %w", err)
		}
		// TODO: reject too big rewardsPerDay which can cause an overflow
	}
	return nil
}
