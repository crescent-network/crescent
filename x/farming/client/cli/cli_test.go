// +build norace

package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmdb "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtestutil "github.com/tendermint/farming/x/farming/client/testutil"
	farmingtypes "github.com/tendermint/farming/x/farming/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	db *tmdb.MemDB // in-memory database is needed for exporting genesis cli integration test
}

// SetupTest creates a new network for _each_ integration test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up integration test suite")

	db := tmdb.NewMemDB()

	cfg := farmingtestutil.NewConfig(db)
	cfg.NumValidators = 1

	var genesisState farmingtypes.GenesisState
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[farmingtypes.ModuleName], &genesisState)
	s.Require().NoError(err)

	genesisState.Params = farmingtypes.DefaultParams()

	cfg.GenesisState[farmingtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&genesisState)
	cfg.AccountTokens = sdk.NewInt(100_000_000_000) // node0token denom
	cfg.StakingTokens = sdk.NewInt(100_000_000_000) // stake denom

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)
	s.db = db

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownTest cleans up the curret test network after each test in the suite.
func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TestIntegrationTestSuite every integration test suite.
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestNewCreateFixedAmountPlanCmd() {
	// TODO: not implemented yet
}

func (s *IntegrationTestSuite) TestNewCreateRatioPlanCmd() {
	// TODO: not implemented yet
}

func (s *IntegrationTestSuite) TestNewStakeCmd() {
	// TODO: not implemented yet
}

func (s *IntegrationTestSuite) TestNewUnstakeCmd() {
	// TODO: not implemented yet
}

func (s *IntegrationTestSuite) TestNewHarvestCmd() {
	// TODO: not implemented yet
}
