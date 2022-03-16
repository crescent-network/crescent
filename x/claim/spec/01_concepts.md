<!-- order: 1 -->

# Concepts

The `claim` module distributes claimable amount of all airdrop recipients as they complete the tasks in core network activities. Full allocation can be claimed within 6 months from launch. Unclaimed amounts from the airdrop quantity within the claim period will be allocated to the community fund. The airdrop recipients are Cosmos Hub stakers based on a snapshot taken at 2022.01.01 00:00 UTC (Block 8902586). The amount of delegation to centralized exchange validators are excluded from eligibility and there is a bonus going towards the airdrop recipients who were supportive of the `liquidity` module, Gravity DEX. 

## Claimable Amount Calculation

<!-- markdown-link-check-disable-next-line -->
Claimable amount for each user is calculated by [this airdrop-calculator program](https://github.com/crescent-network/airdrop-calculator). The program uses the following formulas to calculate the claimable score and amount.

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

The airdrop recipient are required to complete the following conditions (tasks) and there is no order of execution.

- 20% of the initial DEXdrop claimable amount is released in genesis
- 20% of the initial DEXdrop claimable amount is released by executing a deposit transaction to any pool
- 20% of the initial DEXdrop claimable amount is released by executing an order transaction in any pair
- 20% of the initial DEXdrop claimable amount is released by executing a liquid staking transaction
- 20% of the initial DEXdrop claimable amount is released by executing a governance vote transaction 

## Termination

An airdrop ends when the `EndTime` is passed over the current time. Unclaimed amounts from the airdrop quantity within the claim period will be allocated to the community fund.