---
Title: REST APIs
Description: A high-level overview of gRPC-gateway REST routes in farming module.
---

## Swagger Documentation

- Swagger Cosmos SDK Farming Module [REST and gRPC Gateway docs](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/1.0.0)

## gRPC-gateway REST Routes

In order to test out the following REST routes, set up a local node to query from. See the [localnet tutorial](../../Tutorials/localnet) on how to build the `crescentd` binary and bootstrap a local network in your local machine.

- [Params](#Params)
- [Plans](#Plans)
- [Plan](#Plan)
- [Stakings](#Stakings)
- [TotalStakings](#TotalStakings)
- [Rewards](#Rewards)
- [CurrentEpochDays](#CurrentEpochDays)

### Params

Query the values set as farming parameters:
<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/params

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
    "farming_fee_collector": "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x"
  }
}
```

### Plans

Query all the farming plans exist in the network:

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/plans

```json
{
  "plans": [
    {
      "@type": "/cosmos.farming.v1beta1.RatioPlan",
      "base_plan": {
        "id": "1",
        "name": "Second Public Ratio Plan",
        "type": "PLAN_TYPE_PUBLIC",
        "farming_pool_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
        "termination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
        "staking_coin_weights": [
          {
            "denom": "pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C",
            "amount": "0.500000000000000000"
          },
          {
            "denom": "poolE4D2617BFE03E1146F6BBA1D9893F2B3D77BA29E7ED532BB721A39FF1ECC1B07",
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

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/plans/1

```json
{
  "plan": {
    "@type": "/cosmos.farming.v1beta1.RatioPlan",
    "base_plan": {
      "id": "1",
      "name": "Second Public Ratio Plan",
      "type": "PLAN_TYPE_PUBLIC",
      "farming_pool_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "termination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "staking_coin_weights": [
        {
          "denom": "pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C",
          "amount": "0.500000000000000000"
        },
        {
          "denom": "poolE4D2617BFE03E1146F6BBA1D9893F2B3D77BA29E7ED532BB721A39FF1ECC1B07",
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

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/stakings/cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny

```json
{
  "staked_coins": [
    {
      "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
      "amount": "2500000"
    }
  ],
  "queued_coins": [
  ]
}
```

Query for all stakings by a farmer with the given staking coin denom

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/stakings/cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny?staking_coin_denom=poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 

```json
{
  "staked_coins": [
    {
      "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
      "amount": "2500000"
    }
  ],
  "queued_coins": [
  ]
}
```
### TotalStakings

Query for total stakings by a staking coin denom: 

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/total_stakings/poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 

```json
{
  "amount": "2500000"
}
```

### Rewards

Query for all rewards by a farmer:

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/rewards/cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny

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

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/rewards/cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny?staking_coin_denom=poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4

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

<!-- markdown-link-check-disable-next-line -->
http://localhost:1317/cosmos/farming/v1beta1/current_epoch_days

```json
{
  "current_epoch_days": 1
}
```
