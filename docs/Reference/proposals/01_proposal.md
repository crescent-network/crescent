# Signal Proposal

Adopting the `Budget` and `Farming` Modules on Cosmos Hub

By voting yes to this signal proposal, the voter agrees to adopt the Budget Module and Farming Module on Cosmos Hub to allow an incentivization mechanism for userbase growth on Cosmos Hub. The Tendermint team is currently building the two suggested modules, and when this signal proposal passes, the Budget Module and Farming Module will be included in a future Gaia upgrade when both modules are ready to be deployed.

## 1. Introduction

### 1.1 Modern Blockchain Incentivization Scheme: Farming

Lots of modern blockchain ecosystems and decentralized applications incentivize their platform users by distributing coins based on user activities to bootstrap the growth of the userbase. Generally, the incentivization methodology is called "Farming." The source of the farming is various, but the most popular way is to utilize the native coins for current and future platform users.

### 1.2. Cosmos Hub Context: Budget & Farming

With pipelines of new features included on Cosmos Hub, including IBC, bridges, and Gravity DEX, Cosmos Hub needs to incentivize not only dPoS delegators but also platform users to accelerate user adoption. So we need to adopt two kinds of features as below:

- Budget Module: To define and execute budget plans of ATOM inflation for multiple objectives

- Farming Module: To define and distribute incentives for various utility users on Cosmos Hub

## 2. Features

### 2.1. Budget Module

Budget Plan

The Budget Module manages a list of budget plans which describe each proportional distribution of ATOM inflation to different destinations. The Budget Module distributes ATOM inflation according to the existing list of budget plans.

Governance Process

The list of budget plans can be added/removed/modified by a parameter governance proposal.

### 2.2. Farming Module

Farming Plan

A farming plan is a definition of reward distribution plan with two types:

- Public Farming Plan

    - Creation: A public farming plan can be created only by the governance process

    - Farming Pool Address: The source of a public farming plan is an existing module account

- Private Farming Plan

    - Creation: A private farming plan can be created by anyone, submitting a transaction

    - Farming Pool Address: A new farming pool address is assigned to the farming plan. Anyone can fund this farming pool by sending coins to this address.

Reward Distribution

A farming plan defines a list of staking coin weights which are used to calculate the proportional distribution of rewards to each farmer. From the total reward distribution, each staking coin gets the weight proportion defined in the weight list. Then, each farmer who staked this coin receives the amount of corresponding rewards based on their proportion of the staked coin amount from the total staked amount.

Reward Harvest

A farmer can harvest (withdraw) accumulated rewards anytime he/she wants. Rewards are calculated based on a predefined epoch, therefore farmers can harvest rewards accumulated until the last epoch.

### 2.3. Gravity DEX Liquidity Incentivization

Staking Coins as Pool Coins

Staking coins in a farming plan can be defined as a group of pool coins to distribute the farming plan rewards to pool coin holders. Because every liquidity provider on Gravity DEX gets pool coins as evidence of liquidity providing, this methodology naturally provides us the way to incentivize liquidity providing on Gravity DEX

Governance Processes

We need two kinds of governance processes to activate Gravity DEX liquidity incentivization

- Budget: A governance process to decide 

    - percentage of ATOM inflation to be used for Gravity DEX liquidity incentivization

    - time period of the budget plan created by this governance process

- Farming: A governance process to decide

    - list and weights of staking coins (pool coins) to be incentivized

    - time period of the farming plan created by this governance process

## 3. Detail Spec

Detail description of the spec can be found below:

- Budget Module: https://github.com/tendermint/budget/tree/master/x/budget/spec

- Farming Module: https://github.com/tendermint/farming/tree/master/x/farming/spec