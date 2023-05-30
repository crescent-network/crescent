package keeper_test

import (
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.SetupSampleScenario()

	// Export genesis state and verify
	var genState *types.GenesisState
	s.Require().NotPanics(func() {
		genState = s.keeper.ExportGenesis(s.Ctx)
		s.Require().Len(genState.RewardsAuctions, 2)
		s.Require().Len(genState.Bids, 2)
	})
	s.Require().NoError(genState.Validate())

	var genState2 types.GenesisState
	bz := s.App.AppCodec().MustMarshalJSON(genState)
	s.App.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.Ctx, genState2)

	var genState3 *types.GenesisState
	s.Require().NotPanics(func() {
		genState3 = s.keeper.ExportGenesis(s.Ctx)
		s.Require().Equal(*genState, *genState3)
	})
	s.Require().NoError(genState3.Validate())
}
