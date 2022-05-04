---
Title: Claim
Description: A high-level overview of how the command-line interfaces (CLI) works for the claim module.
---

# Claim Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `claim` module. To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it. Run this command under the project root directory `$ ignite chain serve -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

## Command Line Interfaces

- [Transaction](#Transaction)
    * [Claim](#Claim)
- [Query](#Query)
    * [Airdrops](#Airdrops)
    * [Airdrop](#Airdrop)
    * [ClaimRecord](#ClaimRecord)

# Transaction

## Claim

Claim your claimable amount with a condition type.

Before claiming your claimable amount with certain condition, that condition must be met in previous.

Usage 

```bash
crescentd tx claim claim [airdrop-id] [condition-type]
```

| **Argument**      |  **Description**                                            |
| :---------------- | :---------------------------------------------------------- |
| airdrop-id        | airdrop id                                                  | 
| condition-type    | condition (task) type; deposit, swap, liquidstake, and vote |

Example

```bash
# Claim a claimable amount with the liquidity deposit condition
crescentd tx claim claim 1 deposit \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Claim a claimable amount with the liquidity swap condition
crescentd tx claim claim 1 swap \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Claim a claimable amount with the liquidstaking stake condition
crescentd tx claim claim 1 liquidstake \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq

# Claim a claimable amount with the gov vote condition
crescentd tx claim claim 1 vote \
--chain-id localnet \
--from alice \
--keyring-backend test \
--broadcast-mode block \
--yes \
--output json | jq
```

# Query

## Airdrops

Query for all airdrops 

Usage 

```bash
crescentd query claim airdrops
```

Example

```bash
crescentd query claim airdrops -o json | jq
```

## Airdrop

Query details for the particular airdrop

Usage 

```bash
crescentd query claim airdrop [airdrop-id]
```

Example

```bash
crescentd query claim airdrop 1 -o json | jq
```

## ClaimRecord

Query the claim record for an account

Usage 

```bash
crescentd query claim claim-record [airdrop-id] [address]
```

Example

```bash
crescentd query claim claim-record 1 cre1zaavvzxez0elundtn32qnk9lkm8kmcszxclz6p
```