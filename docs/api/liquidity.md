---
Title: Liquidity
Description: A high-level overview of what gRPC-gateway REST routes are supported in the liquidity module.
---

# Liquidity Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the liquidity module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/cosmosquad-labs/squad/blob/main/proto/squad/liquidity/v1beta1/query.proto 

- [Params](#Params)
- [Pairs](#Pairs)
- [Pair](#Pair)
- [Pools](#Pools)
- [Pool](#Pool)
- [PoolByReserveAddress](#PoolByReserveAddress)
- [PoolByPoolCoinDenom](#PoolByPoolCoinDenom)
- [DepositRequests](#DepositRequests)
- [DepositRequest](#DepositRequest)
- [WithdrawRequests](#WithdrawRequests)
- [WithdrawRequest](#WithdrawRequest)
- [Orders](#Orrers)
- [Order](#Order)
- [OrdersByOrderer](#OrdersByOrderer)

## Params

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/params
```

Example Response

```json
{
  "params": {
    "batch_size": 1,
    "tick_precision": 3,
    "fee_collector_address": "cosmos1zdew6yxyw92z373yqp756e0x4rvd2het37j0a2wjp7fj48eevxvqau9aj0",
    "dust_collector_address": "cosmos1suads2mkd027cmfphmk9fpuwcct4d8ys02frk8e64hluswfwfj0se4s8xs",
    "initial_pool_coin_supply": "1000000000000",
    "pair_creation_fee": [
      {
        "denom": "stake",
        "amount": "1"
      }
    ],
    "pool_creation_fee": [
      {
        "denom": "stake",
        "amount": "1"
      }
    ],
    "min_initial_deposit_amount": "1000000",
    "max_price_limit_ratio": "0.100000000000000000",
    "max_order_lifespan": "86400s",
    "swap_fee_rate": "0.000000000000000000",
    "withdraw_fee_rate": "0.000000000000000000"
  }
}
```

## Pairs

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pairs
http://localhost:1317/squad/liquidity/v1beta1/pairs?denoms=uatom
http://localhost:1317/squad/liquidity/v1beta1/pairs?denoms=uatom&denoms=uusd
```

Example Response

```json
{
  "pairs": [
    {
      "id": "1",
      "base_coin_denom": "uatom",
      "quote_coin_denom": "uusd",
      "escrow_address": "cosmos17u9nx0h9cmhypp6cg9lf4q8ku9l3k8mz232su7m28m39lkz25dgqzkypxs",
      "last_order_id": "4",
      "last_price": "0.310500000000000000",
      "current_batch_id": "5"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Pair

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pairs/1
```

Example Response

```json
{
  "pair": {
    "id": "1",
    "base_coin_denom": "uatom",
    "quote_coin_denom": "uusd",
    "escrow_address": "cosmos17u9nx0h9cmhypp6cg9lf4q8ku9l3k8mz232su7m28m39lkz25dgqzkypxs",
    "last_order_id": "4",
    "last_price": "0.310500000000000000",
    "current_batch_id": "5"
  }
}
```

## Pools

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools
http://localhost:1317/squad/liquidity/v1beta1/pools?pair_id=1
http://localhost:1317/squad/liquidity/v1beta1/pools?disabled=false
```

Example Response

```json
{
  "pools": [
    {
      "id": "1",
      "pair_id": "1",
      "reserve_address": "cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy",
      "pool_coin_denom": "pool1",
      "balances": [
        {
          "denom": "uatom",
          "amount": "1636000001"
        },
        {
          "denom": "uusd",
          "amount": "476957901"
        }
      ],
      "last_deposit_request_id": "1",
      "last_withdraw_request_id": "1"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Pool

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/1
```

Example Response

```json
{
  "pool": {
    "id": "1",
    "pair_id": "1",
    "reserve_address": "cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy",
    "pool_coin_denom": "pool1",
    "balances": [
      {
        "denom": "uatom",
        "amount": "1636000001"
      },
      {
        "denom": "uusd",
        "amount": "476957901"
      }
    ],
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1"
  }
}
```

## PoolByReserveAddress

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/reserve_address/cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy
```

Example Response

```json
{
  "pool": {
    "id": "1",
    "pair_id": "1",
    "reserve_address": "cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy",
    "pool_coin_denom": "pool1",
    "balances": [
      {
        "denom": "uatom",
        "amount": "1636000001"
      },
      {
        "denom": "uusd",
        "amount": "476957901"
      }
    ],
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1"
  }
}
```

## PoolByPoolCoinDenom

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/pool_coin_denom/pool1
```

Example Response

```json
{
  "pool": {
    "id": "1",
    "pair_id": "1",
    "reserve_address": "cosmos1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfns20gpwy",
    "pool_coin_denom": "pool1",
    "balances": [
      {
        "denom": "uatom",
        "amount": "1636000001"
      },
      {
        "denom": "uusd",
        "amount": "476957901"
      }
    ],
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1"
  }
}
```

## DepositRequests

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/1/deposit_requests
```

Example Response

```json
{
  "deposit_requests": [
    {
      "id": "2",
      "pool_id": "1",
      "msg_height": "1849",
      "depositor": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "deposit_coins": [
        {
          "denom": "uatom",
          "amount": "30000000"
        },
        {
          "denom": "uusd",
          "amount": "10000000"
        }
      ],
      "accepted_coins": [
        {
          "denom": "uatom",
          "amount": "30000000"
        },
        {
          "denom": "uusd",
          "amount": "8746172"
        }
      ],
      "minted_pool_coin": {
        "denom": "pool1",
        "amount": "9352078233"
      },
      "status": "REQUEST_STATUS_SUCCEEDED"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## DepositRequest

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/1/deposit_requests/1
```

Example Response

```json
{
  "deposit_request": {
    "id": "5",
    "pool_id": "1",
    "msg_height": "1929",
    "depositor": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
    "deposit_coins": [
      {
        "denom": "uatom",
        "amount": "30000000"
      },
      {
        "denom": "uusd",
        "amount": "10000000"
      }
    ],
    "accepted_coins": [
      {
        "denom": "uatom",
        "amount": "30000000"
      },
      {
        "denom": "uusd",
        "amount": "8746172"
      }
    ],
    "minted_pool_coin": {
      "denom": "pool1",
      "amount": "9352078233"
    },
    "status": "REQUEST_STATUS_SUCCEEDED"
  }
}
```

## WithdraRequests

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/1/withdraw_requests
```

Example Response

```json
{
  "withdraw_requests": [
    {
      "id": "2",
      "pool_id": "1",
      "msg_height": "1987",
      "withdrawer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "pool_coin": {
        "denom": "pool1",
        "amount": "10000000"
      },
      "withdrawn_coins": [
        {
          "denom": "uatom",
          "amount": "32078"
        },
        {
          "denom": "uusd",
          "amount": "9352"
        }
      ],
      "status": "REQUEST_STATUS_SUCCEEDED"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## WithdrawRequest

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pools/1/withdraw_requests/1
```

Example Response

```json
{
  "withdraw_request": {
    "id": "3",
    "pool_id": "1",
    "msg_height": "2016",
    "withdrawer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
    "pool_coin": {
      "denom": "pool1",
      "amount": "10000000"
    },
    "withdrawn_coins": [
      {
        "denom": "uatom",
        "amount": "32078"
      },
      {
        "denom": "uusd",
        "amount": "9352"
      }
    ],
    "status": "REQUEST_STATUS_SUCCEEDED"
  }
}
```

## Orders

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/pairs/1/orders
```

Example Response

```json
{
  "orders": [
    {
      "id": "5",
      "pair_id": "1",
      "msg_height": "2129",
      "orderer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "direction": "ORDER_DIRECTION_SELL",
      "offer_coin": {
        "denom": "uatom",
        "amount": "1000000"
      },
      "remaining_offer_coin": {
        "denom": "uatom",
        "amount": "0"
      },
      "received_coin": {
        "denom": "uusd",
        "amount": "291300"
      },
      "price": "0.279500000000000000",
      "amount": "1000000",
      "open_amount": "0",
      "batch_id": "5",
      "expire_at": "2022-03-15T11:30:20.044978Z",
      "status": "ORDER_STATUS_COMPLETED"
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
http://localhost:1317/squad/liquidity/v1beta1/pairs/1/orders/1
```

Example Response

```json
{
  "order": {
    "id": "8",
    "pair_id": "1",
    "msg_height": "2280",
    "orderer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
    "direction": "ORDER_DIRECTION_SELL",
    "offer_coin": {
      "denom": "uatom",
      "amount": "1000000"
    },
    "remaining_offer_coin": {
      "denom": "uatom",
      "amount": "0"
    },
    "received_coin": {
      "denom": "uusd",
      "amount": "290300"
    },
    "price": "0.261700000000000000",
    "amount": "1000000",
    "open_amount": "0",
    "batch_id": "8",
    "expire_at": "2022-03-15T11:33:08.772980Z",
    "status": "ORDER_STATUS_COMPLETED"
  }
}
```

## OrdersByOrderer

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/squad/liquidity/v1beta1/orders/cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
```

Example Response

```json
{
  "orders": [
    {
      "id": "7",
      "pair_id": "1",
      "msg_height": "2242",
      "orderer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "direction": "ORDER_DIRECTION_SELL",
      "offer_coin": {
        "denom": "uatom",
        "amount": "1000000"
      },
      "remaining_offer_coin": {
        "denom": "uatom",
        "amount": "0"
      },
      "received_coin": {
        "denom": "uusd",
        "amount": "290700"
      },
      "price": "0.261900000000000000",
      "amount": "1000000",
      "open_amount": "0",
      "batch_id": "7",
      "expire_at": "2022-03-15T11:32:26.216376Z",
      "status": "ORDER_STATUS_COMPLETED"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```
