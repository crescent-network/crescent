---
Title: LP Farm
Description: A high-level overview of how the command-line interfaces (CLI) works for the lpfarm module.
---

# LP Farm Module

## Synopsis

This document provides a high-level overview of how the command line (CLI)
interface works for the `lpfarm` module. To set up a local testing environment, it requires the latest
[Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine,
see [this guide](https://docs.ignite.com/welcome/install) to install it. Run this command under the project root
directory
`$ ignite chain serve -v -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout
the document.

Make sure that the pairs specified in the plan's reward allocations have the
last price in order to create plans.
You can make some orders to the pairs after creating pools in the pair to
achieve this.

## Command Line Interfaces

- [Transaction](#transaction)
  - [CreatePrivatePlan](#createprivateplan)
  - [Farm](#tx-farm)
  - [Unfarm](#unfarm)
  - [Harvest](#harvest)
- [Query](#query)
  - [Params](#params)
  - [Plans](#plans)
  - [Plan](#plan)
  - [Farm](#query-farm)
  - [Positions](#positions)
  - [Position](#position)
  - [HistoricalRewards](#historicalrewards)
  - [TotalRewards](#totalrewards)
  - [Rewards](#rewards)

### Transaction

#### CreatePrivatePlan

Create a new private farming plan.
The newly created plan's farming pool address is automatically generated and
will have no balances in the account initially.
Manually send enough reward coins to the generated farming pool address to make
sure that the rewards allocation happens.
The plan's termination address is set to the plan creator.

Usage:

```bash
create-private-plan [description] [start-time] [end-time] [reward-allocations...]
```

| **Argument**          | **Description**                                      |
|:----------------------|:-----------------------------------------------------|
| description           | a brief description of the plan                      |
| start-time            | the time at which the plan begins, in RFC3339 format |
| end-time              | the time at which the plan ends, in RFC3339 format   |
| reward-allocations... | whitespace-separated list of the reward allocations  |

A reward allocation is specified in one of the following formats:
1. `<denom>:<rewards_per_day>`
2. `pair<pair-id>:<rewards_per_day>`

Note that the example below assumes that pair 1 and pair 2 exist.

Example:

```bash
crescentd tx lpfarm create-private-plan \
"New Farming Plan" \
2022-01-01T00:00:00Z \
2023-01-01T00:00:00Z \
pair1:10000stake,5000uatom pool2:5000stake \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query plans using the following command
crescentd q lpfarm plans -o json | jq
```

<h4 id="tx-farm">Farm</h4>

Start farming coin.

Usage:

```bash
farm [coin]
```

| **Argument** | **Description** |
|:-------------|:----------------|
| coin         | Coin to farm    |

Example:

```bash
crescentd tx lpfarm farm 1000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query positions of the farmer using the following command
crescentd q lpfarm positions cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

#### Unfarm

Unfarm farming coin.

Usage:

```bash
unfarm [coin]
```

| **Argument** | **Description** |
|:-------------|:----------------|
| coin         | Coin to unfarm  |

Example:

```bash
crescentd tx lpfarm unfarm 1000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query positions of the farmer using the following command
crescentd q lpfarm positions cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

#### Harvest

Harvest farming rewards.

Usage:

```bash
harvest [denom]
```

| **Argument** | **Description**                     |
|:-------------|:------------------------------------|
| denom        | Pool coin denom to withdraw rewards |

Example:

```bash
crescentd tx lpfarm harvest pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# You can query account balances using the following command
crescentd q bank balances cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

### Query

#### Params

Query the current farm parameters.

Usage:

```bash
params
```

Example:

```bash
crescentd q lpfarm params -o json | jq
```

#### Plans

Query all plans.

Usage:

```bash
plans
```

Example:

```bash
crescentd q lpfarm plans -o json | jq
crescentd q lpfarm plans --farming-pool-addr cre1gkvhlzmpxarqwk4jh7k7yemf60r50y55n8ax9kxcx8t28hm0e7cqk52jh9 -o json | jq
crescentd q lpfarm plans --termination-addr cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq
crescentd q lpfarm plans --is-private true -o json | jq
crescentd q lpfarm plans --is-terminated false -o json | jq
```

#### Plan

Query a specific plan.

Usage:

```bash
plan [plan-id]
```

Example:

```bash
crescentd q lpfarm plan 1 -o json | jq
```

<h4 id="query-farm">Farm</h4>

Query a specific farm for the denom.

Usage:

```bash
farm [denom]
```

Example:

```bash
crescentd q lpfarm farm pool1 -o json | jq
```

#### Positions

Query all the positions managed by the farmer.

Usage:

```bash
positions [farmer]
```

Example:

```bash
crescentd q lpfarm positions cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

#### Position

Query a specific position managed by the farmer.

Usage:

```bash
position [farmer] [denom]
```

Example:

```bash
crescentd q lpfarm position cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p pool1 -o json | jq
```

#### HistoricalRewards

Query all historical rewards for the denom.

Usage:

```bash
historical-rewards [denom]
```

Example:

```bash
crescentd q lpfarm historical-rewards pool1 -o json | jq
```

#### TotalRewards

Query total rewards accumulated in all farming assets of the farmer.

Usage:

```bash
total-rewards [farmer]
```

Example:

```bash
crescentd q lpfarm total-rewards cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

#### Rewards

Query rewards accumulated in a farming asset of the farmer.

Usage:

```bash
rewards [farmer] [denom]
```

Example:

```bash
crescentd q lpfarm rewards cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p pool1 -o json | jq
```
