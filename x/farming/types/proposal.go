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
	ProposalTypePublicPlan string = "PublicPlan"
)

// Implements Proposal Interface
var _ gov.Content = &PublicPlanProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypePublicPlan)
	gov.RegisterProposalTypeCodec(&PublicPlanProposal{}, "cosmos-sdk/PublicPlanProposal")
}

// NewPublicPlanProposal creates a new PublicPlanProposal object.
func NewPublicPlanProposal(
	title string,
	description string,
	addReqs []AddPlanRequest,
	modifyReqs []ModifyPlanRequest,
	deleteReqs []DeletePlanRequest,
) *PublicPlanProposal {
	return &PublicPlanProposal{
		Title:              title,
		Description:        description,
		AddPlanRequests:    addReqs,
		ModifyPlanRequests: modifyReqs,
		DeletePlanRequests: deleteReqs,
	}
}

func (p *PublicPlanProposal) GetTitle() string { return p.Title }

func (p *PublicPlanProposal) GetDescription() string { return p.Description }

func (p *PublicPlanProposal) ProposalRoute() string { return RouterKey }

func (p *PublicPlanProposal) ProposalType() string { return ProposalTypePublicPlan }

func (p *PublicPlanProposal) ValidateBasic() error {
	if len(p.AddPlanRequests) == 0 && len(p.ModifyPlanRequests) == 0 && len(p.DeletePlanRequests) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty")
	}

	for _, ap := range p.AddPlanRequests {
		if err := ap.Validate(); err != nil {
			return err
		}
	}

	for _, up := range p.ModifyPlanRequests {
		if err := up.Validate(); err != nil {
			return err
		}
	}

	for _, dp := range p.DeletePlanRequests {
		if err := dp.Validate(); err != nil {
			return err
		}
	}
	return gov.ValidateAbstract(p)
}

func (p PublicPlanProposal) String() string {
	return fmt.Sprintf(`Public Plan Proposal:
  Title:              %s
  Description:        %s
  AddPlanRequests:    %v
  UpdatePlanRequests: %v
  DeletePlanRequests: %v
`, p.Title, p.Description, p.AddPlanRequests, p.ModifyPlanRequests, p.DeletePlanRequests)
}

// NewAddPlanRequest creates a new AddPlanRequest object
func NewAddPlanRequest(
	name string,
	farmingPoolAddr string,
	terminationAddr string,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochAmount sdk.Coins,
	epochRatio sdk.Dec,
) AddPlanRequest {
	return AddPlanRequest{
		Name:               name,
		FarmingPoolAddress: farmingPoolAddr,
		TerminationAddress: terminationAddr,
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
		EpochAmount:        epochAmount,
		EpochRatio:         epochRatio,
	}
}

// IsForFixedAmountPlan returns true if the request is for
// fixed amount plan.
// It checks if EpochAmount is not zero.
func (p *AddPlanRequest) IsForFixedAmountPlan() bool {
	return !p.EpochAmount.Empty()
}

// IsForRatioPlan returns true if the request is for
// ratio plan.
// It checks if EpochRatio is not zero.
func (p *AddPlanRequest) IsForRatioPlan() bool {
	return !p.EpochRatio.IsNil() && !p.EpochRatio.IsZero()
}

// Validate validates AddPlanRequest.
func (p *AddPlanRequest) Validate() error {
	if p.Name == "" {
		return sdkerrors.Wrap(ErrInvalidPlanName, "plan name must not be empty")
	}
	if strings.Contains(p.Name, AccNameSplitter) {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot contain %s", AccNameSplitter)
	}
	if len(p.Name) > MaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot be longer than max length of %d", MaxNameLength)
	}
	if _, err := sdk.AccAddressFromBech32(p.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", p.FarmingPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(p.TerminationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", p.TerminationAddress, err)
	}
	if err := ValidateStakingCoinTotalWeights(p.StakingCoinWeights); err != nil {
		return err
	}
	if !p.EndTime.After(p.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", p.EndTime, p.StartTime)
	}

	isForFixedAmountPlan := p.IsForFixedAmountPlan()
	isForRatioPlan := p.IsForRatioPlan()
	switch {
	case isForFixedAmountPlan == isForRatioPlan:
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "exactly one of epoch amount or epoch ratio must be provided")
	case isForFixedAmountPlan:
		if err := ValidateEpochAmount(p.EpochAmount); err != nil {
			return err
		}
	case isForRatioPlan:
		if err := ValidateEpochRatio(p.EpochRatio); err != nil {
			return err
		}
	}
	return nil
}

// NewModifyPlanRequest creates a new ModifyPlanRequest object.
func NewModifyPlanRequest(
	id uint64,
	name string,
	farmingPoolAddr string,
	terminationAddr string,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochAmount sdk.Coins,
	epochRatio sdk.Dec,
) ModifyPlanRequest {
	return ModifyPlanRequest{
		PlanId:             id,
		Name:               name,
		FarmingPoolAddress: farmingPoolAddr,
		TerminationAddress: terminationAddr,
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        epochAmount,
		EpochRatio:         epochRatio,
	}
}

// IsForFixedAmountPlan returns true if the request is for
// fixed amount plan.
// It checks if EpochAmount is not zero.
func (p *ModifyPlanRequest) IsForFixedAmountPlan() bool {
	return !p.EpochAmount.Empty()
}

// IsForRatioPlan returns true if the request is for
// ratio plan.
// It checks if EpochRatio is not zero.
func (p *ModifyPlanRequest) IsForRatioPlan() bool {
	return !p.EpochRatio.IsNil() && !p.EpochRatio.IsZero()
}

// Validate validates ModifyPlanRequest.
func (p *ModifyPlanRequest) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	if strings.Contains(p.Name, AccNameSplitter) {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot contain %s", AccNameSplitter)
	}
	if len(p.Name) > MaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot be longer than max length of %d", MaxNameLength)
	}
	if p.FarmingPoolAddress != "" {
		if _, err := sdk.AccAddressFromBech32(p.FarmingPoolAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", p.FarmingPoolAddress, err)
		}
	}
	if p.TerminationAddress != "" {
		if _, err := sdk.AccAddressFromBech32(p.TerminationAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", p.TerminationAddress, err)
		}
	}
	if p.StakingCoinWeights != nil {
		if err := ValidateStakingCoinTotalWeights(p.StakingCoinWeights); err != nil {
			return err
		}
	}
	if p.StartTime != nil && p.EndTime != nil {
		if !p.EndTime.After(*p.StartTime) {
			return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", p.EndTime, p.StartTime)
		}
	}
	isForFixedAmountPlan := p.IsForFixedAmountPlan()
	isForRatioPlan := p.IsForRatioPlan()
	switch {
	case isForFixedAmountPlan && isForRatioPlan:
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "at most one of epoch amount or epoch ratio must be provided")
	case isForFixedAmountPlan:
		if err := ValidateEpochAmount(p.EpochAmount); err != nil {
			return err
		}
	case isForRatioPlan:
		if err := ValidateEpochRatio(p.EpochRatio); err != nil {
			return err
		}
	}
	return nil
}

// NewDeletePlanRequest creates a new DeletePlanRequest object.
func NewDeletePlanRequest(id uint64) DeletePlanRequest {
	return DeletePlanRequest{
		PlanId: id,
	}
}

// Validate validates DeletePlanRequest.
func (p *DeletePlanRequest) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	return nil
}
