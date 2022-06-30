package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/app/params"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/client/cli"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	val       *network.Validator
	clientCtx client.Context

	denom1, denom2 string
}

func NewAppConstructor(encodingCfg params.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return chain.NewApp(
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

	encCfg := chain.MakeTestEncodingConfig()

	cfg := network.DefaultConfig()
	cfg.AppConstructor = NewAppConstructor(encCfg)
	cfg.GenesisState = chain.ModuleBasics.DefaultGenesis(cfg.Codec)
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	s.val = s.network.Validators[0]
	s.clientCtx = s.val.ClientCtx

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.denom1, s.denom2 = fmt.Sprintf("%stoken", s.val.Moniker), s.cfg.BondDenom

	s.createPair(s.denom1, s.denom2)
	s.createPool(1, sdk.NewCoins(sdk.NewInt64Coin(s.denom1, 10000000), sdk.NewInt64Coin(s.denom2, 10000000)))
	s.limitOrder(
		1, types.OrderDirectionSell, utils.ParseCoin("1000000node0token"), s.cfg.BondDenom,
		utils.ParseDec("1.0"), sdk.NewInt(1000000), time.Minute)
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

func (s *IntegrationTestSuite) createPool(pairId uint64, depositCoins sdk.Coins) {
	_, err := MsgCreatePool(s.clientCtx, s.val.Address.String(), pairId, depositCoins)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) limitOrder(
	pairId uint64, dir types.OrderDirection, offerCoin sdk.Coin,
	demandCoinDenom string, price sdk.Dec, amt sdk.Int, orderLifespan time.Duration) {
	_, err := MsgLimitOrder(s.clientCtx, s.val.Address.String(), pairId, dir, offerCoin, demandCoinDenom,
		price, amt, orderLifespan)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQueryPairsCmd() {
	val := s.network.Validators[0]

	for _, tc := range []struct {
		name        string
		args        []string
		expectedErr string
		postRun     func(resp types.QueryPairsResponse)
	}{
		{
			"happy case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"",
			func(resp types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 1)
				s.Require().Equal(s.denom1, resp.Pairs[0].BaseCoinDenom)
				s.Require().Equal(s.denom2, resp.Pairs[0].QuoteCoinDenom)
			},
		},
	} {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryPairsCmd()
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, tc.args)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				var resp types.QueryPairsResponse
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryPoolsCmd() {
	val := s.network.Validators[0]

	for _, tc := range []struct {
		name        string
		args        []string
		expectedErr string
		postRun     func(resp types.QueryPoolsResponse)
	}{
		{
			"happy case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"",
			func(resp types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
				s.Require().Equal(uint64(1), resp.Pools[0].PairId)
				s.Require().Equal(uint64(1), resp.Pools[0].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryPoolsCmd()
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, tc.args)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				var resp types.QueryPoolsResponse
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryOrdersCmd() {
	val := s.network.Validators[0]

	for _, tc := range []struct {
		name        string
		args        []string
		expectedErr string
		postRun     func(resp types.QueryOrdersResponse)
	}{
		{
			"happy case",
			[]string{
				fmt.Sprintf("--%s=%d", cli.FlagPairId, 1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"",
			func(resp types.QueryOrdersResponse) {
				s.Require().Len(resp.Orders, 1)
				s.Require().Equal(uint64(1), resp.Orders[0].PairId)
				s.Require().Equal(uint64(1), resp.Orders[0].Id)
			},
		},
		{
			"no arguments",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"either orderer or pair-id must be specified",
			nil,
		},
		{
			"specify both orderer and pair id",
			[]string{
				s.val.Address.String(),
				fmt.Sprintf("--%s=%d", cli.FlagPairId, 1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			"",
			func(resp types.QueryOrdersResponse) {
				s.Require().Len(resp.Orders, 1)
				s.Require().Equal(uint64(1), resp.Orders[0].PairId)
				s.Require().Equal(uint64(1), resp.Orders[0].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryOrdersCmd()
			out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, tc.args)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				var resp types.QueryOrdersResponse
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}
