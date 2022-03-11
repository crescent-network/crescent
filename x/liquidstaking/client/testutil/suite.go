package testutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmdb "github.com/tendermint/tm-db"

	chain "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/client/cli"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	db := tmdb.NewMemDB()
	cfg := chain.NewConfig(db)
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
			`{"liquid_bond_denom":"bstake","whitelisted_validators":[],"unstake_fee_rate":"0.000000000000000000","min_liquid_staking_amount":"1000000"}`,
		},
		{
			"text output",
			[]string{},
			`liquid_bond_denom: bstake
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

	whitelist := types.WhitelistedValidators{
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
			Key:      string(types.KeyWhitelistedValidators),
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
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryStates(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
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

	out, err = clitestutil.ExecTestCLICmd(clientCtx, cli.GetCmdQueryStates(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
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
