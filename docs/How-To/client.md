---
title: Farmingd 
description: A high-level overview of how the command-line (CLI) and REST API interfaces work for the farming module.
---

# Farmingd

This document provides a high-level overview of how the command-line (CLI) and REST API interfaces work for the farming module.

## Command-Line Interface

- [Transaction](#Transaction)
    * [MsgCreateFixedAmountPlan](#MsgCreateFixedAmountPlan)
    * [MsgCreateRatioPlan](#MsgCreateRatioPlan)
    * [MsgStake](#MsgStake)
    * [MsgUnstake](#MsgUnstake)
    * [MsgHarvest](#MsgHarvest)
- [Query](#Query)
    * [Params](#Params)

## Transaction

+++ https://github.com/tendermint/farming/blob/master/proto/tendermint/farming/v1beta1/tx.proto#L13-L29

### MsgCreateFixedAmountPlan

```bash
# Create private fixed amount plan
farmingd tx farming create-private-fixed-plan private-fixed-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes

# An example of private-fixed-plan.json file
#
# This private fixed amount farming plan intends to provide 100ATOM per epoch (measured in day) 
# relative to the rate amount of denoms defined in staking coin weights. 
#
# Parameter Description
#
# name: it can be any name you prefer to be stored in a network. It cannot be overlap with the existing names.
# staking_coin_weights: an amount should be decimal, not an integer. The sum of total weight must be 1.000000000000000000
# epoch_amount: this is an amount that you want to provide as incentive for staking denoms defined in staking coin weights.
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "stake",
      "amount": "0.500000000000000000"
    },
    {
      "denom": "uatom",
      "amount": "0.500000000000000000"
    }
  ],
  "start_time": "2021-07-24T08:41:21.662422Z",
  "end_time": "2022-07-28T08:41:21.662422Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "100000000"
    }
  ]
}
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgCreateFixedAmountPlan",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
        "staking_coin_weights": [
          {
            "denom": "stake",
            "amount": "0.500000000000000000"
          },
          {
            "denom": "uatom",
            "amount": "0.500000000000000000"
          }
        ],
        "start_time": "2021-07-24T08:41:21.662422Z",
        "end_time": "2022-07-28T08:41:21.662422Z",
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
    "xQgne+eEQxK2OZtUppuQI49WqOTIdsPpyOwek4ybveMYdlCgVTK+LOaqN4N6o6gNRfqHN46bCVymc/B59nWBdg=="
  ]
}
```

### MsgCreateRatioPlan

```bash
# Create private ratio plan
farmingd tx farming create-private-ratio-plan private-ratio-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes

# An example of private-ratio-plan.json
#
# This private ratio farming plan intends to provide ratio of all coins that farming pool address has per epoch (measured in day).
# In this example, epoch ratio is 10 percent and 10 percent of all the coins that the creator of this plan has are going to be used as incentives for the stakings.
#
# Parameter Description
#
# name: it can be any name you prefer to be stored in a network. It cannot be overlap with the existing names.
# staking_coin_weights: an amount should be decimal, not an integer. The sum of total weight must be 1.000000000000000000
# epoch_ratio: distributing ratio (of all coins that the creator has) per epoch. The total ratio cannot exceed 1.000000000000000000 (100%)
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "uatom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21.662422Z",
  "end_time": "2022-07-16T08:41:21.662422Z",
  "epoch_ratio": "0.100000000000000000"
}
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgCreateRatioPlan",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
        "staking_coin_weights": [
          {
            "denom": "uatom",
            "amount": "1.000000000000000000"
          }
        ],
        "start_time": "2021-07-15T08:41:21.662422Z",
        "end_time": "2022-07-16T08:41:21.662422Z",
        "epoch_ratio": "0.500000000000000000"
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
        "sequence": "8"
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
    "f5Op+e5QPTXv+akqW1tideS7eNA+zzIry6RgZ1sZOgQgKsmCQu9h5W6JxCQoE7zxq7NbPyJD0zvSgxDiC6Vsjg=="
  ]
}
```

### MsgStake

```bash
# Stake coins into the farming plan
farmingd tx farming stake 10000000uatom \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes

# Stake coins into the farming plan
farmingd tx farming stake 10000000stake \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```

```json
{
      "@type": "/cosmos.tx.v1beta1.Tx",
      "body": {
        "messages": [
          {
            "@type": "/cosmos.farming.v1beta1.MsgStake",
            "farmer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
            "staking_coins": [
              {
                "denom": "uatom",
                "amount": "10000000"
              }
            ]
          }
        ],
        "memo": "",
        "timeout_height": "0",
        "extension_options": [
        ],
        "non_critical_extension_options": [
        ]
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
            "sequence": "7"
          }
        ],
        "fee": {
          "amount": [
          ],
          "gas_limit": "200000",
          "payer": "",
          "granter": ""
        }
      },
      "signatures": [
        "HOi4b+WVWO3O0j8OL3vXI+vfnVX0euh1Z0crbq0CaDFytfJFMJRUsODUslwZotfPamGI5DH/ACXvwJk1daQ4oA=="
      ]
    }
```

### MsgUnstake

```bash
# Unstake coins from the farming plan
farmingd tx farming unstake 50000uatom \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgUnstake",
        "farmer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
        "unstaking_coins": [
          {
            "denom": "uatom",
            "amount": "50000"
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
        "sequence": "6"
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
    "yWQNZyw0k+obrKAdYl1s/RzyKyp+pRHzldHT3HKXBrIZWGpIt26GNHLsH9RQWVXb+KiMd2aSIoBfkuse0eQ7Og=="
  ]
}
```

### MsgHarvest

```bash
# Harvest farming rewards from the farming plan
# Note that epoch_days are meausred in days so you will confront a log stating that
# there is no reward for staking coin denom
farmingd tx farming harvest "uatom,stake" \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes
```

```json
{
  "@type": "/cosmos.tx.v1beta1.Tx",
  "body": {
    "messages": [
      {
        "@type": "/cosmos.farming.v1beta1.MsgHarvest",
        "farmer": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
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

```
