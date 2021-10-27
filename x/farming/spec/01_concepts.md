<!-- order: 1 -->

 # Concepts
## Farming Module

`x/farming` is a Cosmos SDK module that implements farming functionality that keeps track of the staking and provides farming rewards to farmers. A primary use case is to use this module to provide incentives for liquidity pool investors for their pool participation. 

## Plans

There are two types of farming plans in the `farming` module as below.

### 1. Public Farming Plan

A public farming plan can only be created through governance proposal meaning that the proposal must be first agreed and passed in order to create a public plan.
### 2. Private Farming Plan

A private farming plan can be created with any account. The plan creator's account is used as `TerminationAddress`. There is a fee `PlanCreationFee` paid upon plan creation to prevent from spamming attack. 

## Distribution Methods

There are two types of distribution methods  in the `farming` module as below.
### 1. Fixed Amount Plan

A `FixedAmountPlan` distributes fixed amount of coins to farmers for every epoch day. If the plan creators `FarmingPoolAddress` is depleted with distributing coins, then there is no more coins to distribute unless it is filled up again.

### 2. Ratio Plan

A `RatioPlan` distributes to farmers by ratio distribution for every epoch day. If the plan creators `FarmingPoolAddress` is depleted with distributing coins, then there is no more coins to distribute unless it is filled up with more coins.

## Accumulated Reward Calculation

In farming module, farming rewards are calculated per epoch based on plans. The rewards for a single farmer can be calculated by taking the total rewards for the epoch before the staking started, minus the current total rewards. The farming module takes references from [F1 Fee Distribution](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/fee_distribution/f1_fee_distr.pdf) that is used in Cosmos SDK [x/distribution](https://github.com/cosmos/cosmos-sdk/blob/master/x/distribution/spec/01_concepts.md) module.

### Accumulated Unit Reward 

`HistoricalRewards` represents accumulated rewards for each staking coin with amount 1.

### Base Algorithm 

`HistoricalRewards` for each staking coin for every epoch can be calculated as the following algorithm:

- ![](https://latex.codecogs.com/svg.latex?\Large&space;\sum_{i=0}^{now}\frac{TR_i}{TS_i})
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;i) : each epoch
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;now) : `CurrentEpoch`
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TS_i) : total staking amount of the staking coin for epoch i
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;TR_i) : total reward amount of the staking coin for epoch i

Accumulated rewards from any staking position can be calculated from `HistoricalRewards` and the staking amount of the position as the following algorithm:

- ![](https://latex.codecogs.com/svg.latex?\Large&space;x*\(\sum_{i=0}^{now}\frac{TR_i}{TS_i}-\sum_{i=0}^{start}\frac{TR_i}{TS_i}\))
    - assuming constant staking amount for the staking epochs
    - ![](https://latex.codecogs.com/svg.latex?\Large&space;x) : staking amount for the staking period