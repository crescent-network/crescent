<!-- order: 2 -->

# State

## Pool

* LastPoolId: `0x?? -> BigEndian(LastPoolId)`
* Pool: `0x?? | BigEndian(PoolId) -> ProtocolBuffer(Pool)`
* PoolByReserveAddressIndex: `0x?? | AddrLen (1 byte) | ReserveAddress -> BigEndian(PoolId)`
* PoolByMarketIndexKeyPrefix: `0x?? | BigEndian(MarketId) -> BigEndian(PoolId)`
* PoolState: `0x?? | BigEndian(PoolId) -> ProtocolBuffer(PoolState)`

```go
type Pool struct {
    Id             uint64
    MarketId       uint64
    Denom0         string
    Denom1         string
    TickSpacing    uint32
    ReserveAddress string
}

type PoolState struct {
    CurrentTick                int32
    CurrentPrice               sdk.Dec
    CurrentLiquidity           sdk.Int
    FeeGrowthGlobal            sdk.DecCoins
    FarmingRewardsGrowthGlobal sdk.DecCoins
}
```

## Position

* LastPositionId: `0x?? -> BigEndian(LastPositionId)`
* Position: `0x?? | BigEndian(PositionId) -> ProtocoulBuffer(Position)`
* PositionByParamsIndex: `0x?? | BigEndian(PoolId) | AddrLen (1 byte) | Owner | Sign (1 byte) | BigEndian(LowerTick) | Sign (1 byte) | BigEndian(UpperTick) -> BigEndian(PositionId)`
* PositionsByPoolIndex: `0x?? | BigEndian(PoolId) | BigEndian(PositionId) -> nil`

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

* TickInfo: `0x?? | BigEndian(PoolId) | Sign (1 byte) | BigEndian(Tick) -> ProtocolBuffer(TickInfo)`

```go
type TickInfo struct {
    GrossLiquidity              sdk.Int
    NetLiquidity                sdk.Int
    FeeGrowthOutside            sdk.DecCoins
    FarmingRewardsGrowthOutside sdk.DecCoins
}
```

## FarmingPlan

* LastFarmingPlanId: `0x?? -> BigEndian(LastFarmingPlanId)`
* FarmingPlan: `0x?? | BigEndian(FarmingPlanId) -> ProtocolBuffer(FarmingPlan)`
* NumPrivateFarmingPlans: `0x?? -> BigEndian(NumPrivateFarmingPlans)`

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
