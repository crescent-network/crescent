<!-- order: 6 -->

# Parameters

# Parameters

The `marketmaker` module contains the following parameters:

| Key                    | Type               | Example                                                                                                                                                                                          |
|------------------------|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| IncentiveBudgetAddress | string             | cre1ddn66jv0sjpmck0ptegmhmqtn35qsg2vxyk2hn9sqf4qxtzqz3sq3qhhde                                                                                                                                   |
| DepositAmount          | string (sdk.Coins) | [{"denom":"ucre","amount":"1000000000"}]                                                                                                                                                         |
| Common                 | Common             | {"min_open_ratio":"0.500000000000000000","min_open_depth_ratio":"0.100000000000000000","max_downtime":20,"max_total_downtime":100,"min_hours":16,"min_days":22}                                  |
| IncentivePairs         | []IncentivePair    | [{"pair_id":"20","update_time":"2022-12-01T00:00:00Z","incentive_weight":"0.100000000000000000","max_spread":"0.012000000000000000","min_width":"0.002000000000000000","min_depth":"100000000"}] |

## IncentiveBudgetAddress

Address containing the funds used to distribute incentives

## DepositAmount

The amount of deposit to be applied to the market maker, which is calculated per pair and is refunded when the market maker included or rejected through the MarketMaker Proposal

## Common

Common variables used in market maker [scoring system](../../../docs/whitepapers/marketmaker/scoring.md)

```go
types Common struct {
    // Minimum ratio to maintain the tick order
    MinOpenRatio sdk.Dec
    // Minimum ratio of open amount to MinDepth
    MinOpenDepthRatio sdk.Dec
    // Maximum allowable consecutive blocks of outage
    MaxDowntime uint32
    // Maximum allowable sum of blocks in an hour
    MaxTotalDowntime uint32
    // Minimum value of LiveHour to achieve LiveDay
    MinHours uint32
    // Minimum value of LiveDay to maintain MM eligibility
    MinDays uint32
}
```

## IncentivePairs

Include the pairs that are incentive target pairs and the variables used in market maker [scoring system](../../../docs/whitepapers/marketmaker/scoring.md)

```go
type IncentivePair struct {
    // Pair id of liquidity module
    PairId uint64
    // Time the pair variables start to be applied to the scoring system
    UpdateTime time.Time
    // Incentive weights for each pair
    IncentiveWeight sdk.Dec
    // Maximum allowable spread between bid and ask
    MaxSpread sdk.Dec
    // Minimum allowable price difference of high and low on both side of orders
    MinWidth sdk.Dec
    // Minimum allowable order depth on each side
    MinDepth sdk.Int
}
```
