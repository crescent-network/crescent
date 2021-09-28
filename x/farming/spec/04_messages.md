<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps farming module messages from transactions.

## MsgCreateFixedAmountPlan

This is one of the private plan type messages that anyone can create. A fixed amount plan plans to distribute amount of coins by a fixed amount defined in `EpochAmount`. Internally, `PrivatePlanFarmingPoolAddress` is generated and assigned to the plan and the creator should query the plan and send amount of coins to the farming pool address so that the plan distributes as intended. Note that there is a fee `PlanCreationFee` paid upon plan creation to prevent from spamming attack.

```go
type MsgCreateFixedAmountPlan struct {
	Name               string       // name for the plan for display
	Creator            string       // bech32-encoded address of the creator for the private plan
	StakingCoinWeights sdk.DecCoins // staking coin weights for the plan
	StartTime          time.Time    // start time of the plan
	EndTime            time.Time    // end time of the plan
	EpochAmount        sdk.Coins    // distributing amount for every epoch
}
```

## MsgCreateRatioPlan

This is one of the private plan type messages that anyone can create. A ratio plan plans to distribute amount of coins by ratio defined in `EpochRatio`. Internally, `PrivatePlanFarmingPoolAddress` is generated and assigned to the plan and the creator should query the plan and send amount of coins to the farming pool address so that the plan distributes as intended. For a ratio plan, whichever coins that the farming pool address has in balances are used every epoch. Note that there is a fee `PlanCreationFee` paid upon plan creation to prevent from spamming attack.

```go
type MsgCreateRatioPlan struct {
	Name               string       // name for the plan for display
	Creator            string       // bech32-encoded address of the creator for the private plan
	StakingCoinWeights sdk.DecCoins // staking coin weights for the plan
	StartTime          time.Time    // start time of the plan
	EndTime            time.Time    // end time of the plan
	EpochRatio         sdk.Dec      // distributing amount by ratio
}
```

## MsgStake

A farmer must have sufficient amount of coins to stake. If a farmer stakes coin(s) that are defined in staking coin weights of plans, then the farmer becomes eligible to receive rewards.

```go
type MsgStake struct {
	Farmer       string    // bech32-encoded address of the farmer
	StakingCoins sdk.Coins // amount of coins to stake
}
```

## MsgUnstake

A farmer must have some staking coins to trigger this message. Unlike Cosmos SDK's [staking](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/spec/01_state.md) module, there is no concept of unbonding period that requires some time to unstake coins. All the accumulated farming rewards are automatically withdrawn to the farmer once unstaking event is triggered.

```go
type MsgUnstake struct {
    Farmer         string    // bech32-encoded address of the farmer
    UnstakingCoins sdk.Coins // amount of coins to unstake
}
```

## MsgHarvest

The farming rewards are automatically accumulated, but they are not automatically distributed. A farmer should harvest their farming rewards. This mechanism is similar with Cosmos SDK's [distribution](https://github.com/cosmos/cosmos-sdk/blob/master/x/distribution/spec/01_concepts.md) module.

```go
type MsgHarvest struct {
    Farmer            string   // bech32-encoded address of the farmer
    StakingCoinDenoms []string // staking coin denoms that the farmer has staked
}
```
