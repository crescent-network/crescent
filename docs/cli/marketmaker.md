---
Title: Marketmaker
Description: A high-level overview of how the command-line interfaces (CLI) works for the marketmaker module.
---

# Marketmaker Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `marketmaker` module. To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it. Run this command under the project root directory `$ ignite chain serve -v -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
  - [ApplyMarketMaker](#ApplyMarketMaker)
  - [ClaimIncentives](#ClaimIncentives)
- [Query](#Query)
  - [Params](#Params)
  - [MarketMakers](#MarketMakers)
  - [Incentive](#Incentive)

# Transaction

## ApplyMarketMaker

Apply as a market maker with this transaction message. A market maker can apply for a single or multiple pairs, and the community will decide if they are well fit to be included in the whitelisted market makers group to become eligible.

It is important to note that not all pairs in the liquidity module are available in genesis. Each pair must be approved by the community and be registered in params.

For testing purpose, pair 1 and 2 are pre-registered in the `config-test.yml` file.

Usage

```bash
crescentd tx marketmaker apply [pool-ids]
```

| **Argument** | **Description** |
| :----------- | :-------------- |
| pool-ids     | pool id(s)      |

Example

```bash
# Apply as a market maker for pair 1
crescentd tx marketmaker apply 1 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Apply as a market maker for pair 1 and 2
crescentd tx marketmaker apply 1,2 \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#

# Query params to see which pairs are available to apply as a market maker
# For testing purpose, pair 1 and 2 are pre-registered
crescentd q marketmaker params -o json | jq

# Note that there must be governance proposal to include the applied market maker to become eligible
# Now, the returned value must be empty.
crescentd q marketmaker marketmakers -o json | jq
```

## IncludeMarketMaker

For testing purpose, create a `proposal.json` file to include the applied market maker.

```json
{
  "title": "Market Maker Proposal",
  "description": "Include the following market makers",
  "inclusions": [
    {
      "address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "pair_id": "1"
    }
  ],
  "exclusions": [],
  "rejections": [],
  "distributions": [
    {
      "address": "cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p",
      "pair_id": "1",
      "amount": [
        {
          "denom": "stake",
          "amount": "100000000"
        }
      ]
    }
  ]
}
```

Example

```bash
# Create a proposal
crescentd tx gov submit-proposal market-maker-proposal proposal.json \
--gas 400000 \
--chain-id localnet \
--from alice \
--deposit 100000000stake \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

# Vote
# For testing purpose, voting period is shortened to 10 seconds.
crescentd tx gov vote 1 yes \
--chain-id localnet \
--from alice \
--keyring-backend=test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#

# Query all proposals
crescentd q gov proposals -o json | jq

# Query all market makers to see if the applied market maker's eligible is true now
crescentd query marketmaker marketmakers -o json | jq
```

## ClaimIncentives

Usage

```bash
crescentd tx marketmaker claim
```

Example

```bash
# First, query to see if there is any incentive to claim
crescentd tx marketmaker claim \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#

crescentd q marketmaker incentive cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```

# Query

## Params

Query the values set as marketmaker parameters

Usage

```bash
params
```

Example

```bash
crescentd query marketmaker params -o json | jq
```

## MarketMakers

Query the details of market maker(s)

```bash
marketmakers [optional flags]
```

Example

```bash
# Query all market makers
crescentd query marketmaker marketmakers -o json | jq

# Query all market makers for the pair id
crescentd query marketmaker marketmakers --pair-id=1 -o json | jq

# Query specific market maker
crescentd query marketmaker marketmakers --address=cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq

# Query market makers that are eligible
crescentd query marketmaker marketmakers --eligible=true -o json | jq
```

## Incentive

Query claimable incentive of a market maker

```bash
incentive [mm-address]
```

Example

```bash
crescentd q marketmaker incentive cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p -o json | jq
```
