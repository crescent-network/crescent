<!-- order: 7 -->

# Parameters

The `farming` module contains the following parameters:

| Key                     | Type      | Example                                                          |
|-------------------------|-----------|------------------------------------------------------------------|
| PrivatePlanCreationFee  | sdk.Coins | [{"denom":"stake","amount":"1000000000"}]                        |
| NextEpochDays           | uint32    | 1                                                                |
| FarmingFeeCollector     | string    | "cre1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mq4p6cjy" |
| DelayedStakingGasFee    | sdk.Gas   | 60000                                                            |
| MaxNumPrivatePlans      | uint32    | 10000                                                            |


## PrivatePlanCreationFee

Fee paid to create a private farming plan. This fee prevents spamming attack and is reserved in the FarmingFeeCollector. If the plan creator removes the plan, this fee will be refunded to the creator.

## NextEpochDays

`NextEpochDays` is the epoch length in number of days. Internally, the farming module uses `CurrentEpochDays` parameter to process staking and reward distribution in end-blocker because using `NextEpochDays` directly will affect farming rewards allocation.

## FarmingFeeCollector

A farming fee collector is a module account address that collects farming fees, such as staking creation fee and private plan creation fee.

## DelayedStakingGasFee

Since the farming module has adopted F1 reward distribution, changes in staked coins cause withdrawal of accrued rewards.

In addition, the farming module employs a concept of delayed staking. This means that when a farmer stakes coins through `MsgStake`, staked coins are not modified immediately. 

Instead, at the end of the epoch, queued staking coins becomes staked and the rewards are withdrawn. For this reason, the `DelayedStakingGasFee` parameter is available to impose gas fees for the future call of `WithdrawRewards` if a farmer has any staked coins with same
denom of newly staked coin.

## MaxNumPrivatePlans

The maximum number of private plans that are allowed to be created.
It does not include terminated plans.

# Global constants

There are some global constants defined in `x/farming/types/params.go`.

## PrivatePlanMaxNumDenoms

This is the maximum number of denoms in a private plan's staking coin weights and epoch amount.
It's set to `50`.

## PublicPlanMaxNumDenoms

This is the maximum number of denoms in a public plan's staking coin weights and epoch amount.
It's set to `500`.
