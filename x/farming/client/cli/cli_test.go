// +build norace

package cli_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmdb "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/client/cli"
	farmingtestutil "github.com/tendermint/farming/x/farming/client/testutil"
	"github.com/tendermint/farming/x/farming/types"
	farmingtypes "github.com/tendermint/farming/x/farming/types"
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

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownTest cleans up the current test network after each test in the suite.
func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TestIntegrationTestSuite every integration test suite.
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestNewCreateFixedAmountPlanCmd() {
	val := s.network.Validators[0]

	name := "test"
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// happy case
	case1 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
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
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
	}

	// invalid staking coin weights
	case3 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
	}

	// invalid staking coin weights
	case4 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4", sdk.NewInt(2))),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
	}

	// invalid start time and end time
	case5 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("2021-08-13T00:00:00Z"),
		EndTime:            types.ParseTime("2021-08-06T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
	}

	// invalid epoch amount
	case6 := cli.PrivateFixedPlanRequest{
		Name:               name,
		StakingCoinWeights: coinWeights,
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

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
			Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
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
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4", sdk.NewInt(2))),
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

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
			"valid transaction",
			[]string{
				sdk.NewCoin("stake", sdk.NewInt(100000)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid staking coin case #1",
			[]string{
				sdk.NewCoin("stake", sdk.NewInt(0)).String(),
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
			cmd := cli.NewStakeCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

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

	_, err := farmingtestutil.MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoin("stake", sdk.NewInt(10_000_000)).String(),
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
			"valid transaction",
			[]string{
				sdk.NewCoin("stake", sdk.NewInt(100000)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"invalid unstaking coin case #1",
			[]string{
				sdk.NewCoin("stake", sdk.NewInt(0)).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 18,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewUnstakeCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				fmt.Println(txResp)
				fmt.Println(out.String())
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewHarvestCmd() {
	val := s.network.Validators[0]

	// create fixed amount plan
	req := cli.PrivateFixedPlanRequest{
		Name:               "test",
		StakingCoinWeights: sdk.NewDecCoins(sdk.NewDecCoin("stake", sdk.NewInt(1))),
		StartTime:          types.ParseTime("0001-01-01T00:00:00Z"),
		EndTime:            types.ParseTime("9999-01-01T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("node0token", 100_000_000)),
	}

	_, err := farmingtestutil.MsgCreateFixedAmountPlanExec(
		val.ClientCtx,
		val.Address.String(),
		testutil.WriteToNewTempFile(s.T(), req.String()).Name(),
	)
	s.Require().NoError(err)

	// stake coins
	_, err = farmingtestutil.MsgStakeExec(
		val.ClientCtx,
		val.Address.String(),
		sdk.NewCoin("stake", sdk.NewInt(10_000_000)).String(),
	)
	s.Require().NoError(err)

	// TODO: right now, there is no command-line interface that triggers keeeper
	// to increase epoch days by 2 for reward distribution.
	// handle invalid cases for now

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid transaction for no reward for staking coin denom stake",
			[]string{
				"stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"invalid staking coin denoms case #1",
			[]string{
				"!",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 18,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewHarvestCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

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
