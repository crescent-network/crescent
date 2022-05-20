<!-- order: 2 -->

# State

The `farming` module keeps track of the staking and rewards states.

## Plan Interface

The plan interface exposes methods to read and write standard farming plan information. 

Note that all of these methods operate on a plan struct that confirms to the interface. In order to write the plan to the store, the plan keeper is required.

```go
// PlanI is an interface used to store plan records within state.
type PlanI interface {
    proto.Message

    GetId() uint64
    SetId(uint64) error
    
    GetName() string
    SetName(name string) error

    GetType() PlanType
    SetType(PlanType) error

    GetFarmingPoolAddress() sdk.AccAddress
    SetFarmingPoolAddress(sdk.AccAddress) error

    GetTerminationAddress() sdk.AccAddress
    SetTerminationAddress(sdk.AccAddress) error

    GetStakingCoinsWeights() sdk.DecCoins
    SetStakingCoinsWeights(sdk.DecCoins) error

    GetStartTime() time.Time
    SetStartTime(time.Time) error

    GetEndTime() time.Time
    SetEndTime(time.Time) error

    GetTerminated() bool
    SetTerminated(bool) error

    GetLastDistributionTime() *time.Time
    SetLastDistributionTime(*time.Time) error

    GetDistributedCoins() sdk.Coins
    SetDistributedCoins(sdk.Coins) error

    GetBasePlan() *BasePlan

    Validate() error
}
```

## Base Plan

A base plan is the simplest and most common plan type that just stores all requisite fields directly in a struct.

```go
// BasePlan defines a base plan type. It contains all the necessary fields
// for basic farming plan functionality. Any custom farming plan type should extend this
// type for additional functionality (e.g. fixed amount plan, ratio plan).
type BasePlan struct {
    Id                   uint64       // index of the plan
    Name                 string       // name specifies the name for the plan
    Type                 PlanType     // type of the plan; public or private
    FarmingPoolAddress   string       // bech32-encoded farming pool address
    TerminationAddress   string       // bech32-encoded termination address
    StakingCoinWeights   sdk.DecCoins // coin weights for the plan
    StartTime            time.Time    // start time of the plan
    EndTime              time.Time    // end time of the plan
    Terminated           bool         // whether the plan has terminated or not
    LastDistributionTime *time.Time   // last time a distribution happened
    DistributedCoins     sdk.Coins    // total coins distributed
}
```

```go
// FixedAmountPlan defines a fixed amount plan that fixed amount of coins are distributed for every epoch day.
type FixedAmountPlan struct {
    *BasePlan

    EpochAmount sdk.Coins // distributing amount for each epoch
}
```

```go
// RatioPlan defines a ratio plan that ratio of total coins in farming pool address is distributed for every epoch day.
type RatioPlan struct {
    *BasePlan

    EpochRatio sdk.Dec // distributing amount by ratio
}
```

## Plan Types

```go
// PlanType enumerates the valid types of a plan.
type PlanType int32

const (
    // PLAN_TYPE_UNSPECIFIED defines the default plan type.
    PlanTypeNil PlanType = 0
    // PLAN_TYPE_PUBLIC defines the public plan type.
    PlanTypePublic PlanType = 1
    // PLAN_TYPE_PRIVATE defines the private plan type.
    PlanTypePrivate PlanType = 2
)
```

The parameters of the plan state are:

- ModuleName, RouterKey, StoreKey, QuerierRoute: `farming`
- Plan: `0x11 | Id -> ProtocolBuffer(Plan)`
- GlobalPlanIdKey: `[]byte("globalPlanId") -> ProtocolBuffer(uint64)`
  - store latest plan id
- NumPrivatePlans: `[]byte("numPrivatePlans") -> ProtocolBuffer(uint32)`
- ModuleName, RouterKey, StoreKey, QuerierRoute: `farming`

## Epoch

- LastEpochTime: `[]byte("lastEpochTime") -> ProtocolBuffer(Timestamp)`

- CurrentEpochDays: `[]byte("currentEpochDays") -> ProtocolBuffer(uint32)`

## Staking

```go
// Staking defines a farmer's staking information.
type Staking struct {
    Amount        sdk.Int
    StartingEpoch uint64
}
```

