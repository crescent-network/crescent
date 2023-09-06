---
Title: Exchange
Description: A high-level overview of how the command-line interfaces (CLI) works for the exchange module.
---

# Exchange Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `exchange` module. 

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [CreateMarket](#CreateMarket)
  - [PlaceLimitOrder](#PlaceLimitOrder)
  - [PlaceBatchLimitOrder](#PlaceBatchLimitOrder)
  - [PlaceMMLimitOrder](#PlaceMMLimitOrder)
  - [PlaceMMBatchLimitOrder](#PlaceMMBatchLimitOrder)
  - [PlaceMarketOrder](#PlaceMarketOrder)
  - [CancelOrder](#CancelOrder)
  - [CancelAllOrders](#CancelAllOrders)
  - [SwapExactAmountIn](#SwapExactAmountIn)
- [Query](#Query)
  - [Params](#Params)
  - [AllMarkets](#AllMarkets)
  - [Market](#Market)
  - [AllOrders](#AllOrders)
  - [Order](#Order)
  - [BestSwapExactAmountInRoutes](#BestSwapExactAmountInRoutes)
  - [OrderBook](#OrderBook)

# Transaction

## CreateMarket

Create a market for trading assets.

Usage

```bash
create-market [base-denom] [quote-denom]
```

| **Argument**  | **Description**                        |
| :------------ | :------------------------------------- |
| base-denom    | denom of the base coin for the market  |
| quote-denom   | denom of the quote coin for the market |

Example

```bash
# Create a pool
crescentd tx exchange create-market uatom uusd \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query markets using the following command
crescentd q exchange markets -o json | jq
```

## PlaceLimitOrder

Place a limit order to markets. This order will be placed in sequential matching stage.

For buy orders, it allows orders up to 10% above the current price, and for sell orders, it allows orders up to 10% below the current price. This is to prevent users from incurring large financial losses due to simple mistakes.

Usage

```bash
place-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]
```

| **Argument**   | **Description**                                                                          |
| :------------  | :--------------------------------------------------------------------------------------- |
| market-id      | market id                                                                                |
| is-buy         | if this is true, the order is placed to buy base coin                                    | 
| price          | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| quantity       | amount of base coin that the orderer is willing to buy or sell                           |
| lifespan       | duration that the order lives until it is expired                                        |

Example

```bash
# Place a limit order
crescentd tx exchange place-limit-order 1 true 15 100000 1h \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query orders by using the following command
# Query all orders
crescentd q exchange orders -o json | jq

# Query all orders in particular market
crescentd q exchange orders --makret-id 1 -o json | jq

# Query all orders of particular orderer
crescentd q exchange orders --orderer cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceBatchLimitOrder

Place a batch limit order. Batch orders are matched prior to normal orders in a batch matching stage.

For buy orders, it allows orders up to 10% above the current price, and for sell orders, it allows orders up to 10% below the current price. This is to prevent users from incurring large financial losses due to simple mistakes.

Usage

```bash
place-batch-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]
```

| **Argument**   | **Description**                                                                          |
| :------------  | :--------------------------------------------------------------------------------------- |
| market-id      | market id                                                                                |
| is-buy         | if this is true, the order is placed to buy base coin                                    | 
| price          | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| quantity       | amount of base coin that the orderer is willing to buy or sell                           |
| lifespan       | duration that the order lives until it is expired                                        |

Example

```bash
# Place a batch limit order
crescentd tx exchange place-batch-limit-order 1 true 15 100000 1h \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query orders by using the following command
# Query all orders
crescentd q exchange orders -o json | jq

# Query all orders in particular market
crescentd q exchange orders --makret-id 1 -o json | jq

# Query all orders of particular orderer
crescentd q exchange orders --orderer cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceMMLimitOrder

Place a market maker limit order.

For buy orders, it allows orders up to 10% above the current price, and for sell orders, it allows orders up to 10% below the current price. This is to prevent users from incurring large financial losses due to simple mistakes.

Usage

```bash
place-mm-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]
```

| **Argument**   | **Description**                                                                          |
| :------------  | :--------------------------------------------------------------------------------------- |
| market-id      | market id                                                                                |
| is-buy         | if this is true, the order is placed to buy base coin                                    | 
| price          | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| quantity       | amount of base coin that the orderer is willing to buy or sell                           |
| lifespan       | duration that the order lives until it is expired                                        |

Example

```bash
# Place a market maker limit order
crescentd tx exchange place-mm-limit-order 1 true 15 100000 1h \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query orders by using the following command
# Query all orders
crescentd q exchange orders -o json | jq

# Query all orders in particular market
crescentd q exchange orders --makret-id 1 -o json | jq

# Query all orders of particular orderer
crescentd q exchange orders --orderer cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceMMBatchLimitOrder

Place a market maker batch limit order. Batch orders are matched prior to normal orders in a batch matching stage.

For buy orders, it allows orders up to 10% above the current price, and for sell orders, it allows orders up to 10% below the current price. This is to prevent users from incurring large financial losses due to simple mistakes.

Usage
```bash
place-mm-batch-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]
```

| **Argument**   | **Description**                                                                          |
| :------------  | :--------------------------------------------------------------------------------------- |
| market-id      | market id                                                                                |
| is-buy         | if this is true, the order is placed to buy base coin                                    | 
| price          | order price; the exchange ratio is the amount of quote coin over the amount of base coin |
| quantity       | amount of base coin that the orderer is willing to buy or sell                           |
| lifespan       | duration that the order lives until it is expired                                        |

Example

```bash
# Place a market maker batch limit order
crescentd tx exchange place-mm-batch-limit-order 1 true 15 100000 1h \
--chain-id localnet \
--from alice

#
# Tips
#
# You can query orders by using the following command
# Query all orders
crescentd q exchange orders -o json | jq

# Query all orders in particular market
crescentd q exchange orders --makret-id 1 -o json | jq

# Query all orders of particular orderer
crescentd q exchange orders --orderer cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceMarketOrder

Place a market order.

Usage

```bash
place-market-order [market-id] [is-buy] [quantity]
```

| **Argument**   | **Description**                                                                          |
| :------------  | :--------------------------------------------------------------------------------------- |
| market-id      | market id                                                                                |
| is-buy         | if this is true, the order is placed to buy base coin                                    |
| quantity       | amount of base coin that the orderer is willing to buy or sell                           |

Example

```bash
# Place a market order
crescentd tx exchange place-market-order 1 false 100000 \
--chain-id localnet \
--from alice
```

## CancelOrder

Cancel an existing order.

Usage

```bash
cancel-order [order-id]
```

| **Argument**   | **Description**     |
| :------------  | :------------------ |
| order-id       | order id            |

Example

```bash
# Place a market order
crescentd tx exchange cancel-order 1 \
--chain-id localnet \
--from alice
```

## CancelAllOrders

Cancel all orders in a market placed by the sender

Usage

```bash
cancel-all-orders [market-id]
```

| **Argument**   | **Description**     |
| :------------  | :------------------ |
| market-id      | market id           |

Example

```bash
# Place a market order
crescentd tx exchange cancel-all-orders 1 \
--chain-id localnet \
--from alice
```

## SwapExactAmountIn

Swap with exact input amount.

User need to specify swap routes from input to output. If the result of the swap falls short of user's desired output, the request will be reverted.

Usage

```bash
swap-exact-amount-in [routes] [input] [min-output]
```

| **Argument**   | **Description**                                                 |
| :------------  | :-------------------------------------------------------------- |
| routes         | sequential swap routes                                          |
| input          | input token denom and amount                                    |
| min-output     | The denom and minimum amount of the user's desired output token |

Example

```bash
# Place a market order
crescentd tx exchange swap-exact-amount-in 1,2,3 1000000uusd 98000uatom \
--chain-id localnet \
--from alice
```


# Query

## Params

Query the current exchange parameters information

Usage

```bash
params
```

Example

```bash
crescentd q exchange params -o json | jq
```

## AllMarkets

Query for all markets

Usage

```bash
markets
```

Example

```bash
crescentd q exchange markets -o json | jq
````

## Market

Query details for the particular market

Usage

```bash
market [market-id]
```

Example

```bash
crescentd q exchange market 1 -o json | jq
```

## AllOrders

Query for all orders

Usage

```bash
orders
```

Example

```bash
# Query all orders
crescentd q exchange orders -o json | jq

# Query all orders in particular market
crescentd q exchange orders --makret-id 1 -o json | jq

# Query all orders of particular orderer
crescentd q exchange orders --orderer cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## Order

Query details for the particular order

Usage

```bash
order [order-id]
```

Example

```bash
crescentd q exchange order 1 -o json | jq
```

## BestSwapExactAmountInRoutes

Query for the route that can be swapped at the best price given the input token denom and volume, and the denom of the output token.

Usage

```bash
best-swap-exact-amount-in-routes [input] [output-denom]
```

Example

```bash
crescentd q exchange best-swap-exact-amount-in-routes 1000000uusd uatom -o json | jq
```

## Orderbook

Query orderbook of particular market from exisiting orders

Usage

```bash
order-book [market-id]
```

Example

```bash
crescentd q exchange order-book 1 -o json | jq
```
