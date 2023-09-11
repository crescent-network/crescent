package keeper_test

import (
	"github.com/crescent-network/crescent/v5/x/marker/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.nextBlock()

	genState := s.keeper.ExportGenesis(s.ctx)
	bz := s.app.AppCodec().MustMarshalJSON(genState)

	s.SetupTest()
	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)
	genState3 := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(*genState, *genState3)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	genState := s.keeper.ExportGenesis(s.ctx)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.SetupTest()
	s.keeper.InitGenesis(s.ctx, genState2)

	genState3 := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}
