package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)).String()),
}

func MsgLiquidFarmExec(clientCtx client.Context, from, poolId string, amount sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		poolId,
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLiquidFarmCmd(), args)
}

func MsgPlaceBidExec(clientCtx client.Context, from, auctionId, poolId string, amount sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		auctionId,
		poolId,
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewPlaceBidCmd(), args)
}
