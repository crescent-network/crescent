package testutil

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	utilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"

	"github.com/CosmWasm/wasmd/x/wasm"

	chain "github.com/crescent-network/crescent/v4/app"
	"github.com/crescent-network/crescent/v4/app/params"
	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/liquidfarming/client/cli"
	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
	liquiditytestutil "github.com/crescent-network/crescent/v4/x/liquidity/client/testutil"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
)

// emptyWasmOpts defines a type alias for a list of wasm options.
var emptyWasmOpts []wasm.Option = nil

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
			wasm.DisableAllProposals,
			emptyWasmOpts,
			baseapp.SetPruning(store.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

// SetupTest creates a new network for _each_ integration test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up integration test suite")

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}

	encCfg := chain.MakeTestEncodingConfig()

	cfg := network.DefaultConfig()
	cfg.AppConstructor = NewAppConstructor(encCfg)
	cfg.GenesisState = chain.ModuleBasics.DefaultGenesis(cfg.Codec)
	cfg.NumValidators = 1

	var genesisState types.GenesisState
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &genesisState)
	s.Require().NoError(err)

	genesisState.Params = types.DefaultParams()
	genesisState.Params.RewardsAuctionDuration = 1 * time.Hour
	genesisState.Params.LiquidFarms = []types.LiquidFarm{
		{
			PoolId:        1,
			MinFarmAmount: sdk.NewInt(100_000),
			MinBidAmount:  sdk.NewInt(100_000),
		},
	}
	genesisState.LastRewardsAuctionIdRecord = []types.LastRewardsAuctionIdRecord{
		{
			PoolId:    1,
			AuctionId: 1,
		},
	}
	genesisState.RewardsAuctions = []types.RewardsAuction{
		{
			Id:                   1,
			PoolId:               1,
			BiddingCoinDenom:     liquiditytypes.PoolCoinDenom(1),
			PayingReserveAddress: types.PayingReserveAddress(1).String(),
			StartTime:            utils.ParseTime("0001-01-01T00:00:00Z"),
			EndTime:              utils.ParseTime("9999-12-31T23:59:59Z"),
			Status:               types.AuctionStatusStarted,
		},
	}
	cfg.GenesisState[types.ModuleName] = cfg.Codec.MustMarshalJSON(&genesisState)

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	s.val = s.network.Validators[0]
	s.clientCtx = s.val.ClientCtx

	s.denom1, s.denom2 = fmt.Sprintf("%stoken", s.val.Moniker), s.cfg.BondDenom

	s.createPair(s.denom1, s.denom2)
	s.createPool(1, sdk.NewCoins(sdk.NewInt64Coin(s.denom1, 10_000_000), sdk.NewInt64Coin(s.denom2, 10_000_000)))

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownTest cleans up the current test network after each test in the suite.
func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

//
// Helper functions
//

func (s *IntegrationTestSuite) createPair(baseCoinDenom, quoteCoinDenom string) {
	_, err := liquiditytestutil.MsgCreatePair(s.clientCtx, s.val.Address.String(), baseCoinDenom, quoteCoinDenom)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) createPool(pairId uint64, depositCoins sdk.Coins) {
	_, err := liquiditytestutil.MsgCreatePool(s.clientCtx, s.val.Address.String(), pairId, depositCoins)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
}

//
// Query CLI Integration Tests
//

