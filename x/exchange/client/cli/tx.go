package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
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
		NewCreateMarketCmd(),
		NewPlaceLimitOrderCmd(),
		NewPlaceMarketOrderCmd(),
		NewCancelOrderCmd(),
		NewSwapExactInCmd(),
	)

	return cmd
}

func NewCreateMarketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-market [base-denom] [quote-denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a market",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a market.

Example:
$ %s tx %s create-market uatom stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			baseDenom := args[0]
			quoteDenom := args[1]
			msg := types.NewMsgCreateMarket(clientCtx.GetFromAddress(), baseDenom, quoteDenom)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewPlaceLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-limit-order [market-id] [is-buy] [price] [quantity]",
		Args:  cobra.ExactArgs(4),
		Short: "Place a limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a limit order.

Example:
$ %s tx %s place-limit-order 1 true 15 100000 --from mykey
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
			isBuy, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid buy flag: %w", err)
			}
			price, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}
			qty, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid quantity: %s", args[3])
			}
			isBatch := false // TODO: parse arg properly
			msg := types.NewMsgPlaceLimitOrder(clientCtx.GetFromAddress(), marketId, isBuy, price, qty, isBatch)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewPlaceMarketOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-market-order [market-id] [is-buy] [quantity]",
		Args:  cobra.ExactArgs(3),
		Short: "Place a market order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a market order.

Example:
$ %s tx %s place-market-order 1 false 100000 --from mykey
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
			isBuy, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid buy flag: %w", err)
			}
			qty, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid quantity: %s", args[2])
			}
			msg := types.NewMsgPlaceMarketOrder(clientCtx.GetFromAddress(), marketId, isBuy, qty)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCancelOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-order [order-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel an order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel an order by its ID.

Example:
$ %s tx %s cancel-order 1000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid order id: %w", err)
			}
			msg := types.NewMsgCancelOrder(clientCtx.GetFromAddress(), orderId)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewSwapExactInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-exact-in [routes] [input] [min-output]",
		Args:  cobra.ExactArgs(3),
		Short: "Swap with exact amount in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Swap with exact amount in.

Example:
$ %s tx %s swap-exact-in 1,2,3 1000000stake 98000uatom --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			var routes []uint64
			for _, chunk := range strings.Split(args[0], ",") {
				marketId, err := strconv.ParseUint(chunk, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid routes: %w", err)
				}
				routes = append(routes, marketId)
			}
			input, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid input: %w", err)
			}
			minOutput, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid minimum output: %w", err)
			}
			msg := types.NewMsgSwapExactIn(clientCtx.GetFromAddress(), routes, input, minOutput)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
