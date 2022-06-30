package testutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmdb "github.com/tendermint/tm-db"

	chain "github.com/crescent-network/crescent/v2/app"
	"github.com/crescent-network/crescent/v2/app/params"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/client/cli"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewAppConstructor(encodingCfg params.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return chain.NewApp(
			val.Ctx.Logger, tmdb.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(store.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	db := tmdb.NewMemDB()
	cfg := chain.NewConfig(db)
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

func (s *IntegrationTestSuite) TestLiquidStaking() {
	vals := s.network.Validators
	clientCtx := vals[0].ClientCtx

	_, err := clitestutil.ExecTestCLICmd(clientCtx, stakingcli.GetCmdQueryValidators(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)

	whitelist := []types.WhitelistedValidator{
		{
			ValidatorAddress: vals[0].ValAddress.String(),
			TargetWeight:     sdk.NewInt(10),
		},
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
	_, err = MsgParamChangeProposalExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		testutil.WriteToNewTempFile(s.T(), string(paramChangeProp)).Name(),
	)
	s.Require().NoError(err)
	_, err = MsgVote(vals[0].ClientCtx, vals[0].Address.String(), "1", "yes")
	s.Require().NoError(err)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	_, err = clitestutil.ExecTestCLICmd(clientCtx, govcli.GetCmdQueryProposals(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)

	lvs := s.getLiquidValidatorStates()
	s.Require().Len(lvs, 0)

	states := s.getStates()
	s.Require().True(states.BtokenTotalSupply.IsZero())
	s.Require().True(states.TotalLiquidTokens.IsZero())
	s.Require().True(states.TotalDelShares.IsZero())
	s.Require().True(states.NetAmount.IsZero())

	_, err = MsgLiquidStakeExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100000000)).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	lvs = s.getLiquidValidatorStates()
	s.Require().Len(lvs, 1)
	s.Require().True(lvs[0].LiquidTokens.GTE(sdk.NewInt(100000000)))
	s.Require().True(lvs[0].DelShares.GTE(sdk.NewDec(100000000)))
	s.Require().Equal(lvs[0].Status, types.ValidatorStatusActive)
	s.Require().Equal(lvs[0].Weight, sdk.NewInt(10))

	states = s.getStates()
	s.Require().EqualValues(states.BtokenTotalSupply, sdk.NewInt(100000000))
	s.Require().True(states.TotalLiquidTokens.GTE(sdk.NewInt(100000000)))
	s.Require().True(states.TotalDelShares.GTE(sdk.NewDec(100000000)))
	s.Require().True(states.NetAmount.GTE(sdk.NewDec(100000000)))

	_, err = MsgLiquidUnstakeExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(50000000)).String(),
	)
	s.Require().NoError(err)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	lvs = s.getLiquidValidatorStates()
	s.Require().Len(lvs, 1)

	states = s.getStates()
	s.Require().EqualValues(states.BtokenTotalSupply, sdk.NewInt(50000000))
	s.Require().True(states.TotalLiquidTokens.GTE(sdk.NewInt(50000000)))
	s.Require().True(states.TotalDelShares.GTE(sdk.NewDec(50000000)))
	s.Require().True(states.NetAmount.GTE(sdk.NewDec(50000000)))

	_, err = MsgLiquidUnstakeExec(
		vals[0].ClientCtx,
		vals[0].Address.String(),
		sdk.NewCoin(types.DefaultLiquidBondDenom, sdk.NewInt(50000000)).String(),
	)
	s.Require().NoError(err)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	states = s.getStates()
	s.Require().True(states.BtokenTotalSupply.IsZero())
	s.Require().True(states.TotalLiquidTokens.IsZero())
	s.Require().True(states.TotalDelShares.IsZero())
	s.Require().True(states.NetAmount.IsZero())
}

func (s *IntegrationTestSuite) getStates() types.NetAmountState {
	ctx := s.network.Validators[0].ClientCtx
	var states types.QueryStatesResponse
	out, err := clitestutil.ExecTestCLICmd(ctx, cli.GetCmdQueryStates(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(out.Bytes(), &states), out.String())
	return states.NetAmountState
}

func (s *IntegrationTestSuite) getLiquidValidatorStates() []types.LiquidValidatorState {
	ctx := s.network.Validators[0].ClientCtx
	var liquidValsResult types.QueryLiquidValidatorsResponse
	out, err := clitestutil.ExecTestCLICmd(ctx, cli.GetCmdQueryLiquidValidators(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(out.Bytes(), &liquidValsResult), out.String())
	return liquidValsResult.GetLiquidValidators()
}
