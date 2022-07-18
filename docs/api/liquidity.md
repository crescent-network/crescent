---
Title: Liquidity
Description: A high-level overview of what gRPC-gateway REST routes are supported in the liquidity module.
---

# Liquidity Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `liquidity` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->
++https://github.com/crescent-network/crescent/blob/main/proto/crescent/liquidity/v1beta1/query.proto 

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
- [Orders](#Orders)
- [Order](#Order)
- [OrdersByOrderer](#OrdersByOrderer)
- [OrderBooks](#OrderBooks)

## Params

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/params
```

Example Response

```json
{
  "params": {
    "batch_size": 1,
    "tick_precision": 3,
    "fee_collector_address": "cre1zdew6yxyw92z373yqp756e0x4rvd2het37j0a2wjp7fj48eevxvq303p8d",
    "dust_collector_address": "cre1suads2mkd027cmfphmk9fpuwcct4d8ys02frk8e64hluswfwfj0s4xymnj",
    "initial_pool_coin_supply": "1000000000000",
    "pair_creation_fee": [
      {
        "denom": "stake",
        "amount": "1000000"
      }
    ],
    "pool_creation_fee": [
      {
        "denom": "stake",
        "amount": "1000000"
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
http://localhost:1317/crescent/liquidity/v1beta1/pairs
http://localhost:1317/crescent/liquidity/v1beta1/pairs?denoms=uatom
http://localhost:1317/crescent/liquidity/v1beta1/pairs?denoms=uatom&denoms=uusd
```

Example Response

```json
{
  "pairs": [
    {
      "id": "1",
      "base_coin_denom": "uatom",
      "quote_coin_denom": "uusd",
      "escrow_address": "cre17u9nx0h9cmhypp6cg9lf4q8ku9l3k8mz232su7m28m39lkz25dgqw9sanj",
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
http://localhost:1317/crescent/liquidity/v1beta1/pairs/1
```

Example Response

```json
{
  "pair": {
    "id": "1",
    "base_coin_denom": "uatom",
    "quote_coin_denom": "uusd",
    "escrow_address": "cre17u9nx0h9cmhypp6cg9lf4q8ku9l3k8mz232su7m28m39lkz25dgqw9sanj",
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
http://localhost:1317/crescent/liquidity/v1beta1/pools
http://localhost:1317/crescent/liquidity/v1beta1/pools?pair_id=1
http://localhost:1317/crescent/liquidity/v1beta1/pools?disabled=false
```

Example Response

```json
{
  "pools": [
    {
      "type": "POOL_TYPE_BASIC",
      "id": "1",
      "pair_id": "1",
      "creator": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "reserve_address": "cre1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfnsxuuamx",
      "pool_coin_denom": "pool1",
      "pool_coin_supply": "1000000000000",
      "min_price": null,
      "max_price": null,
      "price": "0.291539059112751186",
      "balances": {
        "base_coin": {
          "denom": "uatom",
          "amount": "1636000001"
        },
        "quote_coin": {
          "denom": "uusd",
          "amount": "476957901"
        }
      },
      "last_deposit_request_id": "1",
      "last_withdraw_request_id": "1",
      "disabled": false
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
http://localhost:1317/crescent/liquidity/v1beta1/pools/1
```

Example Response

```json
{
  "pool": {
    "type": "POOL_TYPE_BASIC",
    "id": "1",
    "pair_id": "1",
    "creator": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "reserve_address": "cre1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfnsxuuamx",
    "pool_coin_denom": "pool1",
    "pool_coin_supply": "1000000000000",
    "min_price": null,
    "max_price": null,
    "price": "0.291539059112751186",
    "balances": {
      "base_coin": {
        "denom": "uatom",
        "amount": "1636000001"
      },
      "quote_coin": {
        "denom": "uusd",
        "amount": "476957901"
      }
    },
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1",
    "disabled": false
  }
}
```

## PoolByReserveAddress

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/pools/reserve_address/cre1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfnsxuuamx
```

Example Response

```json
{
  "pool": {
    "type": "POOL_TYPE_BASIC",
    "id": "1",
    "pair_id": "1",
    "creator": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "reserve_address": "cre1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfnsxuuamx",
    "pool_coin_denom": "pool1",
    "pool_coin_supply": "1000000000000",
    "min_price": null,
    "max_price": null,
    "price": "0.291539059112751186",
    "balances": {
      "base_coin": {
        "denom": "uatom",
        "amount": "1636000001"
      },
      "quote_coin": {
        "denom": "uusd",
        "amount": "476957901"
      }
    },
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1",
    "disabled": false
  }
}
```

## PoolByPoolCoinDenom

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/pools/pool_coin_denom/pool1
```

Example Response

```json
{
  "pool": {
    "type": "POOL_TYPE_BASIC",
    "id": "1",
    "pair_id": "1",
    "creator": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "reserve_address": "cre1353ausz7n8arsyf6dp0mq7gvj4ry2c2ht284kzrrft2mx7rdvfnsxuuamx",
    "pool_coin_denom": "pool1",
    "pool_coin_supply": "1000000000000",
    "min_price": null,
    "max_price": null,
    "price": "0.291539059112751186",
    "balances": {
      "base_coin": {
        "denom": "uatom",
        "amount": "1636000001"
      },
      "quote_coin": {
        "denom": "uusd",
        "amount": "476957901"
      }
    },
    "last_deposit_request_id": "1",
    "last_withdraw_request_id": "1",
    "disabled": false
  }
}
```

## DepositRequests

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/pools/1/deposit_requests
```

Example Response

```json
{
  "deposit_requests": [
    {
      "id": "2",
      "pool_id": "1",
      "msg_height": "1849",
      "depositor": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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
http://localhost:1317/crescent/liquidity/v1beta1/pools/1/deposit_requests/1
```

Example Response

```json
{
  "deposit_request": {
    "id": "5",
    "pool_id": "1",
    "msg_height": "1929",
    "depositor": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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

## WithdrawRequests

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/pools/1/withdraw_requests
```

Example Response

```json
{
  "withdraw_requests": [
    {
      "id": "2",
      "pool_id": "1",
      "msg_height": "1987",
      "withdrawer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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
http://localhost:1317/crescent/liquidity/v1beta1/pools/1/withdraw_requests/1
```

Example Response

```json
{
  "withdraw_request": {
    "id": "3",
    "pool_id": "1",
    "msg_height": "2016",
    "withdrawer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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
http://localhost:1317/crescent/liquidity/v1beta1/pairs/1/orders
```

Example Response

```json
{
  "orders": [
    {
      "id": "5",
      "pair_id": "1",
      "msg_height": "2129",
      "orderer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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
http://localhost:1317/crescent/liquidity/v1beta1/pairs/1/orders/1
```

Example Response

```json
{
  "order": {
    "id": "8",
    "pair_id": "1",
    "msg_height": "2280",
    "orderer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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
http://localhost:1317/crescent/liquidity/v1beta1/orders/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```

Example Response

```json
{
  "orders": [
    {
      "id": "7",
      "pair_id": "1",
      "msg_height": "2242",
      "orderer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
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

## OrderBooks

Example Request

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/crescent/liquidity/v1beta1/order_books?pair_ids=1&num_ticks=5
```

Example Response

```json
{
  "pairs": [
    {
      "pair_id": "1",
      "base_price": "1.180000000000000000",
      "order_books": [
        {
          "price_unit": "0.000100000000000000",
          "sells": [
            {
              "price": "1.180500000000000000",
              "user_order_amount": "64712",
              "pool_order_amount": "0"
            },
            {
              "price": "1.180400000000000000",
              "user_order_amount": "19985",
              "pool_order_amount": "0"
            },
            {
              "price": "1.180300000000000000",
              "user_order_amount": "64725",
              "pool_order_amount": "0"
            },
            {
              "price": "1.180200000000000000",
              "user_order_amount": "19993",
              "pool_order_amount": "0"
            },
            {
              "price": "1.180100000000000000",
              "user_order_amount": "64738",
              "pool_order_amount": "0"
            }
          ],
          "buys": [
            {
              "price": "1.180000000000000000",
              "user_order_amount": "20000",
              "pool_order_amount": "0"
            },
            {
              "price": "1.179900000000000000",
              "user_order_amount": "64752",
              "pool_order_amount": "0"
            },
            {
              "price": "1.179800000000000000",
              "user_order_amount": "20009",
              "pool_order_amount": "0"
            },
            {
              "price": "1.179700000000000000",
              "user_order_amount": "64765",
              "pool_order_amount": "0"
            },
            {
              "price": "1.179600000000000000",
              "user_order_amount": "20017",
              "pool_order_amount": "0"
            }
          ]
        },
        {
          "price_unit": "0.001000000000000000",
          "sells": [
            {
              "price": "1.185000000000000000",
              "user_order_amount": "421298",
              "pool_order_amount": "0"
            },
            {
              "price": "1.184000000000000000",
              "user_order_amount": "421830",
              "pool_order_amount": "0"
            },
            {
              "price": "1.183000000000000000",
              "user_order_amount": "422365",
              "pool_order_amount": "0"
            },
            {
              "price": "1.182000000000000000",
              "user_order_amount": "422901",
              "pool_order_amount": "0"
            },
            {
              "price": "1.181000000000000000",
              "user_order_amount": "423440",
              "pool_order_amount": "0"
            }
          ],
          "buys": [
            {
              "price": "1.180000000000000000",
              "user_order_amount": "20000",
              "pool_order_amount": "0"
            },
            {
              "price": "1.179000000000000000",
              "user_order_amount": "424020",
              "pool_order_amount": "0"
            },
            {
              "price": "1.178000000000000000",
              "user_order_amount": "424558",
              "pool_order_amount": "0"
            },
            {
              "price": "1.177000000000000000",
              "user_order_amount": "425099",
              "pool_order_amount": "0"
            },
            {
              "price": "1.176000000000000000",
              "user_order_amount": "425640",
              "pool_order_amount": "0"
            }
          ]
        },
        {
          "price_unit": "0.010000000000000000",
          "sells": [
            {
              "price": "1.230000000000000000",
              "user_order_amount": "3763074",
              "pool_order_amount": "0"
            },
            {
              "price": "1.220000000000000000",
              "user_order_amount": "4377452",
              "pool_order_amount": "0"
            },
            {
              "price": "1.210000000000000000",
              "user_order_amount": "4106019",
              "pool_order_amount": "0"
            },
            {
              "price": "1.200000000000000000",
              "user_order_amount": "4157434",
              "pool_order_amount": "0"
            },
            {
              "price": "1.190000000000000000",
              "user_order_amount": "4210331",
              "pool_order_amount": "0"
            }
          ],
          "buys": [
            {
              "price": "1.180000000000000000",
              "user_order_amount": "20000",
              "pool_order_amount": "0"
            },
            {
              "price": "1.170000000000000000",
              "user_order_amount": "4329993",
              "pool_order_amount": "0"
            },
            {
              "price": "1.160000000000000000",
              "user_order_amount": "4320112",
              "pool_order_amount": "0"
            },
            {
              "price": "1.150000000000000000",
              "user_order_amount": "4111940",
              "pool_order_amount": "0"
            },
            {
              "price": "1.140000000000000000",
              "user_order_amount": "4430076",
              "pool_order_amount": "0"
            }
          ]
        }
      ]
    }
  ]
}
```
