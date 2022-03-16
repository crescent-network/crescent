# Demo

## Changelog

## Demo

### Build from source

```bash
# Clone the demo project and build `crescentd` for testing
git clone https://github.com/crescent-network/crescent.git
cd crescent
make install-testing
```

### Spin up a local node

- Provide 4 different types of pool coins to `user2` in genesis
- Increase inflation rate from default to 33% for better testing environment
- Modify governance parameters to lower threshold and decrease time to reduce governance process

```bash
export BINARY=crescentd
export HOME_APP=$HOME/.crescent
export CHAIN_ID=localnet
export VALIDATOR_1="struggle panic room apology luggage game screen wing want lazy famous eight robot picture wrap act uphold grab away proud music danger naive opinion"
# cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
export USER_1="guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
# cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf
export USER_2="fuel obscure melt april direct second usual hair leave hobby beef bacon solid drum used law mercy worry fat super must ritual bring faculty"
# cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a
export USER_3="smooth bike pool jealous cinnamon seat tiger team canoe almost core bag fluid garbage embrace gorilla wise door toe upon present canal myth corn"
export VALIDATOR_1_GENESIS_COINS=100000000000000000stake,10000000000uatom,10000000000uusd
export USER_1_GENESIS_COINS=10000000000stake,10000000000uatom,10000000000uusd
export USER_2_GENESIS_COINS=10000000000stake,10000000000pool1,10000000000pool2,10000000000pool3,10000000000pool4

# Bootstrap
$BINARY init $CHAIN_ID --chain-id $CHAIN_ID
echo $VALIDATOR_1 | $BINARY keys add val1 --keyring-backend test --recover
echo $USER_1 | $BINARY keys add user1 --keyring-backend test --recover
echo $USER_2 | $BINARY keys add user2 --keyring-backend test --recover
echo $USER_3 | $BINARY keys add user3 --keyring-backend test --recover
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
	sed -i 's/enable = false/enable = true/g' $HOME_APP/config/app.toml
	sed -i 's/swagger = false/swagger = true/g' $HOME_APP/config/app.toml
	sed -i 's%"amount": "10000000"%"amount": "1"%g' $HOME_APP/config/genesis.json
	sed -i 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
	sed -i 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
	sed -i 's%"voting_period": "172800s"%"voting_period": "30s"%g' $HOME_APP/config/genesis.json
  sed -i 's%"inflation": "0.130000000000000000",%"inflation": "0.330000000000000000",%g' $HOME_APP/config/genesis.json
else
	sed -i '' 's/enable = false/enable = true/g' $HOME_APP/config/app.toml
	sed -i '' 's/swagger = false/swagger = true/g' $HOME_APP/config/app.toml
	sed -i '' 's%"amount": "10000000"%"amount": "1"%g' $HOME_APP/config/genesis.json
	sed -i '' 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
	sed -i '' 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOME_APP/config/genesis.json
	sed -i '' 's%"voting_period": "172800s"%"voting_period": "30s"%g' $HOME_APP/config/genesis.json
  sed -i '' 's%"inflation": "0.130000000000000000",%"inflation": "0.330000000000000000",%g' $HOME_APP/config/genesis.json
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
          "validator_address": "crevaloper13w4ueuk80d3kmwk7ntlhp84fk0arlm3mx4uyhq",
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
export BINARY=crescentd

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
$BINARY q bank balances cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
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
$BINARY q bank balances cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--output json | jq

# Query delegations and rewards of liquidstaking proxy module account cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5
$BINARY q staking delegations cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5 --output json | jq
$BINARY q distribution rewards cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5 --output json | jq

# Query voting power of the liquid staking
$BINARY q liquidstaking voting-power cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# normal stake
$BINARY tx staking delegate crevaloper13w4ueuk80d3kmwk7ntlhp84fk0arlm3mx4uyhq 500000000stake \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query voting power of staking and liquid staking
$BINARY q liquidstaking voting-power cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query liquid staking states including net amount, mint rate
$BINARY q liquidstaking states --output json | jq
```


### LiquidUnstake coins

```bash
# Query balance of user2
$BINARY q bank balances cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
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
$BINARY q bank balances cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--output json | jq

# Query liquid validators, you can find del_shares, liquid_tokens 500000000.000000000000000000 + withdrawn and re-staked rewards + UnstakeFee (0.001)
$BINARY q liquidstaking liquid-validators --output json | jq

# Query delegations and rewards of liquidstaking proxy module account cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5
$BINARY q staking delegations cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5 --output json | jq
$BINARY q distribution rewards cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5 --output json | jq

# Query unbonding, 499500000(UnstakeFee(0.001) deducted) + rewards
$BINARY q staking unbonding-delegations cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query balance of liquidstaking proxy module account
$BINARY q bank balances cre1qf3v4kns89qg42xwqhek5cmjw9fsr0ssy7z0jwcjy2dgz6pvjnyqr4aec5 \
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
$BINARY tx gov vote $PROPOSAL no \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq


# Query TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq

{
  "yes": "49999999999999576",
  "abstain": "0",
  "no": "500500352",
  "no_with_veto": "0"
}

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


# Query Tally with TallyLiquidGov calculation, voting power is worth of btoken balance of the user  
$BINARY q gov tally $PROPOSAL --output json | jq

$BINARY q bank balances cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
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

# Query Tally before calculated TallyLiquidGov 
$BINARY q gov tally $PROPOSAL --output json | jq


{
  "yes": "49999999999999557",
  "abstain": "0",
  "no": "500500371",
  "no_with_veto": "0"
}


# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
$BINARY q gov proposals $PROPOSAL --output json | jq

# Query the empty whitelist and inactivated liquid-validators
$BINARY q liquidstaking params --output json | jq
$BINARY q liquidstaking liquid-validators --output json | jq
```



# vesting test

periods.json
```json
{
    "start_time": 1645498130,
    "periods":
    [
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        },
        {
            "coins": "1000000stake",
            "length_seconds": 60
        }
    ]
}
```

```bash
# cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a
export USER_3="smooth bike pool jealous cinnamon seat tiger team canoe almost core bag fluid garbage embrace gorilla wise door toe upon present canal myth corn"

echo $USER_3 | $BINARY keys add user3 --keyring-backend test --recover

# Create periodic vesting account
$BINARY tx vesting create-periodic-vesting-account cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a periods.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json 

# or Create continuous vesting account
$BINARY tx vesting create-vesting-account cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a 100000000stake 1700000000 \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# fail Create vesting account, already existing
$BINARY tx vesting create-vesting-account cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf 100000000stake 1700000000 \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Query vesting account
$BINARY q account cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a
$BINARY q bank balances cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a

# Send test, failed not freed balance
$BINARY tx bank send cre10n3ncmlsaqfuwsmfll8kq6hvt4x7c8czhnv69a cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf 100000000stake \
--chain-id localnet \
--from user3 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# delegate failed custom sdk need spendable balances for delegation
$BINARY tx staking delegate crevaloper13w4ueuk80d3kmwk7ntlhp84fk0arlm3mx4uyhq 100000000stake \
--chain-id localnet \
--from user3 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# fail liquid staking with locked amount
$BINARY tx liquidstaking liquid-stake 100000000stake \
--chain-id localnet \
--from user3 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```