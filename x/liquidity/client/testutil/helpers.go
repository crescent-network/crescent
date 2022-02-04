package testutil

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidity/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)).String()),
}

func MsgCreatePair(clientCtx client.Context, from, baseCoinDenom, quoteCoinDenom string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		baseCoinDenom,
		quoteCoinDenom,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewCreatePairCmd(), args)
}

func MsgCreatePool(clientCtx client.Context, from string, pairId uint64, depositCoins sdk.Coins, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		strconv.FormatUint(pairId, 10),
		depositCoins.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewCreatePoolCmd(), args)
}