func (s *IntegrationTestSuite) TestNewQueryParamsCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"happy case",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
		{
			"with specific height",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryParamsCmd()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				s.Require().NoError(err)

				var params types.Params
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &params))
				s.Require().NotEmpty(params.LiquidFarms)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewQueryLiquidFarmsCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryLiquidFarmsResponse)
	}{
		{
			"happy case",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryLiquidFarmsResponse) {
				s.Require().Len(resp.LiquidFarms, 1)
				s.Require().Equal(uint64(1), resp.LiquidFarms[0].PoolId)
				s.Require().Equal(types.LiquidFarmCoinDenom(1), resp.LiquidFarms[0].LFCoinDenom)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarms[0].MinFarmAmount)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarms[0].MinBidAmount)
			},
		},
		{
			"with specific height",
			[]string{
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryLiquidFarmsResponse) {
				s.Require().Len(resp.LiquidFarms, 1)
				s.Require().Equal(uint64(1), resp.LiquidFarms[0].PoolId)
				s.Require().Equal(types.LiquidFarmCoinDenom(1), resp.LiquidFarms[0].LFCoinDenom)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarms[0].MinFarmAmount)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarms[0].MinBidAmount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryLiquidFarmsCmd()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryLiquidFarmsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewQueryLiquidFarmCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryLiquidFarmResponse)
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryLiquidFarmResponse) {
				s.Require().Equal(uint64(1), resp.LiquidFarm.PoolId)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarm.MinFarmAmount)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarm.MinBidAmount)
			},
		},
		{
			"with specific height",
			[]string{
				strconv.Itoa(1),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryLiquidFarmResponse) {
				s.Require().Equal(uint64(1), resp.LiquidFarm.PoolId)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarm.MinFarmAmount)
				s.Require().Equal(sdk.NewInt(100_000), resp.LiquidFarm.MinBidAmount)
			},
		},
		{
			"pool id not found",
			[]string{
				strconv.Itoa(10),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryLiquidFarmCmd()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryLiquidFarmResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewQueryRewardsAuctionsCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionsResponse)
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 1)
				s.Require().Equal(uint64(1), resp.RewardsAuctions[0].Id)
				s.Require().Equal(uint64(1), resp.RewardsAuctions[0].PoolId)
				s.Require().Equal(types.AuctionStatusStarted, resp.RewardsAuctions[0].Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryRewardsAuctionsCmd()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryRewardsAuctionsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewQueryRewardsAuctionCmd() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionResponse)
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				strconv.Itoa(1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().Equal(uint64(1), resp.RewardsAuction.Id)
				s.Require().Equal(uint64(1), resp.RewardsAuction.PoolId)
				s.Require().Equal(types.AuctionStatusStarted, resp.RewardsAuction.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.NewQueryRewardsAuctionCmd()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryRewardsAuctionResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

//
// Transaction CLI Integration Tests
//

func (s *IntegrationTestSuite) TestNewLiquidFarmCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid case: invalid denom",
			[]string{
				strconv.Itoa(1),
				sdk.NewInt64Coin(s.denom1, 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 18,
		},
		{
			"invalid case: minimum farm amount",
			[]string{
				strconv.Itoa(1),
				sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 100).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 18,
		},
		{
			"invalid case: pool id not found",
			[]string{
				strconv.Itoa(10),
				sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(10), 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 38,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewLiquidFarmCmd()
			clientCtx := val.ClientCtx

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewUnfarmCmd() {
	val := s.network.Validators[0]

	_, err := MsgLiquidFarmExec(
		val.ClientCtx,
		val.Address.String(),
		strconv.Itoa(1),
		sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 100_000_000),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				sdk.NewInt64Coin(types.LiquidFarmCoinDenom(1), 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid case: invalid denom",
			[]string{
				strconv.Itoa(1),
				sdk.NewInt64Coin(s.denom1, 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid case: pool id not found",
			[]string{
				strconv.Itoa(10),
				sdk.NewInt64Coin(types.LiquidFarmCoinDenom(10), 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 38,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewLiquidUnfarmCmd()
			clientCtx := val.ClientCtx

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewPlaceBidCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				strconv.Itoa(1),
				sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid case: minimum bid amount",
			[]string{
				strconv.Itoa(1),
				strconv.Itoa(1),
				sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 100).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 18,
		},
		{
			"invalid case: invalid bidding coin denom",
			[]string{
				strconv.Itoa(1),
				strconv.Itoa(1),
				sdk.NewInt64Coin(s.denom1, 1_000_000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewPlaceBidCmd()
			clientCtx := val.ClientCtx

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewRefundBidCmd() {
	val := s.network.Validators[0]

	_, err := MsgPlaceBidExec(
		val.ClientCtx,
		val.Address.String(),
		strconv.Itoa(1),
		strconv.Itoa(1),
		sdk.NewInt64Coin(liquiditytypes.PoolCoinDenom(1), 10_000_000),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid case",
			[]string{
				strconv.Itoa(1),
				strconv.Itoa(1),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 18,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewRefundBidCmd()
			clientCtx := val.ClientCtx

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}
