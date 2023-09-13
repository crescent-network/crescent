---
Title: Exchange
Description: A high-level overview of what gRPC-gateway REST routes are supported in the exchange module.
---

# Exchange Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `exchange` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/crescent-network/crescent/blob/main/proto/crescent/exchange/v1beta1/query.proto (need check) 

- [Params](#Params)
- [AllMarkets](#AllMarkets)
- [Market](#Market)
- [AllOrders](#AllOrders)
- [Order](#Order)
- [BestSwapExactAmountInRoutes](#BestSwapExactAmountInRoutes)
- [OrderBook](#OrderBook)

## Params

Example Request

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/params
```

Example Response

```json
{
  "params": {
    "market_creation_fee": [
      {
        "denom": "stake",
        "amount": "1000000"
      }
    ],
    "fees": {
      "default_maker_fee_rate": "0.001500000000000000",
      "default_taker_fee_rate": "0.003000000000000000",
      "default_order_source_fee_ratio": "0.500000000000000000"
    },
    "max_order_lifespan": "86400s",
    "max_order_price_ratio": "0.100000000000000000",
    "max_swap_routes_len": 3,
    "max_num_mm_orders": 15
  }
}
```

## AllMarkets

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/markets
```

Example Response

```json
{
  "markets": [
    {
      "id": "1",
      "base_denom": "uatom",
      "quote_denom": "uusd",
      "escrow_address": "cre1whhqyaxuv9vrmr00slaqa6zg9cf30nk4k6ltpqvp6ecn6vks5mfsttzq30",
      "maker_fee_rate": "0.001500000000000000",
      "taker_fee_rate": "0.003000000000000000",
      "order_source_fee_ratio": "0.500000000000000000",
      "last_price": "15.000000000000000000",
      "last_matching_height": "6948"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Market

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/markets/1
```

Example Response

```json
{
  "market": {
    "id": "1",
    "base_denom": "uatom",
    "quote_denom": "uusd",
    "escrow_address": "cre1whhqyaxuv9vrmr00slaqa6zg9cf30nk4k6ltpqvp6ecn6vks5mfsttzq30",
    "maker_fee_rate": "0.001500000000000000",
    "taker_fee_rate": "0.003000000000000000",
    "order_source_fee_ratio": "0.500000000000000000",
    "last_price": "15.000000000000000000",
    "last_matching_height": "6948"
  }
}
```

## AllOrders

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/orders
```

Example Response

```json
{
  "orders": [
    {
      "id": "6",
      "type": "ORDER_TYPE_MM",
      "orderer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "market_id": "1",
      "is_buy": true,
      "price": "15.000000000000000000",
      "quantity": "100000.000000000000000000",
      "msg_height": "7213",
      "open_quantity": "100000.000000000000000000",
      "remaining_deposit": "1500000.000000000000000000",
      "deadline": "2023-08-30T05:30:50.090404Z"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Order

Example Request

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/orders/1
```

Example Response

```json
{
  "order": {
    "id": "1",
    "type": "ORDER_TYPE_LIMIT",
    "orderer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "market_id": "1",
    "is_buy": true,
    "price": "9.000000000000000000",
    "quantity": "100000",
    "msg_height": "6549",
    "open_quantity": "100000",
    "remaining_deposit": "900000",
    "deadline": "2023-06-29T07:28:37.840415Z"
  }
}
```

## BestSwapExactAmountInRoutes

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/best_swap_exact_amount_in_routes?input=100uatom&output_denom=uusd
```

Example Response

```json
{
  "routes": [
    "1"
  ],
  "output": {
    "denom": "uusd",
    "amount": "1016"
  },
  "results": [
    {
      "market_id": "1",
      "input": {
        "denom": "uatom",
        "amount": "100"
      },
      "output": {
        "denom": "uusd",
        "amount": "1016"
      },
      "fee": {
        "denom": "uusd",
        "amount": "4"
      }
    }
  ]
}
```


## Orderbook

Example Request

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/exchange/v1beta1/markets/1/order_book
```

Example Response

```json
{
  "order_books": [
    {
      "price_interval": "0.001000000000000000",
      "sells": [
      ],
      "buys": [
        {
          "p": "15.000000000000000000",
          "q": "100000.000000000000000000"
        }
      ]
    },
    {
      "price_interval": "0.010000000000000000",
      "sells": [
      ],
      "buys": [
        {
          "p": "15.000000000000000000",
          "q": "100000.000000000000000000"
        }
      ]
    },
    {
      "price_interval": "0.100000000000000000",
      "sells": [
      ],
      "buys": [
        {
          "p": "15.000000000000000000",
          "q": "100000.000000000000000000"
        }
      ]
    }
  ]
}
```

