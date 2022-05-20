<!-- order: 3 -->

 # State Transitions

This document describes the state transaction operations for the farming module.

## Plans

As stated in [Concepts](01_concepts.md), both public and private farming plans are available in the `farming` module:

- A public farming plan can be created only through governance proposal.
- A private farming plan can be created with any account. 


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

- Reserves the amount of coins to the staking reserve account for each staking coin denom `address.Module(ModuleName, []byte("StakingReserveAcc|"+stakingCoinDenom))` 
- Creates `QueuedStaking` object, which waits in a queue for `CurrentEpochDays` days and then moves to the `Staking` object
- Imposes more gas if the farmer already has `Staking` with the same coin denom. See [Parameters](07_params.md#DelayedStakingGasFee) for details.

## Unstake

When a farmer unstakes an amount of coins, the following state transitions occur:

- Adds `Staking` and `QueuedStaking` amounts to see if the unstaking amount is sufficient
- Moves accrued rewards for the coin denom to `UnharvestedRewardsReserveAcc` and increases amount of `UnharvestedRewards`
  - When unstaking all staked amount, instead of increasing `UnharvestedRewards`, sends all `UnharvestedRewards` and accrued rewards to the farmer directly
- Subtracts the unstaking amount of coins from `QueuedStaking` first(in last-in-first-out manner), and if not sufficient then subtracts from `Staking`
- Releases the unstaking amount of coins to the farmer

## Harvest (Reward Withdrawal)

- Calculates `CumulativeUnitRewards` in `HistoricalRewards` object in order to get the rewards for the staking coin denom that are accumulated over the last epochs 
- Releases the accumulated rewards to the farmer if it is not zero and decreases the `OutstandingRewards`
- When there is positive `UnharvestedRewards`, sends the rewards to the farmer and deletes the `UnharvestedRewards` object
- Sets `StartingEpoch` in `Staking` object

## Reward Allocation

If the sum of total calculated `EpochAmount` (or `EpochRatio` multiplied by the farming pool balance) exceeds the farming pool balance, then skip the reward allocation for that epoch.

For each abci end block call, the operations to update the rewards allocation are:

- Moves `QueuedStaking` amount to `Staking` when its end time has passed
- Calculates rewards allocation information for the end of the current epoch depending on plan type `FixedAmountPlan` or `RatioPlan`
- Distributes total allocated coins from each plan’s farming pool address `FarmingPoolAddress` to the rewards reserve pool account `RewardsReserveAcc`
- Calculates staking coin weight for each denom in each plan and gets the unit rewards by denom
- Updates `HistoricalRewards` and `CurrentEpoch` based on the allocation information