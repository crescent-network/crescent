package keeper_test

import (
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestPoolParameterChangeProposal() {
	handler := exchange.NewProposalHandler(s.keeper)
	market1 := s.CreateMarket("ucre", "uusd")
	market2 := s.CreateMarket("uatom", "uusd")

	proposal := types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(
				market2.Id, utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("0.3"),
				nil, nil, nil, nil),
		})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().NoError(handler(s.Ctx, proposal))

	market2, _ = s.keeper.GetMarket(s.Ctx, market2.Id)
	s.Require().Equal(utils.ParseDec("0.001"), market2.MakerFeeRate)
	s.Require().Equal(utils.ParseDec("0.002"), market2.TakerFeeRate)
	s.Require().Equal(utils.ParseDec("0.3"), market2.OrderSourceFeeRatio)

	// Untouched
	market1, _ = s.keeper.GetMarket(s.Ctx, market1.Id)
	fees := s.keeper.GetFees(s.Ctx)
	s.Require().Equal(fees.DefaultMakerFeeRate, market1.MakerFeeRate)
	s.Require().Equal(fees.DefaultTakerFeeRate, market1.TakerFeeRate)
	s.Require().Equal(fees.DefaultOrderSourceFeeRatio, market1.OrderSourceFeeRatio)

	// Market not found
	proposal = types.NewMarketParameterChangeProposal(
		"Title", "Description", []types.MarketParameterChange{
			types.NewMarketParameterChange(
				3, utils.ParseDec("0.001"), utils.ParseDec("0.002"), utils.ParseDec("0.5"),
				nil, nil, nil, nil),
		})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().EqualError(handler(s.Ctx, proposal), "market 3 not found: not found")
}
