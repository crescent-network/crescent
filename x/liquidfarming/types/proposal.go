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
	ProposalTypeLiquidFarmCreate                  string = "LiquidFarmCreate"
	ProposalTypeLiquidFarmParameterChangeProposal string = "LiquidFarmParameterChangeProposal"
)

var (
	_ gov.Content = &LiquidFarmCreateProposal{}
	_ gov.Content = &LiquidFarmParameterChangeProposal{}
)

func init() {
	gov.RegisterProposalType(ProposalTypeLiquidFarmCreate)
	gov.RegisterProposalTypeCodec(&LiquidFarmCreateProposal{}, "crescent/LiquidFarmCreateProposal")
	gov.RegisterProposalType(ProposalTypeLiquidFarmParameterChangeProposal)
	gov.RegisterProposalTypeCodec(&LiquidFarmParameterChangeProposal{}, "crescent/LiquidFarmParameterChangeProposal")
}

func NewLiquidFarmCreateProposal(
	title, description string, poolId uint64,
	lowerPrice, upperPrice sdk.Dec, minBidAmt sdk.Int, feeRate sdk.Dec) *LiquidFarmCreateProposal {
	return &LiquidFarmCreateProposal{
		Title:        title,
		Description:  description,
		LowerPrice:   lowerPrice,
		UpperPrice:   upperPrice,
		MinBidAmount: minBidAmt,
		FeeRate:      feeRate,
	}
}

func (p *LiquidFarmCreateProposal) GetTitle() string       { return p.Title }
func (p *LiquidFarmCreateProposal) GetDescription() string { return p.Description }
func (p *LiquidFarmCreateProposal) ProposalRoute() string  { return RouterKey }
func (p *LiquidFarmCreateProposal) ProposalType() string {
	return ProposalTypeLiquidFarmCreate
}

func (p *LiquidFarmCreateProposal) ValidateBasic() error {
	if err := gov.ValidateAbstract(p); err != nil {
		return err
	}
	if !p.LowerPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lower price must be positive: %s", p.LowerPrice)
	}
	if !p.UpperPrice.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upper price must be positive: %s", p.UpperPrice)
	}
	if p.LowerPrice.GTE(p.UpperPrice) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "lower price must be lower than upper price")
	}
	lowerTick, valid := exchangetypes.ValidateTickPrice(p.LowerPrice)
	if !valid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid lower tick price: %s", p.LowerPrice)
	}
	if lowerTick < ammtypes.MinTick {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lower tick must not be lower than the minimum %d", ammtypes.MinTick)
	}
	upperTick, valid := exchangetypes.ValidateTickPrice(p.UpperPrice)
	if !valid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid upper tick price: %s", p.UpperPrice)
	}
	if upperTick > ammtypes.MaxTick {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upper tick must not be higher than the maximum %d", ammtypes.MaxTick)
	}
	liquidFarm := NewLiquidFarm(1, p.PoolId, lowerTick, upperTick, p.MinBidAmount, p.FeeRate)
	if err := liquidFarm.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (p LiquidFarmCreateProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Liquid Farm Create Proposal:
  Title:              %s
  Description:        %s
  Pool Id:            %d
  Lower Price:        %s
  Upper Price:        %s
  Minimum Bid Amount: %s
  Fee Rate:           %s
`, p.Title, p.Description, p.PoolId, p.LowerPrice, p.UpperPrice, p.MinBidAmount, p.FeeRate))
	return b.String()
}

func NewLiquidFarmParameterChangeProposal(
	title, description string, changes []LiquidFarmParameterChange) *LiquidFarmParameterChangeProposal {
	return &LiquidFarmParameterChangeProposal{
		Title:       title,
		Description: description,
		Changes:     changes,
	}
}

func (p *LiquidFarmParameterChangeProposal) GetTitle() string       { return p.Title }
func (p *LiquidFarmParameterChangeProposal) GetDescription() string { return p.Description }
func (p *LiquidFarmParameterChangeProposal) ProposalRoute() string  { return RouterKey }
func (p *LiquidFarmParameterChangeProposal) ProposalType() string {
	return ProposalTypeLiquidFarmParameterChangeProposal
}

func (p *LiquidFarmParameterChangeProposal) ValidateBasic() error {
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

func (p LiquidFarmParameterChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Liquid Farm Parameter Change Proposal:
  Title:       %s
  Description: %s
  Changes:
`, p.Title, p.Description))
	for _, change := range p.Changes {
		b.WriteString(fmt.Sprintf(`    Liquid Farm Parameter Change:
      Liquid Farm Id: %d
      Min Bid Amount: %s
      Fee Rate:       %s
`, change.LiquidFarmId, change.MinBidAmount, change.FeeRate))
	}
	return b.String()
}

func (change LiquidFarmParameterChange) Validate() error {
	if change.LiquidFarmId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "liquid farm id must not be 0")
	}
	if change.MinBidAmount.IsNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "min bid amount must not be negative: %s", change.MinBidAmount)
	}
	if change.FeeRate.IsNegative() || change.FeeRate.GT(utils.OneDec) {
		return fmt.Errorf("fee rate must be in range [0, 1]: %s", change.FeeRate)
	}
	return nil
}
