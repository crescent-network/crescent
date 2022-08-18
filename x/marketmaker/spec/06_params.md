<!-- order: 7 -->

# Parameters

# Parameters

The `marketmaker` module contains the following parameters:

| Key                    | Type               | Example                                                                                                                                                                                          |
|------------------------|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| IncentiveBudgetAddress | string             | cre1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sq3qhhde                                                                                                                                   |
| DepositAmount          | string (sdk.Coins) | [{"denom":"ucre","amount":"1000000000"}]                                                                                                                                                         |
| Common                 | Common             | {"open_threshold":"0.500000000000000000","abs_open_threshold":"0.100000000000000000","max_downtime":20,"max_total_downtime":100,"min_hours":16,"min_days":22}                                    |
| IncentivePairs         | []IncentivePair    | [{"pair_id":"20","update_time":"2022-12-01T00:00:00Z","incentive_weight":"0.100000000000000000","max_spread":"0.012000000000000000","min_width":"0.002000000000000000","min_depth":"100000000"}] |

## IncentiveBudgetAddress

Address containing the funds used to distribute incentives

## DepositAmount

The amount of deposit to be applied to the market maker, which is calculated per pair and is refunded when the market maker included or rejected through the MarketMaker Proposal

## Common

Common variables used in market maker scoring system

```go
types Common struct {
    // Proportion of open amount against original
    OpenThreshold sdk.Dec
    // Absolute number of open amount
    AbsOpenThreshold uint32
    // Threshold of outage
    MaxDowntime uint32
    // Sum of outage time in an hour
    MaxTotalDowntime uint32
    // Minimum live hours to be recognized as live day
    MinHours uint32
    // Minimum live days to be qualified for incentives
    MinDays uint32
}
```

## IncentivePairs

Include the pairs that are incentive target pairs and the variables used in market maker scoring system

```go
type IncentivePair struct {
    // Pair id of liquidity module
    PairId uint64
    // Time the pair variables start to be applied to the scoring system
    UpdateTime time.Time
    // Incentive weights for each pair
    IncentiveWeight sdk.Dec
    // Price difference btw low ask & high bid
    MaxSpread sdk.Dec
    // Price difference btw high & low
    MinWidth sdk.Dec
    // Total amount of order
    MinDepth sdk.Int
}
```
