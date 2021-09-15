---
title: Farmingd 
description: A high-level overview of how the command-line (CLI) and REST API interfaces work for the farming module.
---
# Farmingd

This document provides a high-level overview of how the command-line (CLI) and REST API interfaces work for the farming module.

## Table of Contetns

- [Prerequisite](#Prerequisite)
- [Command-line Interface](#Command-Line-Interface)
- [REST APIs](#REST-APIs)

## Prerequisite 

### Build

```bash
git clone https://github.com/tendermint/farming.git
cd farming
make install
```

### Boostrap

In order to test out the command-line interface, you need to boostrap local network by using the commands below.

```bash
# Configure variables
export BINARY=farmingd
export HOME_FARMINGAPP=$HOME/.farmingapp
export CHAIN_ID=localnet
export VALIDATOR_1="struggle panic room apology luggage game screen wing want lazy famous eight robot picture wrap act uphold grab away proud music danger naive opinion"
export USER_1="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
export USER_2="fuel obscure melt april direct second usual hair leave hobby beef bacon solid drum used law mercy worry fat super must ritual bring faculty"
export VALIDATOR_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_2_GENESIS_COINS=10000000000stake,10000000000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4

# Bootstrap
$BINARY init $CHAIN_ID --chain-id $CHAIN_ID
echo $VALIDATOR_1 | $BINARY keys add val1 --keyring-backend test --recover
echo $USER_1 | $BINARY keys add user1 --keyring-backend test --recover
echo $USER_2 | $BINARY keys add user2 --keyring-backend test --recover
$BINARY add-genesis-account $($BINARY keys show val1 --keyring-backend test -a) $VALIDATOR_1_GENESIS_COINS
$BINARY add-genesis-account $($BINARY keys show user1 --keyring-backend test -a) $USER_1_GENESIS_COINS
$BINARY add-genesis-account $($BINARY keys show user2 --keyring-backend test -a) $USER_2_GENESIS_COINS
$BINARY gentx val1 100000000stake --chain-id $CHAIN_ID --keyring-backend test
$BINARY collect-gentxs

# Modify app.toml
sed -i '' 's/enable = false/enable = true/g' $HOME_FARMINGAPP/config/app.toml
sed -i '' 's/swagger = false/swagger = true/g' $HOME_FARMINGAPP/config/app.toml

# Start
$BINARY start
```
## Command-Line Interface

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
    * [Staking](#Staking)
    * [Rewards](#Rewards)

## Transaction

+++ https://github.com/tendermint/farming/blob/master/proto/tendermint/farming/v1beta1/tx.proto#L13-L29

### MsgCreateFixedAmountPlan

```bash
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

# Send to create a private fixed amount plan
farmingd tx farming create-private-fixed-plan private-fixed-plan.json \
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
        "@type": "/cosmos.farming.v1beta1.MsgCreateFixedAmountPlan",
        "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
        "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
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

```bash
# An example of private-ratio-plan.json
#
# This private ratio farming plan intends to provide ratio of all coins that farming pool address has per epoch (measured in day).
# In this example, epoch ratio is 10 percent and 10 percent of all the coins that the creator of this plan has in balances are used as incentives for the denoms defined in staking_coin_weights.
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

# Send to create a private ratio plan
farmingd tx farming create-private-ratio-plan private-ratio-plan.json \
--chain-id localnet \
--from val1 \
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
        "@type": "/cosmos.farming.v1beta1.MsgCreateRatioPlan",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "farming_pool_address": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
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
--yes

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
--yes
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
# Note that epoch_days are meausred in days so you will confront a log stating that
# there is no reward for staking coin denom
farmingd tx farming harvest "uatom" \
--chain-id localnet \
--from user2 \
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
  "staking_creation_fee": [
    {
      "denom": "stake",
      "amount": "100000"
    }
  ],
  "epoch_days": 1,
  "farming_fee_collector": "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x"
}
```
### Plans 

```bash
# Query for all farmings plans on a network
farmingd q farming plans --output json | jq

# Query for all farmings plans with the given plan type
# plan type must be either public or private
farmingd q farming plans \
--plan-type private \
--output json | jq

# Query for all farmings plans with the given farming pool address
farmingd q farming plans \
--farming-pool-addr cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v \
--output json | jq

# Query for all farmings plans with the given reward pool address
farmingd q farming plans \
--reward-pool-addr cosmos1gshap5099dwjdlxk2ym9z8u40jtkm7hvux45pze8em08fwarww6qc0tvl0 \
--output json | jq

# Query for all farmings plans with the given termination address
farmingd q farming plans \
--termination-addr cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v \
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
      "@type": "/cosmos.farming.v1beta1.FixedAmountPlan",
      "base_plan": {
        "id": "1",
        "name": "This plan intends to provide incentives for Cosmonauts!",
        "type": "PLAN_TYPE_PRIVATE",
        "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
        "reward_pool_address": "cosmos1gshap5099dwjdlxk2ym9z8u40jtkm7hvux45pze8em08fwarww6qc0tvl0",
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
        "end_time": "2021-08-13T09:00:00Z"
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
        "farming_pool_address": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
        "reward_pool_address": "cosmos1qr3xrf66kl594rjtj5mukz2khym4srent0cjafenat6xwym6q8gsq50x7g",
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
        "end_time": "2021-08-13T09:00:00Z"
      },
      "epoch_ratio": "0.100000000000000000"
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
  "plan": {
    "@type": "/cosmos.farming.v1beta1.FixedAmountPlan",
    "base_plan": {
      "id": "1",
      "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
      "type": "PLAN_TYPE_PRIVATE",
      "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "reward_pool_address": "cosmos1gshap5099dwjdlxk2ym9z8u40jtkm7hvux45pze8em08fwarww6qc0tvl0",
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
      "terminated": false
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

```bash
# Query for all stakings on a network
farmingd q farming stakings --output json | jq

# Query for all stakings with the given staking coin denom
farmingd q farming stakings \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--output json | jq

# Query for all stakings with the given farmer address
farmingd q farming stakings \
--farmer-addr cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Query for all stakings with the given params
farmingd q farming stakings \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--farmer-addr cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq
```

```json
{
  "stakings": [
    {
      "id": "1",
      "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
      "staked_coins": [],
      "queued_coins": [
        {
          "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
          "amount": "2500000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### Staking

```bash
# Query for the details of staking with the given staking id
farmingd q farming staking 1 --output json | jq

```

```json
{
  "staking": {
    "id": "1",
    "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
    "staked_coins": [],
    "queued_coins": [
      {
        "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
        "amount": "2500000"
      }
    ]
  }
}
```

### Rewards

```bash
# Query for all rewards on a network
farmingd q farming rewards

# Query for all rewards with the given farmer address
farmingd q farming rewards --farmer-addr cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny --output json

# Query for all rewards with the given staking coin denom
farmingd q farming rewards --staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 --output json

# Query for all rewards with the given params
farmingd q farming rewards \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--farmer-addr cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json
```

```json
{
  "rewards": [
    {
      "farmer": "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny",
      "staking_coin_denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
      "reward_coins": [
        {
          "denom": "uatom",
          "amount": "80000000"
        }
      ]
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "0"
  }
}
```


## REST APIs

- [Swagger Docs v0.1.0](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/0.1.0)
