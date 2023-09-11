<!-- order: 2 -->

# State

## Pool

* LastPoolId: `0x40 -> BigEndian(LastPoolId)`
* Pool: `0x42 | BigEndian(PoolId) -> ProtocolBuffer(Pool)`
* PoolState: `0x43 | BigEndian(PoolId) -> ProtocolBuffer(PoolState)`
* PoolByReserveAddressIndex: `0x44 | AddrLen (1 byte) | ReserveAddress -> BigEndian(PoolId)`
* PoolByMarketIndexKeyPrefix: `0x45 | BigEndian(MarketId) -> BigEndian(PoolId)`

```go
type Pool struct {
    Id               uint64
    MarketId         uint64
    Denom0           string
    Denom1           string
    ReserveAddress   string
    RewardsPool      string
    TickSpacing      uint32
    MinOrderQuantity sdk.Dec
    MinOrderQuote    sdk.Dec
}

type PoolState struct {
    CurrentTick                int32
    CurrentPrice               sdk.Dec
    CurrentLiquidity           sdk.Int
    TotalLiquidity             sdk.Int
    FeeGrowthGlobal            sdk.DecCoins
    FarmingRewardsGrowthGlobal sdk.DecCoins
}
```

## Position

* LastPositionId: `0x41 -> BigEndian(LastPositionId)`
* Position: `0x46 | BigEndian(PositionId) -> ProtocoulBuffer(Position)`
* PositionByParamsIndex: `0x47 | AddrLen (1 byte) | Owner | BigEndian(PoolId) | Sign (1 byte) | BigEndian(LowerTick) | Sign (1 byte) | BigEndian(UpperTick) -> BigEndian(PositionId)`
* PositionsByPoolIndex: `0x48 | BigEndian(PoolId) | BigEndian(PositionId) -> nil`

```go
type Position struct {
    Id                             uint64
    PoolId                         uint64
    Owner                          string
    LowerTick                      int32
    UpperTick                      int32
    Liquidity                      sdk.Int
    LastFeeGrowthInside            sdk.DecCoins
    OwedFee                        sdk.Coins
    LastFarmingRewardsGrowthInside sdk.DecCoins
    OwedFarmingRewards             sdk.Coins
}
```

## TickInfo

* TickInfo: `0x49 | BigEndian(PoolId) | Sign (1 byte) | BigEndian(Tick) -> ProtocolBuffer(TickInfo)`

```go
type TickInfo struct {
    GrossLiquidity              sdk.Int
    NetLiquidity                sdk.Int
    FeeGrowthOutside            sdk.DecCoins
    FarmingRewardsGrowthOutside sdk.DecCoins
}
```

## FarmingPlan

* LastFarmingPlanId: `0x4a -> BigEndian(LastFarmingPlanId)`
* FarmingPlan: `0x4b | BigEndian(FarmingPlanId) -> ProtocolBuffer(FarmingPlan)`
* NumPrivateFarmingPlans: `0x4c -> BigEndian(NumPrivateFarmingPlans)`

```go
type FarmingPlan struct {
    Id                 uint64
    Description        string
    FarmingPoolAddress string
    TerminationAddress string
    RewardAllocations  []FarmingRewardAllocation
    StartTime          time.Time
    EndTime            time.Time
    IsPrivate          bool
    IsTerminated       bool
}

type FarmingRewardAllocation struct {
    PoolId        uint64
    RewardsPerDay sdk.Coins
}
```
