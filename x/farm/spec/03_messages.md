<!-- order: 3 -->

# Messages

## MsgCreatePrivatePlan

A private farming plan can be created with `MsgCreatePrivatePlan`.
See [Plan](02_state.md#plan) for more details about the fields.

```go
type MsgCreatePrivatePlan struct {
    Creator           string
    Description       string
    RewardAllocations []RewardAllocation
    StartTime         time.Time
    EndTime           time.Time
}

type RewardAllocation struct {
    PairId        uint64
    RewardsPerDay sdk.DecCoins
}
```

## MsgFarm

Farmers can start farming on their assets with `MsgFarm`.

```go
type MsgFarm struct {
	Farmer string
	Coin   sdk.Coin
}
```

## MsgUnfarm

Farmers can withdraw their farming assets with `MsgUnfarm`.

```go
type MsgUnfarm struct {
	Farmer string
	Coin   sdk.Coin
}
```

## MsgHarvest

Farmers can withdraw their farming rewards with `MsgHarvest`.

```go
type MsgHarvest struct {
	Farmer string
	Denom  string
}
```
