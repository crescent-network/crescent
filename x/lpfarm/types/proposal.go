package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeFarmingPlan string = "FarmingPlan"
)

var (
	_ gov.Content = &FarmingPlanProposal{}
)

func init() {
	gov.RegisterProposalType(ProposalTypeFarmingPlan)
	gov.RegisterProposalTypeCodec(&FarmingPlanProposal{}, "crescent/FarmingPlanProposal")
}

func NewFarmingPlanProposal(
	title, description string,
	createPlanReqs []CreatePlanRequest,
	terminatePlanReqs []TerminatePlanRequest) *FarmingPlanProposal {
	return &FarmingPlanProposal{
		Title:                 title,
		Description:           description,
		CreatePlanRequests:    createPlanReqs,
		TerminatePlanRequests: terminatePlanReqs,
	}
}

func (p *FarmingPlanProposal) GetTitle() string       { return p.Title }
func (p *FarmingPlanProposal) GetDescription() string { return p.Description }
func (p *FarmingPlanProposal) ProposalRoute() string  { return RouterKey }
func (p *FarmingPlanProposal) ProposalType() string   { return ProposalTypeFarmingPlan }

func (p *FarmingPlanProposal) ValidateBasic() error {
	for _, req := range p.CreatePlanRequests {
		if err := req.Validate(); err != nil {
			return err
		}
	}
	for _, req := range p.TerminatePlanRequests {
		if err := req.Validate(); err != nil {
			return err
		}
	}
	return gov.ValidateAbstract(p)
}

func (p FarmingPlanProposal) String() string {
	return fmt.Sprintf(`Farming Plan Proposal:
  Title:                 %s
  Description:           %s
  CreatePlanRequests:    %v
  TerminatePlanRequests: %v
`, p.Title, p.Description, p.CreatePlanRequests, p.TerminatePlanRequests)
}

func NewCreatePlanRequest(
	description string, farmingPoolAddr sdk.AccAddress,
	rewardAllocs []RewardAllocation, startTime, endTime time.Time) CreatePlanRequest {
	return CreatePlanRequest{
		Description:        description,
		FarmingPoolAddress: farmingPoolAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
	}
}

func (req CreatePlanRequest) Validate() error {
	farmingPoolAddr, err := sdk.AccAddressFromBech32(req.FarmingPoolAddress)
	if err != nil {
		return err
	}
	dummyPlan := NewPlan(
		1, req.Description, farmingPoolAddr, farmingPoolAddr,
		req.RewardAllocations, req.StartTime, req.EndTime, false)
	if err := dummyPlan.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	for _, rewardAlloc := range req.RewardAllocations {
		if rewardAlloc.PairId > 0 {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "pair reward allocation for %d is disabled", rewardAlloc.PairId)
		}
	}
	return nil
}

func NewTerminatePlanRequest(planId uint64) TerminatePlanRequest {
	return TerminatePlanRequest{PlanId: planId}
}

func (req TerminatePlanRequest) Validate() error {
	if req.PlanId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "plan id must not be zero")
	}
	return nil
}
