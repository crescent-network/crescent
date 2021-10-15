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
	addReq []*AddRequestProposal,
	updateReq []*UpdateRequestProposal,
	deleteReq []*DeleteRequestProposal,
) *PublicPlanProposal {
	return &PublicPlanProposal{
		Title:                  title,
		Description:            description,
		AddRequestProposals:    addReq,
		UpdateRequestProposals: updateReq,
		DeleteRequestProposals: deleteReq,
	}
}

func (p *PublicPlanProposal) GetTitle() string { return p.Title }

func (p *PublicPlanProposal) GetDescription() string { return p.Description }

func (p *PublicPlanProposal) ProposalRoute() string { return RouterKey }

func (p *PublicPlanProposal) ProposalType() string { return ProposalTypePublicPlan }

func (p *PublicPlanProposal) ValidateBasic() error {
	if len(p.AddRequestProposals) == 0 && len(p.UpdateRequestProposals) == 0 && len(p.DeleteRequestProposals) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty")
	}

	for _, ap := range p.AddRequestProposals {
		if err := ap.Validate(); err != nil {
			return err
		}
	}

	for _, up := range p.UpdateRequestProposals {
		if err := up.Validate(); err != nil {
			return err
		}
	}

	for _, dp := range p.DeleteRequestProposals {
		if err := dp.Validate(); err != nil {
			return err
		}
	}
	return gov.ValidateAbstract(p)
}

func (p PublicPlanProposal) String() string {
	return fmt.Sprintf(`Public Plan Proposal:
  Title:       			  %s
  Description: 		      %s
  AddRequestProposals: 	  %s
  UpdateRequestProposals: %s
  DeleteRequestProposals: %s
`, p.Title, p.Description, p.AddRequestProposals, p.UpdateRequestProposals, p.DeleteRequestProposals)
}

// NewAddRequestProposal creates a new AddRequestProposal object
func NewAddRequestProposal(
	name string,
	farmingPoolAddr string,
	terminationAddr string,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochAmount sdk.Coins,
	epochRatio sdk.Dec,
) *AddRequestProposal {
	return &AddRequestProposal{
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
func (p *AddRequestProposal) IsForFixedAmountPlan() bool {
	return !p.EpochAmount.Empty()
}

// IsForRatioPlan returns true if the request is for
// ratio plan.
// It checks if EpochRatio is not zero.
func (p *AddRequestProposal) IsForRatioPlan() bool {
	return !p.EpochRatio.IsNil() && !p.EpochRatio.IsZero()
}

// Validate validates AddRequestProposal.
func (p *AddRequestProposal) Validate() error {
	if p.Name == "" {
		return sdkerrors.Wrap(ErrInvalidPlanName, "plan name must not be empty")
	}
	if strings.Contains(p.Name, PoolAddrSplitter) {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot contain %s", PoolAddrSplitter)
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

// NewUpdateRequestProposal creates a new UpdateRequestProposal object.
func NewUpdateRequestProposal(
	id uint64,
	name string,
	farmingPoolAddr string,
	terminationAddr string,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochAmount sdk.Coins,
	epochRatio sdk.Dec,
) *UpdateRequestProposal {
	return &UpdateRequestProposal{
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
func (p *UpdateRequestProposal) IsForFixedAmountPlan() bool {
	return !p.EpochAmount.Empty()
}

// IsForRatioPlan returns true if the request is for
// ratio plan.
// It checks if EpochRatio is not zero.
func (p *UpdateRequestProposal) IsForRatioPlan() bool {
	return !p.EpochRatio.IsNil() && !p.EpochRatio.IsZero()
}

// Validate validates UpdateRequestProposal.
func (p *UpdateRequestProposal) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	if strings.Contains(p.Name, PoolAddrSplitter) {
		return sdkerrors.Wrapf(ErrInvalidPlanName, "plan name cannot contain %s", PoolAddrSplitter)
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

// NewDeleteRequestProposal creates a new DeleteRequestProposal object.
func NewDeleteRequestProposal(id uint64) *DeleteRequestProposal {
	return &DeleteRequestProposal{
		PlanId: id,
	}
}

// Validate validates DeleteRequestProposal.
func (p *DeleteRequestProposal) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	return nil
}
