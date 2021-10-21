# MVP Demo (Legacy)

⚠ **This MVP demo is written before [F1 Distribution](https://github.com/tendermint/farming/issues/91) is implemented in farming module** ⚠

## Disclaimer

Farming module is in active development by the Gravity DEX team in Tendermint. Today's demo is MVP version meaning that it is not stable version. There are a lot of rooms for improvement at this stage and there may be breaking changes along the way. Also, it can have unexpected issues or bugs. We welcome any outside contributors to contribute to the module and we are happy to receive any feedback or suggestion.

## What does farming module do?

> A farming module is a Cosmos SDK based module that implements farming functionality that provides farming rewards to staking participants called farmers. One use case is to use the module to provide incentives for liquidity pool investors for their pool participation.

## Resources

- [Github Repo](https://github.com/tendermint/farming)
- [Spec Docs](https://github.com/tendermint/farming/blob/master/x/farming/spec/01_concepts.md)
- MVP branch uses [local-testing](https://github.com/tendermint/farming/tree/local-testing) 
- Other useful resources are available in [docs](https://github.com/tendermint/farming/blob/master/docs) folder
- [Swagger Docs v0.1.0](https://app.swaggerhub.com/apis-docs/gravity-devs/farming/0.1.0)

## Build

```bash
# Clone the demo project 
git clone -b local-testing https://github.com/tendermint/farming.git
cd farming
make install
```

## Explanation about the design

### Farming Plan Types

There are two different farming plan types. One is public and the other one is private. Public farming plan can only be created through on-chain governance proposal meaning that the proposal must be first submitted, agreed, and passed in order for the plan to be created. It allows three different actions, which are adding a new plan, updating the plan, and deleting the plan. 

Unlike public farming plan, private farming plan can be created by anyone with some creation fee. As you can see, there are five different messages. You can create fixed amount plan, ratio plan, stake your coins to participate farming, unstake coins, and harvest your farming rewards. Lastly, there is custom message called MsgAdvanceEpoch and this is customized message just for today's demo. I'll explain more about how to use these messages during the demo. 

1. PublicPlan: can only be created by governance proposal.
    - PublicPlanProposal (A single message accepts three different objects)
        - AddRequestProposal
        - UpdateRequestProposal
        - DeleteRequestProposal
2. PrivatePlan: can be created by anyone (account).
    - MsgCreateFixedAmountPlan
    - MsgCreateRatioPlan

Common Messages

    - MsgStake
    - MsgUnstake
    - MsgHarvest
    - MsgAdvanceEpoch (Custom message for the demo)

## Demo

### Step 1. Bootstrap

First step is to bootstrap the local network by using the prepared commands below. To explain about the commands briefly, there are three different accounts; one is validator and the other two are normal accounts. Initialize the chain with a single validator and modify the parameters for the governance proposal. After running the commands, you will have a running chain in your local machine.

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

# Check OS for sed -i option value
export SED_I=""
if [[ "$OSTYPE" == "darwin"* ]]; then 
    export SED_I="''"
fi 

# Modify app.toml
sed -i $SED_I 's/enable = false/enable = true/g' $HOME_FARMINGAPP/config/app.toml
sed -i $SED_I 's/swagger = false/swagger = true/g' $HOME_FARMINGAPP/config/app.toml

# Modify parameters for the governance proposal
sed -i $SED_I 's%"amount": "10000000"%"amount": "1"%g' $HOME_FARMINGAPP/config/genesis.json
sed -i $SED_I 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
sed -i $SED_I 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_FARMINGAPP/config/genesis.json
sed -i $SED_I 's%"voting_period": "172800s"%"voting_period": "30s"%g' $HOME_FARMINGAPP/config/genesis.json

# Start
$BINARY start
```

### Step 2. Query Params

Let's see what values are set as farming parameters. 

There are two different types of fee parameters `private_plan_creation_fee` is the fee for private plan creating and `staking_creation_fee` is for when staking coins. These fees are there to prevent from spamming attacks.

The `epoch_days` is for reward distribution. It is measured in days. So, EndBlocker on the first block after every UTC 00:00 calculates rewards and distributes to global reward pool address. Farmers should harvest their rewards from the pool.

The last parameter `farming_fee_collector` is the module account that collects fees within the farming module. It is a global parameter because the module can be used by any Cosmos SDK based chains and chains can update the address for their need.

```bash
# Query params
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

### Step 3. Create a public fixed amount plan

PublicPlanProposal command-line interface accepts the following actions:

`add-public-fixed-plan-proposal.json`

> This plan intends to create a public fixed amount plan that provides `100ATOMs` as incentives for those who stake the denoms defines in `staking_coin_weights` field

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
          "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
          "amount": "0.800000000000000000"
        },
        {
          "denom": "uatom",
          "amount": "0.200000000000000000"
        }
      ],
      "start_time": "2021-08-04T09:00:00Z",
      "end_time": "2021-08-13T09:00:00Z",
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

`update-public-fixed-plan-proposal.json`

> This plan intends to update the plan with the following fields

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
          "denom": "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-08-06T09:00:00Z",
      "end_time": "2021-08-13T09:00:00Z",
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

`delete-public-fixed-plan-proposal.json`

> This plan intends to delete the plan

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
# Create public fixed amount plan through governance proposal
farmingd tx gov submit-proposal public-farming-plan public-fixed-plan-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--deposit 100000stake \
--broadcast-mode block \
--yes

# Query proposals
farmingd q gov proposals --output json | jq

# Vote
farmingd tx gov vote 1 yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes

# Wait for a while to pass the proposal (30s)
# Query for all plans on a network
farmingd q farming plans --output json | jq
```

### Step 4. Stake coins

Let's stake some pool coin with `user2` account because it is the one that has some pool coin (pool investor). 

```bash
# Stake pool coin
farmingd tx farming stake 5000000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes

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

### Step 5. (Custom Message) Send AdvanceEpoch for Reward Distribution

`epoch_days` is measured in days. To simulate reward distribution for this demo, we prepared the custom message to manually increase `epoch_days` by 1.

```bash
# Custom message to increase epoch_days by 1 to simulate reward distribution
farmingd tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes

# Send again to increase epoch_days by 1
farmingd tx farming advance-epoch \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes

# Query rewards
farmingd q farming rewards --output json | jq
```

### Step 6. Harvest farming rewards

```bash
# Query user2 balance
farmingd q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Harvest farming rewards from the farming plan with the staking coin
# Note that epoch_days are meausred in days so you will confront a log stating that 
# there is no reward for staking coin denom
farmingd tx farming harvest poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes
```

### Step 7. Unstake coins

This step is just to provide demonstration that you can unstake your coins. 

```bash
# Unstake coins from the farming plan
farmingd tx farming unstake 2500000poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes

# Query for all stakings on a network
farmingd q farming stakings --output json | jq
```

