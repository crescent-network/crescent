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

	"github.com/crescent-network/crescent/v3/x/liquidity/types"
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
		NewCreateRangedPoolCmd(),
		NewDepositCmd(),
		NewWithdrawCmd(),
		NewLimitOrderCmd(),
		NewMarketOrderCmd(),
		NewMMOrderCmd(),
		NewCancelOrderCmd(),
		NewCancelAllOrdersCmd(),
		NewCancelMMOrderCmd(),
	)

	return cmd
}

func NewCreatePairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pair [base-coin-denom] [quote-coin-denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a pair(market) for trading",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a pair(market) for trading.

Example:
$ %s tx %s create-pair uatom stake --from mykey
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
		Short: "Create a basic liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a basic liquidity pool with coins.

Example:
$ %s tx %s create-pool 1 1000000000uatom,50000000000stake --from mykey
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
				return fmt.Errorf("invalid deposit coins: %w", err)
			}

			msg := types.NewMsgCreatePool(clientCtx.GetFromAddress(), pairId, depositCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateRangedPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-ranged-pool [pair-id] [deposit-coins] [min-price] [max-price] [initial-price]",
		Args:  cobra.ExactArgs(5),
		Short: "Create a ranged liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a ranged liquidity pool with coins.

Example:
$ %s tx %s create-ranged-pool 1 1000000000uatom,10000000000stake 0.001 100 1.0 --from mykey
$ %s tx %s create-ranged-pool 1 1000000000uatom,10000000000stake 0.9 10000 1.0 --from mykey
$ %s tx %s create-ranged-pool 1 1000000000uatom,10000000000stake 1.3 2.5 1.5 --from mykey
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
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
				return fmt.Errorf("invalid deposit coins: %w", err)
			}

			minPrice, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid min price: %w", err)
			}

			maxPrice, err := sdk.NewDecFromStr(args[3])
			if err != nil {
				return fmt.Errorf("invalid max price: %w", err)
			}

			initialPrice, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid initial price: %w", err)
			}

			msg := types.NewMsgCreateRangedPool(
				clientCtx.GetFromAddress(), pairId, depositCoins,
				minPrice, maxPrice, initialPrice)

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
$ %s tx %s deposit 1 1000000000uatom,50000000000stake --from mykey
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
		Use:   "limit-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [price] [amount]",
		Args:  cobra.ExactArgs(6),
		Short: "Make a limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a limit order.

Example:
$ %s tx %s limit-order 1 buy 5000stake uatom 0.5 10000 --from mykey
$ %s tx %s limit-order 1 b 5000stake uatom 0.5 10000 --from mykey
$ %s tx %s limit-order 1 sell 10000uatom stake 2.0 10000 --order-lifespan=10m --from mykey
$ %s tx %s limit-order 1 s 10000uatom stake 2.0 10000 --order-lifespan=10m --from mykey

[pair-id]: pair id to swap with
[direction]: order direction (one of: buy,b,sell,s)
[offer-coin]: the amount of offer coin to swap
[demand-coin-denom]: the denom to exchange with the offer coin
[price]: the limit order price for the swap; the exchange ratio is X/Y where X is the amount of quote coin and Y is the amount of base coin
[amount]: the amount of base coin to buy or sell
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
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

			dir, err := parseOrderDirection(args[1])
			if err != nil {
				return fmt.Errorf("parse order direction: %w", err)
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

			orderLifespan, _ := cmd.Flags().GetDuration(FlagOrderLifespan)

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

	cmd.Flags().AddFlagSet(flagSetOrder())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewMarketOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-order [pair-id] [direction] [offer-coin] [demand-coin-denom] [amount]",
		Args:  cobra.ExactArgs(5),
		Short: "Make a market order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a market order.

Example:
$ %s tx %s market-order 1 buy 5000stake uatom 10000 --from mykey
$ %s tx %s market-order 1 b 5000stake uatom 10000 --from mykey
$ %s tx %s market-order 1 sell 10000uatom stake 10000 --order-lifespan=10m --from mykey
$ %s tx %s market-order 1 s 10000uatom stake 10000 --order-lifespan=10m --from mykey

[pair-id]: pair id to swap with
[direction]: order direction (one of: buy,b,sell,s)
[offer-coin]: the amount of offer coin to swap
[demand-coin-denom]: the denom to exchange with the offer coin
[amount]: the amount of base coin to buy or sell
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
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

			dir, err := parseOrderDirection(args[1])
			if err != nil {
				return fmt.Errorf("parse order direction: %w", err)
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

			orderLifespan, _ := cmd.Flags().GetDuration(FlagOrderLifespan)

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

	cmd.Flags().AddFlagSet(flagSetOrder())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewMMOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mm-order [pair-id] [max-sell-price] [min-sell-price] [sell-amount] [max-buy-price] [min-buy-price] [buy-amount]",
		Args:  cobra.ExactArgs(7),
		Short: "Make a market making order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a market making order.
A market making order is a set of limit orders for each buy/sell side.
You can leave one side(but not both) empty by passing 0 as its arguments.

Example:
$ %s tx %s mm-order 1 102 101 10000 100 99 10000 --from mykey
$ %s tx %s mm-order 1 0 0 0 100 99 10000 --from mykey
$ %s tx %s mm-order 1 102 101 10000 0 0 0 --from mykey

[pair-id]: pair id to make order
[max-sell-price]: maximum price of sell orders
[min-sell-price]]: minimum price of sell orders
[sell-amount]: total amount of sell orders
[max-buy-price]: maximum price of buy orders
[min-buy-price]: minimum price of buy orders
[buy-amount]: the total amount of buy orders
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
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

			maxSellPrice, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid max sell price: %w", err)
			}

			minSellPrice, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid min sell price: %w", err)
			}

			sellAmt, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid sell amount: %s", args[3])
			}

			maxBuyPrice, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid max buy price: %w", err)
			}

			minBuyPrice, err := sdk.NewDecFromStr(args[5])
			if err != nil {
				return fmt.Errorf("invalid min buy price: %w", err)
			}

			buyAmt, ok := sdk.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("invalid buy amount: %s", args[3])
			}

			orderLifespan, _ := cmd.Flags().GetDuration(FlagOrderLifespan)

			msg := types.NewMsgMMOrder(
				clientCtx.GetFromAddress(),
				pairId,
				maxSellPrice, minSellPrice, sellAmt,
				maxBuyPrice, minBuyPrice, buyAmt,
				orderLifespan,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetOrder())
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

func NewCancelMMOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-mm-order [pair-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel the mm order in a pair",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel the mm order in a pair.
This will cancel all limit orders in the pair made by the mm order.

Example:
$ %s tx %s cancel-mm-order 1 --from mykey
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

			msg := types.NewMsgCancelMMOrder(clientCtx.GetFromAddress(), pairId)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
