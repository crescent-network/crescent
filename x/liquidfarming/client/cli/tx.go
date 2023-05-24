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

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
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
		NewMintShareCmd(),
		NewBurnShareCmd(),
		NewPlaceBidCmd(),
		NewCancelBidCmd(),
	)
	return cmd
}

// NewMintShareCmd implements the mint share command handler.
func NewMintShareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint-share [liquid-farm-id] [desired-amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Mint liquid farm share for auto compounding rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint liquid farm share for auto compounding rewards. 
			
Example:
$ %s tx %s mint-share 1 100000000ucre,500000000uusd --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			desiredAmt, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid desired amount: %w", err)
			}
			msg := types.NewMsgMintShare(clientCtx.GetFromAddress(), liquidFarmId, desiredAmt)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewBurnShareCmd implements the burn share command handler.
func NewBurnShareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-share [liquid-farm-id] [share]",
		Args:  cobra.ExactArgs(2),
		Short: "Burn liquid farm share",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn liquid farm share.
			
Example:
$ %s tx %s burn-share 1 10000000000lfshare1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			share, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid share: %w", err)
			}
			msg := types.NewMsgBurnShare(clientCtx.GetFromAddress(), liquidFarmId, share)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewPlaceBidCmd implements the place bid command handler.
func NewPlaceBidCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-bid [liquid-farm-id] [auction-id] [share]",
		Args:  cobra.ExactArgs(3),
		Short: "Place a bid for a rewards auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a bid for a rewards auction.
			
Example:
$ %s tx %s place-bid 1 1 10000000lfshare1 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			auctionId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction id: %w", err)
			}
			share, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid share: %w", err)
			}
			msg := types.NewMsgPlaceBid(
				clientCtx.GetFromAddress(), liquidFarmId, auctionId, share)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCancelBidCmd implements the refund bid command handler.
func NewCancelBidCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund-bid [liquid-farm-id] [auction-id]",
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
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			auctionId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction id: %w", err)
			}
			msg := types.NewMsgCancelBid(clientCtx.GetFromAddress(), liquidFarmId, auctionId)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
