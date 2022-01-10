package testutil

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	crescentapp "github.com/crescent-network/crescent/app"
	crescentcli "github.com/crescent-network/crescent/x/farming/client/cli"
)

// NewConfig returns config that defines the necessary testing requirements
// used to bootstrap and start an in-process local testing network.
func NewConfig(dbm *dbm.MemDB) network.Config {
	encCfg := simapp.MakeTestEncodingConfig()

	cfg := network.DefaultConfig()
	cfg.AppConstructor = NewAppConstructor(encCfg, dbm)                   // the ABCI application constructor
	cfg.GenesisState = crescentapp.ModuleBasics.DefaultGenesis(cfg.Codec) // farming genesis state to provide
	return cfg
}

// NewAppConstructor returns a new network AppConstructor.
func NewAppConstructor(encodingCfg params.EncodingConfig, db *dbm.MemDB) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return crescentapp.NewCrescentApp(
			val.Ctx.Logger, db, nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			crescentapp.MakeEncodingConfig(),
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

// MsgCreateFixedAmountPlanExec creates a transaction for creating a private fixed amount plan.
func MsgCreateFixedAmountPlanExec(clientCtx client.Context, from string, file string,
	extraArgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		file,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, commonArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, crescentcli.NewCreateFixedAmountPlanCmd(), args)
}

// MsgStakeExec creates a transaction for staking coin.
func MsgStakeExec(clientCtx client.Context, from string, stakingCoins string,
	extraArgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		stakingCoins,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, commonArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, crescentcli.NewStakeCmd(), args)
}

// MsgAdvanceEpochExec creates a transaction to advance epoch by 1.
func MsgAdvanceEpochExec(clientCtx client.Context, from string,
	extraAtgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
	}, commonArgs...)

	args = append(args, commonArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, crescentcli.NewAdvanceEpochCmd(), args)
}

// MsgSendExec creates a transaction to transfer coins.
func MsgSendExec(clientCtx client.Context, from string, to string, amount string,
	extraAtgs ...string) (testutil.BufferWriter, error) {

	args := append([]string{
		from,
		to,
		amount,
	}, commonArgs...)

	args = append(args, commonArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(), args)
}
