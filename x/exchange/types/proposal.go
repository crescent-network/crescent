package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func (p MarketParameterChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Market Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, p.Title, p.Description))
	for _, change := range p.Changes {
		b.WriteString(fmt.Sprintf(`    Market Parameter Change:
      Market Id:              %d
      Maker Fee Rate:         %s
      Taker Fee Rate:         %s
      Order Source Fee Ratio: %s
      Min Order Quantity:     %s
      Min Order Quote:        %s
      Max Order Quantity:     %s
      Max Order Quote:        %s
`, change.MarketId, change.MakerFeeRate, change.TakerFeeRate, change.OrderSourceFeeRatio,
			change.MinOrderQuantity, change.MinOrderQuote, change.MaxOrderQuantity, change.MaxOrderQuote))
	}
	return b.String()
}

func NewMarketParameterChange(
	marketId uint64, makerFeeRate, takerFeeRate, orderSourceRatio sdk.Dec,
	minOrderQty, minOrderQuote, maxOrderQty, maxOrderQuote *sdk.Int) MarketParameterChange {
	return MarketParameterChange{
		MarketId:            marketId,
		MakerFeeRate:        makerFeeRate,
		TakerFeeRate:        takerFeeRate,
		OrderSourceFeeRatio: orderSourceRatio,
		MinOrderQuantity:    minOrderQty,
		MinOrderQuote:       minOrderQuote,
		MaxOrderQuantity:    maxOrderQty,
		MaxOrderQuote:       maxOrderQuote,
	}
}

func (change MarketParameterChange) Validate() error {
	if change.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
	}
	if err := ValidateFees(
		change.MakerFeeRate, change.TakerFeeRate, change.OrderSourceFeeRatio); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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
	if change.MaxOrderQuantity != nil {
		if change.MaxOrderQuantity.IsNegative() {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "max order quantity must not be negative: %s", change.MaxOrderQuantity)
		}
	}
	if change.MaxOrderQuote != nil {
		if change.MaxOrderQuote.IsNegative() {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "max order quote must not be negative: %s", change.MaxOrderQuote)
		}
	}
	return nil
}
