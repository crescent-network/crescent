# Farming Plans

There are two different types of farming plans in the farming module. Whereas a public farming plan can only be created through governance proposal, a private farming plan can be created with any account or an entity. Read [spec](https://github.com/tendermint/farming/blob/master/x/farming/spec/01_concepts.md) documentation for more information about the plan types.

In this tutorial, some sample data in JSON structure are provided. We will use command-line interfaces to test the functionality. 

## Table of Contetns

- [Bootstrap Local Network](#Boostrap)
- [Public Farming Plan](#Public-Farming-Plan)
  * [AddPublicFarmingFixedAmountPlan](#AddPublicFarmingFixedAmountPlan)
  * [AddPublicFarmingRatioPlan](#AddPublicFarmingRatioPlan)
  * [AddMultiplePublicPlans](#AddMultiplePublicPlans)
  * [UpdatePublicFarmingFixedAmountPlan](#UpdatePublicFarmingFixedAmountPlan)
  * [DeletePublicFarmingFixedAmountPlan](#DeletePublicFarmingFixedAmountPlan)
- [Private Farming Plan](#Private-Farming-Plan)
  * [PrivateFarmingFixedAmountPlan](#PrivateFarmingFixedAmountPlan)
  * [PrivateFarmingRatioPlan](#PrivateFarmingRatioPlan)
- [REST APIs](#REST-APIs)

# Bootstrap

```bash
# Clone the project 
git clone https://github.com/tendermint/farming.git
cd cosmos-sdk
make install

# Configure variables
export BINARY=farmingd
export HOME_1=$HOME/.farmingapp
export CHAIN_ID=localnet
export VALIDATOR_1="struggle panic room apology luggage game screen wing want lazy famous eight robot picture wrap act uphold grab away proud music danger naive opinion"
export USER_1="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
export GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd

# Boostrap
$BINARY init $CHAIN_ID --chain-id $CHAIN_ID
echo $VALIDATOR_1 | $BINARY keys add val1 --keyring-backend test --recover
echo $USER_1 | $BINARY keys add user1 --keyring-backend test --recover
$BINARY add-genesis-account $($BINARY keys show val1 --keyring-backend test -a) $GENESIS_COINS
$BINARY add-genesis-account $($BINARY keys show user1 --keyring-backend test -a) $GENESIS_COINS
$BINARY gentx val1 100000000stake --chain-id $CHAIN_ID --keyring-backend test
$BINARY collect-gentxs

# Modify app.toml
sed -i '' 's/enable = false/enable = true/g' $HOME_1/config/app.toml
sed -i '' 's/swagger = false/swagger = true/g' $HOME_1/config/app.toml

# Modify governance proposal for testing purpose
sed -i '' 's%"amount": "10000000"%"amount": "1"%g' $HOME_1/config/genesis.json
sed -i '' 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_1/config/genesis.json
sed -i '' 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_1/config/genesis.json
sed -i '' 's%"voting_period": "172800s"%"voting_period": "60s"%g' $HOME_1/config/genesis.json

# Start
$BINARY start
```

# Public Farming Plan
## AddPublicFarmingFixedAmountPlan

Create `public-fixed-plan-proposal.json` file in your local directory and copy the below json into the file. 

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_request_proposals": [
    {
      "name": "First Public Fixed Amount Plan",
      "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1"
        }
      ]
    }
  ]
}
```

```bash
# Create public fixed amount plan through governance proposal
farmingd tx gov submit-proposal public-farming-plan public-fixed-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
farmingd tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

## AddPublicFarmingRatioPlan

Create `public-ratio-plan-proposal.json` file in your local directory and copy the below json into the file. 

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_request_proposals": [
    {
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
      "epoch_ratio": "1.000000000000000000"
    }
  ]
}
```

```bash
# Create public ratio plan through governance proposal
farmingd tx gov submit-proposal public-farming-plan public-ratio-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
farmingd tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

## AddMultiplePublicPlans

Create `public-multiple-plans-proposal.json` file in your local directory and copy the below json into the file. 

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_request_proposals": [
    {
      "name": "First Public Fixed Amount Plan",
      "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1"
        }
      ]
    },
    {
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "termination_address": "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
      "epoch_ratio": "1.000000000000000000"
    }
  ]
}
```

```bash
# Create public multiple plans through governance proposal
farmingd tx gov submit-proposal public-farming-plan public-multiple-plans-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
farmingd tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```
## UpdatePublicFarmingFixedAmountPlan

Create `update-plan-proposal.json` file in your local directory and copy the below json into the file. 

```json
{
  "title": "Update the Farming Plan 1",
  "description": "FarmingPoolAddress needs to be changed",
  "update_request_proposals": [
    {
      "plan_id": 1,
      "farming_pool_address": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
      "termination_address": "cosmos13w4ueuk80d3kmwk7ntlhp84fk0arlm3mqf0w08",
      "staking_coin_weights": [
        {
          "denom": "uatom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1"
        }
      ]
    }
  ]
}
```

```bash
# Update public plan through governance proposal
farmingd tx gov submit-proposal public-farming-plan update-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
farmingd tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

## DeletePublicFarmingFixedAmountPlan

Create `delete-plan-proposal.json` file in your local directory and copy the below json into the file. 

```json
{
  "title": "Delete Public Farming Plan 1",
  "description": "This plan is no longer needed",
  "delete_request_proposals": [
    {
      "plan_id": 1
    }
  ]
}
```

```bash
# Update public plan through governance proposal
farmingd tx gov submit-proposal public-farming-plan delete-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
farmingd tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

# Private Farming Plan

Create `create-private-fixed-plan.json` file in your local directory and copy the below json into the file. 

## PrivateFarmingFixedAmountPlan

```json
{
  "name": "This Farming Plan intends to incentivize ATOM HODLERS!",
  "staking_coin_weights": [
    {
      "denom": "uatom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21.662422Z",
  "end_time": "2022-07-16T08:41:21.662422Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "1"
    }
  ]
}
```

```bash
# Create private fixed amount plan
farmingd tx farming create-private-fixed-plan create-private-fixed-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```

## PrivateFarmingRatioPlan

Create `create-private-ratio-plan.json` file in your local directory and copy the below json into the file. 

```json
{
  "name": "This Farming Plan intends to incentivize ATOM HODLERS!",
  "staking_coin_weights": [
    {
      "denom": "uatom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21.662422Z",
  "end_time": "2022-07-16T08:41:21.662422Z",
  "epoch_ratio": "1.000000000000000000"
}
```

```bash
# Create private ratio plan
farmingd tx farming create-private-fixed-plan create-private-ratio-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```

## REST APIs

- http://localhost:1317/cosmos/bank/v1beta1/balances/{ADDRESS}
- http://localhost:1317/cosmos/gov/v1beta1/proposals
- http://localhost:1317/cosmos/farming/v1beta1/plans
- http://localhost:1317/cosmos/tx/v1beta1/txs/{TX_HASH}