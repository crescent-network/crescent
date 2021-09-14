<!-- order: 8 -->

# Parameters

The farming module contains the following parameters:

| Key                        | Type      | Example                                                             |
| -------------------------- | --------- | ------------------------------------------------------------------- |
| PrivatePlanCreationFee     | sdk.Coins | [{"denom":"stake","amount":"100000000"}]                            |
| EpochDays                  | uint32    | 1                                                                   |
| FarmingFeeCollector        | string    | "cosmos1h292smhhttwy0rl3qr4p6xsvpvxc4v05s6rxtczwq3cs6qc462mqejwy8x" |

## PrivatePlanCreationFee

Fee paid for to create a Private type Farming plan. This fee prevents spamming and is collected in in the community pool of the distribution module.

## EpochDays

The universal epoch length in number of days. Every process for staking and reward distribution is executed with this `EpochDays` frequency.

## FarmingFeeCollector

A farming fee collector is a module account address that collects farming fees, such as staking creation fee and private plan creation fee.