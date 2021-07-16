<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps farming module messages from transactions.

## MsgCreateFixedAmountPlan

```go
type MsgCreateFixedAmountPlan struct {
    FarmingPoolAddress  string
    StakingCoinWeights  sdk.DecCoins
    StartTime           time.Time
    EndTime             time.Time
    EpochAmount         sdk.Coins
}
```

## MsgCreateRatioPlan

```go
type MsgCreateRatioPlan struct {
    FarmingPoolAddress  string
    StakingCoinWeights  sdk.DecCoins
    StartTime           time.Time
    EndTime             time.Time
    EpochRatio          sdk.Dec
}
```

## MsgStake

A farmer must have sufficient coins to stake into a farming plan. The farmer becomes eligible to receive rewards once the farmer stakes some coins.

```go
type MsgStake struct {
    Farmer       string
    StakingCoins sdk.Coins
}
```

## MsgUnstake

A farmer must have some staking coins in the plan to trigger this message. Unlike `x/staking` module, there is no unbonding period of time required to unstake coins from the plan. All accumulated farming rewards are automatically withdrawn to the farmer once unstaking event is triggered.

```go
type MsgUnstake struct {
    Farmer         string
    UnstakingCoins sdk.Coins
}

```

## MsgHarvest

A farmer should harvest their farming rewards. The rewards are not automatically distributed. This is similar mechanism with `x/distribution` module.

```go
type MsgHarvest struct {
    Farmer             string
    StakingCoinDenom   string
}
```
