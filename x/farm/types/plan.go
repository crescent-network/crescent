package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewPlan returns a new Plan.
func NewPlan(
	id uint64, description string, srcAddr, termAddr sdk.AccAddress,
	rewardAllocs []RewardAllocation, startTime, endTime time.Time,
	isPrivate bool) Plan {
	return Plan{
		Id:                 id,
		Description:        description,
		SourceAddress:      srcAddr.String(),
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
		IsPrivate:          isPrivate,
		IsTerminated:       false,
	}
}

// NewRewardAllocation returns a new RewardAllocation.
func NewRewardAllocation(pairId uint64, rewardsPerDay sdk.DecCoins) RewardAllocation {
	return RewardAllocation{
		PairId:        pairId,
		RewardsPerDay: rewardsPerDay,
	}
}