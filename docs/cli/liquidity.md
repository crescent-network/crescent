---
Title: Liquidity
Description: A high-level overview of how the command-line interfaces (CLI) works for the liquidity module.
---

# Liquidity Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidity` module. To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it. Run this command under the project root directory `$ ignite chain serve -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
    * [CreatePair](#CreatePair)
    * [CreatePool](#CreatePool)
    * [CreateRangedPool](#CreateRangedPool)
    * [Deposit](#Deposit)
    * [Withdraw](#Withdraw)
    * [LimitOrder](#LimitOrder)
    * [MarketOrder](#MarketOrder)
    * [CancelOrder](#CancelOrder)
    * [CancelAllOrders](#CancelAllOrders)
- [Query](#Query)
    * [Params](#Params)
    * [Pairs](#Pairs)
    * [Pair](#Pair)
    * [Pools](#Pools)
    * [Pool](#Pool)
    * [DepositRequests](#DepositRequests)
    * [DepositRequest](#DepositRequest)
    * [WithdrawRequests](#WithdrawRequests)
    * [WithdrawRequest](#WithdrawRequest)
    * [Orders](#Orders)
    * [Order](#Order)
    * [OrderBooks](#OrderBooks)

# Transaction

## CreatePair

Create a pair (market) for trading. 

A pair consists of a base coin and a quote coin and you can think of a pair in an order book. An orderer can request a limit or market order once a pair is created. Anyone can create a pair by paying a fee `PairCreationFee` (default is 1000000stake).

Usage

```bash
create-pair [base-coin-denom] [quote-coin-denom]
```

| **Argument**      |  **Description**                     |
| :---------------- | :----------------------------------- |
| base-coin-denom   | denom of the base coin for the pair  | 
| quote-coin-deom   | denom of the quote coin for the pair |

Example

```bash
# Create a pair ATOM/UST
crescentd tx liquidity create-pair uatom uusd \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query pairs using the following command
crescentd q liquidity pairs -o json | jq
```

## CreatePool

Create a liquidity pool in existing pair. 

Pool(s) belong to a pair. Therefore, a pair must exist in order to create a pool. Anyone can create a pool by paying a fee PoolCreationFee (default is 1000000stake).

Usage

```bash
create-pool [pair-id] [deposit-coins]
```

| **Argument**  |  **Description**                       |
| :------------ | :------------------------------------- |
| pair-id       | pair id                                | 
| deposit-coins | deposit amount of base and quote coins |

Example

```bash
# Create a pool 1000ATOM/3000UST
crescentd tx liquidity create-pool 1 1000000000uatom,3000000000uusd \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query pools using the following command
crescentd q liquidity pools -o json | jq
```

## CreateRangedPool

Create a ranged liquidity pool in existing pair.

Pool(s) belong to a pair. Therefore, a pair must exist in order to create a pool. Anyone can create a pool by paying a fee PoolCreationFee (default is 1000000stake).

Usage

```bash
create-ranged-pool [pair-id] [deposit-coins] [min-price] [max-price] [initial-price]
```

| **Argument**  | **Description**                        |
|:--------------|:---------------------------------------|
| pair-id       | pair id                                |
| deposit-coins | deposit amount of base and quote coins |
| min-price     | minimum price of the pool              |
| max-price     | maximum price of the pool              |
| initial-price | initial pool price                     |

Example

```bash
# Create a ranged pool with 1000ATOM/1000UST with price range of [2.5, 10.0],
# with initial price set to 3.0
crescentd tx liquidity create-ranged-pool 1 1000000000uatom,1000000000uusd 2.5 10.0 3.0 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query pools using the following command
crescentd q liquidity pools -o json | jq
```

## Deposit

Deposit coins to a liquidity pool.  

Deposit uses a batch execution methodology. Deposit requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch. A minimum deposit amount is 1000000 for each denomination.

Note that in an order book system, a pool is considered as an orderer. A liquidity in the pool places orders conservatively. What that means is that it places buy orders lower than the pool price and places sell orders higher than the pool price.

Usage

```bash
deposit [pool-id] [deposit-coins]
```

| **Argument**  |  **Description**                       |
| :------------ | :------------------------------------- |
| pool-id       | pool id                                |
| deposit-coins | deposit amount of base and quote coins |

Example

```bash
# Deposit 10ATOM/30UST to the pool
crescentd tx liquidity deposit 1 10000000uatom,30000000uusd \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query deposit requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
crescentd q liquidity deposit-requests 1 -o json | jq
```

## Withdraw

Withdraw coins from the liquidity pool.

Withdraw uses a batch execution methodology. Withdraw requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.

Usage

```bash
withdraw [pool-id] [pool-coin]
```

| **Argument** |  **Description**                |
| :----------- | :------------------------------ |
| pool-id      | pool id                         |
| pool-coin    | amount of pool coin to withdraw |

Example

```bash
# Withdraw pool coin from the pool
crescentd tx liquidity withdraw 1 500000000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query withdraw requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
crescentd q liquidity withdraw-requests 1 -o json | jq
```

## LimitOrder

Make a limit order.

Buy limit order will be matched at lower than or equal to the defined order price whereas sell limit order will be matched at higher than or equal to the defined order price.

Order uses a batch execution methodology. Order requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.

Usage

```bash
limit-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [price] [amount]
```

| **Argument**        |  **Description**                |
| :------------------ | :------------------------------ |
| pair-id             | pair id                         |
| direction           | swap direction; buy or sell |
| offer-coin          | amount of coin that the orderer offers to swap with; buy direction requires quote coin whereas sell direction requires base coin. For buy direction, quote coin amount must be greater than or equal to price * amount. For sell direction, base coin amount must be greater than or equal to the amount value.  |
| demand-coin-denom   | demand coin denom that the orderer is willing to swap for |
| price               | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| amount              | amount of base coin that the orderer is willing to buy or sell |

| **Optional Flag**  |  **Description**                                     |
| :------------------- | :--------------------------------------------------- |
| order-lifespan       | duration that the order lives until it is expired; an order will be executed for at least one batch, even if the lifespan is 0; valid time units are ns|us|ms|s|m|h|

Example

```bash
# Make a limit order to swap
crescentd tx liquidity limit-order 1 sell 50000000uatom uusd 3.3 50000000 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Make a limit order to swap with order-lifespan flag
crescentd tx liquidity limit-order 1 sell 50000000uatom uusd 3.3 50000000 \
--chain-id localnet \
--order-lifespan 30s \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query order requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
crescentd q liquidity orders cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## MarketOrder

Make a market order.

Unlike a limit order, there is no need to input order price. 

Buy market order uses `MaxPriceLimitRatio` of the last price, which is `LastPrice * (1+MaxPriceLimitRatio)`. 

Sell market order uses negative MaxPriceLimitRatio of the last price, which is `LastPrice * (1-MaxPriceLimitRatio)`.

Order uses a batch execution methodology. Order requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.

Usage

```bash
market-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [amount]
```

| **Argument**      |  **Description**                |
| :---------------- | :------------------------------ |
| pair-id           | pair id                         |
| direction         | swap direction; buy or sell |
| offer-coin        | amount of coin that the orderer offers to swap with; buy direction requires quote coin whereas sell direction requires base coin. For buy direction, quote coin amount must be greater than or equal to price * amount. For sell direction, base coin amount must be greater than or equal to the amount value.  |
| demand-coin-denom | demand coin denom that the orderer is willing to swap for |
| amount            | amount of base coin that the orderer is willing to buy or sell |

| **Optional Flag**  |  **Description**                                     |
| :------------------- | :--------------------------------------------------- |
| order-lifespan       | duration that the order lives until it is expired; an order will be executed for at least one batch, even if the lifespan is 0; valid time units are ns|us|ms|s|m|h|

Example

```bash
# Make a market order to swap 
crescentd tx liquidity market-order 1 sell 100000000uatom uusd 100000000 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Make a limit order to swap with order-lifespan flag
crescentd tx liquidity market-order 1 sell 100000000uatom uusd 100000000 \
--chain-id localnet \
--order-lifespan 30s \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query order requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
crescentd q liquidity orders cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## CancelOrder

Cancel an order.

Usage

```bash
cancel-order [pair-id] [order-id]
```

| **Argument** |  **Description** |
| :----------- | :----------------|
| pair-id      | pair id          |
| order-id     | order id         |

Example

```bash
crescentd tx liquidity cancel-order 1 1 \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq
```

## CancelAllOrders

Cancel all orders.

This command provides a convenient way to cancel all orders.

Usage

```bash
cancel-all-orders [pair-ids]
```

| **Argument** |  **Description**                |
| :----------- | :------------------------------ |
| pool-id      | pool id                         |
| pool-coin    | amount of pool coin to withdraw |

Example

```bash
crescentd tx liquidity cancel-all-orders 1,2,3 \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq
```

# Query

## Params

Query the current liquidity parameters information

Usage

```bash
params
```

Example

```bash
crescentd q liquidity params -o json | jq
```

## Pairs

Query for all pairs

Usage

```bash
pairs
```

Example

```bash
# Query all pairs
crescentd q liquidity pairs -o json | jq

# Query all pairs that has the defined denom
crescentd q liquidity pairs --denoms=uatom -o json | jq

# Query all pairs that has the defined denoms
crescentd q liquidity pairs --denoms=uatom,uusd -o json | jq
```

## Pair

Query details for the particular pair

Usage

```bash
pair [pair-id]
```

Example

```bash
crescentd q liquidity pair 1 -o json | jq
```

## Pools

Query for all pools

Usage

```bash
pools
```

Example

```bash
# Query all pools
crescentd q liquidity pools -o json | jq

# Query all pools that has the pair id
crescentd q liquidity pools -o json --pair-id=1 | jq

# Query all pools with disabled flag
crescentd q liquidity pools -o json --disabled=false | jq
```

## Pool

Query details for the particular pool

Usage

```bash
pool [pool-id]
```

Example

```bash
# Query the specific pool
crescentd q liquidity pool 1 -o json | jq

# Query the specific pool that has the defined pool coin denom
crescentd q liquidity pool --pool-coin-denom=pool1 -o json | jq
```

## DepositRequests

Query for all deposit requests in the pool

Usage

```bash
deposit-requests [pool-id]
```

Example

```bash
crescentd q liquidity deposit-requests 1 -o json | jq
```

## DepositRequest

Query details for the particular deposit request in the pool

Usage

```bash
deposit-request [pool-id] [id]
```

Example

```bash
crescentd q liquidity deposit-request 1 1 -o json | jq
```

## WithdrawRequests

Query for all withdraw requests in the pool

Usage

```bash
withdraw-requests [pool-id]
```

Example

```bash
crescentd q liquidity withdraw-requests 1 -o json | jq
```

## WithdrawRequest

Query details for the particular withdraw request in the pool

Usage

```bash
withdraw-request [pool-id] [id]
```

Example

```bash
crescentd q liquidity withdraw-request 1 1 -o json | jq
```

## Orders

Query for all orders made by an orderer or in the pair.

Usage

```bash
orders
```

Example

```bash
crescentd q liquidity orders cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p \
-o json | jq

crescentd q liquidity orders cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p \
--pair-id=1 \
-o json | jq

crescentd q liquidity orders \
--pair-id=1 \
-o json | jq
```

## Order

Query details for the particular order

Usage

```bash
order [pair-id] [id]
```

Example

```bash
crescentd q liquidity order 1 1
```

## OrderBooks

Query order books for given pairs and tick precisions.

Usage

```bash
order-books [pair-ids] [tick-precisions]
```

Example

```bash
crescentd order-books 1 --num-ticks=10

crescentd order-books 1,2,3
```


