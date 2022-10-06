---
Title: Farm
Description: A high-level overview of what gRPC-gateway REST routes are supported in the farm module.
---

# Farm Module

## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the `farm` module.

## gRPC-gateway REST Routes

<!-- markdown-link-check-disable -->

++https://github.com/crescent-network/crescent/blob/main/proto/crescent/farm/v1beta1/query.proto

- [Farm Module](#farm-module)
  - [Synopsis](#synopsis)
  - [gRPC-gateway REST Routes](#grpc-gateway-rest-routes)
  - [Params](#params)

## Params

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/params
```

Example Response

```json
{
  "params": {
    "private_plan_creation_fee": [
      {
        "denom": "stake",
        "amount": "100000000"
      }
    ],
    "fee_collector": "cre1jclh5ymhug04qr2julz25m2yqv4ughnuuy65exx36mwurtcwnrzqsg9tgn",
    "max_num_private_plans": 50,
    "max_block_duration": "10s"
  }
}
```

## Plans

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/plans
```

Example Response

```json
{
  "plans": [
    {
      "id": "1",
      "description": "New Farming Plan",
      "farming_pool_address": "cre1gkvhlzmpxarqwk4jh7k7yemf60r50y55n8ax9kxcx8t28hm0e7cqk52jh9",
      "termination_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "reward_allocations": [
        {
          "pair_id": "1",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "1000000"
            }
          ]
        },
        {
          "pair_id": "2",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "2000000"
            }
          ]
        }
      ],
      "start_time": "2022-01-01T00:00:00Z",
      "end_time": "2023-01-01T00:00:00Z",
      "is_private": true,
      "is_terminated": false
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Plan

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/plans/1
```

Example Response

```json
{
  "plan": {
    "id": "1",
    "description": "New Farming Plan",
    "farming_pool_address": "cre1gkvhlzmpxarqwk4jh7k7yemf60r50y55n8ax9kxcx8t28hm0e7cqk52jh9",
    "termination_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
    "reward_allocations": [
      {
        "pair_id": "1",
        "rewards_per_day": [
          {
            "denom": "stake",
            "amount": "1000000"
          }
        ]
      },
      {
        "pair_id": "2",
        "rewards_per_day": [
          {
            "denom": "stake",
            "amount": "2000000"
          }
        ]
      }
    ],
    "start_time": "2022-01-01T00:00:00Z",
    "end_time": "2023-01-01T00:00:00Z",
    "is_private": true,
    "is_terminated": false
  }
}
```

## Farm

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/farms/pool1
```

Example Response

```json
{
  "farm": {
    "total_farming_amount": "1000000",
    "current_rewards": [
      {
        "denom": "stake",
        "amount": "281.000000000000000000"
      }
    ],
    "outstanding_rewards": [
      {
        "denom": "stake",
        "amount": "281.000000000000000000"
      }
    ],
    "period": "2"
  }
}
```

## Positions

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/positions/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```

Example Response

```json
{
  "positions": [
    {
      "farmer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "denom": "pool1",
      "farming_amount": "1000000",
      "previous_period": "1",
      "starting_block_height": "382"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## Position

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/positions/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p/pool1
```

Example Response

```json
{
  "position": {
    "farmer": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
    "denom": "pool1",
    "farming_amount": "1000000",
    "previous_period": "1",
    "starting_block_height": "382"
  }
}
```

## HistoricalRewards

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/historical_rewards/pool1
```

Example Response

```json
{
  "historical_rewards": [
    {
      "period": "2",
      "cumulative_unit_rewards": [
        {
          "denom": "stake",
          "amount": "0.004716000000000000"
        }
      ],
      "reference_count": 2
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

## AllRewards

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/rewards/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```

Example Response

```json
{
  "rewards": [
    {
      "denom": "stake",
      "amount": "780.000000000000000000"
    }
  ]
}
```

## Rewards

Example Request:

<!-- markdown-link-check-disable -->

```bash
http://localhost:1317/crescent/farm/v1beta1/rewards/cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p/pool1
```

Example Response

```json
{
  "rewards": [
    {
      "denom": "stake",
      "amount": "780.000000000000000000"
    }
  ]
}
```
