package simulation_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/simulation"
)

func (s *SimTestSuite) TestProposalContents() {
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 50)

	// Create all possible markets
	var denoms []string
	s.App.BankKeeper.IterateTotalSupply(s.Ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	for _, denomA := range denoms {
		for _, denomB := range denoms {
			if denomA != denomB {
				s.CreateMarket(utils.TestAddress(0), denomA, denomB, true)
			}
		}
	}

	// execute ProposalContents function
	weightedProposalContent := simulation.ProposalContents(s.keeper)
	s.Require().Len(weightedProposalContent, 1)
	w0 := weightedProposalContent[0]

	// tests w0 interface:
	s.Require().Equal(simulation.OpWeightSubmitMarketParameterChangeProposal, w0.AppParamsKey())
	s.Require().Equal(simulation.DefaultWeightMarketParameterChangeProposal, w0.DefaultWeight())

	content := w0.ContentSimulatorFn()(r, s.Ctx, accs)

	s.Require().Equal("PQrNbTwxsGdwuduvibtYndRwpdsvyCktRHFalvUuEKMqXbItfGcNGWsGzubdPMYayOUOINjpcFBeESdwpdlTYmrPsLsVDhpTzoMe", content.GetDescription())
	s.Require().Equal("VZkfJRPuDC", content.GetTitle())
	s.Require().Equal("exchange", content.ProposalRoute())
	s.Require().Equal("MarketParameterChange", content.ProposalType())
}
