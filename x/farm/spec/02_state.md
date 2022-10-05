<!-- order: 2 -->

# State

## Last Block Time

To calculate elapsed time between each block, the last block's block time is
stored under unique key.

* LastBlockTime: `0xd1 -> format(LastBlockTime)`

## Plan

* LastPlanId: `0xd0 -> BigEndian(LastPlanId)`
* Plan: `0xd2 | BigEndian(PlanId) -> ProtocoulBuffer(Plan)`

```go
type Plan struct {
	Id                 uint64
	Description        string
	FarmingPoolAddress string
	TerminationAddress string
	RewardAllocations  []RewardAllocation
	StartTime          time.Time
	EndTime            time.Time
	IsPrivate          bool
	IsTerminated       bool
}

type RewardAllocation struct {
	PairId        uint64
	RewardsPerDay sdk.DecCoins
}
```

## Farm

* Farm: `0xd3 | Denom -> ProtocolBuffer(Farm)`

```go
type Farm struct {
	TotalFarmingAmount sdk.Int
	CurrentRewards     sdk.DecCoins
	OutstandingRewards sdk.DecCoins
	Period             uint64
}
```

## Position

* Position: `0xd4 | FarmerAddrLen (1 byte) | FarmerAddr | Denom -> ProtocolBuffer(Position)`

```go
type Position struct {
	Farmer              string
	Denom               string
	FarmingAmount       sdk.Int
	PreviousPeriod      uint64
	StartingBlockHeight int64
}
```

## HistoricalRewards

* HistoricalRewards: `0xd5 | DenomLen (1 byte) | Denom | BigEndian(Period) -> ProtocolBuffer(HistoricalRewards)`

```go
type HistoricalRewards struct {
	CumulativeUnitRewards sdk.DecCoins
	ReferenceCount        uint32
}
```
