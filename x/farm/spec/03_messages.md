<!-- order: 3 -->o

# Messages

## MsgCreatePrivatePlan

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

```go
type MsgFarm struct {
	Farmer string
	Coin   sdk.Coin
}
```

## MsgUnfarm

```go
type MsgUnfarm struct {
	Farmer string
	Coin   sdk.Coin
}
```

## MsgHarvest

```go
type MsgHarvest struct {
	Farmer string
	Denom  string
}
```
