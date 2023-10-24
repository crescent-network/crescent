package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

const (
	ProposalTypePublicPositionCreate          string = "PublicPositionCreate"
	ProposalTypePublicPositionParameterChange string = "PublicPositionParameterChange"
)

var (
	_ gov.Content = &PublicPositionCreateProposal{}
	_ gov.Content = &PublicPositionParameterChangeProposal{}
)

func init() {
	gov.RegisterProposalType(ProposalTypePublicPositionCreate)
	gov.RegisterProposalTypeCodec(&PublicPositionCreateProposal{}, "crescent/PublicPositionCreateProposal")
	gov.RegisterProposalType(ProposalTypePublicPositionParameterChange)
	gov.RegisterProposalTypeCodec(&PublicPositionParameterChangeProposal{}, "crescent/PublicPositionParameterChangeProposal")
}

func NewPublicPositionCreateProposal(
	title, description string, poolId uint64,
	lowerPrice, upperPrice sdk.Dec, feeRate sdk.Dec) *PublicPositionCreateProposal {
	return &PublicPositionCreateProposal{
		Title:       title,
		Description: description,
		PoolId:      poolId,
		LowerPrice:  lowerPrice,
		UpperPrice:  upperPrice,
		FeeRate:     feeRate,
	}
}

func (p *PublicPositionCreateProposal) GetTitle() string       { return p.Title }
func (p *PublicPositionCreateProposal) GetDescription() string { return p.Description }
func (p *PublicPositionCreateProposal) ProposalRoute() string  { return RouterKey }
func (p *PublicPositionCreateProposal) ProposalType() string {
	return ProposalTypePublicPositionCreate
}

func (p *PublicPositionCreateProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}
	if err := ammtypes.ValidatePriceRange(p.LowerPrice, p.UpperPrice); err != nil {
		return err
	}
	publicPosition := NewPublicPosition(
		1, p.PoolId,
		exchangetypes.TickAtPrice(p.LowerPrice), exchangetypes.TickAtPrice(p.UpperPrice), p.FeeRate)
	if err := publicPosition.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (p PublicPositionCreateProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Public Position Create Proposal:
  Title:       %s
  Description: %s
  Pool Id:     %d
  Lower Price: %s
  Upper Price: %s
  Fee Rate:    %s
`, p.Title, p.Description, p.PoolId, p.LowerPrice, p.UpperPrice, p.FeeRate))
	return b.String()
}

func NewPublicPositionParameterChangeProposal(
	title, description string, changes []PublicPositionParameterChange) *PublicPositionParameterChangeProposal {
	return &PublicPositionParameterChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

func (p *PublicPositionParameterChangeProposal) GetTitle() string       { return p.Title }
func (p *PublicPositionParameterChangeProposal) GetDescription() string { return p.Description }
func (p *PublicPositionParameterChangeProposal) ProposalRoute() string  { return RouterKey }
func (p *PublicPositionParameterChangeProposal) ProposalType() string {
	return ProposalTypePublicPositionParameterChange
}

func (p *PublicPositionParameterChangeProposal) ValidateBasic() error {
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

func (p PublicPositionParameterChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Public Position Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, p.Title, p.Description))
	for _, change := range p.Changes {
		b.WriteString(fmt.Sprintf(`    Public Position Parameter Change:
      Public Position Id: %d
      Fee Rate:           %s
`, change.PublicPositionId, change.FeeRate))
	}
	return b.String()
}

func NewPublicPositionParameterChange(publicPositionId uint64, feeRate sdk.Dec) PublicPositionParameterChange {
	return PublicPositionParameterChange{
		PublicPositionId: publicPositionId,
		FeeRate:          feeRate,
	}
}

func (change PublicPositionParameterChange) Validate() error {
	if change.PublicPositionId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "public position id must not be 0")
	}
	if change.FeeRate.IsNegative() || change.FeeRate.GT(utils.OneDec) {
		return fmt.Errorf("fee rate must be in range [0, 1]: %s", change.FeeRate)
	}
	return nil
}
