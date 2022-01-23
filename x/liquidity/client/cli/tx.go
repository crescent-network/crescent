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
		NewCreatePairCmd(),
		NewCreatePoolCmd(),
		NewDepositBatchCmd(),
		NewWithdrawBatchCmd(),
		NewSwapBatchCmd(),
		NewCancelSwapBatchCmd(),
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
$ %s tx %s create-pair uatom ucsnt --from mykey
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
$ %s tx %s create-pool 1 1000000000uatom,50000000000ucsnt --from mykey
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

func NewDepositBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [pool-id] [deposit-coins]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit coins to a liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit coins to a liquidity pool.
Example:
$ %s tx %s deposit 1 1000000000uatom,50000000000ucsnt --from mykey
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

			msg := types.NewMsgDepositBatch(clientCtx.GetFromAddress(), poolId, depositCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [pool-id] [pool-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw coins from the specified liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw coins from the specified liquidity pool.
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			poolCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawBatch(
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

func NewSwapBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [x-coin-denom] [y-coin-denom] [offer-coin] [demand-coin-denom] [order-price] [order-life-span]",
		Args:  cobra.ExactArgs(6),
		Short: "Swap coins within a pair",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Swap coins within a pair.
Example:
$ %s tx %s swap ucsnt uatom 10000uatom ucsnt 1.0 10s --from mykey

[x-coin-denom]: x coin denomination
[y-coin-denom]: y coin denomination
[offer-coin]: the amount of offer coin to swap 
[demand-coin-denom]: the denom to exchange with the offer coin
[order-price]: the limir order price for the swap; the exchange ratio is X/Y where X is the amount of first coin and Y is the amount of second coin
[order-life-span]: the time duration that the swap order request lives until it is executed; valid time units are "ns", "us", "ms", "s", "m", and "h" 
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			xCoinDenom := args[0]
			if err := sdk.ValidateDenom(xCoinDenom); err != nil {
				return err
			}

			yCoinDenom := args[1]
			if err := sdk.ValidateDenom(yCoinDenom); err != nil {
				return err
			}

			offerCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			demandCoinDenom := args[3]
			if err := sdk.ValidateDenom(demandCoinDenom); err != nil {
				return err
			}

			orderPrice, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return err
			}

			orderLifespan, err := time.ParseDuration(args[5])
			if err != nil {
				return err
			}

			msg := types.NewMsgSwapBatch(
				clientCtx.GetFromAddress(),
				xCoinDenom,
				yCoinDenom,
				offerCoin,
				demandCoinDenom,
				orderPrice,
				orderLifespan,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCancelSwapBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-swap [pair-id] [swap-request-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Cancel a swap request",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel a swap request.
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

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			swapRequestId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelSwapBatch(
				clientCtx.GetFromAddress(),
				pairId,
				swapRequestId,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
