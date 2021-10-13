<!-- order: 9 -->

# Proposal

The farming module contains the following public plan governance proposal that receives one of the following requests. 

- `AddRequestProposal` is the request proposal that requests the module to create a public farming plan. You can either input epoch amount `EpochAmount` or epoch ratio `EpochRatio`. Depending on which value of the parameter you input, it creates the following plan type `FixedAmountPlan` or `RatioPlan`.

- `UpdateRequestProposal` is the request proposal that requests the module to update the plan. You can also update the plan type. 

- `DeleteRequestProposal` is the request proposal that requests the module to delete the plan. It sends all remaining coins in the plan's farming pool `FarmingPoolAddress` to the termination address `TerminationAddress` and mark the plan as terminated.

## PublicPlanProposal

```go
// PublicPlanProposal defines a public farming plan governance proposal that receives one of the following requests:
// A request that creates a public farming plan, a request that updates the plan, and a request that deletes the plan.
// For public plan creation, depending on which field is passed, either epoch amount or epoch ratio, it creates a fixed amount plan or ratio plan.
type PublicPlanProposal struct {
	// title specifies the title of the plan
	Title string 
	// description specifies the description of the plan
	Description string 
	// add_request_proposals specifies AddRequestProposal object
	AddRequestProposals []*AddRequestProposal 
	// update_request_proposals specifies UpdateRequestProposal object
	UpdateRequestProposals []*UpdateRequestProposal 
	// delete_request_proposals specifies DeleteRequestProposal object
	DeleteRequestProposals []*DeleteRequestProposal 
}
```

## AddRequestProposal

You can either input epoch amount `EpochAmount` or epoch ratio `EpochRatio`. Depending on which value of the parameter you input, it creates the following plan type `FixedAmountPlan` or `RatioPlan`.

```go
// AddRequestProposal details a proposal for creating a public plan.
type AddRequestProposal struct {
	// name specifies the name of the plan 
	Name string
	// farming_pool_address defines the bech32-encoded address of the farming pool
	FarmingPoolAddress string   
	// termination_address defines the bech32-encoded address that terminates plan
	// when the plan ends after the end time, the balance of farming pool address
	// is transferred to the termination address
	TerminationAddress string 
	// staking_coin_weights specifies coin weights for the plan
	StakingCoinWeights sdk.DecCoins 
	// start_time specifies the start time of the plan
	StartTime time.Time 
	// end_time specifies the end time of the plan
	EndTime time.Time 
	// epoch_amount specifies the distributing amount for each epoch
	EpochAmount sdk.Coins 
	// epoch_ratio specifies the distributing amount by ratio
	EpochRatio sdk.Dec
}
```

## UpdateRequestProposal

```go
// UpdateRequestProposal details a proposal for updating an existing public plan.
type UpdateRequestProposal struct {
	// plan_id specifies index of the farming plan
	PlanId uint64 
	// name specifies the name of the plan 
	Name string
	// farming_pool_address defines the bech32-encoded address of the farming pool
	FarmingPoolAddress string 
	// termination_address defines the bech32-encoded address that terminates plan
	// when the plan ends after the end time, the balance of farming pool address
	// is transferred to the termination address
	TerminationAddress string 
	// staking_coin_weights specifies coin weights for the plan
	StakingCoinWeights sdk.DecCoins 
	// start_time specifies the start time of the plan
	StartTime time.Time 
	// end_time specifies the end time of the plan
	EndTime time.Time 
	// epoch_amount specifies the distributing amount for each epoch
	EpochAmount sdk.Coins 
	// epoch_ratio specifies the distributing amount by ratio
	EpochRatio sdk.Dec 
}
```

## DeleteRequestProposal

The request requests the module to delete the plan and it sends all remaining coins in the plan's farming pool `FarmingPoolAddress` to the termination address `TerminationAddress` and mark the plan as terminated.

```go
// DeleteRequestProposal details a proposal for deleting an existing public plan.
type DeleteRequestProposal struct {
	// plan_id specifies index of the farming plan
	PlanId uint64 
}
```