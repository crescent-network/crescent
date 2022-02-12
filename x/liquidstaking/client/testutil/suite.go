package testutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/client/cli"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
	"github.com/stretchr/testify/suite"
	tmdb "github.com/tendermint/tm-db"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	db := tmdb.NewMemDB()
	cfg := app.NewConfig(db)
	//cfg.TimeoutCommit = 3 * time.Second
	cfg.NumValidators = 1
	s.cfg = cfg

	genesisStateLiquidStaking := types.DefaultGenesisState()
	genesisStateLiquidStaking.Params.UnstakeFeeRate = sdk.ZeroDec()
	bz, _ := cfg.Codec.MarshalJSON(genesisStateLiquidStaking)
	cfg.GenesisState["liquidstaking"] = bz

	genesisStateGov := govtypes.DefaultGenesisState()
	genesisStateGov.DepositParams = govtypes.NewDepositParams(sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, govtypes.DefaultMinDepositTokens)), time.Duration(15)*time.Second)
	genesisStateGov.VotingParams = govtypes.NewVotingParams(time.Duration(3) * time.Second)
	genesisStateGov.TallyParams.Quorum = sdk.MustNewDecFromStr("0.01")
	bz, err := cfg.Codec.MarshalJSON(genesisStateGov)
	s.Require().NoError(err)
	cfg.GenesisState["gov"] = bz

	//var genesisState types.GenesisState
	//err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[types.ModuleName], &genesisState)
	//s.Require().NoError(err)
	//
	//genesisState.Params = types.DefaultParams()
	//cfg.GenesisState[types.ModuleName] = cfg.Codec.MustMarshalJSON(&genesisState)
	//cfg.AccountTokens = sdk.NewInt(100_000_000_000) // node0token denom
	//cfg.StakingTokens = sdk.NewInt(100_000_000_000) // stake denom

	s.network = network.New(s.T(), s.cfg)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCmdParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"bonded_bond_denom":"bstake","whitelisted_validators":[],"unstake_fee_rate":"0.000000000000000000","min_liquid_staking_amount":"1000000"}`,
		},
		{
			"text output",
			[]string{},
			`bonded_bond_denom: bstake
