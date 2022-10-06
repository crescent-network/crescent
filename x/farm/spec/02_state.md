<!-- order: 2 -->

# State

## Last Block Time

To calculate elapsed time between each block, the last block's block time is
stored under the unique key.

* LastBlockTime: `0xd0 -> format(LastBlockTime)`

## Plan

`Plan` represents a farming plan.
A plan's `FarmingPoolAddress` is the source of farming rewards.
When a plan allocates rewards, the rewards are moved from the farming pool to
the global `RewardsPoolAddress`.
If a plan's farming pool doesn't have enough balances for rewards allocation,
then the plan is silently ignored for the current block.
`TerminationAddress` is the address where all balances in the farming pool of
a plan are moved to when the plan gets terminated.
`RewardAllocations` describes how a pool allocates rewards to each market pair.
A plan is active(able to allocate rewards) when the current block time is
between the plan's `StartTime` and `EndTime`.
`IsPrivate` indicates whether the plan is private(created by individuals) or
public(created through a governance proposal).

The module keeps track of the last(the most recently created) plan's
ID(`LastPlanId`) to generate a new plan's ID.
Also, the module manages a global `NumPrivatePlans` counter which has the
current number of non-terminated private plans to prevent spamming attack.
The reason we use a separate counter rather than iterating through all plans
and counting the numbers each time is obvious: not to impose too much gas cost
to a plan creator.

* LastPlanId: `0xd1 -> BigEndian(LastPlanId)`
* NumPrivatePlans: `0xd2 -> BigEndian(NumPrivatePlans)`
* Plan: `0xd3 | BigEndian(PlanId) -> ProtocoulBuffer(Plan)`

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

`Farm` holds the information about a farm, which represents a single farming
asset.
`CurrentRewards` is the farm's accumulated rewards for the current period and
becomes zero when the farm's period is incremented.
`OutstandingRewards` keeps track of un-withdrawn rewards for the farm remaining
in the `RewardsPoolAddress`.

* Farm: `0xd4 | Denom -> ProtocolBuffer(Farm)`

```go
type Farm struct {
	TotalFarmingAmount sdk.Int
	CurrentRewards     sdk.DecCoins
	OutstandingRewards sdk.DecCoins
	Period             uint64
}
```

## Position

`Position` represents a farmer's farming position for a specific farming asset.
`PreviousPeriod` is the period just before the farmer started farming, and is
for calculating farming rewards using F1 algorithm along with
`StartingBlockHeight`.
`StartingBlockHeight` is the height of the block where the farmer started
farming.

* Position: `0xd5 | FarmerAddrLen (1 byte) | FarmerAddr | Denom -> ProtocolBuffer(Position)`

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

`HistoricalRewards` holds the historical information of a farm's cumulative
unit farming rewards, which is used in F1 algorithm.
`ReferenceCount` means how many other objects are referencing the
`HistoricalRewards` object.
When `ReferenceCount` goes down to zero, the object is safely deleted.

* HistoricalRewards: `0xd6 | DenomLen (1 byte) | Denom | BigEndian(Period) -> ProtocolBuffer(HistoricalRewards)`

```go
type HistoricalRewards struct {
	CumulativeUnitRewards sdk.DecCoins
	ReferenceCount        uint32
}
```
