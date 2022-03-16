---
Title: REST APIs
Description: A high-level overview of gRPC-gateway REST routes in farming module.
---

# Farming Module
 
## Synopsis

This document provides a high-level overview of what gRPC-gateway REST routes are supported in the farming module.


## Swagger Documentation

- Swagger Cosmos SDK Farming Module [REST and gRPC Gateway docs](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/1.0.0)

## gRPC-gateway REST Routes

- [Params](#Params)
- [Plans](#Plans)
- [Plan](#Plan)
- [Stakings](#Stakings)
- [TotalStakings](#TotalStakings)
- [Rewards](#Rewards)
- [CurrentEpochDays](#CurrentEpochDays)

### Params

Query the values set as farming parameters:

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/params
```


```json
{
  "params": {
    "private_plan_creation_fee": [
      {
        "denom": "stake",
        "amount": "100000000"
      }
    ],
    "next_epoch_days": 1,
    "farming_fee_collector": "cre1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mq4p6cjy"
  }
}
```

### Plans

Query all the farming plans exist in the network:


Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/plans
```

```json
{
  "plans": [
    {
      "@type": "/crescent.farming.v1beta1.MsgCreateRatioPlan",
      "base_plan": {
        "id": "1",
        "name": "Second Public Ratio Plan",
        "type": "PLAN_TYPE_PUBLIC",
        "farming_pool_address": "cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx",
        "termination_address": "cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx",
        "staking_coin_weights": [
          {
            "denom": "pool1",
            "amount": "0.500000000000000000"
          },
          {
            "denom": "pool2",
            "amount": "0.500000000000000000"
          }
        ],
        "start_time": "2021-09-10T00:00:00Z",
        "end_time": "2021-10-01T00:00:00Z",
        "terminated": false,
        "last_distribution_time": "2021-09-17T01:00:43.410373Z",
        "distributed_coins": [
          {
            "denom": "stake",
            "amount": "2399261190929"
          }
        ]
      },
      "epoch_ratio": "0.500000000000000000"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### Plan

Query a particular plan:


Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/plans/1
```

```json
{
  "plan": {
    "@type": "/crescent.farming.v1beta1.MsgCreateRatioPlan",
    "base_plan": {
      "id": "1",
      "name": "Second Public Ratio Plan",
      "type": "PLAN_TYPE_PUBLIC",
      "farming_pool_address": "cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx",
      "termination_address": "cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "0.500000000000000000"
        },
        {
          "denom": "pool2",
          "amount": "0.500000000000000000"
        }
      ],
      "start_time": "2021-09-10T00:00:00Z",
      "end_time": "2021-10-01T00:00:00Z",
      "terminated": false,
      "last_distribution_time": "2021-09-17T01:00:43.410373Z",
      "distributed_coins": [
        {
          "denom": "stake",
          "amount": "2399261190929"
        }
      ]
    },
    "epoch_ratio": "0.500000000000000000"
  }
}
```

### Stakings

Query for all stakings by a farmer: 


Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/stakings/cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf
```

```json
{
  "staked_coins": [
    {
      "denom": "pool1",
      "amount": "2500000"
    }
  ],
  "queued_coins": [
  ]
}
```

Query for all stakings by a farmer with the given staking coin denom

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/stakings/cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf?staking_coin_denom=pool1 
```

```json
{
  "staked_coins": [
    {
      "denom": "pool1",
      "amount": "2500000"
    }
  ],
  "queued_coins": [
  ]
}
```
### TotalStakings

Query for total stakings by a staking coin denom: 


Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/total_stakings/pool1 
```

```json
{
  "amount": "2500000"
}
```

### Rewards

Query for all rewards by a farmer:

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/rewards/cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf
```

```json
{
  "rewards": [
    {
      "denom": "stake",
      "amount": "2346201014138"
    }
  ]
}
```


Query for all rewards by a farmer with the staking coin denom:

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/rewards/cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf?staking_coin_denom=pool1
```

```json
{
  "rewards": [
    {
      "denom": "stake",
      "amount": "2346201014138"
    }
  ]
}
```

### CurrentEpochDays

Query for the current epoch days:

Example Request 

<!-- markdown-link-check-disable -->
```bash
http://localhost:1317/cosmos/farming/v1beta1/current_epoch_days
```

```json
{
  "current_epoch_days": 1
}
```
