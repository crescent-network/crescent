<!-- order: 4 -->

# Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps `farming` module messages from transactions.

## MsgCreateFixedAmountPlan

Anyone can create this private plan type message. 

- A fixed amount plan distributes the amount of coins by a fixed amount that is defined in `EpochAmount`. 
- Internally, the private plan's farming pool address is derived and assigned to the plan. 
- The plan's `TerminationAddress` is set to the plan creator's address.
- All the coin denoms specified in `StakingCoinWeights` and `EpochAmount` must have positive supply on chain.

The creator must query the plan and send the amount of coins to the farming pool address so that the plan distributes as intended. 

**Note:** The `PlanCreationFee` must be paid on plan creation to prevent spamming attacks. This fee is refunded when the creator removes the plan by sending `MsgRemovePlan`.

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

***This message is disabled by default, you have to build the binary with `make install-testing` to activate this message.***

Anyone can create this private plan type message. 

- A ratio plan plans to distribute amount of coins by ratio defined in `EpochRatio`.
- Internally, the private plan's farming pool address is derived and assigned to the plan.
- The plan's `TerminationAddress` is set to the plan creator's address.
- All the coin denoms specified in `StakingCoinWeights` must have positive supply on chain.

The creator must query the plan and send the amount of coins to the farming pool address so that the plan distributes as intended. 

For a ratio plan, whichever coins the farming pool address has in balances are used every epoch. 

**Note:** The `PlanCreationFee` must be paid on plan creation to prevent spamming attacks. This fee is refunded when the creator removes the plan by sending `MsgRemovePlan`.


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

A farmer must have sufficient amount of coins to stake. If a farmer stakes coin or coins that are defined in staking the coin weights of plans, then the farmer becomes eligible to receive rewards.

```go
type MsgStake struct {
	Farmer       string    // bech32-encoded address of the farmer
	StakingCoins sdk.Coins // amount of coins to stake
}
```

## MsgUnstake

A farmer must have some staking coins to trigger this message.

In contrast to the Cosmos SDK [staking](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/x/staking/spec/01_state.md) module, there is no concept of an unbonding period where some time is required to unstake coins. 

All of the accumulated farming rewards are automatically withdrawn to the farmer after an unstaking event is triggered.

```go
type MsgUnstake struct {
    Farmer         string    // bech32-encoded address of the farmer
    UnstakingCoins sdk.Coins // amount of coins to unstake
}
```

## MsgHarvest

The farming rewards are automatically accumulated, but they are not automatically distributed. 
A farmer must harvest their farming rewards. This mechanism is similar to the Cosmos SDK [distribution](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/x/distribution/spec/01_concepts.md) module.
Also, if there is `UnharvestedRewards`, unharvested rewards are also withdrawn and the object is deleted.

```go
type MsgHarvest struct {
    Farmer            string   // bech32-encoded address of the farmer
    StakingCoinDenoms []string // staking coin denoms that the farmer has staked
}
```

## MsgRemovePlan

After a private plan is terminated, the plan's creator should remove the plan by sending `MsgRemovePlan`.
By removing a plan, the plan is deleted in the store and the creator gets `PrivatePlanCreationFee` refunded.

```go
type MsgRemovePlan struct {
	Creator string // bech32-encoded address of the plan creator
	PlanId  uint64 // id of the plan that is going to be removed
}
```

## MsgAdvanceEpoch

***This message is disabled by default, you have to build the binary with `make install-testing` to activate this message.***

For testing purposes only, this custom message is used to advance epoch by 1.

When you send the `MsgAdvanceEpoch` message to the network, epoch increases by 1.

```go
type MsgAdvanceEpoch struct {
	Requester string // requester defines the bech32-encoded address of the requester
}
```