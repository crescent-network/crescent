<!-- order: 1 -->

# Concepts

The `claim` module distributes claimable amount of all airdrop recipients when they complete participating in core network activities. The airdrop recipients are Cosmos Hub stakers based on 8902586 (2022.01.01 UTC 00:00) block height and there is a bonus going towards the airdrop recipients who were supportive of the `liquidity` module. 

## Claimable Amount Calculation

<!-- markdown-link-check-disable-next-line -->
Claimable amount for each user is calculated by [this airdrop-calculator program](https://github.com/cosmosquad-labs/airdrop-calculator). The program uses the following formulas to calculate the claimable score and amount.

```
// Sqaure root of the delegation shares a user had at the snapshot date and multiply by 2 for each criteria
Claimable Score = SQRT(DelegationAmount) * M1(x2) * M2(x2) * M3(x2)

// Total airdrop amount multiply by a user's share of the claimable score
Claimable Amount = int(TotalAirdropAmount * (ClaimableScore / TotalClaimableScore))
```

Each of the following criteria is a multiplier of 2 in calculating the claimable score.

- M1: when a user has ever deposited coins in Cosmos Hub by block height 8902586
- M2: when a user has ever swapped coins in Cosmos Hub by block height 8902586
- M3: when a user has voted governance proposal either [38](https://www.mintscan.io/cosmos/proposals/38) or [58](https://www.mintscan.io/cosmos/proposals/58)

## Core Network Activities

The airdrop recipient are required to complete the following Conditions. There is no order of execution. The recipients can execute any order the prefer.

- 1/3 of the initial claimable amount is released by executing a deposit transaction to any pool
- 1/3 of the initial claimable amount is released by executing a swap transaction from any pool
- 1/3 of the initial claimable amount is released by executing a staking transaction for any farming plan


## Termination

An airdrop ends when the `EndTime` is passed over the current time. All the unclaimed airdrop coins are sent to the community pool.