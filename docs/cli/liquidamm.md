---
Title: LiquidAMM
Description: A high-level overview of how the command-line interfaces (CLI) works for the liquidamm module.
---

# LiquidAMM Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidamm` module.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [MintShare](#MintShare)
  - [BurnShare](#BurnShare)
  - [PlaceBid](#PlaceBid)
- [Query](#Query)
  - [Params](#Params)
  - [PublicPositions](#PublicPositions)
  - [PublicPosition](#PublicPosition)
  - [RewardsAuctions](#RewardsAuctions)
  - [RewardsAuction](#RewardsAuction)
  - [Bids](#Bids)
  - [Rewards](#Rewards)

# Transaction

## MintShare

Mint public position share for auto compounding rewards. The module mints the corresponding amount of `sbCoin` and sends it to the sender when the execution is complete.

Usage

```bash
mint-share [public-position-id] [desired-amount]
```

| **Argument**       | **Description**                         |
|:-------------------|:----------------------------------------|
| public-position-id | public position id                      |
| desired-amount     | deposit amounts of base and quote coins |

Example

```bash
# In order to fully test the module in your local network, public positions must be set up by governance proposal.
#
# For example, 

crescentd tx gov submit-proposal public-position-create proposal.json --chain-id localnet --from alice

Where proposal.json contains:
{
  "title": "Public Position Create Proposal",
  "description": "Let's start new liquid amm",
  "pool_id": "1",
  "lower_price": "4.5",
  "upper_price": "5.5",
  "fee_rate": "0.003"
}

# mint share
crescentd tx liquidamm mint-share 1 100000000uatom,500000000uusd \
--chain-id localnet \
--from alice

#
# Tips
#
# Query all the registered public position objects
crescentd q liquidamm public-positions -o json | jq
#
# Query account balances to see if Alice has sb coin.
crescentd q bank balances cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## BurnShare

Burn liquid amm share to withdraw underlying tokens.

Usage

```bash
burn-share [public-position-id] [share]
```

| **Argument**       | **Description**                 |
|:-------------------|:--------------------------------|
| public-position-id | public position id              |
| share              | desired amount of burning share |

Example

```bash
crescentd tx liquidamm burn-share 1 10000000000sb1 \
--chain-id localnet \
--from alice \
#
# Tips
#
# Query account balances
crescentd q bank balances cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

## PlaceBid

Place a bid for a rewards auction. Bidders estimate how much rewards for the next epoch will be accumulated and place their bids accordingly with sb coin amount.

Usage

```bash
place-bid [public-position-id] [auction-id] [share]
```

| **Argument**       | **Description**        |
|:-------------------|:-----------------------|
| public-position-id | public position id     |
| auction-id         | auction id             |
| share              | bid amount for auction |

Example

```bash
crescentd tx liquidamm place-bid 1 1 10000000sb1 \
--chain-id localnet \
--from alice \

#
# Tips
#
crescentd q liquidamm bids 1 -o json | jq
```


# Query

## Params

Query the current liquidamm parameters information.

Usage

```bash
params
```

Example

```bash
crescentd query liquidamm params -o json | jq
```

## PublicPositions

Query for all public positions.

Usage

```bash
public-positions
```

Example

```bash
crescentd query liquidamm public-positions -o json | jq
```

## PublicPosition

Query the specific public position with id.

Usage

```bash
public-position [public-position-id]
```

Example

```bash
crescentd query liquidamm public-position 1 -o json | jq
```

## RewardsAuctions

Query all rewards auctions for specific public position.

Usage

```bash
rewards-auctions [public-position-id]
```

Example

```bash
# The "rewards_auction_duration" param is the duration that is used to create new rewards auction in begin blocker.
# You can adjust the value in config-test.yml file to make it faster or slower.
# By default, the value is set to 8 hours but for local testing purpose it is set to 120 seconds.
# If you wait 120 seconds (2 minutes) after starting a local network, the module automatically creates new rewards auction.
crescentd query liquidamm rewards-auctions 1 -o json | jq
crescentd query liquidamm rewards-auctions 1 --status AUCTION_STATUS_STARTED -o json | jq
crescentd query liquidamm rewards-auctions 1 --status AUCTION_STATUS_FINISHED -o json | jq
crescentd query liquidamm rewards-auctions 1 --status AUCTION_STATUS_SKIPPED -o json | jq
```

## RewardsAuction

Query the specific reward auction

Usage

```bash
rewards-auction [public-position-id] [auction-id]
```

Example

```bash
crescentd query liquidamm rewards-auction 1 1 -o json | jq
```

## Bids

Query all bids for the rewards auction

Usage

```bash
bids [public-position-id] [auction-id]
```

Example

```bash
crescentd query liquidamm bids 1 1 -o json | jq
```

## Rewards

Query current rewards for the particular public position

Usage

```bash
rewards [public-position-id]
```

Example

```bash
crescentd query liquidamm rewards 1 -o json | jq
```
