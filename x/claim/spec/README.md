<!-- order: 0 title: Overview parent: title: "claim" -->

# `claim`

## Abstract

The `claim` module initializes genesis states with `Airdrops` and `ClaimRecords`. They are extracted and calculated by using this program called `airdrop-calculator`. You can find more information about the program in this link. Once the airdrop information and its claim records are stored in the network, the module distributes claimable amount of coins to each of the airdrop recipients as they perform certain condition(s). Each of the condition triggers the module to distribute a proportionate amount of coins. The first airdrop event is for the Cosmos Hub stakers who had delegated their tokens on the block height `8902586`, which is `2022.01.01 UTC 00:00`. Moreover, there is a bonus going towards them if they were supportive of the `liquidity` module. 

## Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
