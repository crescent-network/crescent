<!-- order: 3 -->

 # State Transitions

This document describes the state transaction operations pertaining to the farming module.

## Plan

As stated in [01_concepts.md](01_concepts.md), there are public and private farming plans available in the `farming` module. Private plan can be created by any account whereas public plan can only be created through governance proposal.

```go
// PlanType enumerates the valid types of a plan.
type PlanType int32

const (
    // PLAN_TYPE_UNSPECIFIED defines the default plan type.
    PlanTypeNil PlanType = 0
    // PLAN_TYPE_PUBLIC defines the public plan type.
    PlanTypePublic PlanType = 1
    // PLAN_TYPE_PRIVATE defines the private plan type.
    PlanTypePrivate PlanType = 2
)
```

## Stake

When a farmer stakes an amount of coins, the following state transitions occur:

- it reserves the amount of coins to the staking reserve pool account `StakingReservePoolAcc` 
- it creates `QueuedStaking` object and stores the staking coins in `QueueStaking`, which are waiting in a queue until the end of epoch to move to `Staking` object
- it imposes more gas if the farmer already has `Staking` with the same coin denom(see [07_params.md](07_params.md#DelayedStakingGasFee) for details)

## Unstake

When a farmer unstakes an amount of coins, the following state transitions occur:

- it adds `Staking` and `QueueStaking` amounts to see if the unstaking amount is sufficient
- it automatically withdraws rewards for the coin denom which are accumulated over the last epochs
- it subtracts the unstaking amount of coins from `QueueStaking` first and if it is not sufficient then it subtracts from `Staking`
- it releases the unstaking amount of coins to the farmer

## Harvest (Reward Withdrawal)

- it calculates `CumulativeUnitRewards` in `HistoricalRewards` object in order to get the rewards for the staking coin denom which are accumulated over the last epochs for the farmer
- it releases the accumulated rewards to the farmer if it is not zero and decreases the `OutstandingRewards`
- it sets `StartingEpoch` in `Staking` object

## Reward Allocation

If the sum of total calculated `EpochAmount` (or `EpochRatio` multiplied by the farming pool's balance) exceeds the farming pool's balance, then the reward allocation is skipped for that epoch.

Each abci end block call, the operations to update rewards allocation are to execute:

++ https://github.com/tendermint/farming/blob/69db071ce30b99617b8ba9bb6efac76e74cd100b/x/farming/keeper/reward.go#L363-L426

- it calculates rewards allocation information for the end of the current epoch depending on plan type `FixedAmountPlan` or `RatioPlan`
- it distributes total allocated coins from each plan’s farming pool address `FarmingPoolAddress` to the rewards reserve pool account `RewardsReserveAcc`
- it calculates staking coin weight for each denom in each plan and gets the unit rewards by denom
- it updates `HistoricalRewards` and `CurrentEpoch` based on the allocation information
- it deletes `QueueStaking` object after moving `QueueCoins` to `StakedCoins` in `Staking` object