---
Title: LiquidFarming
Description: A high-level overview of how the command-line interfaces (CLI) works for the liquidfarming module.
---

# LiquidFarming Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidfarming` module.
To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/).
If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it.
Run this command under the project root directory `$ ignite chain serve -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [LiquidFarm](#LiquidFarm)
  - [LiquidUnfarm](#LiquidUnfarm)
  - [PlaceBid](#PlaceBid)
  - [RefundBid](#RefundBid)
- [Query](#Query)
  - [Params](#Params)
  - [LiquidFarms](#LiquidFarms)
  - [LiquidFarm](#LiquidFarm)
  - [RewardsAuctions](#RewardsAuctions)
  - [RewardsAuction](#RewardsAuction)
  - [Bids](#Bids)

# Transaction

## LiquidFarm

Farm pool coin for liquid farming. The module mints the corresponding amount of `LFCoin` and sends it to the farmer when the execution is complete.

Usage

```bash
liquid-farm [pool-id] [amount]
```

| **Argument** | **Description**                    |
| :----------- | :--------------------------------- |
| pool-id      | pool id for the liquid farm        |
| amount       | amount of pool coin to liquid farm |

Example

```bash
# Note that Alice must have some pool coin.
# Reference docs/cli/liquidity.md page to get to know how to
# create a pair, a pool, and deposit coins into the pool.
crescentd tx liquidfarming liquid-farm 1 500000000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--gas 1000000 \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# Query account balances
crescentd q bank balances cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## LiquidUnfarm

Unfarm liquid farming coin to get the corresponding pool coin in return.

Usage

```bash
liquid-unfarm [pool-id] [amount]
```

| **Argument** | **Description**             |
| :----------- | :-------------------------- |
| pool-id      | pool id for the liquid farm |
| amount       | amount of lf coin to unfarm |

Example

```bash
crescentd tx liquidfarming liquid-unfarm 1 300000000000lf1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# Query account balances
crescentd q bank balances cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceBid

Place a bid for a rewards auction. Bidders estimate how much rewards for the next epoch will be accumulated and place their bids accordingly with pool coin amount.

Usage

```bash
place-bid [auction-id] [pool-id] [amount]
```

| **Argument** | **Description**                                    |
| :----------- | :------------------------------------------------- |
| auction-id   | auction id for the liquid unfarm                   |
| pool-id      | pool id for the liquid unfarm                      |
| amount       | amount of pool coin to bid for the rewards auction |

Example

```bash
crescentd tx liquidfarming place-bid 1 1 1000000000pool1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
crescentd q liquidfarming bids 1 -o json | jq
```

## RefundBid

Refund the placed bid for the rewards auction. Bidders use this transaction message to refund their bid; however, it is important to note that if the bid is currently winning bid, it can't be refunded.

Usage

```bash
refund-bid [auction-id] [pool-id]
```

| **Argument** | **Description**                |
| :----------- | :----------------------------- |
| auction-id   | auction id for the liquid farm |
| pool-id      | pool id for the liquid farm    |

Example

```bash
crescentd tx liquidfarming refund-bid 1 1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

# Query

## Params

Query the current liquidfarming parameters information.

Usage

```bash
params
```

Example

```bash
crescentd query liquidfarming params -o json | jq
```

## LiquidFarms

Query for all liquidfarms.

Usage

```bash
liquidfarms
```

Example

```bash
crescentd query liquidfarming liquidfarms -o json | jq
```

## LiquidFarm

Query the specific liquidfarm with pool id.

Usage

```bash
liquidfarm [pool-id]
```

| **Argument** | **Description**           |
| :----------- | :------------------------ |
| pool-id      | pool id of the liquidfarm |

Example

```bash
crescentd query liquidfarming liquidfarm 1 -o json | jq
```

## RewardsAuctions

Query all rewards auctions for the liquidfarm.

Usage

```bash
rewards-auctions
```

Example

```bash
# The "rewards_auction_duration" param is the duration that is used to create new rewards auction in begin blocker.
# You can adjust the value in config-test.yml file to make it faster or slower.
# By default, the value is set to 12 hours but for local testing purpose it is set to 120 seconds.
# If you wait 120 seconds (2 minutes) after starting a local network, the module automatically creates new rewards auction.
crescentd query liquidfarming rewards-auctions 1 -o json | jq
```

## RewardsAuction

Query the specific reward auction

Usage

```bash
rewards-auction [pool-id] [auction-id]
```

| **Argument** | **Description**                               |
| :----------- | :-------------------------------------------- |
| pool-id      | pool id of the liquidfarm                     |
| auction-id   | auction id of the liquidfarm with the pool id |

Example

```bash
crescentd query liquidfarming rewards-auction 1 1 -o json | jq
```

## Bids

Query all bids for the rewards auction

Usage

```bash
bids [pool-id]
```

| **Argument** | **Description**           |
| :----------- | :------------------------ |
| pool-id      | pool id of the liquidfarm |

Example

```bash
crescentd query liquidfarming bids 1 -o json | jq
```
