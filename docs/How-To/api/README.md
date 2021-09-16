---
Title: REST APIs
Description: A high-level overview of how the REST API interfaces work for the farming module.
---

## Swagger Documentation

- [Swagger Docs v0.1.0](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/0.1.0)

## REST APIs

In order to test out the following REST APIs, you need to set up a local node to query from. You can refer to this [localnet tutorial](./Tutorials/localnet) on how to build `farmingd` binary and bootstrap a local network in your local machine.

- [Params](#Params)
- [Plans](#Plans)
- [Plan](#Plan)
- [Stakings](#Stakings)
- [TotalStakings](#TotalStakings)
- [Rewards](#Rewards)

### Params

Query the values set as farming parameters
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
    "epoch_days": 1,
    "farming_fee_collector": "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x"
  }
}
```

### Plans

Query all the farming plans exist in the network

http://localhost:1317/cosmos/farming/v1beta1/plans

```json
{
  "plans": [
    {
      "@type": "/cosmos.farming.v1beta1.FixedAmountPlan",
      "base_plan": {
        "id": "1",
        "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
        "type": "PLAN_TYPE_PRIVATE",
        "farming_pool_address": "cosmos10tmtlj9au53ws34ycu95p0k67mxxus6h4q4z0a5pw77vc4n93nmqfp58g0",
        "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
        "staking_coin_weights": [
          {
            "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
            "amount": "0.800000000000000000"
          },
          {
            "denom": "uatom",
            "amount": "0.200000000000000000"
          }
        ],
        "start_time": "2021-08-06T09:00:00Z",
        "end_time": "2021-08-13T09:00:00Z",
        "terminated": true,
        "last_distribution_time": null,
        "distributed_coins": []
      },
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "100000000"
        }
      ]
    },
    {
      "@type": "/cosmos.farming.v1beta1.RatioPlan",
      "base_plan": {
        "id": "2",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "type": "PLAN_TYPE_PRIVATE",
        "farming_pool_address": "cosmos1tye4hfxt57r65sshv4e4hmq22un9rrmg26v23dyl5grqdn0fsews9uqtfl",
        "termination_address": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
        "staking_coin_weights": [
          {
            "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
            "amount": "0.800000000000000000"
          },
          {
            "denom": "uatom",
            "amount": "0.200000000000000000"
          }
        ],
        "start_time": "2021-08-06T09:00:00Z",
        "end_time": "2021-08-13T09:00:00Z",
        "terminated": true,
        "last_distribution_time": null,
        "distributed_coins": []
      },
      "epoch_ratio": "0.100000000000000000"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### Plan

Query a particular plan 

http://localhost:1317/cosmos/farming/v1beta1/plans/1

```json
{
  "plan": {
    "@type": "/cosmos.farming.v1beta1.FixedAmountPlan",
    "base_plan": {
      "id": "1",
      "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
      "type": "PLAN_TYPE_PRIVATE",
      "farming_pool_address": "cosmos10tmtlj9au53ws34ycu95p0k67mxxus6h4q4z0a5pw77vc4n93nmqfp58g0",
      "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "staking_coin_weights": [
        {
          "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
          "amount": "0.800000000000000000"
        },
        {
          "denom": "uatom",
          "amount": "0.200000000000000000"
        }
      ],
      "start_time": "2021-08-06T09:00:00Z",
      "end_time": "2021-08-13T09:00:00Z",
      "terminated": true,
      "last_distribution_time": null,
      "distributed_coins": []
    },
    "epoch_amount": [
      {
        "denom": "uatom",
        "amount": "100000000"
      }
    ]
  }
}
```

### Stakings

### TotalStakings

### Rewards
