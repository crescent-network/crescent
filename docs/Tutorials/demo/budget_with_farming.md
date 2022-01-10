# Demo

## Changelog

- 2021.09.27: initial document
- 2021.10.15: update farming, budget version
- 2021.11.12: update farming, budget version to v1.0.0-rc1
- 2021.11.26: update farming, budget version to v1.0.0

## What does budget module do?

The budget module is a Cosmos SDK module that implements budget functionality. It is an independent module from other SDK modules and core functionality is to enable anyone to create a budget plan through governance param change proposal. After it is agreed within the community, voted, and passed, it uses the budget source address to distribute amount of coins by the rate that is defined in the plan to the destination address. Collecting all budgets and distribution take place every epoch blocks that can be modified by a governance proposal.

One use case is the Gravity DEX farming plan. The budget module can be used to create a budget plan that defines the Cosmos Hub FeeCollector module account where transaction gas fees and part of ATOM inflation are collected as budget source address and uses a custom module account (created by budget creator) as the collection address. 

## What does the farming module do?

The farming module is a Cosmos SDK-based module that implements farming functionality that provides farming rewards to staking participants called farmers. 

One use case is to use the module to provide incentives for liquidity pool investors for their pool participation.

## Resources

**Budget module**

- [tendermint/budget GitHub repo](https://github.com/tendermint/budget)
- [x/budget spec docs](https://github.com/tendermint/budget/blob/main/x/budget/spec/01_concepts.md)
- Audit-ready [demo version](https://github.com/tendermint/budget/releases)
- Other useful resources are available in the [budget docs](https://github.com/tendermint/budget/blob/main/docs) folder
- Swagger Cosmos SDK Budget Module [REST and gRPC Gateway docs v1.0.0](https://app.swaggerhub.com/apis-docs/gravity-devs/budget/1.0.0)

**Farming module**

- [tendermint/farming GitHub repo](https://github.com/tendermint/farming)
- [x/farming spec docs](https://github.com/tendermint/farming/blob/main/x/farming/spec/01_concepts.md)
- Audit-ready [demo version](https://github.com/tendermint/farming/releases)
- Other useful resources are available in the [farming docs](https://github.com/tendermint/farming/blob/main/docs) folder
- Swagger Cosmos SDK Budget Module [REST and gRPC Gateway docs v1.0.0](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/1.0.0)

## Demo

### Step 1. Build from source

```bash
# Clone the demo project and build `crescentd` for testing
git clone https://github.com/crescent-network/crescent.git
cd crescent
make install-testing
```

### Step 2. Spin up a local node

- Provide 4 different types of pool coins to `user2` in genesis
- Increase inflation rate from default to 33% for better testing environment
- Modify governance parameters to lower threshold and decrease time to reduce governance process

```bash
export BINARY=crescentd
export HOME_FARMINGAPP=$HOME/.crescentapp
export CHAIN_ID=localnet
export VALIDATOR_1="struggle panic room apology luggage game screen wing want lazy famous eight robot picture wrap act uphold grab away proud music danger naive opinion"
export USER_1="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
export USER_2="fuel obscure melt april direct second usual hair leave hobby beef bacon solid drum used law mercy worry fat super must ritual bring faculty"
export VALIDATOR_1_GENESIS_COINS=100000000000000000stake,10000000000uatom,10000000000uusd
export USER_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_2_GENESIS_COINS=10000000000stake,10000000000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4,10000000000pool93E069B333B5ECEBFE24C6E1437E814003248E0DD7FF8B9F82119F4587449BA5,10000000000pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C,10000000000poolE4D2617BFE03E1146F6BBA1D9893F2B3D77BA29E7ED532BB721A39FF1ECC1B07

# Bootstrap
$BINARY init $CHAIN_ID --chain-id $CHAIN_ID
echo $VALIDATOR_1 | $BINARY keys add val1 --keyring-backend test --recover
echo $USER_1 | $BINARY keys add user1 --keyring-backend test --recover
echo $USER_2 | $BINARY keys add user2 --keyring-backend test --recover
$BINARY add-genesis-account $($BINARY keys show val1 --keyring-backend test -a) $VALIDATOR_1_GENESIS_COINS
$BINARY add-genesis-account $($BINARY keys show user1 --keyring-backend test -a) $USER_1_GENESIS_COINS
$BINARY add-genesis-account $($BINARY keys show user2 --keyring-backend test -a) $USER_2_GENESIS_COINS
$BINARY gentx val1 50000000000000000stake --chain-id $CHAIN_ID --keyring-backend test
$BINARY collect-gentxs

# Check platform
platform='unknown'
unamestr=`uname`
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

# Enable API and swagger docs and modify parameters for the governance proposal and
# inflation rate from 13% to 33%
if [ $platform = 'linux' ]; then
	sed -i 's/enable = false/enable = true/g' $HOME_FARMINGAPP/config/app.toml
	sed -i 's/swagger = false/swagger = true/g' $HOME_FARMINGAPP/config/app.toml
	sed -i 's%"amount": "10000000"%"amount": "1"%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i 's%"voting_period": "172800s"%"voting_period": "30s"%g' $HOME_FARMINGAPP/config/genesis.json
  sed -i 's%"inflation": "0.130000000000000000",%"inflation": "0.330000000000000000",%g' $HOME_FARMINGAPP/config/genesis.json
else
	sed -i '' 's/enable = false/enable = true/g' $HOME_FARMINGAPP/config/app.toml
	sed -i '' 's/swagger = false/swagger = true/g' $HOME_FARMINGAPP/config/app.toml
	sed -i '' 's%"amount": "10000000"%"amount": "1"%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i '' 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i '' 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
	sed -i '' 's%"voting_period": "172800s"%"voting_period": "30s"%g' $HOME_FARMINGAPP/config/genesis.json
  sed -i '' 's%"inflation": "0.130000000000000000",%"inflation": "0.330000000000000000",%g' $HOME_FARMINGAPP/config/genesis.json
fi

# Start
$BINARY start
```

### Step 3. Send a governance proposal to create a budget plan 

Create the `budget-proposal.json` file and copy the following JSON contents into the file. Depending on what budget plan you create, you can customize values of the fields. 

In this demo, you create a budget plan that distributes partial amount of coins from the [FeeCollector module account](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/keys.go#L15) that collects gas fees and ATOM inflation in Cosmos Hub. This budget plan will be used for Gravity DEX farming plan to `GravityDEXFarmingBudget` account. 

The `GravityDEXFarmingBudget` account is derived using the following query.
```bash
$BINARY query budget address GravityDEXFarmingBudget --module-name farming
# > address: cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky
```

This code snippet is how the module derives the account.

```go
// cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky
sdk.AccAddress(address.Module("farming", []byte("GravityDEXFarmingBudget")))
```

where the fields in the JSON file are:

- `name`: the name of the budget plan used for display
- `description`: the description of the budget plan used for display
- `rate`: the distributing amount by ratio of the total budget source
- `source_address`: the address where the source of budget comes from
- `destination_address`: the address that collects budget from the budget source address
- `start_time`: start time of the budget plan
- `end_time`: end time of the budget plan

Use the following values for the fields:

```json
{
  "title": "Create a Budget Plan",
  "description": "An example of Budget Plan for Gravtiy DEX Farming",
  "changes": [
    {
      "subspace": "budget",
      "key": "Budgets",
      "value": [
        {
          "name": "gravity-dex-farming-20213Q-20313Q",
          "rate": "0.500000000000000000",
          "source_address": "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
          "destination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
          "start_time": "2021-09-01T00:00:00Z",
          "end_time": "2031-09-30T00:00:00Z"
        }
      ]
    }
  ],
  "deposit": "10000000stake"
}
```

Now, run each command one at a time. You can copy and paste each command in the command line in your terminal:

```bash
# Submit a param-change governance proposal
$BINARY tx gov submit-proposal param-change budget-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Vote
$BINARY tx gov vote 1 yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Wait a while (30s) for the proposal to pass
#

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals --output json | jq

# Query the values set as budget parameters
$BINARY q budget params --output json | jq
```

### Step 4. Query `GravityDEXFarmingBudget` account to see if coins are accrued

```bash
# Query balances of the Gravity DEX budget collector account address
$BINARY q bank balances cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky \
--output json | jq
```

### Step 5. Send a governance proposal to create a public ratio amount plan

Now, create the `public-ratio-plan-proposal.json` file and copy the following JSON contents into the file. 

In this demo, you create a public farming ratio plan to distribute 90% of of the balance of the farming pool address to the accounts who stake the coins that are defined in staking coin weights, the time period is from Sept. 01, 2021 to Sept. 24, 2021. 

where the fields in the JSON file are:

- `name`: is the name of the farming plan used for display. It allows duplicate value.
- `farming_pool_address`: is the faucet address for the plan
- `termination_address`: is the address that the remaining coins are transferred to when the plan ends
- `staking_coin_weights`: are the coin weights for the plan. The weights must add up to 1
- `start_time`: is start time of the farming plan
- `end_time`: is start time of the farming plan
- `epoch_ratio`: is the distributing amount by ratio per epoch

where the fields in the JSON file are:

```json
{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "add_plan_requests": [
    {
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "termination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "staking_coin_weights": [
        {
          "denom": "pool93E069B333B5ECEBFE24C6E1437E814003248E0DD7FF8B9F82119F4587449BA5",
          "amount": "0.500000000000000000"
        },
        {
          "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
          "amount": "0.500000000000000000"
        }
      ],
      "start_time": "2021-09-01T00:00:00Z",
      "end_time": "2031-09-30T00:00:00Z",
      "epoch_ratio": "0.900000000000000000"
    }
  ]
}
```

Now, run each command one at a time. You can copy and paste each command in the command line in your terminal:

```bash
# Submit a public ratio plan governance proposal
$BINARY tx gov submit-proposal public-farming-plan public-ratio-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes \
--output json | jq

# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Vote
$BINARY tx gov vote 2 yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Wait for a while to pass the proposal (30s)
#

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals --output json | jq

# Query for all plans on a network
$BINARY q farming plans --output json | jq
```

### Step 6. Stake coins

```bash
# Query balance of user2
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Stake pool coin
$BINARY tx farming stake 5000000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query for all stakings by a staker address
$BINARY q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# You can also query using the following command
# Query for all stakings by a staker address with the given staking coin denom
$BINARY q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--output json | jq
```

### Step 7. (Custom Message) Send AdvanceEpoch for Reward Distribution

To simulate reward distribution for this demo, enable a custom transaction message `AdvanceEpoch` when you build the binary `crescentd` with the `make install-testing` command. 

When you send the `AdvanceEpoch` message to the network, it increases epoch by day 1.

In this step, you might wonder why you need to increase 2 epochs by sending two transactions to the network. The reason is to ensure fairness of distribution. The global parameter called `next_epoch_days` can be updated through a param change governance proposal. If the value of `next_epoch_days` is changed, it can lead to an edge case. Let's say `next_epoch_days` is 7 and it is changed to 1 although it hasn't proceeded up to 7 days before it is changed. Therefore, the internal state `current_epoch_days` is used to process staking and reward distribution in an end blocker. This technical decision has been made by the Gravity DEX team. To understand more about this decision, feel free to jump right into [the code](https://github.com/tendermint/farming/blob/main/x/farming/abci.go#L13).

```bash
# Increase epoch by 1 
$BINARY tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query for all stakings by a staker address
# Queued coins should have been moved to staked coins 
$BINARY q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Increase epoch by 1 again to distribute rewards
$BINARY tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query rewards
$BINARY q farming rewards cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq
```

### Step 8. Harvest farming rewards

```bash
# Query balance of user2 account 
# There should be no rewards claimed yet
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Harvest all with all flag
$BINARY tx farming harvest \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--all \
--yes \
--output json | jq

# Query the balance again to see if stake coin has increased
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# You can also query with the following command
# Harvest farming rewards from the farming plan with the staking coin
$BINARY tx farming harvest poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

### Step 9. Modify the public farming ratio plan

Now create the `multiple-public-ratio-plan-proposals.json` file and copy the JSON contents into the file. 

This step is intended to demonstrate the fact that you don't need to create another public ratio plan by sending another governance proposal. You can just modify the existing proposal and add another ratio plan.

Update the following values of the fields:

- `plan_id`: 1
- `staking_coin_weights`
    - `pool93E069B333B5ECEBFE24C6E1437E814003248E0DD7FF8B9F82119F4587449BA5` weight 50% → 100%
    - `poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4` weight 50% → 0% ( deleted)
- `epoch_ratio`: 0.500000000000000000 (50%)

Add a second public ratio plan proposal:

- `name`: Second Public Ratio Plan
- `farming_pool_address`: the Gravity DEX budget collector account address
- `termination_address`: the Gravity DEX budget collector account address
- `staking_coin_weights`
    - `pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C`
    - `poolE4D2617BFE03E1146F6BBA1D9893F2B3D77BA29E7ED532BB721A39FF1ECC1B07`
- `start_time`: 2021-09-11T00:00:00Z
- `end_time`: 2031-09-30T00:00:00Z
- `epoch_ratio`: 0.500000000000000000 (50%)

```json
{
  "title": "Update Public Farming Plan",
  "description": "Are you ready to farm?",
  "modify_plan_requests": [
    {
      "plan_id": 1,
      "name": "First Public Ratio Plan",
      "farming_pool_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "termination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "staking_coin_weights": [
        {
          "denom": "pool93E069B333B5ECEBFE24C6E1437E814003248E0DD7FF8B9F82119F4587449BA5",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-09-01T00:00:00Z",
      "end_time": "2031-09-30T00:00:00Z",
      "epoch_ratio": "0.500000000000000000"
    }
  ],
  "add_plan_requests": [
    {
      "name": "Second Public Ratio Plan",
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
      "start_time": "2021-09-11T00:00:00Z",
      "end_time": "2031-09-30T00:00:00Z",
      "epoch_ratio": "0.500000000000000000"
    }
  ]
}
```

```bash
# Submit a public plan governance proposal
$BINARY tx gov submit-proposal public-farming-plan multiple-public-ratio-plan-proposals.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--gas 100000000 \
--yes \
--output json | jq

# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Vote
$BINARY tx gov vote 3 yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Wait for a while to pass the proposal (30s)
#

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals --output json | jq

# Query for all plans on a network
$BINARY q farming plans --output json | jq

#
# Plans are updated!
#

# Increase epoch by 1
$BINARY tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query rewards
# Rewards should be empty because user2's staking coin is not defined
# in staking coin weights any more
$BINARY q farming rewards cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Stake another pool coin that is defined in the second plan
$BINARY tx farming stake 5000000pool3036F43CB8131A1A63D2B3D3B11E9CF6FA2A2B6FEC17D5AD283C25C939614A8C \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query what user2 is staking
$BINARY q farming stakings cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Increase epoch by 2 again to distribute rewards
$BINARY tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query rewards
$BINARY q farming rewards cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq
```

## Conclusion

Throughout the demo, you learn what the budget and farming modules are and how they work to provide eseential functionalities in the Cosmos ecosystem. Thank you for taking your time and your interest. 
