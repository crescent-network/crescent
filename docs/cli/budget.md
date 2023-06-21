---
Title: Budget
Description: Description of governance proposal for a budget module command line interface (CLI).
---

# Budget Module

The budget module includes query commands, but does not support transaction commands. Users can query the values set as budget parameters, query budget plans, and query an address that can be used as source or destination address. 

This document describes a governance proposal for a budget module command line interface (CLI).

## Command Line Interface

- [Budget Module](#budget-module)
  - [Command Line Interface](#command-line-interface)
  - [Transaction](#transaction)
    - [Propose a Budget Plan](#propose-a-budget-plan)
  - [Query](#query)
    - [Address](#address)
    - [Params](#params)
    - [Budgets](#budgets)

## Transaction

The budget module does not support transaction commands. 

The budget module supports only query commands. The ability to query budget parameters and plans requires a CLI. 

### Propose a Budget Plan

Create a `proposal.json` file for a budget plan. 

The field values are dependent on which budget plan you plan to create. 

This example creates a budget plan that distributes partial amount of coins from the Cosmos Hub's gas fees and ATOM inflation accrued in the [FeeCollector](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/keys.go#L15) module account for Gravity DEX farming plan to GravityDEXFarmingBudget account:

```go
// cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky
sdk.AccAddress(address.Module("farming", []byte("GravityDEXFarmingBudget")))
```

- `name`: display name of the budget plan
- `description`: display description of the budget plan
- `rate`: distributing amount by ratio of the total budget source
- `source_address`: address where the source of budget comes from
- `destination_address`: address that collects budget from the source address
- `start_time`: start time of the budget plan
- `end_time`: end time of the budget plan

```json
{
  "title": "Create a Budget Plan",
  "description": "Here is an example of how to add a budget plan by using ParameterChangeProposal",
  "changes": [
    {
      "subspace": "budget",
      "key": "Budgets",
      "value": [
        {
          "name": "gravity-dex-farming-20213Q-20221Q",
          "rate": "0.300000000000000000",
          "source_address": "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
          "destination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
          "start_time": "2021-10-01T00:00:00Z",
          "end_time": "2022-04-01T00:00:00Z"
        }
      ]
    }
  ],
  "deposit": "10000000stake"
}
```

```bash
# Submit a parameter changes proposal to create a budget plan
crescentd tx gov submit-proposal param-change proposal.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes

# Query the proposal to check the status PROPOSAL_STATUS_VOTING_PERIOD
crescentd q gov proposals --output json | jq

# Vote
crescentd tx gov vote 1 yes \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes

#
# Wait a while (30s) for the proposal to pass
#

# Query the proposal again to check the status PROPOSAL_STATUS_PASSED
crescentd q gov proposals --output json | jq
 
# Query the balances of destination_address for a couple times
# the balances should increase over time as gas fees and part of ATOM inflation flow in
crescentd q bank balances cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky --output json | jq
```

## Query

### Address

```bash
# Query an address that derived can be used as source and destination
# Derived according to the given name, module name, and type

# Default flag:
# $ [--type 0] - ADDRESS_TYPE_32_BYTES of ADR 028
# $ [--module-name budget] - When B, the default module name is budget

crescentd query budget address testSourceAddr
# address: cosmos1hg0v9u92ztzecpmml26206wwtghggx0flpwn5d4qc3r6dvuanxeqs4mnk5
crescentd query budget address fee_collector --type 1
# address: cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta
crescentd query budget address GravityDEXFarmingBudget --module-name farming
# address: cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky
```

### Params 

```bash
# Query the values set as budget parameters
# Note that default params are empty. You need to submit governance proposal to create budget plan
# Reference the Transaction section in this documentation
crescentd q budget params --output json | jq
```

```json
{
  "epoch_blocks": 1,
  "budgets": [
    {
      "name": "gravity-dex-farming-20213Q-20221Q",
      "rate": "0.300000000000000000",
      "source_address": "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
      "destination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
      "start_time": "2021-10-01T00:00:00Z",
      "end_time": "2022-04-01T00:00:00Z"
    }
  ]
}
```

### Budgets

```bash
# Query all the budget plans exist in the network
crescentd q budget budgets --output json | jq
```

```json
{
  "budgets": [
    {
      "budget": {
        "name": "gravity-dex-farming-20213Q-20221Q",
        "rate": "0.300000000000000000",
        "source_address": "cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta",
        "destination_address": "cosmos1228ryjucdpdv3t87rxle0ew76a56ulvnfst0hq0sscd3nafgjpqqkcxcky",
        "start_time": "2021-10-01T00:00:00Z",
        "end_time": "2022-04-01T00:00:00Z"
      },
      "total_collected_coins": [
        {
          "denom": "stake",
          "amount": "2220"
        }
      ]
    }
  ]
}
```
