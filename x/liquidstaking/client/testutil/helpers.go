package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	paramscli "github.com/cosmos/cosmos-sdk/x/params/client/cli"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))).String()),
	fmt.Sprintf("--%s=%s", flags.FlagGas, "1000000"),
	//fmt.Sprintf("--%s=%s", flags.FlagLogLevel, "trace"),
}

// MsgSubmitProposal creates a tx for submit proposal
func MsgSubmitProposal(clientCtx client.Context, from, title, description, proposalType string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append([]string{
		fmt.Sprintf("--%s=%s", govcli.FlagTitle, title),
		fmt.Sprintf("--%s=%s", govcli.FlagDescription, description),
		fmt.Sprintf("--%s=%s", govcli.FlagProposalType, proposalType),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, govcli.NewCmdSubmitProposal(), args)
}

// MsgParamChangeProposalExec creates a transaction for submitting param change proposal
func MsgParamChangeProposalExec(clientCtx client.Context, from string, file string) (testutil.BufferWriter, error) {

	args := append([]string{
		file,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	paramChangeCmd := paramscli.NewSubmitParamChangeProposalTxCmd()
	flags.AddTxFlagsToCmd(paramChangeCmd)

	return clitestutil.ExecTestCLICmd(clientCtx, paramChangeCmd, args)
}

// MsgVote votes for a proposal
func MsgVote(clientCtx client.Context, from, id, vote string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append([]string{
		id,
		vote,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, govcli.NewCmdWeightedVote(), args)
}

// MsgLiquidStakeExec creates a transaction for liquid-staking coin.
func MsgLiquidStakeExec(clientCtx client.Context, from string, stakingCoin string,
	extraArgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		stakingCoin,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, commonArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLiquidStakeCmd(), args)
}

// MsgLiquidUnstakeExec creates a transaction for liquid-unstaking coin.
func MsgLiquidUnstakeExec(clientCtx client.Context, from string, unstakingCoin string,
	extraArgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		unstakingCoin,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLiquidUnstakeCmd(), args)
}
