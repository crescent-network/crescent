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
				market := s.CreateMarket(utils.TestAddress(0), denomA, denomB, true)
				price := utils.RandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
				s.CreatePool(utils.TestAddress(0), market.Id, price, true)
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
	s.Require().Equal("ibtYndRwpdsvyCktRHFalvUuEKMqXbItfGcNGWsGzubdPMYayOUOINjpcFBeESdwpdlTYmrPsLsVDhpTzoMegKrytNVZkfJRPuDC", content.GetDescription())
	s.Require().Equal("SthOohmsux", content.GetTitle())
	s.Require().Equal("amm", content.ProposalRoute())
	s.Require().Equal("PoolParameterChange", content.ProposalType())

	w1 := weightedProposalContent[1]
	s.Require().Equal(simulation.OpWeightSubmitPublicFarmingPlanProposal, w1.AppParamsKey())
	s.Require().Equal(simulation.DefaultWeightPublicFarmingPlanProposal, w1.DefaultWeight())
	content = w1.ContentSimulatorFn()(r, s.Ctx, accs)
	s.Require().Equal("jVUZZCgqDrSeltJGXPMgZnGDZqISrGDOClxXCxMjmKqEPwKHoOfOeyGmqWqihqjINXLqnyTesZePQRqaWDQNqpLgNrAUKulklmck", content.GetDescription())
	s.Require().Equal("TijUltQKuW", content.GetTitle())
	s.Require().Equal("amm", content.ProposalRoute())
	s.Require().Equal("PublicFarmingPlan", content.ProposalType())
}
