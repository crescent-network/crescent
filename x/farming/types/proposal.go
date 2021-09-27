package types

import (
	"fmt"
	time "time"

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
	if p.AddRequestProposals == nil && p.UpdateRequestProposals == nil && p.DeleteRequestProposals == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty")
	}

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

func (p *AddRequestProposal) Validate() error {
	if len(p.Name) > MaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", MaxNameLength)
	}
	if _, err := sdk.AccAddressFromBech32(p.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", p.FarmingPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(p.TerminationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", p.TerminationAddress, err)
	}
	if p.StakingCoinWeights.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty")
	}
	if err := p.StakingCoinWeights.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid staking coin weights: %v", err)
	}
	if ok := ValidateStakingCoinTotalWeights(p.StakingCoinWeights); !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1")
	}
	if !p.EndTime.After(p.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", p.EndTime, p.StartTime)
	}
	if !p.EpochAmount.IsZero() && !p.EpochRatio.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio should be provided")
	}
	if p.EpochAmount.IsZero() && p.EpochRatio.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio must not be zero")
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

func (p *UpdateRequestProposal) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	if len(p.Name) > MaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", MaxNameLength)
	}
	if _, err := sdk.AccAddressFromBech32(p.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", p.FarmingPoolAddress, err)
	}
	if _, err := sdk.AccAddressFromBech32(p.TerminationAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid termination address %q: %v", p.TerminationAddress, err)
	}
	if p.StakingCoinWeights.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty")
	}
	if err := p.StakingCoinWeights.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid staking coin weights: %v", err)
	}
	if ok := ValidateStakingCoinTotalWeights(p.StakingCoinWeights); !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1")
	}
	if !p.EndTime.After(*p.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", p.EndTime, p.StartTime)
	}
	if !p.EpochAmount.IsZero() && !p.EpochRatio.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio should be provided")
	}
	if p.EpochAmount.IsZero() && p.EpochRatio.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio must not be zero")
	}
	return nil
}

// NewDeleteRequestProposal creates a new DeleteRequestProposal object.
func NewDeleteRequestProposal(id uint64) *DeleteRequestProposal {
	return &DeleteRequestProposal{
		PlanId: id,
	}
}

func (p *DeleteRequestProposal) Validate() error {
	if p.PlanId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", p.PlanId)
	}
	return nil
}
