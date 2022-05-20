# Farming Plans

<!-- markdown-link-check-disable-next-line -->
There are two different types of farming plans in the farming module. Whereas a public farming plan can only be created through governance proposal, a private farming plan can be created with any account or an entity. Read [spec](https://github.com/crescent-network/crescent/blob/main/x/farming/spec/01_concepts.md) documentation for more information about the plan types.

In this documentation, some sample data in JSON are provided. They will be used to test out farming plan functionality.

## Table of Contents

- [Bootstrap Local Network](#Bootstrap)
- [Public Farming Plan](#Public-Farming-Plan)
  * [AddPublicFarmingFixedAmountPlan](#AddPublicFarmingFixedAmountPlan)
  * [AddPublicFarmingRatioPlan](#AddPublicFarmingRatioPlan)
  * [AddMultiplePublicPlans](#AddMultiplePublicPlans)
  * [UpdatePublicFarmingFixedAmountPlan](#UpdatePublicFarmingFixedAmountPlan)
  * [DeletePublicFarmingFixedAmountPlan](#DeletePublicFarmingFixedAmountPlan)
- [Private Farming Plan](#Private-Farming-Plan)
  * [PrivateFarmingFixedAmountPlan](#PrivateFarmingFixedAmountPlan)
  * [PrivateFarmingRatioPlan](#PrivateFarmingRatioPlan)

# Bootstrap

***Since the creation of ratio plans through msg server or gov proposal is disabled by default, you have to build the binary with `make install-testing` to activate it.***

```bash
# Clone the project 
git clone https://github.com/crescent-network/crescent.git
cd crescent
make install-testing

# Configure variables
export BINARY=crescentd
export HOME_APP=$HOME/.crescent
export CHAIN_ID=localnet
export VALIDATOR_1="struggle panic room apology luggage game screen wing want lazy famous eight robot picture wrap act uphold grab away proud music danger naive opinion"
export USER_1="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
export USER_2="fuel obscure melt april direct second usual hair leave hobby beef bacon solid drum used law mercy worry fat super must ritual bring faculty"
export VALIDATOR_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_2_GENESIS_COINS=10000000000stake,10000000000pool1

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

# Check OS for sed -i option value
export SED_I=""
if [[ "$OSTYPE" == "darwin"* ]]; then 
    export SED_I="''"
fi 

# Modify app.toml
sed -i $SED_I 's/enable = false/enable = true/g' $HOME_APP/config/app.toml
sed -i $SED_I 's/swagger = false/swagger = true/g' $HOME_APP/config/app.toml

# (Optional) Modify governance proposal for testing public plan proposal
sed -i $SED_I 's%"amount": "10000000"%"amount": "1"%g' $HOME_APP/config/genesis.json
sed -i $SED_I 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
sed -i $SED_I 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
sed -i $SED_I 's%"voting_period": "172800s"%"voting_period": "60s"%g' $HOME_APP/config/genesis.json

# Start
$BINARY start
```

# Public Farming Plan

The `cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx` address for `farming_pool_address` is derived using the following query or code snippet.
```bash
$BINARY query budget address GravityDEXFarmingBudget --module-name farming
# > address: cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx
```

```go
// cre1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqq6tjyrx
sdk.AccAddress(address.Module("farming", []byte("GravityDEXFarmingBudget")))
```

## AddPublicFarmingFixedAmountPlan

Create `public-fixed-plan-proposal.json` file in your local directory and copy the below JSON into a file. To explain about what this public plan does is that i want to create a public fixed amount plan that provides `100000000uatom` as incentives for those who stake the denoms of `pool1` and `uatom` defined in `staking_coin_weights`.

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_plan_requests": [
    {
      "name": "First Public Fixed Amount Plan",
      "farming_pool_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "0.800000000000000000"
        },
        {
          "denom": "uatom",
          "amount": "0.200000000000000000"
        }
      ],
      "start_time": "0001-01-01T00:00:00Z",
      "end_time": "9999-12-31T00:00:00Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "100000000"
        }
      ]
    }
  ]
}
```

```bash
# Send governance proposal to the network
$BINARY tx gov submit-proposal public-farming-plan public-fixed-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
$BINARY tx gov vote 1 yes \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--yes
```

## AddPublicFarmingRatioPlan

Create `public-ratio-plan-proposal.json` file in your local directory and copy the below JSON into a file. To explain about what this public plan does is that i want to create a public ratio plan that provides 10% of what `farming_pool_address` has in balances as incentives for every epoch to those who stake the `pool1` denom defined in `staking_coin_weights`.

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_plan_requests": [
    {
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "0001-01-01T00:00:00Z",
      "end_time": "9999-12-31T00:00:00Z",
      "epoch_ratio": "0.100000000000000000"
    }
  ]
}
```

```bash
# Send governance proposal to the network
$BINARY tx gov submit-proposal public-farming-plan public-ratio-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
$BINARY tx gov vote 2 yes \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--yes
```

## AddMultiplePublicPlans

Create `public-multiple-plans-proposal.json` file in your local directory and copy the below JSON into a file. 

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_plan_requests": [
    {
      "name": "First Public Fixed Amount Plan",
      "farming_pool_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "0001-01-01T00:00:00Z",
      "end_time": "9999-12-31T00:00:00Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "100000000"
        }
      ]
    },
    {
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "termination_address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "staking_coin_weights": [
        {
          "denom": "uatom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "0001-01-01T00:00:00Z",
      "end_time": "9999-12-31T00:00:00Z",
      "epoch_ratio": "0.100000000000000000"
    }
  ]
}
```

```bash
# Send governance proposal to the network
$BINARY tx gov submit-proposal public-farming-plan public-multiple-plans-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
$BINARY tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

## UpdatePublicFarmingFixedAmountPlan

Create `update-plan-proposal.json` file in your local directory and copy the below JSON into a file. 

```json
{
  "title": "Update the Farming Plan 1",
  "description": "FarmingPoolAddress needs to be changed",
  "modify_plan_requests": [
    {
      "plan_id": 1,
      "farming_pool_address": "cre1j7aapnvqq8jg4vjlsgz5cl38t3qh2pw0mayjfz",
      "termination_address": "cre1j7aapnvqq8jg4vjlsgz5cl38t3qh2pw0mayjfz",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "0001-01-01T00:00:00Z",
      "end_time": "9999-12-31T00:00:00Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1000000000"
        }
      ]
    }
  ]
}
```

```bash
# Send governance proposal to the network
$BINARY tx gov submit-proposal public-farming-plan update-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
$BINARY tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

## DeletePublicFarmingFixedAmountPlan

Create `delete-plan-proposal.json` file in your local directory and copy the below JSON into a file. 

```json
{
  "title": "Delete Public Farming Plan 1",
  "description": "This plan is no longer needed",
  "delete_plan_requests": [
    {
      "plan_id": 1
    }
  ]
}
```

```bash
# Send governance proposal to the network
$BINARY tx gov submit-proposal public-farming-plan delete-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Vote
# Make sure to change proposal-id for the proposal
$BINARY tx gov vote <proposal-id> yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--yes
```

# Private Farming Plan

Create `private-fixed-plan.json` file in your local directory and copy the below JSON into a file. 

## PrivateFarmingFixedAmountPlan

```json
{
  "name": "This Farming Plan intends to incentivize ATOM HODLERS!",
  "staking_coin_weights": [
    {
      "denom": "pool1",
      "amount": "0.200000000000000000"
    },
    {
      "denom": "stake",
      "amount": "0.400000000000000000"
    },
    {
      "denom": "ukava",
      "amount": "0.400000000000000000"
    }
  ],
  "start_time": "0001-01-01T00:00:00Z",
  "end_time": "9999-12-31T00:00:00Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "100000000"
    }
  ]
}
```

```bash
# Send to create a private fixed amount plan
$BINARY tx farming create-private-fixed-plan private-fixed-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```

## PrivateFarmingRatioPlan

Create `private-ratio-plan.json` file in your local directory and copy the below JSON into a file. 

```json
{
  "name": "This Farming Plan intends to incentivize ATOM HODLERS!",
  "staking_coin_weights": [
    {
      "denom": "pool1",
      "amount": "0.200000000000000000"
    },
    {
      "denom": "stake",
      "amount": "0.400000000000000000"
    },
    {
      "denom": "ukava",
      "amount": "0.400000000000000000"
    }
  ],
  "start_time": "0001-01-01T00:00:00Z",
  "end_time": "9999-12-31T00:00:00Z",
  "epoch_ratio": "1.000000000000000000"
}
```

```bash
# Send to create a private ratio plan
$BINARY tx farming create-private-fixed-plan private-ratio-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--yes
```
