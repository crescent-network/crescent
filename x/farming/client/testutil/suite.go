package testutil

import (
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	utilcli "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmdb "github.com/tendermint/tm-db"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming/client/cli"
	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

// SetupTest creates a new network for _each_ integration test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up integration test suite")

	keeper.EnableAdvanceEpoch = true
	keeper.EnableRatioPlan = true

	db := tmdb.NewMemDB()
	cfg := chain.NewConfig(db)
	cfg.NumValidators = 1

	var genesisState types.GenesisState
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &genesisState)
	s.Require().NoError(err)

	genesisState.Params = types.DefaultParams()
	cfg.GenesisState[types.ModuleName] = cfg.Codec.MustMarshalJSON(&genesisState)
	cfg.AccountTokens = sdk.NewInt(100_000_000_000) // node0token denom
	cfg.StakingTokens = sdk.NewInt(100_000_000_000) // stake denom

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownTest cleans up the current test network after each test in the suite.
func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestNewCreateFixedAmountPlanCmd() {
	val := s.network.Validators[0]

	name := "test"
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "node0token",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// happy case
	case1 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000_000)),
	}

	// invalid name
	case2 := cli.PrivateFixedPlanRequest{
		Name: `OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`,
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000_000)),
	}

	// invalid staking coin weights
	case3 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000_000)),
	}

	// invalid staking coin weights
	case4 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("node0token", sdk.NewInt(2))),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000_000)),
	}

	// invalid start time and end time
	case5 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("2021-08-13T00:00:00Z"),
		EndTime:            types.ParseTime("2021-08-06T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000_000)),
	}

	// invalid epoch amount
	case6 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("stake", 0)),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"valid transaction",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case1.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid name case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case2.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin weights case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case3.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin weights case #2",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case4.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid start time & end time case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case5.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 1,
		},
		{
			"invalid epoch amount case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case6.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreateFixedAmountPlanCmd()
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

func (s *IntegrationTestSuite) TestNewCreateRatioPlanCmd() {
	val := s.network.Validators[0]

	name := "test"
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "node0token",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// happy case
	case1 := cli.PrivateRatioPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("0.1"),
	}

	// invalid name
	case2 := cli.PrivateRatioPlanRequest{
		Name: `OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`,
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("0.1"),
	}

	// invalid staking coin weights
	case3 := cli.PrivateRatioPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("0.1"),
	}

	// invalid staking coin weights
	case4 := cli.PrivateRatioPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("node0token", sdk.NewInt(2))),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("0.1"),
	}

	// invalid start time and end time
	case5 := cli.PrivateRatioPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("2021-08-13T00:00:00Z"),
		EndTime:            types.ParseTime("2021-08-06T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("0.1"),
	}

	// invalid epoch ratio
	case6 := cli.PrivateRatioPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochRatio:         sdk.MustNewDecFromStr("1.1"),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"valid transaction",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case1.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid name case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case2.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin weights case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case3.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin weights case #2",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case4.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid start time & end time case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case5.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid epoch ratio case #1",
			[]string{
				testutil.WriteToNewTempFile(s.T(), case6.String()).Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreateRatioPlanCmd()
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

func (s *IntegrationTestSuite) TestNewStakeCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"valid transaction case #1",
			[]string{
				sdk.NewInt64Coin("stake", 100000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid transaction case #2",
			[]string{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 100000), sdk.NewInt64Coin("node0token", 100000)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin case #1",
			[]string{
				sdk.NewInt64Coin("stake", 0).String(),
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
			cmd := cli.NewStakeCmd()
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

func (s *IntegrationTestSuite) TestNewUnstakeCmd() {
	val := s.network.Validators[0]

	_, err := MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoins(
			sdk.NewInt64Coin("stake", 10_000_000),
			sdk.NewInt64Coin("node0token", 10_000_000),
		).String(),
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
			"valid transaction case #1",
			[]string{
				sdk.NewInt64Coin("stake", 100000).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid transaction case #2",
			[]string{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 100000), sdk.NewInt64Coin("node0token", 100000)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid unstaking coin case #1",
			[]string{
				sdk.NewInt64Coin("stake", 0).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 18,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewUnstakeCmd()
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

func (s *IntegrationTestSuite) TestNewHarvestCmd() {
	val := s.network.Validators[0]

	req := cli.PrivateFixedPlanRequest{
		Name:               "test",
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("stake", sdk.NewInt(1))),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("node0token", 100_000_000)),
	}

	// create a fixed amount plan
	_, err := MsgCreateFixedAmountPlanExec(
		val.ClientCtx,
		val.Address.String(),
		testutil.WriteToNewTempFile(s.T(), req.String()).Name(),
	)
	s.Require().NoError(err)

	// stake coin
	_, err = MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoins(
			sdk.NewInt64Coin("stake", 10_000_000),
			sdk.NewInt64Coin("node0token", 10_000_000),
		).String(),
	)
	s.Require().NoError(err)

	// advance epoch by 1
	_, err = MsgAdvanceEpochExec(
		val.ClientCtx,
		val.Address.String(),
	)
	s.Require().NoError(err)

	// advance epoch by 1
	_, err = MsgAdvanceEpochExec(
		val.ClientCtx,
		val.Address.String(),
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
			"valid transaction case #1",
			[]string{
				"stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid transaction case #2",
			[]string{
				"stake,node0token",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"valid transaction case #3",
			[]string{
				fmt.Sprintf("--%s", cli.FlagAll),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin denoms case #1",
			[]string{
				"invaliddenom!",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			true, &sdk.TxResponse{}, 18,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewHarvestCmd()
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

type QueryCmdTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *QueryCmdTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	keeper.EnableAdvanceEpoch = true

	db := tmdb.NewMemDB()
	cfg := chain.NewConfig(db)
	cfg.NumValidators = 1

	var genesisState types.GenesisState
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &genesisState)
	s.Require().NoError(err)

	genesisState.Params = types.DefaultParams()
	cfg.GenesisState[types.ModuleName] = cfg.Codec.MustMarshalJSON(&genesisState)
	cfg.AccountTokens = sdk.NewInt(100_000_000_000) // node0token denom
	cfg.StakingTokens = sdk.NewInt(100_000_000_000) // stake denom

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	req := cli.PrivateFixedPlanRequest{
		Name:               "test",
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewInt64DecCoin(sdk.DefaultBondDenom, 1)),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("node0token", 100_000_000)),
	}

	// create a fixed amount plan
	_, err = MsgCreateFixedAmountPlanExec(
		val.ClientCtx,
		val.Address.String(),
		testutil.WriteToNewTempFile(s.T(), req.String()).Name(),
	)
	s.Require().NoError(err)

	// query the farming pool address that is assigned to the pool and
	// transfer some amount of coins to the address
	s.fundFarmingPool(1, sdk.NewCoins(sdk.NewInt64Coin("node0token", 1_000_000_000)))

	_, err = MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)).String(),
	)
	s.Require().NoError(err)

	_, err = MsgAdvanceEpochExec(val.ClientCtx, val.Address.String())
	s.Require().NoError(err)

	_, err = MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500000)).String(),
	)
	s.Require().NoError(err)

	_, err = MsgAdvanceEpochExec(val.ClientCtx, val.Address.String())
	s.Require().NoError(err)

	_, err = MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 500000)).String(),
	)
	s.Require().NoError(err)
}

