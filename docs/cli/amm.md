---
Title: AMM
Description: A high-level overview of how the command-line interfaces (CLI) works for the amm module.
---

# AMM Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `amm` module. 

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [CreatePool](#CreatePool)
  - [AddLiquidity](#AddLiquidity)
  - [RemoveLiquidity](#RemoveLiquidity)
  - [Collect](#Collect)
  - [CreatePrivateFarmingPlan](#CreatePrivateFarmingPlan)
  - [TerminatePrivateFarmingPlan](#TerminatePrivateFarmingPlan)
- [Query](#Query)
  - [Params](#Params)
  - [AllPools](#AllPools)
  - [Pool](#Pool)
  - [AllPositions](#AllPositions)
  - [Position](#Position)
  - [SimulateAddLiquidity](#SimulateAddLiquidity)
  - [SimulateRemoveLiquidity](#SimulateRemoveLiquidity)
  - [CollectibleCoins](#CollectibleCoins)
  - [AllTickInfos](#AllTickInfos)
  - [TickInfo](#TickInfo)
  - [AllFarmingPlans](#AllFarmingPlans)
  - [FarmingPlan](#FarmingPlan)

# Transaction

## CreatePool

Create a pool to market for trading.

A pool is tied to a single market and places orders to market based on the preset logic. Once a pool is created, liquidity providers can create positions.

Usage

```bash
create-pool [market-id] [price]
```

| **Argument**     | **Description**                                        |
| :--------------- | :----------------------------------------------------- |
| market-id        | id of the market where the pool's order will be placed |
| price            | initial pool price                                     |

Example

```bash
# Create a pool
crescentd tx amm create-pool 1 10 \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query pools by using the following command
crescentd q amm pools -o json | jq
```

## AddLiquidity

Add liquidity to existing position or make a new position.

Position(s) belong to a pool. Therefore, a pool must exist in order to create a position. Anyone can create a position with custom range.

Usage

```bash
add-liquidity [pool-id] [lower-price] [upper-price] [desired-amount]
```

| **Argument**   | **Description**                                    |
| :------------  | :------------------------------------------------- |
| pool-id        | pool id                                            |
| lower-price    | lower bound for price range of liquidity providing |
| upper-price    | upper bound for price range of liquidity providing |
| desired-amount | deposit amounts of base and quote coins            |

Example

```bash
# Create a position with 10ATOM/10USD which provide liquidity to price range [9,11] 
crescentd tx amm add-liquidity 1 9 11 10000000uatom,10000000uusd \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query expected result of add-liquidity by using the following command
crescentd q amm add-liquidity-simulation 1 9 11 10000000uatom,10000000uusd -o json | jq
```

## RemoveLiquidity

Withdraw coins from the liquidity providing position.

Withdrawal requests are typically processed in the order they are received, rather than being delayed until the end of a batch.

Usage

```bash
remove-liquidity [position-id] [liquidity]
```

| **Argument**  | **Description**                                     |
| :------------ | :-------------------------------------------------- |
| position-id   | position id                                         |
| liquidity     | amount of liquidity to be removed from the position |

Example

```bash
# Remove 10000 liquidity from position with id 1
crescentd tx amm remove-liquidity 1 10000 \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query expected result of remove-liquidity by using the following command
crescentd q amm remove-liquidity-simulation 1 10000 -o json | jq
```

## Collect

Collect rewards accumulated in a position.

The reward consists of the swap fee and farming reward accumulated in the position.

Usage

```bash
collect [position-id] [amount]
```

| **Argument**  | **Description**                   |
| :------------ | :-------------------------------- |
| position-id   | position id                       |
| amount        | amounts of reward to be withdrawn |

Example

```bash
# Withdraw 10uATOM and 10uUSD of reward from the positino
crescentd tx amm collect 1 10uatom,10uusd \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query collectible reward by using the following command
crescentd crescentd q amm collectible-coins --position-id 1 -o json | jq
```

## CreatePrivateFarmingPlan

Create a new private farming plan.

The newly created plan's farming pool address is automatically generated and will have no balances in the account initially.

Manually send enough reward coins to the generated farming pool address to make sure that the rewards allocation happens.

The plan's termination address is set to the plan creator.

```bash
create-private-farming-plan [description] [termination-address] [start-time] [end-time] [reward-allocations...]
```

| **Argument**          | **Description**                                                                                         |
| :-------------------- | :------------------------------------------------------------------------------------------------------ |
| description           | a brief description of the plan                                                                         |
| termination-address   | address where the remaining farming rewards in the farming pool transferred when the plan is terminated |
| start-time            | the time at which the plan begins, in RFC3339 format                                                    |
| end-time              | the time at which the plan ends, in RFC3339 format                                                      |
| reward-allocations... | whitespace-separated list of the reward allocations                                                     |

Example

```bash
# Create private farming plan
crescentd tx amm create-private-farming-plan "New farming plan" cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p 2023-01-01T00:00:00Z 2024-01-01T00:00:00Z 1:1000000uatom \
--chain-id localnet \
--from alice \

#
# Tips
#
# You can query farming plans by using the following command
crescentd q amm farming-plans -o json | jq
```

## TerminatePrivateFarmingPlan

Terminate a private farming plan.

The plan's termination address must be same with the message sender.

Usage

```bash
terminate-private-farming-plan [farming-plan-id]
```

| **Argument**      | **Description** |                                                                                                                                                                                                                                                                                                  |
|:------------------|:--------------- |
| farming-plan-id   | farming plan id |

Example

```bash
# Withdraw pool coin from the pool
crescentd tx amm terminate-private-farming-plan 1 \
--chain-id localnet \
--from alice \

#
# Tips
#
# You can query farming plans by using the following command
crescentd q amm farming-plans -o json | jq
```


# Query

## Params

Query the current amm parameters information

Usage

```bash
params
```

Example

```bash
crescentd q amm params -o json | jq
```

## AllPools

Query for all pools

Usage

```bash
pools
```

Example

```bash
crescentd q amm pools -o json | jq
````

## Pool

Query details for the particular pool

Usage

```bash
pool [pool-id]
```

Example

```bash
crescentd q amm pool 1 -o json | jq
```

## AllPositions

Query for all positions

Usage

```bash
positions
```

Example

```bash
# Query all positions
crescentd q amm positions -o json | jq

# Query all positions that has the pool id
crescentd q amm positions --pool-id 1 -o json | jq

# Query all positions of particular address
crescentd q amm position --owner cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## Position

Query details for the particular position

Usage

```bash
position [position-id]
```

Example

```bash
crescentd q amm position 1 -o json | jq
```

## SimulateAddLiquidity

Query expected result for add-liquidity

Usage

```bash
simulate-add-liquidity [pool-id] [lower-price] [upper-price] [desired-amount]
```

Example

```bash
crescentd q amm simulate-add-liquidity 1 9 11 10000000uatom,10000000uusd -o json | jq
```

## SimulateRemoveLiquidity

Query expected result for remove-liquidity

Usage

```bash
simulate-remove-liquidity [position-id] [liquidity]
```

Example

```bash
crescentd q amm simulate-remove-liquidity 1 10000 -o json | jq
```

## CollectibleCoins

Query collectible coins(fees, rewards) in the position.

Usage

```bash
collectible-coins
```

Example

```bash
# Query collectible coins with address
crescentd q amm collectible-coins --owner cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq

# Query collectible coins with position id
crescentd q amm collectible-coins --position-id 1 -o json | jq
```

## AllTickInfos

Query for information of all ticks in the particular pool

Usage

```bash
tick-infos [pool-id]
```

Example

```bash
# Query all ticks in pool
crescentd q amm tick-infos 1 -o json | jq

# Query all ticks above designated lower tick
crescentd q amm tick-infos 1 --lower-tick 10000 -o json | jq

# Query all ticks below designated upper tick
crescentd q amm tick-infos 1 --upper-tick 10000 -o json | jq
```

## TickInfo

Query details for the particular tick in the pool

Usage

```bash
tick-info [pool-id] [tick]
```

Example

```bash
crescentd q amm tick-info 1 10000 -o json | jq
```


## AllFarmingPlans

Query for all farming plans

Usage

```bash
farming-plans
```

Example

```bash
# Query all farming plans
crescentd q amm farming-plans -o json | jq

# Query farming plans based on private status
crescentd q amm farming-plans --is-private true -o json | jq

# Query farming plans based on private status
crescentd q amm farming-plans --is-terminated true -o json | jq
```

## FarmingPlan

Query details for the particular farming plan

Usage

```bash
farming-plan [plan-id]
```

Example

```bash
crescentd q amm farming-plan 1 -o json | jq
```