package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crescentapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity/keeper"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

type GenesisTestSuite struct {
	suite.Suite
	app    *crescentapp.CrescentApp
	ctx    sdk.Context
	keeper keeper.Keeper
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	s.app = crescentapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidityKeeper
}

func (s *GenesisTestSuite) TestDefaultGenesis() {
	genesisState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genesisState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genesisState, *got)
}
