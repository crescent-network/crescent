<!-- order: 3 -->

 # State Transitions

This document describes the state transaction operations pertaining to the farming module. 

As stated in [01_concepts.md](01_concepts.md), there are public and private farming plans available in the `farming` module. Public plan can be created by any account whereas private plan can only be created through governance proposal.

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
    - Each `farmingPlan` has its own `epochDays` : farming rewards distribution frequency
- Distribution Method
    - `FixedAmountPlan`
        - fixed amount of coins are distributed for each `epochDays`
        - amount in `sdk.Coins`
    - `RatioPlan`
        - `epochRatio` of total assets in `farmingPoolAddress` is distributed for each `epochDays`
        - `epochRatio` in percentage
- Termination Address
    - When the plan ends after the `endTime`, transfer the balance of `farmingPoolAddress` to  `terminationAddress`.