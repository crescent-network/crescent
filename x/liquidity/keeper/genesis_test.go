package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	squadapp "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidity/keeper"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

type GenesisTestSuite struct {
	suite.Suite
	app    *squadapp.SquadApp
	ctx    sdk.Context
	keeper keeper.Keeper
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	s.app = squadapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidityKeeper
}

func (s *GenesisTestSuite) TestDefaultGenesis() {
	genesisState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genesisState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genesisState, *got)
}
