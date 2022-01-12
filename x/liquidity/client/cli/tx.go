package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetTxCmd returns the transaction commands for the module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreatePoolCmd(),
		NewDepositBatchCmd(),
		NewWithdrawBatchCmd(),
		NewSwapBatchCmd(),
		NewCancelSwapBatchCmd(),
	)

	return cmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [x-coin] [y-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Create liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create liquidity pool with deposit coins of x and y.
Example:
$ %s tx %s create-pool 1000000000uatom 50000000000ucsnt --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fmt.Println(clientCtx)

			// TODO: not implemented yet

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDepositBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [pool-id] [x-coin] [y-coin]",
		Args:  cobra.ExactArgs(3),
		Short: "Deposit x and y coins to the liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit x and y coins to the liquidity pool.
Example:
$ %s tx %s deposit 1 1000000000uatom 50000000000ucsnt --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fmt.Println(clientCtx)

			// TODO: not implemented yet

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [pool-id] [pool-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw pool coin from the specified liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw pool coin from the specified liquidity pool.
Example:
$ %s tx %s withdraw 1 10000pool96EF6EA6E5AC828ED87E8D07E7AE2A8180570ADD212117B2DA6F0B75D17A6295 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fmt.Println(clientCtx)

			// TODO: not implemented yet

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSwapBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [pool-id] [pool-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Swap x coin to y coin from the specified liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Swap x coin to y coin from the specified liquidity pool.
Example:
$ %s tx %s swap 1 10000pool96EF6EA6E5AC828ED87E8D07E7AE2A8180570ADD212117B2DA6F0B75D17A6295 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fmt.Println(clientCtx)

			// TODO: not implemented yet

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCancelSwapBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-swap [swap-request-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel swap request",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel swap request.
Example:
$ %s tx %s cancel-swap 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fmt.Println(clientCtx)

			// TODO: not implemented yet

			// return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
