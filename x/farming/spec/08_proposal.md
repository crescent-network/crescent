<!-- order: 8 -->

# Proposal

The `farming` module contains the following public plan governance proposal that receives one of the following requests. 

- `AddPlanRequest` is the request proposal that requests the module to create a public farming plan. You can either input epoch amount `EpochAmount` or epoch ratio `EpochRatio`. Depending on which value of the parameter you input, it creates the following plan type `FixedAmountPlan` or `RatioPlan`.

- `ModifyPlanRequest` is the request proposal that requests the module to update the plan. You can also update the plan type. 

- `DeletePlanRequest` is the request proposal that requests the module to delete the plan. It sends all remaining coins in the plan's farming pool `FarmingPoolAddress` to the termination address `TerminationAddress` and mark the plan as terminated.

Note that adding or modifying `RatioPlan`s is disabled by default.
The binary should be built using `make install-testing` command to enable that.

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
	// add_plan_requests specifies AddPlanRequest object
	AddPlanRequests []AddPlanRequest
	// modify_plan_requests specifies ModifyPlanRequest object
	ModifyPlanRequests []ModifyPlanRequest
	// delete_plan_requests specifies DeletePlanRequest object
	DeletePlanRequests []DeletePlanRequest
}
```

## AddPlanRequest

Request the module to create a public farming plan. 

- For each request, you must specify epoch amount `EpochAmount` or epoch ratio `EpochRatio`. 
- Depending on the value, the plan type `FixedAmountPlan` or `RatioPlan` is created.

```go
// AddPlanRequest details a proposal for creating a public plan.
type AddPlanRequest struct {
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

## ModifyPlanRequest

Request the module to update the plan or the plan type.

```go
// ModifyPlanRequest details a proposal for updating an existing public plan.
type ModifyPlanRequest struct {
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

## DeletePlanRequests

Request the module to delete the plan. All remaining coins in the plan's farming pool `FarmingPoolAddress` are sent to the termination address `TerminationAddress` and the plan is marked as terminated.

```go
// DeletePlanRequests details a proposal for deleting an existing public plan.
type DeletePlanRequests struct {
	// plan_id specifies index of the farming plan
	PlanId uint64 
}
```