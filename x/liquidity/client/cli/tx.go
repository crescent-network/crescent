package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
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
		NewCreatePairCmd(),
		NewCreatePoolCmd(),
		NewDepositCmd(),
		NewWithdrawCmd(),
		NewLimitOrderCmd(),
		NewMarketOrderCmd(),
		NewCancelOrderCmd(),
		NewCancelAllOrdersCmd(),
	)

	return cmd
}

func NewCreatePairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pair [base-coin-denom] [quote-coin-denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a denom pair for an order book",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a denom pair for an order book.
Example:
$ %s tx %s create-pair uatom usquad --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			baseCoinDenom := args[0]
			quoteCoinDenom := args[1]

			msg := types.NewMsgCreatePair(clientCtx.GetFromAddress(), baseCoinDenom, quoteCoinDenom)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [pair-id] [deposit-coins]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a liquidity pool with coins.
Example:
$ %s tx %s create-pool 1 1000000000uatom,50000000000usquad --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse pair id: %w", err)
			}

			depositCoins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid deposit coints: %w", err)
			}

			msg := types.NewMsgCreatePool(clientCtx.GetFromAddress(), pairId, depositCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [pool-id] [deposit-coins]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit coins to a liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit coins to a liquidity pool.
Example:
$ %s tx %s deposit 1 1000000000uatom,50000000000usquad --from mykey
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

			depositCoins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid deposit coins: %w", err)
			}

			msg := types.NewMsgDeposit(clientCtx.GetFromAddress(), poolId, depositCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [pool-id] [pool-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw coins from the specified liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw coins from the specified liquidity pool.
Example:
$ %s tx %s withdraw 1 10000pool1 --from mykey
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
				return err
			}

			poolCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdraw(
				clientCtx.GetFromAddress(),
				poolId,
				poolCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limit-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [price] [base-coin-amount] [order-lifespan]",
		Args:  cobra.ExactArgs(7),
		Short: "Make a limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a limit order.
Example:
$ %s tx %s limit-order 1 ORDER_DIRECTION_BUY 10000usquad uatom 1.0 10000 10s --from mykey

[pair-id]: pair id to swap with
[direction]: order direction (one of: ORDER_DIRECTION_BUY,ORDER_DIRECTION_SELL)
[offer-coin]: the amount of offer coin to swap
[demand-coin-denom]: the denom to exchange with the offer coin
[price]: the limit order price for the swap; the exchange ratio is X/Y where X is the amount of quote coin and Y is the amount of base coin
[base-coin-amount]: the amount of base coin to buy or sell
[order-lifespan]: the time duration that the order lives until it is expired; valid time units are "ns", "us", "ms", "s", "m", and "h" 
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse pair id: %w", err)
			}

			rawDir, ok := types.OrderDirection_value[args[1]]
			if !ok {
				return fmt.Errorf("unknown order direction: %s", args[1])
			}
			dir := types.OrderDirection(rawDir)
			if dir != types.OrderDirectionBuy && dir != types.OrderDirectionSell {
				return fmt.Errorf("invalid order direction: %s", dir)
			}

			offerCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid offer coin: %w", err)
			}

			demandCoinDenom := args[3]
			if err := sdk.ValidateDenom(demandCoinDenom); err != nil {
				return fmt.Errorf("invalid demand coin denom: %w", err)
			}

			price, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}

			amt, ok := sdk.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[5])
			}

			orderLifespan, err := time.ParseDuration(args[6])
			if err != nil {
				return fmt.Errorf("invalid order lifespan: %w", err)
			}

			msg := types.NewMsgLimitOrder(
				clientCtx.GetFromAddress(),
				pairId,
				dir,
				offerCoin,
				demandCoinDenom,
				price,
				amt,
				orderLifespan,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewMarketOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [base-coin-amount] [order-lifespan]",
		Args:  cobra.ExactArgs(6),
		Short: "Make a market order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a market order.
Example:
$ %s tx %s market-order 1 ORDER_DIRECTION_BUY 10000usquad uatom 10000 10s --from mykey

[pair-id]: pair id to swap with
[direction]: order direction (one of: ORDER_DIRECTION_BUY,ORDER_DIRECTION_SELL)
[offer-coin]: the amount of offer coin to swap
[demand-coin-denom]: the denom to exchange with the offer coin
[base-coin-amount]: the amount of base coin to buy or sell
[order-lifespan]: the time duration that the order lives until it is expired; valid time units are "ns", "us", "ms", "s", "m", and "h" 
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse pair id: %w", err)
			}

			rawDir, ok := types.OrderDirection_value[args[1]]
			if !ok {
				return fmt.Errorf("unknown order direction: %s", args[1])
			}
			dir := types.OrderDirection(rawDir)
			if dir != types.OrderDirectionBuy && dir != types.OrderDirectionSell {
				return fmt.Errorf("invalid order direction: %s", dir)
			}

			offerCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid offer coin: %w", err)
			}

			demandCoinDenom := args[3]
			if err := sdk.ValidateDenom(demandCoinDenom); err != nil {
				return fmt.Errorf("invalid demand coin denom: %w", err)
			}

			amt, ok := sdk.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[4])
			}

			orderLifespan, err := time.ParseDuration(args[5])
			if err != nil {
				return fmt.Errorf("invalid order lifespan: %w", err)
			}

			msg := types.NewMsgMarketOrder(
				clientCtx.GetFromAddress(),
				pairId,
				dir,
				offerCoin,
				demandCoinDenom,
				amt,
				orderLifespan,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCancelOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-order [pair-id] [order-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Cancel an order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel an order.
Example:
$ %s tx %s cancel-order 1 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			orderId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelOrder(
				clientCtx.GetFromAddress(),
				pairId,
				orderId,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCancelAllOrdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-all-orders [pair-ids]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Cancel all orders",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel all orders.
Example:
$ %s tx %s cancel-all-orders --from mykey
$ %s tx %s cancel-all-orders 1,3 --from mykey
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var pairIds []uint64
			for _, pairIdStr := range strings.Split(args[0], ",") {
				pairId, err := strconv.ParseUint(pairIdStr, 10, 64)
				if err != nil {
					return fmt.Errorf("parse pair id: %w", err)
				}
				pairIds = append(pairIds, pairId)
			}

			msg := types.NewMsgCancelAllOrders(clientCtx.GetFromAddress(), pairIds)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
