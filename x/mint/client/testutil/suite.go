package testutil

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmdb "github.com/tendermint/tm-db"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/x/mint/client/cli"
	minttypes "github.com/crescent-network/crescent/v2/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	db := tmdb.NewMemDB()
	cfg := chain.NewConfig(db)
	s.cfg = cfg

	genesisState := s.cfg.GenesisState

	var mintData minttypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	mintDataBz, err := s.cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	s.cfg.GenesisState = genesisState

	s.network = network.New(s.T(), s.cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"mint_denom":"stake","mint_pool_address":"cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta","block_time_threshold":"10s","inflation_schedules":[{"start_time":"2022-01-01T00:00:00Z","end_time":"2023-01-01T00:00:00Z","amount":"300000000000000"},{"start_time":"2023-01-01T00:00:00Z","end_time":"2024-01-01T00:00:00Z","amount":"200000000000000"}]}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`block_time_threshold: 10s
inflation_schedules:
- amount: "300000000000000"
  end_time: "2023-01-01T00:00:00Z"
  start_time: "2022-01-01T00:00:00Z"
- amount: "200000000000000"
  end_time: "2024-01-01T00:00:00Z"
  start_time: "2023-01-01T00:00:00Z"
mint_denom: stake
mint_pool_address: cosmos17xpfvakm2amg962yls6f84z3kell8c5lserqta`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
