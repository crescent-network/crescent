<!-- order: 9 -->

# Proposal

The `farming` module contains the following public plan governance proposal that receives one of the following requests. 

- `AddRequestProposal` is the request proposal that requests the module to create a public farming plan. 

- `UpdateRequestProposal` is the request proposal that requests the module to update the plan. Updating the the plan type is also supported. 

- `DeleteRequestProposal` is the request proposal that requests the module to delete the plan. 

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

Request the module to create a public farming plan. 

- For each request, you must specify epoch amount `EpochAmount` or epoch ratio `EpochRatio`. 
- Depending on the value, the plan type `FixedAmountPlan` or `RatioPlan` is created.

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

Request the module to update the plan or the plan type.

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

Request the module to delete the plan. All remaining coins in the plan's farming pool `FarmingPoolAddress` are sent to the termination address `TerminationAddress` and the plan is marked as terminated.

```go
// DeleteRequestProposal details a proposal for deleting an existing public plan.
type DeleteRequestProposal struct {
	// plan_id specifies index of the farming plan
	PlanId uint64 
}
```