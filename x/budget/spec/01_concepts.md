<!-- order: 1 -->

# Concepts

## Budget Module

`x/budget` is a simple Cosmos SDK module that implements budget functionality. 

After the module is agreed within the community, voted, and passed, the core functionality of this independent module enables anyone to create a budget plan through parameter change governance proposal. 

High level overview: 

- The budget module uses `SourceAddress` to distribute to the `DestinationAddress`.
- The amount of coins is relative to the rate defined in the plan. 
- The module collects all budgets at each `BeginBlock`.
- Distribution takes place at every `EpochBlocks`, a global parameter that can be modified by a governance proposal.

A primary use case for the budget module is for the Gravity DEX farming plan. 

Use the budget module to create a budget plan that has `SourceAddress` for the Cosmos Hub [FeeCollector](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/auth/types/keys.go#L15) module account that collects transaction gas fees and part of ATOM inflation. Then, `SourceAddress` plans to distribute some amount of coins to `DestinationAddress` for farming plan.

### Budget Plan for ATOM Inflation Use Case

The Cosmos SDK current reward workflow:

- In AnteHandler:

    - Gas fees are collected in ante handler and are sent to the `FeeCollectorName` module account

    - Reference the following lines of code:

      +++ https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/auth/ante/fee.go#L112-L140

- In `x/mint` module:

  - ATOM inflation is minted in `x/mint` module and is sent to the `FeeCollectorName` module account

  - Reference the following lines of code:

    +++ https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/mint/abci.go#L27-L40

    +++ https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/mint/keeper/keeper.go#L108-L110

- In `x/distribution` module:

  - Send all rewards in `FeeCollectorName` to distribution module account
  
  - From `distributionModuleAccount`, substitute `communityTax`

  - Remaining rewards are distributed to proposer and validator reward pools

  - Substituted amount for community budget is saved in key-value store

  - Reference the following lines of code:

    +++ https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/distribution/keeper/allocation.go#L13-L102

Implementation with `x/budget` module:

  - A budget module is independent of all other Cosmos SDK modules

  - In chains that where there will be budget plans with `SourceAddress` set to `FeeCollectorName`, it should be set as follows:

    - BeginBlock processing order should be mint module → budget module → distribution module

    - if inflation and gas fees occur every block, `params.EpochBlocks` should be set to 1

    - It should be noted that if the rate sum of these budget plans is 1.0 (100%), inflation and gas fees can not go to validators

  - Distribute ATOM inflation and transaction gas fees to different budget purposes:

    - ATOM inflation and gas fees are accumulated in `FeeCollectorName` module account

    - Distribute budget amounts from `FeeCollectorName` module account to each `DestinationAddress` for budget plans with `SourceAddress` set to `FeeCollectorName`

    - Remaining amounts stay in `FeeCollectorName` so that distribution module can use them for community fund and staking rewards distribution (no change to current `FeeCollectorName` implementation)

  - Create, modify or remove budget plans by using governance process:
  
    - A budget plan can be created, modified, or removed by parameter change governance proposal
