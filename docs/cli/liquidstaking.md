---
Title: Liquidstaking
Description: A high-level overview of how the command-line interfaces (CLI) works for the liquidstaking module.
---

# Liquidstaking Module

## Synopsis

This document provides a high-level overview of how the command line (CLI) interface works for the `liquidstaking` module. To set up a local testing environment, it requires the latest [Ignite CLI](https://docs.ignite.com/). If you don't have Ignite CLI set up in your local machine, see [this guide](https://docs.ignite.com/guide/install.html) to install it. Run this command under the project root directory `$ ignite chain serve -c config-test.yml`.

Note that [jq](https://stedolan.github.io/jq/) is recommended to be installed as it is used to process JSON throughout the document.

- [Transaction](#Transaction)
    * [LiquidStake](#LiquidStake)
    * [LiquidUnstake](#LiquidUnstake)
- [Query](#Query)
    * [Params](#Params)
    * [LiquidValidators](#LiquidValidators)
    * [States](#States)
    * [VotingPower](#VotingPower)

# Transaction

## LiquidStake

Liquid stake coin.

It requires `whitelisted_validators` to be registered. The [config.yml](https://github.com/crescent-network/crescent/blob/main/config.yml) file registers a single whitelist validator for testing purpose. 

Usage

```bash
liquid-stake [amount]
```

| **Argument** |  **Description**                                          |
| :----------- | :-------------------------------------------------------- |
| amount       | amount of coin to liquid stake; it must be the bond denom |

Example

```bash
crescentd tx liquidstaking liquid-stake 5000000000stake \
--chain-id localnet \
--from bob \
--keyring-backend test \
--gas 1000000 \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# Query account balances
# Notice the newly minted bToken
crescentd q bank balances cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq

# Query the voter's liquid staking voting power
crescentd q liquidstaking voting-power cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq
```

## LiquidUnstake

Unstake coin.

Usage

```bash
liquid-unstake [amount]
```

| **Argument**  |  **Description**                                      |
| :------------ | :---------------------------------------------------- |
| amount        | amount of coin to unstake; it must be the bToken denom|

Example

```bash
crescentd tx liquidstaking liquid-unstake 1000000000bstake \
--chain-id localnet \
--from bob \
--keyring-backend test \
--gas 1000000 \
--broadcast-mode block \
--yes \
--output json | jq

#
# Tips
#
# Query account balances
# Notice the newly minted bToken
crescentd q bank balances cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq

# Query the voter's liquid staking voting power
crescentd q liquidstaking voting-power cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq
```

# Query

## Params

Query the current liquidstaking parameters information.

Usage

```bash
params
```

Example

```bash
crescentd query liquidstaking params -o json | jq
```

## LiquidValidators

Query all liquid validators.

Usage

```bash
liquid-validators
```

Example

```bash
crescentd query liquidstaking liquid-validators -o json | jq
```
## States

Query net amount state.

Usage

```bash
states
```

Example

```bash
crescentd query liquidstaking states -o json | jq
```

## VotingPower

Query the voter’s staking and liquid staking voting power. 

Usage

```bash
voting-power [voter]
```

| **Argument** |  **Description**      |
| :----------- | :-------------------- |
| voter        | voter account address |

Example

```bash
crescentd query liquidstaking voting-power cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3 -o json | jq
```