package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

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
		NewPlaceBatchLimitOrderCmd(),
		NewPlaceMMLimitOrderCmd(),
		NewPlaceMMBatchLimitOrderCmd(),
		NewPlaceMarketOrderCmd(),
		NewCancelOrderCmd(),
		NewSwapExactAmountInCmd(),
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
		Use:   "place-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]",
		Args:  cobra.ExactArgs(5),
		Short: "Place a limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a limit order.

Example:
$ %s tx %s place-limit-order 1 true 15 100000 1h --from mykey
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
			lifespan, err := time.ParseDuration(args[4])
			if err != nil {
				return fmt.Errorf("invalid lifespan: %w", err)
			}
			msg := types.NewMsgPlaceLimitOrder(
				clientCtx.GetFromAddress(), marketId, isBuy, price, qty, lifespan)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewPlaceBatchLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-batch-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]",
		Args:  cobra.ExactArgs(5),
		Short: "Place a batch limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a batch limit order.
Batch orders are matched prior to normal orders in a batch matching stage.

Example:
$ %s tx %s place-batch-limit-order 1 true 15 100000 1h --from mykey
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
			lifespan, err := time.ParseDuration(args[4])
			if err != nil {
				return fmt.Errorf("invalid lifespan: %w", err)
			}
			msg := types.NewMsgPlaceBatchLimitOrder(
				clientCtx.GetFromAddress(), marketId, isBuy, price, qty, lifespan)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewPlaceMMLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-mm-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]",
		Args:  cobra.ExactArgs(5),
		Short: "Place a market maker limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a market maker limit order.

Example:
$ %s tx %s place-mm-limit-order 1 true 15 100000 1h --from mykey
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
			lifespan, err := time.ParseDuration(args[4])
			if err != nil {
				return fmt.Errorf("invalid lifespan: %w", err)
			}
			msg := types.NewMsgPlaceMMLimitOrder(
				clientCtx.GetFromAddress(), marketId, isBuy, price, qty, lifespan)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewPlaceMMBatchLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-mm-batch-limit-order [market-id] [is-buy] [price] [quantity] [lifespan]",
		Args:  cobra.ExactArgs(5),
		Short: "Place a market maker batch limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a market maker batch limit order.
Batch orders are matched prior to normal orders in a batch matching stage.

Example:
$ %s tx %s place-mm-batch-limit-order 1 true 15 100000 1h --from mykey
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
			lifespan, err := time.ParseDuration(args[4])
			if err != nil {
				return fmt.Errorf("invalid lifespan: %w", err)
			}
			msg := types.NewMsgPlaceMMBatchLimitOrder(
				clientCtx.GetFromAddress(), marketId, isBuy, price, qty, lifespan)
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

func NewSwapExactAmountInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-exact-amount-in [routes] [input] [min-output]",
		Args:  cobra.ExactArgs(3),
		Short: "Swap with exact amount in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Swap with exact amount in.

Example:
$ %s tx %s swap-exact-amount-in 1,2,3 1000000stake 98000uatom --from mykey
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
			msg := types.NewMsgSwapExactAmountIn(clientCtx.GetFromAddress(), routes, input, minOutput)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdSubmitMarketParameterChangeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-parameter-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a market parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a market parameter change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal market-parameter-change <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Market parameter change",
  "description": "Change fee rates",
  "changes": [
    {
      "market_id": "1",
      "maker_fee_rate"": "0.0005",
      "taker_fee_rate": "0.001"
    },
    {
      "market_id": "2",
      "maker_fee_rate"": "-0.001",
      "taker_fee_rate": "0.002"
    }
  ]
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			depositStr, _ := cmd.Flags().GetString(cli.FlagDeposit)
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return fmt.Errorf("invalid deposit: %w", err)
			}
			var proposal types.MarketParameterChangeProposal
			bz, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read proposal: %w", err)
			}
			if err = clientCtx.Codec.UnmarshalJSON(bz, &proposal); err != nil {
				return fmt.Errorf("unmarshal proposal: %w", err)
			}
			msg, err := gov.NewMsgSubmitProposal(&proposal, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	return cmd
}