func (s *QueryCmdTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *QueryCmdTestSuite) TestCmdQueryParams() {
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
			cmd := cli.GetCmdQueryParams()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				s.Require().NoError(err)

				var params types.Params
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &params))
				s.Require().NotEmpty(params.FarmingFeeCollector)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryPlans() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	types.RegisterInterfaces(clientCtx.InterfaceRegistry)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryPlansResponse)
	}{
		{
			"happy case",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryPlansResponse) {
				plans, err := types.UnpackPlans(resp.Plans)
				s.Require().NoError(err)
				s.Require().Len(plans, 1)
				s.Require().Equal(uint64(1), plans[0].GetId())
			},
		},
		{
			"invalid plan type",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPlanType, "invalid"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
		{
			"invalid farming pool addr",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagFarmingPoolAddr, "invalid"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
		{
			"invalid termination addr",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagTerminationAddr, "invalid"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
		{
			"invalid staking coin denom",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagStakingCoinDenom, "!"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPlans()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryPlansResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryPlan() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	types.RegisterInterfaces(clientCtx.InterfaceRegistry)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryPlanResponse)
	}{
		{
			"happy case",
			[]string{
				strconv.Itoa(1),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryPlanResponse) {
				plan, err := types.UnpackPlan(resp.Plan)
				s.Require().NoError(err)
				s.Require().Equal(uint64(1), plan.GetId())
			},
		},
		{
			"id not found",
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
			cmd := cli.GetCmdQueryPlan()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryPlanResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryPosition() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryPositionResponse)
	}{
		{
			"happy case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryPositionResponse) {
				s.Require().True(coinsEq(utils.ParseCoins("1500000stake"), resp.StakedCoins))
				s.Require().True(coinsEq(utils.ParseCoins("500000stake"), resp.QueuedCoins))
				s.Require().True(coinsEq(utils.ParseCoins("199999999node0token"), resp.Rewards))
			},
		},
		{
			"invalid farmer addr",
			[]string{
				"invalid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPosition()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryPositionResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryStakings() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryStakingsResponse)
	}{
		{
			"happy case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryStakingsResponse) {
				s.Require().Len(resp.Stakings, 1)
				s.Require().Equal(sdk.DefaultBondDenom, resp.Stakings[0].StakingCoinDenom)
				s.Require().True(intEq(sdk.NewInt(1500000), resp.Stakings[0].Amount))
				s.Require().EqualValues(2, resp.Stakings[0].StartingEpoch)
			},
		},
		{
			"invalid farmer addr",
			[]string{
				"invalid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryStakings()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryStakingsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryQueuedStakings() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryQueuedStakingsResponse)
	}{
		{
			"happy case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryQueuedStakingsResponse) {
				s.Require().Len(resp.QueuedStakings, 1)
				s.Require().Equal(sdk.DefaultBondDenom, resp.QueuedStakings[0].StakingCoinDenom)
				s.Require().True(intEq(sdk.NewInt(500000), resp.QueuedStakings[0].Amount))
				// Omitted EndTime check.
			},
		},
		{
			"invalid farmer addr",
			[]string{
				"invalid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryQueuedStakings()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryQueuedStakingsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryTotalStakings() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryTotalStakingsResponse)
	}{
		{
			"happy case",
			[]string{
				sdk.DefaultBondDenom,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryTotalStakingsResponse) {
				s.Require().True(intEq(sdk.NewInt(1500000), resp.Amount))
			},
		},
		{
			"invalid staking coin denom",
			[]string{
				"!",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryTotalStakings()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryTotalStakingsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryRewards() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryRewardsResponse)
	}{
		{
			"happy case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryRewardsResponse) {
				s.Require().Len(resp.Rewards, 1)
				s.Require().Equal("stake", resp.Rewards[0].StakingCoinDenom)
				s.Require().True(coinsEq(sdk.NewCoins(sdk.NewInt64Coin("node0token", 99_999_999)), resp.Rewards[0].Rewards))
			},
		},
		{
			"invalid farmer addr",
			[]string{
				"invalid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryRewards()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryRewardsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryUnharvestedRewards() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryUnharvestedRewardsResponse)
	}{
		{
			"happy case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryUnharvestedRewardsResponse) {
				s.Require().Len(resp.UnharvestedRewards, 1)
				s.Require().Equal(sdk.DefaultBondDenom, resp.UnharvestedRewards[0].StakingCoinDenom)
				s.Require().True(coinsEq(utils.ParseCoins("100000000node0token"), resp.UnharvestedRewards[0].Rewards))
			},
		},
		{
			"invalid farmer addr",
			[]string{
				"invalid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryUnharvestedRewards()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryUnharvestedRewardsResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) TestCmdQueryCurrentEpochDays() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		postRun   func(*types.QueryCurrentEpochDaysResponse)
	}{
		{
			"happy case",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			func(resp *types.QueryCurrentEpochDaysResponse) {
				s.Require().Equal(uint32(1), resp.CurrentEpochDays)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryCurrentEpochDays()

			out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryCurrentEpochDaysResponse
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
				tc.postRun(&resp)
			}
		})
	}
}

func (s *QueryCmdTestSuite) fundFarmingPool(planId uint64, amount sdk.Coins) {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	types.RegisterInterfaces(clientCtx.InterfaceRegistry)

	cmd := cli.GetCmdQueryPlan()
	args := []string{
		strconv.FormatUint(planId, 10),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	out, err := utilcli.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	var resp types.QueryPlanResponse
	clientCtx.Codec.MustUnmarshalJSON(out.Bytes(), &resp)
	plan, err := types.UnpackPlan(resp.Plan)
	s.Require().NoError(err)

	_, err = MsgSendExec(
		val.ClientCtx,
		val.Address.String(),
		plan.GetFarmingPoolAddress().String(),
		amount.String(),
	)
	s.Require().NoError(err)
}

func intEq(exp, got sdk.Int) (bool, string, string, string) {
	return exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
