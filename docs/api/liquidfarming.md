---
Title: Liquidfarming
Description: A high-level overview of what gRPC-gateway REST routes are supported in the liquidfarming module.
---

# Liquidfarming Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `liquidfarming` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->

++https://github.com/crescent-network/crescent/blob/main/proto/crescent/liquidfarming/v1beta1/query.proto


- [Params](#params)
- [LiquidFarms](#liquidfarms)
- [LiquidFarm](#liquidfarm)
- [RewardsAuctions](#rewardsauctions)
- [RewardsAuction](#rewardsauction)
- [Bids](#bids)

## Params

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/params
```

Example Response

```json
{
  "params": {
    "liquid_farms": [
      {
        "pool_id": "1",
        "min_farm_amount": "1",
        "min_bid_amount": "1",
        "fee_rate": "0.000000000000000000"
      }
    ],
    "rewards_auction_duration": "120s"
  }
}
```

## LiquidFarms

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/liquidfarms
```

Example Response

```json
{
  "liquid_farms": [
    {
      "pool_id": "1",
      "liquid_farm_reserve_address": "cre1zyyf855slxure4c8dr06p00qjnkem95d2lgv8wgvry2rt437x6ts363hdt",
      "lf_coin_denom": "lf1",
      "min_farm_amount": "1",
      "min_bid_amount": "1",
      "total_farming_amount": "500000000000"
    }
  ]
}
```

## LiquidFarm

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/liquidfarms/1
```

Example Response

```json
{
  "liquid_farm": {
    "pool_id": "1",
    "liquid_farm_reserve_address": "cre1zyyf855slxure4c8dr06p00qjnkem95d2lgv8wgvry2rt437x6ts363hdt",
    "lf_coin_denom": "lf1",
    "minimum_farm_amount": "1",
    "minimum_bid_amount": "1",
    "total_farming_amount": "500000000000"
  }
}
```

## RewardsAuctions

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/pools/1/rewards_auctions
```

Example Response

```json
{
  "reward_auctions": [
    {
      "id": "1",
      "pool_id": "1",
      "bidding_coin_denom": "pool1",
      "paying_reserve_address": "cre1h72q3pkvsz537kj08hyv20tun3apampxhpgad97t3ls47nukgtxq4nw9fg",
      "start_time": "2022-09-27T06:06:52.627872Z",
      "end_time": "2022-09-27T06:08:52.627872Z",
      "status": "AUCTION_STATUS_STARTED",
      "winner": "",
      "rewards": []
    }
  ]
}
```

## RewardsAuction

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/pools/1/rewards_auctions/1
```

Example Response

```json
{
  "reward_auction": {
    "id": "1",
    "pool_id": "1",
    "bidding_coin_denom": "pool1",
    "paying_reserve_address": "cre1h72q3pkvsz537kj08hyv20tun3apampxhpgad97t3ls47nukgtxq4nw9fg",
    "start_time": "2022-08-05T08:56:22.237454Z",
    "end_time": "2022-08-06T08:56:22.237454Z",
    "status": "AUCTION_STATUS_FINISHED",
    "winner": "",
    "rewards": []
  }
}
```

## Bids

Example Request

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/liquidfarming/v1beta1/pools/1/bids
```

Example Response

```json
{
  "bids": [
    {
      "pool_id": "1",
      "bidder": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "amount": {
        "denom": "pool1",
        "amount": "1000000000"
      }
    }
  ]
}
```
