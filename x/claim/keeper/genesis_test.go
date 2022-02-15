package keeper_test

import (
	"github.com/cosmosquad-labs/squad/x/claim/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, *genState)
	// got := s.keeper.ExportGenesis(s.ctx)
	// s.Require().Equal(genState, got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	// TODO: not implemented yet
}
