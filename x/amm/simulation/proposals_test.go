package simulation_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/simulation"
)

func (s *SimTestSuite) TestProposalContents() {
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 50)

	// Create all possible markets and pools
	var denoms []string
	s.App.BankKeeper.IterateTotalSupply(s.Ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	for _, denomA := range denoms {
		for _, denomB := range denoms {
			if denomA != denomB {
				market := s.CreateMarket(denomA, denomB)
				price := utils.SimRandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
				s.CreatePool(market.Id, price)
			}
		}
	}

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(s.App.BankKeeper, s.keeper)
	s.Require().Len(weightedProposalContent, 2)

	w0 := weightedProposalContent[0]
	s.Require().Equal(simulation.OpWeightSubmitPoolParameterChangeProposal, w0.AppParamsKey())
	s.Require().Equal(simulation.DefaultWeightPoolParameterChangeProposal, w0.DefaultWeight())
	content := w0.ContentSimulatorFn()(r, s.Ctx, accs)
	s.Assert().Equal("amm", content.ProposalRoute())
	s.Assert().Equal("PoolParameterChange", content.ProposalType())

	w1 := weightedProposalContent[1]
	s.Require().Equal(simulation.OpWeightSubmitPublicFarmingPlanProposal, w1.AppParamsKey())
	s.Require().Equal(simulation.DefaultWeightPublicFarmingPlanProposal, w1.DefaultWeight())
	content = w1.ContentSimulatorFn()(r, s.Ctx, accs)
	s.Assert().Equal("amm", content.ProposalRoute())
	s.Assert().Equal("PublicFarmingPlan", content.ProposalType())
}
