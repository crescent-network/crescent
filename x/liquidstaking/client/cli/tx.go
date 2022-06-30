package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// GetTxCmd returns a root CLI command handler for all x/liquidstaking transaction commands.
func GetTxCmd() *cobra.Command {
	liquidstakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Liquid-staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquidstakingTxCmd.AddCommand(
		NewLiquidStakeCmd(),
		NewLiquidUnstakeCmd(),
	)

	return liquidstakingTxCmd
}

// NewLiquidStakeCmd implements the liquid stake coin command handler.
func NewLiquidStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-stake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Liquid-stake coin",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid-stake coin. 
			
Example:
$ %s tx %s liquid-stake 1000stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			liquidStaker := clientCtx.GetFromAddress()

			stakingCoin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidStake(liquidStaker, stakingCoin)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewLiquidUnstakeCmd implements the liquid unstake coin command handler.
func NewLiquidUnstakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-unstake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Liquid-unstake coin",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid-unstake coin. 
			
Example:
$ %s tx %s liquid-unstake 500stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			liquidStaker := clientCtx.GetFromAddress()

			unstakingCoin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidUnstake(liquidStaker, unstakingCoin)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
