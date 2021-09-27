---
Title: Farmingd
Description: A high-level overview of how the command-line (CLI) interfaces work for the farming module.
---

# Farmingd

This document provides a high-level overview of how the command-line (CLI) interfaces work for the farming module.

## Command-Line Interfaces

In order to test out the following command-line interfaces, you need to set up a local node to either send transaction or query from. You can refer to this [localnet tutorial](../../Tutorials/localnet) on how to build `farmingd` binary and bootstrap a local network in your local machine.

- [Transaction](#Transaction)
    * [MsgCreateFixedAmountPlan](#MsgCreateFixedAmountPlan)
    * [MsgCreateRatioPlan](#MsgCreateRatioPlan)
    * [MsgStake](#MsgStake)
    * [MsgUnstake](#MsgUnstake)
    * [MsgHarvest](#MsgHarvest)
- [Query](#Query)
    * [Params](#Params)
    * [Plans](#Plans)
    * [Plan](#Plan)
    * [Stakings](#Stakings)
    * [TotalStakings](#TotalStakings)
    * [Rewards](#Rewards)
    * [CurrentEpochDays](#CurrentEpochDays)

## Transaction

+++ https://github.com/tendermint/farming/blob/master/proto/tendermint/farming/v1beta1/tx.proto#L13-L29

### MsgCreateFixedAmountPlan

Create a file name `private-fixed-plan.json`. This private fixed amount farming plan intends to provide 100ATOM per epoch (measured in day) relative to the rate amount of denoms defined in staking coin weights.

- `name`: is the name of the farming plan. It can be any name you prefer to be stored in a blockchain network; however it cannot overlap with the existing plan names.
- `staking_coin_weights`: is the distributing amount for each epoch. An amount should be decimal, not an integer. The sum of total weight must be 1.000000000000000000
- `start_time`: is start time of the farming plan 
- `end_time`: is end time of the farming plan
- `epoch_amount`: is an amount that will be distributed per epoch as an incentive for staking denoms defined in the staking coin weights.

```json
{
  "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
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
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "100000000"
    }
  ]
}
```

```bash
# Create a private fixed amount plan
farmingd tx farming create-private-fixed-plan private-fixed-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgCreateFixedAmountPlan",
        "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
        "creator": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
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
        "epoch_amount": [
          {
            "denom": "uatom",
            "amount": "100000000"
          }
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "AvzwBOriY8sVwEXrXf1gXanhT9imlfWeUWLQ8pMxrRsg"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "0"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "O/BIpLWPJngobN3PAFzzzI1grKWmg5kMa6XfAP96k1lF2x4p8A6ob1grwKSvN1lILgEPMe6d2V5SDHmZ6fZBNg=="
  ]
}
```

### MsgCreateRatioPlan

Create a file name `private-fixed-plan.json`. This private ratio farming plan intends to provide ratio of all coins that farming pool address has per epoch (measured in day). In this example, epoch ratio is 10 percent and 10 percent of all the coins that the creator of this plan has in balances are used as incentives for the denoms defined in the staking coin weights.

- `name`: is the name of the farming plan. It can be any name you prefer to be stored in a blockchain network; however it cannot overlap with the existing plan names.
- `staking_coin_weights`: is the distributing amount for each epoch. An amount should be decimal, not an integer. The sum of total weight must be 1.000000000000000000
- `start_time`: is start time of the farming plan 
- `end_time`: is end time of the farming plan
- `epoch_ratio`: is a ratio that will be distributed per epoch as an incentive for staking denoms defined in staking coin weights. The ratio refers to all coins that the creator has in his/her account. Note that the total ratio cannot exceed 1.0 (100%). 

```json
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
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
  "epoch_ratio": "0.100000000000000000"
}
```

```bash
# Create a private ratio plan
farmingd tx farming create-private-ratio-plan private-ratio-plan.json \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgCreateRatioPlan",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "creator": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
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
        "epoch_ratio": "0.100000000000000000"
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "A7B7KsceK2UklO3tyH2XkPBZGzEpvOf+35vwTUisxVKV"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "1"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "y1PFSvxUfsqEbAQacLgAICTNYWfITHAlzjlzUkBj2rABWnbcU2NQZSUQKz6oYiHCKfWm7gOSPIL1pDD6Am+xtg=="
  ]
}
```

### MsgStake

```bash
# Stake pool coin
farmingd tx farming stake 5000000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgStake",
        "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
        "staking_coins": [
          {
            "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
            "amount": "5000000"
          }
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "AuFUt9g9uckLNgVlO7BCzqUCOL8OUg+zIgeHTxxeG4Fy"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "0"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "hYDIb0leA1RJm8Lcu0Mt1uPXIMP7lWVUquC3uqAls8FWMUL3Fy+OBejmjpcjp9Fh+y/YjsbakLT5IZixkVLuzw=="
  ]
}
```

### MsgUnstake

```bash
# Unstake coins from the farming plan
farmingd tx farming unstake 2500000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgUnstake",
        "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
        "unstaking_coins": [
          {
            "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
            "amount": "2500000"
          }
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "AuFUt9g9uckLNgVlO7BCzqUCOL8OUg+zIgeHTxxeG4Fy"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "1"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "VYErDdkw3EWiahPK7vW/VYzKTHS4RPoLc30ZxpAONm1UFdRBlRrsibUngQlK+EmtkrJHlxopzEPhv5ekxlC8Dg=="
  ]
}
```

### MsgHarvest

```bash
# Harvest farming rewards from the farming plan
# Note that there won't be any rewards if the time hasn't passed by the epoch days
farmingd tx farming harvest uatom \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# or

# Harvest all with --all flag
farmingd tx farming harvest \
--all \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgHarvest",
        "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
        "staking_coin_denoms": [
          "uatom",
          "stake"
        ]
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "AvzwBOriY8sVwEXrXf1gXanhT9imlfWeUWLQ8pMxrRsg"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "3"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "Pm5PxNyE/r5U4P2i3ihAlfot3HQFR0lKvVNEj35vHAMn5mYTq9rVEYhp3Oux710GspueXqax5wKkeAhhr8tz9Q=="
  ]
}
```

## Query

https://github.com/tendermint/farming/blob/master/proto/tendermint/farming/v1beta1/query.proto#L15-L40

### Params 

```bash
# Query the values set as farming parameters
farmingd q farming params --output json | jq
```

```json
{
  "private_plan_creation_fee": [
    {
      "denom": "stake",
      "amount": "100000000"
    }
  ],
  "next_epoch_days": 1,
  "farming_fee_collector": "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x"
}
```
### Plans 

```bash
# Query for all farmings plans on a network
farmingd q farming plans --output json | jq

# Query for all farmings plans with the given plan type
# plan type must be either PLAN_TYPE_PUBLIC or PLAN_TYPE_PRIVATE
farmingd q farming plans \
--plan-type PLAN_TYPE_PUBLIC \
--output json | jq

# Query for all farmings plans with the given farming pool address
farmingd q farming plans \
--farming-pool-addr cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08 \
--output json | jq

# Query for all farmings plans with the given reward pool address
farmingd q farming plans \
--reward-pool-addr cosmos1gshap5099dwjdlxk2ym9z8u40jtkm7hvux45pze8em08fwarww6qc0tvl0 \
--output json | jq

# Query for all farmings plans with the given termination address
farmingd q farming plans \
--termination-addr cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Query for all farmings plans with the given staking coin denom
farmingd q farming plans \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--output json | jq
```

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
    "total": "0"
  }
}
```
### Plan 

```bash
# Query plan with the given plan id
farmingd q farming plan 1 --output json | jq
```

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
    "total": "0"
  }
}
```

### Stakings 

```bash
# Query for all stakings by a farmer 
farmingd q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny --output json | jq

# Query for all stakings by a farmer with the given staking coin denom
farmingd q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--output json | jq
```

```json
{
  "staked_coins": [],
  "queued_coins": [
    {
      "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
      "amount": "5000000"
    }
  ]
}
```

### TotalStakings

```bash
# Query for total stakings by a staking coin denom 
farmingd q farming total-stakings poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 --output json | jq
```

```json
{
  "amount": "2500000"
}
```

### Rewards

```bash
# Query for all rewards by a farmer 
farmingd q farming rewards cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny --output json | jq

# Query for all rewards by a farmer with the staking coin denom
farmingd q farming rewards cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--output json | jq
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

```bash
# Query for the current epoch days
farmingd q farming current-epoch-days --output json | jq
```

```json
{
  "current_epoch_days": 1
}
```