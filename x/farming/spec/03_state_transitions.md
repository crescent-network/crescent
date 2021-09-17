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
    - fixed amount of coins are distributed per `CurrentEpochDays`
    - `epochAmount` is `sdk.Coins`
  - `RatioPlan`
    - ratio of total assets in `farmingPoolAddress` is distributed per `CurrentEpochDays`
    - `epochRatio` is in percentage
- Termination Address
  - When the plan ends after the `endTime`, transfer the balance of `farmingPoolAddress` to `terminationAddress`.

## Staking

- New `Staking` object is created when a farmer creates a staking, and when the farmer does not have existing `Staking`.
- When a farmer add/remove stakings to/from existing `Staking`, `StakedCoins` and `QueuedCoins` are updated in the corresponding `Staking`.
- `QueuedCoins` : newly staked coins are in this status until end of current epoch, and then migrated to `StakedCoins` at the end of current epoch.
- When a farmer unstakes, `QueuedCoins` are unstaked first, and then `StakedCoins`.

## Reward Withdrawal

To assume constant staking amount for reward withdrawal, automatic withdrawal is designed as below:
- Add staking position : When there exists `QueuedCoins` in `Staking` at the end of current epoch
  - accumulated rewards until current epoch are automatically withdrawn
  - `StartEpochId` is modified to the `EpochId` of the next epoch
  - `QueuedCoins` is migrated to `StakedCoins`
- Remove staking position : When a farmer remove stakings from `StakedCoins`
  - accumulated rewards until last epoch are immediately withdrawn
  - `StartEpochId` is modified to the `EpochId` of the current epoch
  - unstake executed immediately and `StakedCoins` are reduced accordingly
- Manual reward withdrawal : When a farmer request a reward withdrawal
  - accumulated rewards until last epoch are immediately withdrawn
  - `StartEpochId` is modified to the `EpochId` of the current epoch

## Accumulated Reward Calculation

- Accumulated Unit Reward : AUR represents accumulated rewards(for each staking coin) of a staking position with amount 1.
- AUR for each staking coin for each epoch can be calculated as below
  - ![](https://latex.codecogs.com/svg.latex?\Large&space;\sum_{i=0}^{now}\frac{TR_i}{TS_i})
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;i) : each `EpochId`
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;now) : current `EpochId`
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TS_i) : total staking amount of the staking coin for epoch i
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TR_i) : total reward amount of the staking coin for epoch i
- Accumulated rewards from any staking position can be calculated from AUR and the staking amount of the position as below
  - ![](https://latex.codecogs.com/svg.latex?\Large&space;x*\(\sum_{i=0}^{now}\frac{TR_i}{TS_i}-\sum_{i=0}^{start}\frac{TR_i}{TS_i}\))
    - assuming constant staking amount for the staking epochs
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;x) : staking amount for the staking period