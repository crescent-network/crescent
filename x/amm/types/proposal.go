package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypePoolParameterChange string = "PoolParameterChange"
	ProposalTypePublicFarmingPlan   string = "PublicFarmingPlan"
)

var (
	_ gov.Content = &PoolParameterChangeProposal{}
	_ gov.Content = &PublicFarmingPlanProposal{}
)

func init() {
	gov.RegisterProposalType(ProposalTypePoolParameterChange)
	gov.RegisterProposalTypeCodec(&PoolParameterChangeProposal{}, "crescent/PoolParameterChangeProposal")
	gov.RegisterProposalType(ProposalTypePublicFarmingPlan)
	gov.RegisterProposalTypeCodec(&PublicFarmingPlanProposal{}, "crescent/PublicFarmingPlanProposal")
}

func NewPoolParameterChangeProposal(title, description string, changes []PoolParameterChange) *PoolParameterChangeProposal {
	return &PoolParameterChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

func (p *PoolParameterChangeProposal) GetTitle() string       { return p.Title }
func (p *PoolParameterChangeProposal) GetDescription() string { return p.Description }
func (p *PoolParameterChangeProposal) ProposalRoute() string  { return RouterKey }
func (p *PoolParameterChangeProposal) ProposalType() string {
	return ProposalTypePoolParameterChange
}

func (p *PoolParameterChangeProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}
	if len(p.Changes) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "changes must not be empty")
	}
	for _, change := range p.Changes {
		if err := change.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p PoolParameterChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Pool Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, p.Title, p.Description))
	for _, change := range p.Changes {
		b.WriteString(fmt.Sprintf(`    Pool Parameter Change:
      Pool Id:            %d
      Tick Spacing:       %d
      Min Order Quantity: %s
      Min Order Quote:    %s
`, change.PoolId, change.TickSpacing, change.MinOrderQuantity, change.MinOrderQuote))
	}
	return b.String()
}

func NewPoolParameterChange(
	poolId uint64, tickSpacing uint32, minOrderQty, minOrderQuote *sdk.Dec) PoolParameterChange {
	return PoolParameterChange{
		PoolId:           poolId,
		TickSpacing:      tickSpacing,
		MinOrderQuantity: minOrderQty,
		MinOrderQuote:    minOrderQuote,
	}
}

func (change PoolParameterChange) Validate() error {
	if change.PoolId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "pool id must not be 0")
	}
	if change.TickSpacing == 0 && change.MinOrderQuantity == nil && change.MinOrderQuote == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no changes")
	}
	if change.TickSpacing != 0 {
		if !IsAllowedTickSpacing(change.TickSpacing) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "tick spacing %d is not allowed", change.TickSpacing)
		}
	}
	if change.MinOrderQuantity != nil {
		if change.MinOrderQuantity.IsNegative() {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "min order quantity must not be negative: %s", change.MinOrderQuantity)
		}
	}
	if change.MinOrderQuote != nil {
		if change.MinOrderQuote.IsNegative() {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "min order quote must not be negative: %s", change.MinOrderQuote)
		}
	}
	return nil
}

func NewPublicFarmingPlanProposal(
	title, description string,
	createReqs []CreatePublicFarmingPlanRequest, termReqs []TerminateFarmingPlanRequest) *PublicFarmingPlanProposal {
	return &PublicFarmingPlanProposal{
		Title:             title,
		Description:       description,
		CreateRequests:    createReqs,
		TerminateRequests: termReqs,
	}
}

func (p *PublicFarmingPlanProposal) GetTitle() string       { return p.Title }
func (p *PublicFarmingPlanProposal) GetDescription() string { return p.Description }
func (p *PublicFarmingPlanProposal) ProposalRoute() string  { return RouterKey }
func (p *PublicFarmingPlanProposal) ProposalType() string {
	return ProposalTypePublicFarmingPlan
}

func (p *PublicFarmingPlanProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}
	if len(p.CreateRequests) == 0 && len(p.TerminateRequests) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "requests must not be empty")
	}
	for _, createReq := range p.CreateRequests {
		if err := createReq.Validate(); err != nil {
			return err
		}
	}
	for _, termReq := range p.TerminateRequests {
		if err := termReq.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p PublicFarmingPlanProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Public Farming Plan Proposal:
  Title:       %s
  Description: %s
  Create Requests:
`, p.Title, p.Description))
	for _, createReq := range p.CreateRequests {
		b.WriteString(fmt.Sprintf(`    Create Public Farming Plan Request:
      Description:          %s
      Farming Pool Address: %s
      Termination Address:  %s
      Start Time:           %s
      End Time:             %s
      Reward Allocations:
`, createReq.Description, createReq.FarmingPoolAddress, createReq.TerminationAddress,
			createReq.StartTime, createReq.EndTime))
		for _, rewardAlloc := range createReq.RewardAllocations {
			b.WriteString(fmt.Sprintf(`        Reward Allocation:
          Pool Id:         %d
          Rewards Per Day: %s
`, rewardAlloc.PoolId, rewardAlloc.RewardsPerDay))
		}
	}
	b.WriteString("  Terminate Farming Plan Request:\n")
	for _, termReq := range p.TerminateRequests {
		b.WriteString(fmt.Sprintf(`    Terminate Public Farming Plan Request:
      Farming Plan Id: %d
`, termReq.FarmingPlanId))
	}
	return b.String()
}

func NewCreatePublicFarmingPlanRequest(
	description string, farmingPoolAddr, termAddr sdk.AccAddress,
	rewardAllocs []FarmingRewardAllocation, startTime, endTime time.Time) CreatePublicFarmingPlanRequest {
	return CreatePublicFarmingPlanRequest{
		Description:        description,
		FarmingPoolAddress: farmingPoolAddr.String(),
		TerminationAddress: termAddr.String(),
		RewardAllocations:  rewardAllocs,
		StartTime:          startTime,
		EndTime:            endTime,
	}
}

func (req CreatePublicFarmingPlanRequest) Validate() error {
	dummyPlan := FarmingPlan{
		Id:                 1,
		Description:        req.Description,
		FarmingPoolAddress: req.FarmingPoolAddress,
		TerminationAddress: req.TerminationAddress,
		RewardAllocations:  req.RewardAllocations,
		StartTime:          req.StartTime,
		EndTime:            req.EndTime,
		IsPrivate:          false,
		IsTerminated:       false,
	}
	if err := dummyPlan.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func NewTerminateFarmingPlanRequest(planId uint64) TerminateFarmingPlanRequest {
	return TerminateFarmingPlanRequest{FarmingPlanId: planId}
}

func (req TerminateFarmingPlanRequest) Validate() error {
	if req.FarmingPlanId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "farming plan id must not be zero")
	}
	return nil
}
