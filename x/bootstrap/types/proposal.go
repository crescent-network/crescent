package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ProposalTypeBootstrap string = "Bootstrap"
)

// Implements Proposal Interface
var _ gov.Content = &BootstrapProposal{}

func init() {
	gov.RegisterProposalType(ProposalTypeBootstrap)
	gov.RegisterProposalTypeCodec(&BootstrapProposal{}, "crescent/BootstrapProposal")
}

// NewBootstrapProposal creates a new BootstrapProposal object.
func NewBootstrapProposal(
	title string,
	description string,
	inclusions []BootstrapHandle,
	exclusions []BootstrapHandle,
	rejections []BootstrapHandle,
	distributions []IncentiveDistribution,
) *BootstrapProposal {
	return &BootstrapProposal{
		Title:         title,
		Description:   description,
		Inclusions:    inclusions,
		Exclusions:    exclusions,
		Rejections:    rejections,
		Distributions: distributions,
	}
}

func (p *BootstrapProposal) GetTitle() string { return p.Title }

func (p *BootstrapProposal) GetDescription() string { return p.Description }

func (p *BootstrapProposal) ProposalRoute() string { return RouterKey }

func (p *BootstrapProposal) ProposalType() string { return ProposalTypeBootstrap }

func (p *BootstrapProposal) ValidateBasic() error {
	if len(p.Inclusions) == 0 && len(p.Exclusions) == 0 && len(p.Rejections) == 0 && len(p.Distributions) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty")
	}

	// checking duplicated market maker for inclusion, exclusion, rejection
	addrMap := make(map[BootstrapHandle]struct{})

	for _, mm := range p.Inclusions {
		if _, ok := addrMap[mm]; ok {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market maker can't be duplicated")
		}
		addrMap[mm] = struct{}{}
		if err := mm.Validate(); err != nil {
			return err
		}
	}

	for _, mm := range p.Exclusions {
		if _, ok := addrMap[mm]; ok {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market maker can't be duplicated")
		}
		addrMap[mm] = struct{}{}
		if err := mm.Validate(); err != nil {
			return err
		}
	}

	for _, mm := range p.Rejections {
		if _, ok := addrMap[mm]; ok {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market maker can't be duplicated")
		}
		addrMap[mm] = struct{}{}
		if err := mm.Validate(); err != nil {
			return err
		}
	}

	for _, dp := range p.Distributions {
		if err := dp.Validate(); err != nil {
			return err
		}
	}
	return gov.ValidateAbstract(p)
}

func (p BootstrapProposal) String() string {
	return fmt.Sprintf(`Market Maker Proposal:
  Title:         %s
  Description:   %s
  Inclusions:    %v
  Exclusions:    %v
  Rejections:    %v
  Distributions: %v
`, p.Title, p.Description, p.Inclusions, p.Exclusions, p.Rejections, p.Distributions)
}
