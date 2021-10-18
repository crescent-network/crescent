<!-- order: 8 -->

# Parameters

The farming module contains the following parameters:

| Key                        | Type      | Example                                                             |
| -------------------------- | --------- | ------------------------------------------------------------------- |
| PrivatePlanCreationFee     | sdk.Coins | [{"denom":"stake","amount":"100000000"}]                            |
| NextEpochDays              | uint32    | 1                                                                   |
| FarmingFeeCollector        | string    | "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x" |
| DelayedStakingGasFee       | sdk.Gas   | 60000                                                               |

## PrivatePlanCreationFee

Fee paid for to create a Private type Farming plan. This fee prevents spamming and is collected in in the community pool of the distribution module.

## NextEpochDays

`NextEpochDays` is the epoch length in number of days. Internally, the farming module uses `CurrentEpochDays` parameter to process staking and reward distribution in end blocker because using `NextEpochDays` directly will affect farming rewards allocation.

## FarmingFeeCollector

A farming fee collector is a module account address that collects farming fees, such as staking creation fee and private plan creation fee.

## DelayedStakingGasFee

Since the farming module has adopted F1 reward distribution,
changes in staked coins cause withdrawal of accrued rewards.
In addition, the farming module employs a concept of delayed staking which means
when a farmer stakes coins through `MsgStake`, staked coins are not modified immediately.
Instead, at the end of the epoch, queued staking coins becomes staked and the rewards will be withdrawn.
For this reason, we added `DelayedStakingGasFee` to impose gas fee for
the future call of `WithdrawRewards` if a farmer has any staked coins with same
denom of newly staking coin.