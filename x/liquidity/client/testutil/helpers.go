package testutil

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidity/client/cli"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
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

//nolint
func MsgDeposit(clientCtx client.Context, from string, poolId uint64, depositCoins sdk.Coins, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		strconv.FormatUint(poolId, 10),
		depositCoins.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewDepositCmd(), args)
}

//nolint
func MsgWithdraw(clientCtx client.Context, from string, poolId uint64, poolCoin sdk.Coin, extraArgs ...string) (testutil.BufferWriter, error) {
	args := append(append([]string{
		strconv.FormatUint(poolId, 10),
		poolCoin.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewWithdrawCmd(), args)
}

func MsgLimitOrder(
	clientCtx client.Context, from string, pairId uint64, dir types.OrderDirection, offerCoin sdk.Coin,
	demandCoinDenom string, price sdk.Dec, amt sdk.Int, orderLifespan time.Duration, extraArgs ...string) (testutil.BufferWriter, error) {
	var dirStr string
	switch dir {
	case types.OrderDirectionBuy:
		dirStr = "buy"
	case types.OrderDirectionSell:
		dirStr = "sell"
	}
	args := append(append([]string{
		strconv.FormatUint(pairId, 10),
		dirStr,
		offerCoin.String(),
		demandCoinDenom,
		price.String(),
		amt.String(),
		fmt.Sprintf("--%s=%s", cli.FlagOrderLifespan, orderLifespan.String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...), extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewLimitOrderCmd(), args)
}
