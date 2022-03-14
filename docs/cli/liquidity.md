---
Title: Liquidity
Description: A high-level overview of how the command-line interfaces (CLI) work for the liquidity module.
---

# Liquidity Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidity` module. To set up a local testing environment, it requires the latest [Starport](https://starport.com/). If you don't have Starport set up in your local machine, see [this Starport guide](https://docs.starport.network/) to install it. Run this command under the project root directory `$ starport chain serve`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
    * [CreatePair](#CreatePair)
    * [CreatePool](#CreatePool)
    * [Deposit](#Deposit)
    * [Withdraw](#Withdrawool)
    * [LimitOrder](#LimitOrderol)
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
    * [Orders](#Orrers)
    * [Order](#Order)

## Transaction

### CreatePair

Create a pair (market) for trading. 

A pair consists of a base coin and a quote coin and you can think of a pair in an order book. An orderer can request a limit or market order once a pair is created. Anyone can create a pair by paying a fee `PairCreationFee` (default is 1000000stake).

```bash
create-pair [base-coin-denom] [quote-coin-denom]
```

| **Argument**      |  **Description**                     |
| :---------------- | :----------------------------------- |
| base-coin-denom   | denom of the base coin for the pair  | 
| quote-coin-deom   | denom of the quote coin for the pair |
| | | 

```bash
# Create a pair ATOM/UST
squad tx liquidity create-pair uatom uusd \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query pairs using the following command
squad q liquidity pairs -o json | jq
```

### CreatePool

Create a liquidity pool in existing pair. 

Pool(s) belong to a pair. Therefore, a pair must exist in order to create a pool. Anyone can create a pool by paying a fee PoolCreationFee (default is 1000000stake).

```bash
create-pool [pair-id] [deposit-coins]
```

| **Argument**  |  **Description**                       |
| :------------ | :------------------------------------- |
| pair-id       | pair id                                | 
| deposit-coins | deposit amount of base and quote coins |
| | | 


```bash
# Create a pool 3000ATOM/1000UST
squad tx liquidity create-pool 1 3000000000uatom,1000000000uusd \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query pairs using the following command
squad q liquidity pools -o json | jq
```

### Deposit

Deposit coins to a liquidity pool.  

Deposit uses a batch execution methodology. Deposit requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch. A minimum deposit amount is 1000000stake.

Note that in an order book system, a pool is considered as an orderer. A liquidity in the pool places orders conservatively. What that means it that it places buy orders lower than the pool price and places sell order higher than the pool price.

```bash
deposit [pool-id] [deposit-coins]
```

| **Argument**  |  **Description**                       |
| :------------ | :------------------------------------- |
| pool-id       | pool id                                |
| deposit-coins | deposit amount of base and quote coins |
| | | 

```bash
# Deposit 30ATOM/10UST to the pool
squad tx liquidity deposit 1 30000000uatom,10000000uusd \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query deposit requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
squad q liquidity deposit-requests 1 -o json | jq
```

### Withdraw

Withdraw coins from the liquidity pool.

Withdraw uses a batch execution methodology. Withdraw requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.

```bash
withdraw [pool-id] [pool-coin]
```

| **Argument** |  **Description**                |
| :----------- | :------------------------------ |
| pool-id      | pool id                         |
| pool-coin    | amount of pool coin to withdraw |
| | | 


```bash
# Withdraw pool coin from the pool
squad tx liquidity withdraw 1 500000000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query withdraw requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
squad q liquidity withdraw-requests 1 -o json | jq
```

### LimitOrder

Make a limit order.

Buy limit order will be matched at lower than or equal to the defined order price whereas sell limit order will be matched at higher than or equal to the defined order price.

Order uses a batch execution methodology. Order requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.

```bash
limit-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [price] [amount] [order-lifespan]
```

| **Argument**      |  **Description**                |
| :---------------- | :------------------------------ |
| pair-id           | pair id                         |
| direction         | swap direction; buy or sell |
| offer-coin        | amount of coin that the orderer offers to swap with; buy direction requires quote coin whereas sell direction requires base coin. For buy direction, quote coin amount must be greater than or equal to price * amount. For sell direction, base coin amount must be greater than or equal to the amount value.  |
| demand-coin-denom | demand coin denom that the orderer is willing to swap for |
| price             | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| amount            | amount of base coin that the orderer is willing to buy or sell |
| order-lifespan    | order lifespan; how long it resides within the chain. Maximum duration is a day. |
| | | 


```bash
# Make a limit order to swap 
squad tx liquidity limit-order 1 sell 100000uatom uusd 0.99 100000 10s \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query order requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
squad q liquidity orders 1 -o json | jq
```

### MarketOrder

Make a market order.

Unlike a limit order, there is no need to input order price. Buy market order uses `MaxPriceLimitRatio` of the last price, which is `LastPrice * (1+MaxPriceLimitRatio)`. Sell market order uses negative MaxPriceLimitRatio of the last price, which is `LastPrice * (1-MaxPriceLimitRatio)`.

Order uses a batch execution methodology. Order requests are accumulated in a batch for a pre-defined period (default is 1 block) and they are executed at the end of the batch.


```bash
market-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [amount] [order-lifespan]
```


| **Argument**      |  **Description**                |
| :---------------- | :------------------------------ |
| pair-id           | pair id                         |
| direction         | swap direction; buy or sell |
| offer-coin        | amount of coin that the orderer offers to swap with; buy direction requires quote coin whereas sell direction requires base coin. For buy direction, quote coin amount must be greater than or equal to price * amount. For sell direction, base coin amount must be greater than or equal to the amount value.  |
| demand-coin-denom | demand coin denom that the orderer is willing to swap for |
| amount            | amount of base coin that the orderer is willing to buy or sell |
| order-lifespan    | order lifespan; how long it resides within the chain. Maximum duration is a day. |
| | | 

```bash
# Make a market order to swap 
squad tx liquidity market-order 1 buy 100000usquad uatom 10000 10s \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query order requests by using the following command
# You must query this right away to get the result
# Otherwise, it is removed as it is executed.
squad q liquidity orders 1 -o json | jq
```

### CancelOrder


```bash

```

| **Argument** |  **Description**                |
| :----------- | :------------------------------ |
| pool-id      | pool id                         |
| pool-coin    | amount of pool coin to withdraw |
| | | 


```bash

```

### CancelAllOrders


```bash

```

| **Argument** |  **Description**                |
| :----------- | :------------------------------ |
| pool-id      | pool id                         |
| pool-coin    | amount of pool coin to withdraw |
| | | 


```bash

```

## Queries

### Params
### Pairs
### Pair
### Pools
### Pool
### DepositRequests
### DepositRequest
### WithdrawRequests
### WithdrawRequest
### Orders
### Order