min_liquid_staking_amount: "1000000"
unstake_fee_rate: "0.000000000000000000"
whitelisted_validators: []
`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
		})
	}
}

// TODO: WIP add assertion
func (s *IntegrationTestSuite) TestLiquidStaking() {
	vals := s.network.Validators
	clientCtx := vals[0].ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, stakingcli.GetCmdQueryValidators(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)
	fmt.Println(out)

	whitelist := []types.WhitelistedValidator{
		{
			ValidatorAddress: vals[0].ValAddress.String(),
			TargetWeight:     sdk.NewInt(10),
		},
		//{
		//	ValidatorAddress: vals[2].ValAddress.String(),
		//	TargetWeight:     sdk.NewInt(10),
		//},
		//{
		//	ValidatorAddress: vals[3].ValAddress.String(),
		//	TargetWeight:     sdk.NewInt(10),
		//},
		//{
		//	ValidatorAddress: vals[4].ValAddress.String(),
		//	TargetWeight:     sdk.NewInt(10),
		//},
	}
	whitelistStr, err := json.Marshal(&whitelist)
	if err != nil {
		panic(err)
	}

	paramChange := paramscutils.ParamChangeProposalJSON{
		Title:       "test",
		Description: "test",
		Changes: []paramscutils.ParamChangeJSON{{
			Subspace: types.ModuleName,
			Key:      "WhitelistedValidators",
			Value:    whitelistStr,
		},
		},
		Deposit: sdk.NewCoin(s.cfg.BondDenom, govtypes.DefaultMinDepositTokens).String(),
	}
	paramChangeProp, err := json.Marshal(&paramChange)
	if err != nil {
		panic(err)
	}

	//create a proposal with deposit
	res, err := MsgParamChangeProposalExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		testutil.WriteToNewTempFile(s.T(), string(paramChangeProp)).Name(),
	)
	fmt.Println(res, err)
	s.Require().NoError(err)
	_, err = MsgVote(vals[0].ClientCtx, vals[0].Address.String(), "1", "yes")
	s.Require().NoError(err)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)
	//_, err = MsgVote(vals[0].ClientCtx, vals[1].Address.String(), "1", "yes")
	//s.Require().NoError(err)
	//_, err = MsgVote(vals[2].ClientCtx, vals[2].Address.String(), "1", "yes")
	//s.Require().NoError(err)

	out, err = clitestutil.ExecTestCLICmd(clientCtx, govcli.GetCmdQueryProposals(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryLiquidValidators(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryParams(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)

	res, err = MsgLiquidStakeExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100000000)).String(),
	)
	fmt.Println(res, err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	out, err = clitestutil.ExecTestCLICmd(clientCtx, stakingcli.GetCmdQueryDelegations(), []string{types.LiquidStakingProxyAcc.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)

	out, err = clitestutil.ExecTestCLICmd(clientCtx, bankcli.GetBalancesCmd(), []string{vals[0].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryLiquidValidators(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	fmt.Println(out, err)

	// TODO: fix timed out waiting for tx to be included in a block
	//res, err = MsgLiquidUnstakeExec(
	//	vals[0].ClientCtx,
	//	vals[0].Address.String(),
	//	sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(100000000)).String(),
	//)
	//fmt.Println(res, err)
	//s.Require().NoError(err)
	//fmt.Println(res)
	//err = s.network.WaitForNextBlock()
	//s.Require().NoError(err)

	//_, err = s.network.WaitForHeight(1)
	//s.Require().NoError(err)

	//// create a proposal without deposit
	//_, err = MsgSubmitProposal(val.ClientCtx, val.Address.String(),
	//	"Text Proposal 2", "Where is the title!?", govtypes.ProposalTypeText)
	//s.Require().NoError(err)
	//_, err = s.network.WaitForHeight(1)
	//s.Require().NoError(err)
	//
	//// create a proposal3 with deposit
	//_, err = MsgSubmitProposal(val.ClientCtx, val.Address.String(),
	//	"Text Proposal 3", "Where is the title!?", govtypes.ProposalTypeText,
	//	fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, govtypes.DefaultMinDepositTokens).String()))
	//s.Require().NoError(err)
	//_, err = s.network.WaitForHeight(1)
	//s.Require().NoError(err)
	//
	//// vote for proposal3 as val
	//_, err = MsgVote(val.ClientCtx, val.Address.String(), "3", "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05")
	//s.Require().NoError(err)
}

//func (s *IntegrationTestSuite) TestCmdGovParams() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name           string
//		args           []string
//		expectedOutput string
//	}{
//		{
//			"json output",
//			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
//			`{"voting_params":{"voting_period":"172800000000000"},"tally_params":{"quorum":"0.334000000000000000","threshold":"0.500000000000000000","veto_threshold":"0.334000000000000000"},"deposit_params":{"min_deposit":[{"denom":"stake","amount":"10000000"}],"max_deposit_period":"172800000000000"}}`,
//		},
//		{
//			"text output",
//			[]string{},
//			`
//deposit_params:
//  max_deposit_period: "172800000000000"
//  min_deposit:
//  - amount: "10000000"
//    denom: stake
//tally_params:
//  quorum: "0.334000000000000000"
//  threshold: "0.500000000000000000"
//  veto_threshold: "0.334000000000000000"
//voting_params:
//  voting_period: "172800000000000"
//	`,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryParams()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//			s.Require().NoError(err)
//			s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdProposer() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name           string
//		args           []string
//		expectErr      bool
//		expectedOutput string
//	}{
//		{
//			"without proposal id",
//			[]string{
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			true,
//			``,
//		},
//		{
//			"json output",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//			fmt.Sprintf("{\"proposal_id\":\"%s\",\"proposer\":\"%s\"}", "1", val.Address.String()),
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryProposer()
//			clientCtx := val.ClientCtx
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				s.Require().Equal(strings.TrimSpace(tc.expectedOutput), strings.TrimSpace(out.String()))
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdTally() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name           string
//		args           []string
//		expectErr      bool
//		expectedOutput govtypes.TallyResult
//	}{
//		{
//			"without proposal id",
//			[]string{
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			true,
//			govtypes.TallyResult{},
//		},
//		{
//			"json output",
//			[]string{
//				"2",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//			govtypes.NewTallyResult(sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0)),
//		},
//		{
//			"json output",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//			govtypes.NewTallyResult(s.cfg.BondedTokens, sdk.NewInt(0), sdk.NewInt(0), sdk.NewInt(0)),
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryTally()
//			clientCtx := val.ClientCtx
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				var tally govtypes.TallyResult
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &tally), out.String())
//				s.Require().Equal(tally, tc.expectedOutput)
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestNewCmdSubmitProposal() {
//	val := s.network.Validators[0]
//	invalidProp := `{
//  "title": "",
//	"description": "Where is the title!?",
//	"type": "Text",
//  "deposit": "-324foocoin"
//}`
//	invalidPropFile := testutil.WriteToNewTempFile(s.T(), invalidProp)
//	validProp := fmt.Sprintf(`{
//  "title": "Text Proposal",
//	"description": "Hello, World!",
//	"type": "Text",
//  "deposit": "%s"
//}`, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)))
//	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
//	testCases := []struct {
//		name         string
//		args         []string
//		expectErr    bool
//		expectedCode uint32
//		respType     proto.Message
//	}{
//		{
//			"invalid proposal (file)",
//			[]string{
//				fmt.Sprintf("--%s=%s", govcli.FlagProposal, invalidPropFile.Name()),
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			true, 0, nil,
//		},
//		{
//			"invalid proposal",
//			[]string{
//				fmt.Sprintf("--%s='Where is the title!?'", govcli.FlagDescription),
//				fmt.Sprintf("--%s=%s", govcli.FlagProposalType, govtypes.ProposalTypeText),
//				fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			true, 0, nil,
//		},
//		{
//			"valid transaction (file)",
//			[]string{
//				fmt.Sprintf("--%s=%s", govcli.FlagProposal, validPropFile.Name()),
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0, &sdk.TxResponse{},
//		},
//		{
//			"valid transaction",
//			[]string{
//				fmt.Sprintf("--%s='Text Proposal'", govcli.FlagTitle),
//				fmt.Sprintf("--%s='Where is the title!?'", govcli.FlagDescription),
//				fmt.Sprintf("--%s=%s", govcli.FlagProposalType, govtypes.ProposalTypeText),
//				fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)).String()),
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0, &sdk.TxResponse{},
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.NewCmdSubmitProposal()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
//				txResp := tc.respType.(*sdk.TxResponse)
//				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdGetProposal() {
//	val := s.network.Validators[0]
//
//	title := "Text Proposal 1"
//
//	testCases := []struct {
//		name      string
//		args      []string
//		expectErr bool
//	}{
//		{
//			"get non existing proposal",
//			[]string{
//				"10",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			true,
//		},
//		{
//			"get proposal with json response",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryProposal()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				var proposal govtypes.Proposal
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposal), out.String())
//				s.Require().Equal(title, proposal.GetTitle())
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdGetProposals() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name      string
//		args      []string
//		expectErr bool
//	}{
//		{
//			"get proposals as json response",
//			[]string{
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//		},
//		{
//			"get proposals with invalid status",
//			[]string{
//				"--status=unknown",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			true,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryProposals()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				var proposals govtypes.QueryProposalsResponse
//
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposals), out.String())
//				s.Require().Len(proposals.Proposals, 3)
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdQueryDeposits() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name      string
//		args      []string
//		expectErr bool
//	}{
//		{
//			"get deposits of non existing proposal",
//			[]string{
//				"10",
//			},
//			true,
//		},
//		{
//			"get deposits(valid req)",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryDeposits()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//
//				var deposits govtypes.QueryDepositsResponse
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &deposits), out.String())
//				s.Require().Len(deposits.Deposits, 1)
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdQueryDeposit() {
//	val := s.network.Validators[0]
//	depositAmount := sdk.NewCoin(s.cfg.BondDenom, govtypes.DefaultMinDepositTokens)
//
//	testCases := []struct {
//		name      string
//		args      []string
//		expectErr bool
//	}{
//		{
//			"get deposit with no depositer",
//			[]string{
//				"1",
//			},
//			true,
//		},
//		{
//			"get deposit with wrong deposit address",
//			[]string{
//				"1",
//				"wrong address",
//			},
//			true,
//		},
//		{
//			"get deposit (valid req)",
//			[]string{
//				"1",
//				val.Address.String(),
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryDeposit()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//
//				var deposit govtypes.Deposit
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &deposit), out.String())
//				s.Require().Equal(depositAmount.String(), deposit.Amount.String())
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestNewCmdDeposit() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name         string
//		args         []string
//		expectErr    bool
//		expectedCode uint32
//	}{
//		{
//			"without proposal id",
//			[]string{
//				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			true, 0,
//		},
//		{
//			"without deposit amount",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			true, 0,
//		},
//		{
//			"deposit on non existing proposal",
//			[]string{
//				"10",
//				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 2,
//		},
//		{
//			"deposit on non existing proposal",
//			[]string{
//				"1",
//				sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10)).String(), // 10stake
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		var resp sdk.TxResponse
//
//		s.Run(tc.name, func() {
//			cmd := govcli.NewCmdDeposit()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
//				s.Require().Equal(tc.expectedCode, resp.Code, out.String())
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdQueryVotes() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name      string
//		args      []string
//		expectErr bool
//	}{
//		{
//			"get votes with no proposal id",
//			[]string{},
//			true,
//		},
//		{
//			"get votes of non existed proposal",
//			[]string{
//				"10",
//			},
//			true,
//		},
//		{
//			"vote for invalid proposal",
//			[]string{
//				"1",
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryVotes()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//
//				var votes govtypes.QueryVotesResponse
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &votes), out.String())
//				s.Require().Len(votes.Votes, 1)
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestCmdQueryVote() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name           string
//		args           []string
//		expectErr      bool
//		expVoteOptions govtypes.WeightedVoteOptions
//	}{
//		{
//			"get vote of non existing proposal",
//			[]string{
//				"10",
//				val.Address.String(),
//			},
//			true,
//			govtypes.NewNonSplitVoteOption(govtypes.OptionYes),
//		},
//		{
//			"get vote by wrong voter",
//			[]string{
//				"1",
//				"wrong address",
//			},
//			true,
//			govtypes.NewNonSplitVoteOption(govtypes.OptionYes),
//		},
//		{
//			"vote for valid proposal",
//			[]string{
//				"1",
//				val.Address.String(),
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//			govtypes.NewNonSplitVoteOption(govtypes.OptionYes),
//		},
//		{
//			"split vote for valid proposal",
//			[]string{
//				"3",
//				val.Address.String(),
//				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
//			},
//			false,
//			govtypes.WeightedVoteOptions{
//				govtypes.WeightedVoteOption{Option: govtypes.OptionYes, Weight: sdk.NewDecWithPrec(60, 2)},
//				govtypes.WeightedVoteOption{Option: govtypes.OptionNo, Weight: sdk.NewDecWithPrec(30, 2)},
//				govtypes.WeightedVoteOption{Option: govtypes.OptionAbstain, Weight: sdk.NewDecWithPrec(5, 2)},
//				govtypes.WeightedVoteOption{Option: govtypes.OptionNoWithVeto, Weight: sdk.NewDecWithPrec(5, 2)},
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.GetCmdQueryVote()
//			clientCtx := val.ClientCtx
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//
//				var vote govtypes.Vote
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &vote), out.String())
//				s.Require().Equal(len(vote.Options), len(tc.expVoteOptions))
//				for i, option := range tc.expVoteOptions {
//					s.Require().Equal(option.Option, vote.Options[i].Option)
//					s.Require().True(option.Weight.Equal(vote.Options[i].Weight))
//				}
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestNewCmdVote() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name         string
//		args         []string
//		expectErr    bool
//		expectedCode uint32
//	}{
//		{
//			"invalid vote",
//			[]string{},
//			true, 0,
//		},
//		{
//			"vote for invalid proposal",
//			[]string{
//				"10",
//				"yes",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 2,
//		},
//		{
//			"valid vote",
//			[]string{
//				"1",
//				"yes",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.NewCmdVote()
//			clientCtx := val.ClientCtx
//			var txResp sdk.TxResponse
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
//				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
//			}
//		})
//	}
//}
//
//func (s *IntegrationTestSuite) TestNewCmdWeightedVote() {
//	val := s.network.Validators[0]
//
//	testCases := []struct {
//		name         string
//		args         []string
//		expectErr    bool
//		expectedCode uint32
//	}{
//		{
//			"invalid vote",
//			[]string{},
//			true, 0,
//		},
//		{
//			"vote for invalid proposal",
//			[]string{
//				"10",
//				"yes",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 2,
//		},
//		{
//			"valid vote",
//			[]string{
//				"1",
//				"yes",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0,
//		},
//		{
//			"invalid valid split vote string",
//			[]string{
//				"1",
//				"yes/0.6,no/0.3,abstain/0.05,no_with_veto/0.05",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			true, 0,
//		},
//		{
//			"valid split vote",
//			[]string{
//				"1",
//				"yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05",
//				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
//				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
//				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
//				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
//			},
//			false, 0,
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//		s.Run(tc.name, func() {
//			cmd := govcli.NewCmdWeightedVote()
//			clientCtx := val.ClientCtx
//			var txResp sdk.TxResponse
//
//			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
//
//			if tc.expectErr {
//				s.Require().Error(err)
//			} else {
//				s.Require().NoError(err)
//				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
//				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
//			}
//		})
//	}
//}
