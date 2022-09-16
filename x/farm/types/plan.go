package types

import (
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

func (plan Plan) IsActiveAt(t time.Time) bool {
	return !plan.StartTime.After(t) && plan.EndTime.After(t)
}

// NewRewardAllocation returns a new RewardAllocation.
func NewRewardAllocation(pairId uint64, rewardsPerDay sdk.DecCoins) RewardAllocation {
	return RewardAllocation{
		PairId:        pairId,
		RewardsPerDay: rewardsPerDay,
	}
}
