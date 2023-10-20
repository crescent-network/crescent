package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestMarketParameterChangeProposal() {
	handler := exchange.NewProposalHandler(s.keeper)
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "uusd")

	proposal := types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(
				market2.Id,
				types.NewFees(
					utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("0.3")),
				types.NewAmountLimits(sdk.NewInt(100), sdk.NewIntWithDecimal(1, 20)),
				types.NewAmountLimits(sdk.NewInt(1000), sdk.NewIntWithDecimal(1, 25))),
		})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().NoError(handler(s.Ctx, proposal))

	market2, _ = s.keeper.GetMarket(s.Ctx, market2.Id)
	s.AssertEqual(utils.ParseDec("0.001"), market2.Fees.MakerFeeRate)
	s.AssertEqual(utils.ParseDec("0.002"), market2.Fees.TakerFeeRate)
	s.AssertEqual(utils.ParseDec("0.3"), market2.Fees.OrderSourceFeeRatio)
	s.AssertEqual(sdk.NewInt(100), market2.OrderQuantityLimits.Min)
	s.AssertEqual(sdk.NewIntWithDecimal(1, 20), market2.OrderQuantityLimits.Max)
	s.AssertEqual(sdk.NewInt(1000), market2.OrderQuoteLimits.Min)
	s.AssertEqual(sdk.NewIntWithDecimal(1, 25), market2.OrderQuoteLimits.Max)

	// Market 1 is untouched
	market1, _ = s.keeper.GetMarket(s.Ctx, market1.Id)
	s.Require().Equal(s.keeper.GetDefaultFees(s.Ctx), market1.Fees)
	s.Require().Equal(s.keeper.GetDefaultOrderQuantityLimits(s.Ctx), market1.OrderQuantityLimits)
	s.Require().Equal(s.keeper.GetDefaultOrderQuoteLimits(s.Ctx), market1.OrderQuoteLimits)

	// Market not found
	proposal = types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(
				3,
				types.NewFees(
					utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("0.3")),
				types.NewAmountLimits(sdk.NewInt(100), sdk.NewIntWithDecimal(1, 20)),
				types.NewAmountLimits(sdk.NewInt(1000), sdk.NewIntWithDecimal(1, 25))),
		})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().EqualError(handler(s.Ctx, proposal), "market 3 not found: not found")
}
