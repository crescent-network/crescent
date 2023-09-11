---
Title: Liquidamm
Description: A high-level overview of what gRPC-gateway REST routes are supported in the liquidamm module.
---

# liquidamm Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `liquidamm` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->

++https://github.com/crescent-network/crescent/blob/main/proto/crescent/liquidamm/v1beta1/query.proto

- [Params](#Params)
- [PublicPositions](#PublicPositions)
- [PublicPosition](#PublicPosition)
- [RewardsAuctions](#Rewardsauctions)
- [RewardsAuction](#Rewardsauction)
- [Bids](#Bids)
- [Rewards](#Rewards)

## Params

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/params
```

Example Response

```json
{
  "params": {
    "rewards_auction_duration": "3600s",
    "max_num_recent_rewards_auctions": 10
  }
}
```

## PublicPositions

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions
```

Example Response

```json
{
  "public_positions": [
    {
      "id": "1",
      "pool_id": "1",
      "lower_tick": 35000,
      "upper_tick": 45000,
      "bid_reserve_address": "cre1rkln8d74uhfyc9qp3645xnwks0pd8rsterguf6uugd2g60m37dmqwcapvh",
      "min_bid_amount": "100000",
      "fee_rate": "0.003000000000000000",
      "last_rewards_auction_id": "0",
      "liquidity": "0",
      "position_id": "0",
      "total_share": {
        "denom": "sb1",
        "amount": "0"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## PublicPosition

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1
```

Example Response

```json
{
  "public_position": {
    "id": "1",
    "pool_id": "1",
    "lower_tick": 35000,
    "upper_tick": 45000,
    "bid_reserve_address": "cre1rkln8d74uhfyc9qp3645xnwks0pd8rsterguf6uugd2g60m37dmqwcapvh",
    "min_bid_amount": "100000",
    "fee_rate": "0.003000000000000000",
    "last_rewards_auction_id": "0",
    "liquidity": "0",
    "position_id": "0",
    "total_share": {
      "denom": "sb1",
      "amount": "0"
    }
  }
}
```

## RewardsAuctions

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions?status=AUCTION_STATUS_STARTED
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions?status=AUCTION_STATUS_FINISHED
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions?status=AUCTION_STATUS_SKIPPED
```

Example Response

```json
{
  "rewards_auctions": [
    {
      "public_position_id": "1",
      "id": "1",
      "start_time": "2023-07-05T01:59:28.180826Z",
      "end_time": "2023-07-06T00:00:00Z",
      "status": "AUCTION_STATUS_STARTED",
      "winning_bid": null,
      "rewards": [
      ],
      "fees": [
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## RewardsAuction

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions/1
```

Example Response

```json
{
  "rewards_auction": {
    "public_position_id": "1",
    "id": "1",
    "start_time": "2023-07-05T01:59:28.180826Z",
    "end_time": "2023-07-06T00:00:00Z",
    "status": "AUCTION_STATUS_STARTED",
    "winning_bid": null,
    "rewards": [
    ],
    "fees": [
    ]
  }
}
```

## Bids

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards_auctions/1/bids
```

Example Response

```json
{
  "bids": [
    {
      "public_position_id": "1",
      "rewards_auction_id": "1",
      "bidder": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "share": {
        "denom": "sb1",
        "amount": "1000000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Rewards

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidamm/v1beta1/public_positions/1/rewards
```

Example Response

```json
{
  "rewards": [
    {
      "denom": "uatom",
      "amount": "50456"
    },
    {
      "denom": "uusd",
      "amount": "749866"
    }
  ]
}
```
