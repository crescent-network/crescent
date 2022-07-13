---
Title: Farming
Description: A high-level overview of how the command-line interfaces (CLI) works for the farming module.
---

# Farming Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `farming` module. To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it. Run this command under the project root directory `$ ignite chain serve -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
    * [CreateFixedAmountPlan](#CreateFixedAmountPlan)
    * [CreateRatioPlan](#CreateRatioPlan)
    * [Stake](#Stake)
    * [Unstake](#Unstake)
    * [Harvest](#Harvest)
    * [RemovePlan](#RemovePlan)
- [Query](#Query)
    * [Params](#Params)
    * [Plans](#Plans)
    * [Plan](#Plan)
    * [Position](#Position)
    * [Stakings](#Stakings)
    * [QueuedStakings](#QueuedStakings)
    * [TotalStakings](#TotalStakings)
    * [Rewards](#Rewards)
    * [UnharvestedRewards](#UnharvestedRewards)
    * [CurrentEpochDays](#CurrentEpochDays)
    * [HistoricalRewards](#HistoricalRewards)

# Transaction

<!-- markdown-link-check-disable -->
++ https://github.com/crescent-network/crescent/blob/main/proto/crescent/farming/v1beta1/tx.proto

## CreateFixedAmountPlan

Anyone can create a private plan by paying a fee. A fixed amount plan plans to distribute amount of coins by a fixed amount defined in `EpochAmount`. Internally, `PrivatePlanFarmingPoolAddress` is generated and assigned to the plan. The creator queries the plan and sends amount of coins to the farming pool address so that the plan distributes as intended. To prevent spamming attacks, a `PlanCreationFee` fee must be paid on plan creation.

Create a `private-fixed-plan.json` file. This private fixed amount farming plan intends to provide 100ATOM per epoch (measured in day), relative to the rate amount of denoms that is defined in staking coin weights.

Usage

```bash
create-private-fixed-plan [plan-file]
```

- `name`: the name of the farming plan can be any name to store in a blockchain network, duplicate values are allowed
- `staking_coin_weights`: the distributing amount for each epoch. An amount must be decimal, not an integer. The sum of total weight must be 1.000000000000000000
- `start_time`: start time of the farming plan 
- `end_time`: end time of the farming plan
- `epoch_amount`: the amount to distribute per epoch as an incentive for staking denoms that are defined in the staking coin weights

Example JSON:

```json
{
  "name": "This plan intends to provide incentives for liquidity pool investors and ATOM holders",
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

Example

```bash
# Create a private fixed amount plan
crescentd tx farming create-private-fixed-plan private-fixed-plan.json \
--chain-id localnet \
--from user1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

## CreateRatioPlan

***This message is disabled by default, you have to build the binary with `make install-testing` to activate this message.***

Anyone can create this private plan type message. A ratio plan plans to distribute amount of coins by ratio that is defined in `EpochRatio`. Internally, `PrivatePlanFarmingPoolAddress` is generated and assigned to the plan. The creator must query the plan and send amount of coins to the farming pool address so that the plan distributes as intended. For a ratio plan, whichever coins that the farming pool address has in balances are used every epoch. To prevent spamming attacks, a `PlanCreationFee` fee must be paid on plan creation.

Create the `private-ratio-plan.json` file. This private ratio farming plan intends to provide ratio of all coins that farming pool address has per epoch (measured in day). In this example, epoch ratio is 10 percent and 10 percent of all the coins that the creator of this plan has in balances are used as incentives for the denoms that are defined in the staking coin weights.

Usage

```bash
create-private-ratio-plan [plan-file]
```

- `name`: the name of the farming plan can be any name to store in a blockchain network, duplicate values are allowed
- `staking_coin_weights`: the distributing amount for each epoch. An amount must be decimal, not an integer. The sum of total weight must be 1.000000000000000000
- `start_time`: start time of the farming plan 
- `end_time`: end time of the farming plan
- `epoch_ratio`: a ratio to distribute per epoch as an incentive for staking denoms that are defined in staking coin weights. The ratio refers to all coins that the creator has in their account. Note that the total ratio cannot exceed 1.0 (100%). 

Example JSON

```json
{
  "name": "This plan intends to provide incentives for Cosmonauts!",
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
  "start_time": "2021-08-06T09:00:00Z",
  "end_time": "2021-08-13T09:00:00Z",
  "epoch_ratio": "0.100000000000000000"
}
```

Example

```bash
# Create a private ratio plan
crescentd tx farming create-private-ratio-plan private-ratio-plan.json \
--chain-id localnet \
--from val1 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

## Stake

Stake coins to receive farming rewards.

Usage

```bash
stake [amount]
```

Example

```bash
# Stake pool coin
crescentd tx farming stake 5000000pool1 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

## Unstake

Unstake coins from the network.

Usage

```bash
unstake [amount]
```

Example

```bash
# Unstake coins from the farming plan
crescentd tx farming unstake 2500000pool1 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

## Harvest

Harvest farming rewards.

Usage

```bash
harvest [staking-coin-denoms]
```

Example

```bash
# Harvest farming rewards from the farming plan
# Note that there won't be any rewards if the time hasn't passed by the epoch days
crescentd tx farming harvest uatom \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Harvest all with --all flag
crescentd tx farming harvest \
--all \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

## RemovePlan

Remove farming plan.

Usage

```bash
remove-plan [plan-id]
```

Example

```bash
crescentd tx farming remove-plan 1 \
--chain-id localnet \
--from user2 \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```


# Query

<!-- markdown-link-check-disable -->
++ https://github.com/crescent-network/crescent/blob/main/proto/crescent/farming/v1beta1/query.proto

## Params 

Usage

```bash
params
```

Example

```bash
# Query the values set as farming parameters
crescentd q farming params --output json | jq
```

## Plans 

Usage

```bash
plans [optional flags]
```

Example

```bash
# Query for all farmings plans on a network
crescentd q farming plans --output json | jq

# Query for all farmings plans with the given plan type
# plan type must be either PLAN_TYPE_PUBLIC or PLAN_TYPE_PRIVATE
crescentd q farming plans \
--plan-type PLAN_TYPE_PUBLIC \
--output json | jq

# Query for all farmings plans with the given farming pool address
crescentd q farming plans \
--farming-pool-addr cre13w4ueuk80d3kmwk7ntlhp84fk0arlm3myput62 \
--output json | jq

# Query for all farmings plans with the given termination address
crescentd q farming plans \
--termination-addr cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--output json | jq

# Query for all farmings plans with the given staking coin denom
crescentd q farming plans \
--staking-coin-denom pool1 \
--output json | jq
```

## Plan 

Usage

```bash
plan [plan-id]
```

Example

```bash
# Query plan with the given plan id
crescentd q farming plan 1 --output json | jq
```

## Position

Usage

```bash
position [farmer]
```

Example

```bash
# Query for farming position of a farmer
crescentd q farming position cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query for farming position of a farmer with the given staking coin denom
crescentd q farming position cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--staking-coin-denom pool1 \
--output json | jq
```

## Stakings

Usage

```bash
stakings [farmer]
```

Example

```bash
# Query for all stakings by a farmer
crescentd q farming stakings cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query for all stakings by a farmer with the given staking coin denom
crescentd q farming stakings cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--staking-coin-denom pool1 \
--output json | jq
```

## QueuedStakings

Usage

```bash
queued-stakings [farmer]
```

Example

```bash
# Query for all queued stakings by a farmer
crescentd q farming queued-stakings cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query for all queued stakings by a farmer with the given staking coin denom
crescentd q farming queued-stakings cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--staking-coin-denom pool1 \
--output json | jq
```

## TotalStakings

Usage

```bash
total-stakings [staking-coin-denom]
```

Example

```bash
# Query for total stakings by a staking coin denom 
crescentd q farming total-stakings pool1 --output json | jq
```

### Rewards

Usage

```bash
rewards [farmer]
```

Example

```bash
# Query for all rewards by a farmer 
crescentd q farming rewards cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query for all rewards by a farmer with the staking coin denom
crescentd q farming rewards cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--staking-coin-denom pool1 \
--output json | jq
```

### UnharvestedRewards

Usage

```bash
unharvested-rewards [farmer]
```

Example

```bash
# Query for unharvested rewards for a farmer
crescentd q farming unharvested-rewards cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf --output json | jq

# Query for unharvested rewards for a farmer with the staking coin denom
crescentd q farming unharvested-rewards cre185fflsvwrz0cx46w6qada7mdy92m6kx4vg42xf \
--staking-coin-denom pool1 \
--output json | jq
```

## CurrentEpochDays 

Usage

```bash
current-epoch-days
```

Example

```bash
# Query for the current epoch days
crescentd q farming current-epoch-days --output json | jq
```

## HistoricalRewards

Usage

```bash
historical-rewards [staking-coin-denom]
```

Example

```bash
# Query for historical rewards for pool1
crescentd q farming staking-coin-denom pool1 --output json | jq
```
