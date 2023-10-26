package simulation_test

import (
	"math/rand"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/simulation"
)

func (s *SimTestSuite) TestProposalContents() {
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 50)

	market := s.CreateMarket("denom1", "denom2")
	s.CreatePool(market.Id, utils.ParseDec("5"))

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(s.App.AMMKeeper, s.keeper)
	s.Require().Len(weightedProposalContent, 1)
	w0 := weightedProposalContent[0]

	// tests w0 interface:
	s.Require().Equal(simulation.OpWeightSubmitPublicPositionCreateProposal, w0.AppParamsKey())
	s.Require().Equal(simulation.DefaultWeightPublicPositionCreateProposal, w0.DefaultWeight())

	// NOTE: currently the proposal is not returned. instead, it's run
	// directly inside the proposal generator.
	w0.ContentSimulatorFn()(r, s.Ctx, accs)

	publicPosition, found := s.keeper.GetPublicPosition(s.Ctx, 1)
	s.Require().True(found)
	s.Assert().Equal(int32(30000), publicPosition.LowerTick)
	s.Assert().Equal(int32(52500), publicPosition.UpperTick)
}
