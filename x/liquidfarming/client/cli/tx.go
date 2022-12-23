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

	"github.com/crescent-network/crescent/v4/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
)

// GetTxCmd returns the cli transaction commands for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Transaction commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewLiquidFarmCmd(),
		NewLiquidUnfarmCmd(),
		NewLiquidUnfarmAndWithdrawCmd(),
		NewPlaceBidCmd(),
		NewRefundBidCmd(),
	)

	if keeper.EnableAdvanceAuction {
		cmd.AddCommand(NewAdvanceAuctionCmd())
	}

	return cmd

}

// NewLiquidFarmCmd implements the liquid farm command handler.
func NewLiquidFarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-farm [pool-id] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Liquid farm pool coin for auto compounding rewards and receive LFCoin by mint rate",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid farm pool coin for auto compounding and receive LFCoin by mint rate. 
The module farms your pool coin to the farm module for you and auto compounds rewards for every auction period.
LFCoin opens up other opportunities to make profits like providing liquidity for lending protocol.
			
Example:
$ %s tx %s liquid-farm 1 100000000pool1 --from mykey
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
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			farmingCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid coin: %w", err)
			}

			msg := types.NewMsgLiquidFarm(
				poolId,
				clientCtx.GetFromAddress().String(),
				farmingCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewLiquidUnfarmCmd implements the liquid unfarm command handler.
func NewLiquidUnfarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-unfarm [pool-id] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Liquid unfarm liquid farming coin (LFCoin) ",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid unfarm liquid farming coin (LFCoin) to receive corresponding pool coin by burn rate.
			
Example:
$ %s tx %s liquid-unfarm 1 100000lf1 --from mykey
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
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			unfarmingCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid coin: %w", err)
			}

			msg := types.NewMsgLiquidUnfarm(
				poolId,
				clientCtx.GetFromAddress().String(),
				unfarmingCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewLiquidUnfarmAndWithdrawCmd implements the liquid unfarm and withdraw command handler.
func NewLiquidUnfarmAndWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-unfarm-and-withdraw [pool-id] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Liquid unfarm liquid farming coin (LFCoin) and withdraw from the liquidity module",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Liquid unfarm liquid farming coin (LFCoin) to receive corresponding pool coin by burn rate and withdraw from the liquidity module.
			
Example:
$ %s tx %s liquid-unfarm-and-withdraw 1 100000lf1 --from mykey
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
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			unfarmingCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid coin: %w", err)
			}

			msg := types.NewMsgLiquidUnfarmAndWithdraw(
				poolId,
				clientCtx.GetFromAddress().String(),
				unfarmingCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewPlaceBidCmd implements the place bid command handler.
func NewPlaceBidCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-bid [auction-id] [pool-id] [amount]",
		Args:  cobra.ExactArgs(3),
		Short: "Place a bid for a rewards auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a bid for a rewards auction.
			
Example:
$ %s tx %s place-bid 1 1 10000000pool1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse auctionId id: %w", err)
			}

			poolId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid bidding amount: %w", err)
			}

			msg := types.NewMsgPlaceBid(
				auctionId,
				poolId,
				clientCtx.GetFromAddress().String(),
				amount,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewRefundBidCmd implements the refund bid command handler.
func NewRefundBidCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund-bid [auction-id] [pool-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Refund a bid",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Refund a bid.
			
Example:
$ %s tx %s refund-bid 1 1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse auctionId id: %w", err)
			}

			poolId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			msg := types.NewMsgRefundBid(
				auctionId,
				poolId,
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewAdvanceAuctionCmd implements the advance auction by 1 command handler.
func NewAdvanceAuctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance-auction",
		Args:  cobra.NoArgs,
		Short: "Advance auction by 1 to simulate rewards auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Advance auction by 1 to simulate rewards auction.
This message is available for testing purpose and it can only be enabled when you build the binary with "make install-testing" command. 

Example:
$ %s tx %s advance-auction --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			requesterAcc := clientCtx.GetFromAddress()

			msg := types.NewMsgAdvanceAuction(requesterAcc)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
