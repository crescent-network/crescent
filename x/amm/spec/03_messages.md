<!-- order: 3 -->

# Messages

## MsgCreatePool

```go
type MsgCreatePool struct {
    Sender   string
    MarketId uint64
    Price    sdk.Dec
}
```

## MsgAddLiquidity

```go
type MsgAddLiquidity struct {
    Sender        string
    PoolId        uint64
    LowerPrice    sdk.Dec
    UpperPrice    sdk.Dec
    DesiredAmount sdk.Coins
}
```

## MsgRemoveLiquidity

```go
type MsgRemoveLiquidity struct {
    Sender     string
    PositionId uint64
    Liquidity  sdk.Int
}
```

## MsgCollect

```go
type MsgCollect struct {
    Sender     string
    PositionId uint64
    Amount     sdk.Coins
}
```

## MsgCreatePrivateFarmingPlan

```go
type MsgCreatePrivateFarmingPlan struct {
    Sender             string
    Description        string
    TerminationAddress string
    RewardAllocations  []FarmingRewardAllocation
    StartTime          time.Time
    EndTime            time.Time
}

type FarmingRewardAllocation struct {
    PoolId        uint64
    RewardsPerDay sdk.Coins
}
```

## MsgTerminatePrivateFarmingPlan

```go
type MsgTerminatePrivateFarmingPlan struct {
    Sender        string
    FarmingPlanId uint64
}
```