The parameters of the staking state are:

- Staking: `0x21 | StakingCoinDenomLen (1 byte) | StakingCoinDenom | FarmerAddr -> ProtocolBuffer(Staking)`
- StakingIndex: `0x22 | FarmerAddrLen (1 byte) | FarmerAddr | StakingCoinDenom -> nil`

```go
type QueuedStaking struct {
    Amount sdk.Int
}
```

- QueuedStaking: `0x23 | EndTimeLen (1 byte) | sdk.FormatTimeBytes(EndTime) | StakingCoinDenomLen (1 byte) | StakingCoinDenom | FarmerAddr -> ProtocolBuffer(QueuedStaking)`
- QueuedStakingIndex: `0x24 | FarmerAddrLen (1 byte) | FarmerAddr | StakingCoinDenomLen (1 byte) | StakingCoinDenom | sdk.FormatTimeBytes(EndTime) -> nil`

```go
type TotalStakings struct {
    Amount sdk.Int
}
```

- TotalStakings: `0x25 | StakingCoinDenom -> ProtocolBuffer(TotalStakings)`

## Historical Rewards

The `HistoricalRewards` struct holds the cumulative unit rewards for each epoch that are required for the reward calculation.

```go
type HistoricalRewards struct {
    CumulativeUnitRewards sdk.DecCoins
}
```

- HistoricalRewards: `0x31 | StakingCoinDenomLen (1 byte) | StakingCoinDenom | Epoch -> ProtocolBuffer(HistoricalRewards)`
- CurrentEpoch: `0x32 | StakingCoinDenom -> ProtocolBuffer(uint64)`
  - CurrentEpoch remains unchanged after all farmers has unstaked their coins.
## Outstanding Rewards

The `OutstandingRewards` struct holds outstanding (un-withdrawn) rewards for a staking denom.

```go
type OutstandingRewards struct {
    Rewards sdk.DecCoins
}
```

- OutstandingRewards: `0x33 | StakingCoinDenom -> ProtocolBuffer(OutstandingRewards)`

## Unharvested Rewards

The `UnharvestedRewards` struct holds unharvested rewards of a farmer for a staking coin denom.
Unharvested rewards are accumulated when there was a change in staked coin amount, as a result of
the withdrawal of previous rewards.

```go
type UnharvestedRewards struct {
    Rewards          sdk.Coins
}
```

- UnharvestedRewards: `0x34 | FarmerAddrLen (1 byte) | FarmerAddr | StakingCoinDenom -> ProtocolBuffer(UnharvestedRewards)`

## Examples

An example of `FixedAmountPlan`:

```json
{
  "base_plan": {
    "id": 1,
    "name": "testPlan",
    "type": 1,
    "farming_pool_address": "cre1...",
    "termination_address": "cre1...",
    "staking_coin_weights": [
      {
        "denom": "xxx",
        "amount": "0.200000000000000000"
      },
      {
        "denom": "yyy",
        "amount": "0.300000000000000000"
      },
      {
        "denom": "zzz",
        "amount": "0.500000000000000000"
      }
    ],
    "start_time": "2021-10-01T00:00:00Z",
    "end_time": "2022-04-01T00:00:00Z",
    "terminated": false,
    "last_distribution_time": "2021-10-11T00:00:00Z",
    "distributed_coins": [
      {
        "denom": "uatom",
        "amount": "10000000"
      }
    ]
  },
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "10000000"
    }
  ]
}
```

An example of `RatioPlan`:

```json
{
  "base_plan": {
    "id": 1,
    "name": "testPlan",
    "type": 1,
    "farming_pool_address": "cre1...",
    "termination_address": "cre1...",
    "staking_coin_weights": [
      {
        "denom": "xxx",
        "amount": "0.200000000000000000"
      },
      {
        "denom": "yyy",
        "amount": "0.300000000000000000"
      },
      {
        "denom": "zzz",
        "amount": "0.500000000000000000"
      }
    ],
    "start_time": "2021-10-01T00:00:00Z",
    "end_time": "2022-04-01T00:00:00Z",
    "terminated": false,
    "last_distribution_time": null,
    "distributed_coins": []
  },
  "epoch_ratio": "0.010000000000000000"
}
```
