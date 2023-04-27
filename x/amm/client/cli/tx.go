package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/crescent-network/crescent/v5/x/amm/types"
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
		NewAddLiquidityCmd(),
		NewRemoveLiquidityCmd(),
		NewCollectCmd(),
	)

	return cmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [market-id] [price]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a pool.

Example:
$ %s tx %s create-pool 1 10 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			marketId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid market id: %w", err)
			}
			price, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}
			msg := types.NewMsgCreatePool(clientCtx.GetFromAddress(), marketId, price)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewAddLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [pool-id] [lower-price] [upper-price] [desired-amount0] [desired-amount1] [min-amt0] [min-amt1]",
		Args:  cobra.ExactArgs(7),
		Short: "Add liquidity to a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add liquidity to a pool.

Example:
$ %s tx %s add-liquidity 1 9.5 10.5 1000000 10000000 900000 9000000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}
			lowerPrice, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid lower price: %w", err)
			}
			upperPrice, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid upper price: %w", err)
			}
			desiredAmt0, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid desired amount0: %s", args[3])
			}
			desiredAmt1, ok := sdk.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid desired amount1: %s", args[4])
			}
			minAmt0, ok := sdk.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("invalid minimum amount0: %s", args[5])
			}
			minAmt1, ok := sdk.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("invalid minimum amount1: %s", args[6])
			}
			msg := types.NewMsgAddLiquidity(
				clientCtx.GetFromAddress(), poolId, lowerPrice, upperPrice,
				desiredAmt0, desiredAmt1, minAmt0, minAmt1)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewRemoveLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [position-id] [liquidity] [min-amt0] [min-amt1]",
		Args:  cobra.ExactArgs(4),
		Short: "Remove liquidity from a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove liquidity from a pool.

Example:
$ %s tx %s remove-liquidity 1 10000000000000 500000 5000000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			liquidity, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid liquidity: %w", err)
			}
			minAmt0, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid minimum amount0: %s", args[2])
			}
			minAmt1, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid minimum amount1: %s", args[3])
			}
			msg := types.NewMsgRemoveLiquidity(
				clientCtx.GetFromAddress(), positionId, liquidity, minAmt0, minAmt1)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCollectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect [position-id] [min-amt0] [min-amt1]",
		Args:  cobra.ExactArgs(3),
		Short: "Collect fees accrued in the position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Collect fees accrued in the position.

Example:
$ %s tx %s collect 1 100000 1000000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			minAmt0, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid minimum amount0: %s", args[1])
			}
			minAmt1, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid minimum amount1: %s", args[2])
			}
			msg := types.NewMsgCollect(
				clientCtx.GetFromAddress(), positionId, minAmt0, minAmt1)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
