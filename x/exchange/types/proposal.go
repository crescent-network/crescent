package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	ProposalTypeMarketParameterChange string = "MarketParameterChange"
)

var (
	_ gov.Content = &MarketParameterChangeProposal{}
)

func init() {
	gov.RegisterProposalType(ProposalTypeMarketParameterChange)
	gov.RegisterProposalTypeCodec(&MarketParameterChangeProposal{}, "crescent/MarketParameterChangeProposal")
}

func NewMarketParameterChangeProposal(
	title, description string, changes []MarketParameterChange) *MarketParameterChangeProposal {
	return &MarketParameterChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

func (p *MarketParameterChangeProposal) GetTitle() string       { return p.Title }
func (p *MarketParameterChangeProposal) GetDescription() string { return p.Description }
func (p *MarketParameterChangeProposal) ProposalRoute() string  { return RouterKey }
func (p *MarketParameterChangeProposal) ProposalType() string {
	return ProposalTypeMarketParameterChange
}

func (p *MarketParameterChangeProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}
	for _, change := range p.Changes {
		if err := change.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p MarketParameterChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Market Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, p.Title, p.Description))
	for _, change := range p.Changes {
		b.WriteString(fmt.Sprintf(`    Market Parameter Change:
      Market Id:      %d
      Maker Fee Rate: %s
      Taker Fee Rate: %s
`, change.MarketId, change.MakerFeeRate, change.TakerFeeRate))
	}
	return b.String()
}

func NewMarketParameterChange(
	marketId uint64, makerFeeRate, takerFeeRate sdk.Dec) MarketParameterChange {
	return MarketParameterChange{
		MarketId:     marketId,
		MakerFeeRate: makerFeeRate,
		TakerFeeRate: takerFeeRate,
	}
}

func (change MarketParameterChange) Validate() error {
	if change.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	if change.MakerFeeRate.GT(utils.OneDec) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "maker fee rate must not be greater than 1.0: %s", change.MakerFeeRate)
	}
	if change.MakerFeeRate.LT(utils.OneDec.Neg()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "maker fee rate must not be less than -1.0: %s", change.MakerFeeRate)
	}
	if change.TakerFeeRate.GT(utils.OneDec) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "taker fee rate must not be greater than 1.0: %s", change.TakerFeeRate)
	}
	if change.TakerFeeRate.IsNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "taker fee rate must not be negative: %s", change.TakerFeeRate)
	}
	if change.MakerFeeRate.IsNegative() && change.MakerFeeRate.Neg().GT(change.TakerFeeRate) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minus maker fee rate must not be greater than taker fee rate")
	}
	return nil
}
