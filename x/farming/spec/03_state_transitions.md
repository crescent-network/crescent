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

- Staking Coins for Farming
  - Each `farmingPlan` predefines list of `stakingCoinWeights` using `sdk.DecCoins`
  - `weight` mean that each group of stakers with each coin `denom` will receive each predefined `weight` of the total rewards
- Multiple Farming Coins within a `farmingPoolAddress`
  - If `farmingPoolAddress` has multiple kinds of coins, then all coins are identically distributed following the given `farmingPlan`
- Time Parameters
  - Each `farmingPlan` has its own `startTime` and `endTime`
- Distribution Method
  - `FixedAmountPlan`
    - fixed amount of coins are distributed for each `epochDays`
    - amount in `sdk.Coins`
  - `RatioPlan`
    - `epochRatio` of total assets in `farmingPoolAddress` is distributed for each `epochDays`
    - `epochRatio` in percentage
- Termination Address
  - When the plan ends after the `endTime`, transfer the balance of `farmingPoolAddress` to `terminationAddress`.

## Staking

- New `Staking` object is created when a farmer creates a staking, and when the farmer does not have existing `Staking`.
- When a farmer creates new staking, the farmer should pay `StakingCreationFee` to prevent spamming.
- When a farmer add/remove stakings to/from existing `Staking`, `StakedCoins` and `QueuedCoins` are updated in the corresponding `Staking`.
- `QueuedCoins` : newly staked coins are in this status until end of current epoch, and then migrated to `StakedCoins` at the end of current epoch.
- When a farmer unstakes, `QueuedCoins` are unstaked first, and then `StakedCoins`.

## Reward

- At every end of epoch, `Reward` are created or updated by the calculation from each `Plan` and corresponding `Staking`.
- Every `StakedCoins` in `Staking` which is eligible for any alive `Plan` accumulates rewards in `RewardCoins`.
- Reward for specific `Plan` and `StakingCoinDenom` = total_reward_of_the_plan_for_this_epoch _ weight_for_this_staking_coin _ (this_denom_staked_coins_for_this_farmer)/(this_denom_total_staked_coins)
- Accumulated `RewardCoins` are withdrawable anytime when the farmer request the withdrawal from the `Reward`.
