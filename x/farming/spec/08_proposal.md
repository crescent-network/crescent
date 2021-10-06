<!-- order: 9 -->

# Proposal

The farming module contains the following public plan governance proposal that receives one of the following requests. A request `AddRequestProposal` that creates a public farming plan, a request `UpdateRequestProposal` that updates the plan, and a request `DeleteRequestProposal` that deletes the plan. For most cases, it expects that a single request is used, but note that `PublicPlanProposal` accepts more than a single request to cover broader use cases. Also,

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

Note that when requesting `AddRequestProposal` depending on which field is passed, either `EpochAmount` or `EpochRatio`, it creates a `FixedAmountPlan` or `RatioPlan`.

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

```go
// DeleteRequestProposal details a proposal for deleting an existing public plan.
type DeleteRequestProposal struct {
	// plan_id specifies index of the farming plan
	PlanId uint64 
}
```