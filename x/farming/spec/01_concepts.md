<!-- order: 1 -->

 # Concepts
## Farming Module

`x/farming` is a Cosmos SDK module that implements farming functionality that keeps track of the staking and provides farming rewards to farmers. One use case is to use this module to provide incentives for liquidity pool investors for their pool participation. 

## Plans

There are two types of farming plans in the `farming` module as below.

### 1. Public Farming Plan

A public farming plan can only be created through governance proposal meaning that the proporsal must be first agreed and passed in order to create a public plan.
### 2. Private Farming Plan

A private farming plan can be created with any account. The plan creator's account is used as distributing account `FarmingPoolAddress` that will be distributed to farmers automatically. There is a fee `PlanCreationFee` paid upon plan creation to prevent from spamming attack. 

## Distribution Methods

There are two types of distribution methods  in the `farming` module as below.
### 1. Fixed Amount Plan

A `FixedAmountPlan` distributes fixed amount of coins to farmers for every epoch day. If the plan creators `FarmingPoolAddress` is depleted with distributing coins, then there is no more coins to distribute unless it is filled up again.

### 2. Ratio Plan

A `RatioPlan` distributes to farmers by ratio distribution for every epoch day. If the plan creators `FarmingPoolAddress` is depleted with distributing coins, then there is no more coins to distribute unless it is filled up with more coins.

