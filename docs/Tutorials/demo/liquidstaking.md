# Demo

## Changelog

## Demo

### Build from source

```bash
# Clone the demo project and build `squad` for testing
git clone https://github.com/cosmosquad-labs/squad.git
cd squad
make install-testing
```

### Spin up a local node

- Provide 4 different types of pool coins to `user2` in genesis
- Increase inflation rate from default to 33% for better testing environment
- Modify governance parameters to lower threshold and decrease time to reduce governance process

```bash
export BINARY=squad
export HOME_FARMINGAPP=$HOME/.squadapp
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

### Send a governance proposal to set WhitelistedValidators

Create the `liquidstaking-param-change-proposal.json` file and copy the following JSON contents into the file. Depending on what whitelisted validators you create, you can customize values of the fields.

Use the following values for the fields:

```json
{
  "title": "test",
  "description": "test",
  "changes":
  [
    {
      "subspace": "liquidstaking",
      "key": "WhitelistedValidators",
      "value":
      [
        {
          "validator_address": "cosmosvaloper13w4ueuk80d3kmwk7ntlhp84fk0arlm3m9ammr5",
          "target_weight": "10"
        }
      ]
    }
  ],
  "deposit": "10000000stake"
}
```

Now, run each command one at a time. You can copy and paste each command in the command line in your terminal:

```bash
export BINARY=squad

# Submit a param-change governance proposal
$BINARY tx gov submit-proposal param-change liquidstaking-param-change-proposal.json \
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

# Query the values set as liquidstaking parameters and liquid-validators
$BINARY q liquidstaking params --output json | jq
$BINARY q liquidstaking liquid-validators --output json | jq
```


### LiquidStake coins

```bash
# Query balance of user2
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Liquid Stake
$BINARY tx liquidstaking liquid-stake 1000000000stake \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query liquid validators
$BINARY q liquidstaking liquid-validators --output json | jq

# Query balance of user2, you can find 1000000000bstake balance
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Query delegations and rewards of liquidstaking proxy module account cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk
$BINARY q staking delegations cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk --output json | jq
$BINARY q distribution rewards cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk --output json | jq

# reward trigger is 0.001
```


### LiquidUnstake coins

```bash
# Query balance of user2
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Liquid UnStake
$BINARY tx liquidstaking liquid-unstake 500000000bstake \
--gas 400000 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query balance of user2, you can find 500000000bstake balance left
$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Query liquid validators, you can find del_shares, liquid_tokens 500000000.000000000000000000 + withdrawn and re-staked rewards
$BINARY q liquidstaking liquid-validators --output json | jq

# Query delegations and rewards of liquidstaking proxy module account cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk
$BINARY q staking delegations cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk --output json | jq
$BINARY q distribution rewards cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk --output json | jq

# Query balance of liquidstaking proxy module account
$BINARY q bank balances cosmos1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyq0xf9dk \
--output json | jq

# reward trigger is 0.001
```



### Send a governance proposal to remove WhitelistedValidators

Create the `liquidstaking-param-change-proposal-2.json` file and copy the following JSON contents into the file. Depending on what whitelisted validators you create, you can customize values of the fields.

Use the following values for the fields:

```json
{
  "title": "test",
  "description": "test",
  "changes":
  [
    {
      "subspace": "liquidstaking",
      "key": "WhitelistedValidators",
      "value":
      []
    }
  ],
  "deposit": "10000000stake"
}
```


```bash
# Submit a param-change governance proposal
$BINARY tx gov submit-proposal param-change liquidstaking-param-change-proposal-2.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Set Proposal ID for the submitted proposal
export PROPOSAL=2

# Vote
$BINARY tx gov vote $PROPOSAL yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq


# Vote of user2
$BINARY tx gov vote $PROPOSAL yes \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq


# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals $PROPOSAL --output json | jq

# Query the empty whitelist and inactivated liquid-validators
$BINARY q liquidstaking params --output json | jq
$BINARY q liquidstaking liquid-validators --output json | jq


```

# Add the whitelist validator again

```bash
$BINARY tx gov submit-proposal param-change liquidstaking-param-change-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Set Proposal ID for the submitted proposal
export PROPOSAL=3

# Vote
$BINARY tx gov vote $PROPOSAL yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query Tally before calculated TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq

{
  "yes": "50000000500499928",
  "abstain": "0",
  "no": "0",
  "no_with_veto": "0"
}

# Vote of user2
$BINARY tx gov vote $PROPOSAL no \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query Tally with TallyLiquidGov calculation, voting power is worth of btoken balance of the user  
$BINARY q gov tally $PROPOSAL --output json | jq

{
  "yes": "49999999999999576",
  "abstain": "0",
  "no": "500500352",
  "no_with_veto": "0"
}

$BINARY q bank balances cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny \
--output json | jq

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals $PROPOSAL --output json | jq

# Query the empty whitelist and inactivated liquid-validators
$BINARY q liquidstaking params --output json | jq
$BINARY q liquidstaking liquid-validators --output json | jq

```




# Add the whitelist validator again

```bash
$BINARY tx gov submit-proposal param-change liquidstaking-param-change-proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
$BINARY q gov proposals --output json | jq

# Set Proposal ID for the submitted proposal
export PROPOSAL=4


# Vote of user2(liquidstaked)
$BINARY tx gov vote $PROPOSAL no \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query Tally with TallyLiquidGov calculation, voting power is worth of btoken balance of the user
$BINARY q gov tally $PROPOSAL --output json | jq

{
  "yes": "0",
  "abstain": "0",
  "no": "500500370",
  "no_with_veto": "0"
}

# Vote with vali
$BINARY tx gov vote $PROPOSAL yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

{
  "yes": "49999999999999557",
  "abstain": "0",
  "no": "500500371",
  "no_with_veto": "0"
}

# Query Tally before calculated TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals $PROPOSAL --output json | jq

# Query the empty whitelist and inactivated liquid-validators
$BINARY q liquidstaking params --output json | jq
$BINARY q liquidstaking liquid-validators --output json | jq
```

