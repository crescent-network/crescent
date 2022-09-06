<!-- order: 1 -->

 # Concepts

## Farming Module

The `x/farming` Cosmos SDK module implements farming functionality that keeps track of staking and provides farming rewards to farmers. A primary use case of this module is to provide incentives for liquidity pool investors for their pool participation. 

## Plans

There are two types of farming plans in the `farming` module:

### Public Farming Plan

A public farming plan can be created only through governance proposal. 

- The proposal must be first agreed and passed before a public farming plan can be created. 
- A creation fee is not required.

### Private Farming Plan

A private farming plan can be created with any account. 

- The address of the plan creator account is used as the `TerminationAddress`. When the plan ends after its end time, the balance of the farming pool address is transferred to the termination address.
- To prevent spamming attacks, the `PlanCreationFee` fee must be paid on plan creation. 
- Internally, the private plan's farming pool address is derived from the following derivation rule of `address.Module(ModuleName, []byte("PrivatePlan|{planId}|{planName}"))` and it is assigned to the plan. 
- After creation, need to query the plan and send the amount of coins to the farming pool address so that the plan distributes as intended.

## Distribution Methods

There are two types of reward distribution methods in the `farming` module:

### Fixed Amount Plan

A `FixedAmountPlan` distributes a fixed amount of coins to farmers for every epoch day. 

When the plan creator's `FarmingPoolAddress` is depleted, then there are no more coins to distribute until more coins are added to the account.

### Ratio Plan

A `RatioPlan` distributes coins to farmers by ratio distribution for every epoch day. 

If the plan creator's `FarmingPoolAddress` is depleted, then there are no more coins to distribute until more coins are added to the account.

## Accumulated Reward Calculation

In the farming module, farming rewards are calculated per epoch based on the distribution plan. 

To calculate the rewards for a single farmer, take the total rewards for the epoch before the staking started, minus the current total rewards. 

The farming module takes references from [F1 Fee Distribution](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/fee_distribution/f1_fee_distr.pdf) that is used in the Cosmos SDK [x/distribution](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/x/distribution/spec/01_concepts.md) module.

### Accumulated Unit Reward 

`HistoricalRewards` represents accumulated rewards for each staking coin with amount 1.

### Base Algorithm 

`HistoricalRewards` for each staking coin for every epoch can be calculated as the following algorithm:
<!-- markdown-link-check-disable -->
- ![](https://latex.codecogs.com/svg.latex?\Large&space;\sum_{i=0}^{now}\frac{TR_i}{TS_i})
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;i) : each epoch
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;now) : `CurrentEpoch`
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TS_i) : total staking amount of the staking coin for epoch i
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TR_i) : total reward amount of the staking coin for epoch i

Accumulated rewards from any staking position can be calculated from `HistoricalRewards` and the staking amount of the position as the following algorithm:

- ![](https://latex.codecogs.com/svg.latex?\Large&space;x*\(\sum_{i=0}^{now}\frac{TR_i}{TS_i}-\sum_{i=0}^{start}\frac{TR_i}{TS_i}\))
    - assuming constant staking amount for the staking epochs
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;x) : staking amount for the staking period