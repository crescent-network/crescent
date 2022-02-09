package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	squadapp "github.com/cosmosquad-labs/squad/app"
	squadparams "github.com/cosmosquad-labs/squad/app/params"
	"github.com/cosmosquad-labs/squad/x/liquidity/client/cli"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	val       *network.Validator
	clientCtx client.Context
}

func NewAppConstructor(encodingCfg squadparams.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return squadapp.NewSquadApp(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(store.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	encCfg := squadapp.MakeTestEncodingConfig()

	cfg := network.DefaultConfig()
	cfg.AppConstructor = NewAppConstructor(encCfg)
	cfg.GenesisState = squadapp.ModuleBasics.DefaultGenesis(cfg.Codec)
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	s.val = s.network.Validators[0]
	s.clientCtx = s.val.ClientCtx

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.createPair("node0token", s.cfg.BondDenom)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) createPair(baseCoinDenom, quoteCoinDenom string) {
	_, err := MsgCreatePair(s.clientCtx, s.val.Address.String(), baseCoinDenom, quoteCoinDenom)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

//nolint
func (s *IntegrationTestSuite) createPool(pairId uint64, depositCoins sdk.Coins) {
	_, err := MsgCreatePool(s.clientCtx, s.val.Address.String(), pairId, depositCoins)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestGetPairsCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	cmd := cli.QueryPairs()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{"--output=json"})
	s.Require().NoError(err)

	var resp types.QueryPairsResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().NotNil(resp.Pairs)
}
