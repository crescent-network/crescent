package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func NewCmdSubmitLiquidFarmCreateProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-farm-create [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a liquid farm create proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public liquid farm create proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal liquid-farm-create <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Liquid Farm Create Proposal",
  "description": "Let's start new liquid farming",
  "pool_id": "1",
  "lower_price": "4.5",
  "upper_price": "5.5",
  "min_bid_amount": "100000000",
  "fee_rate": "0.003"
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
			var proposal types.LiquidFarmCreateProposal
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

func NewCmdSubmitLiquidFarmParameterChangeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-farm-parameter-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a liquid farm parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public liquid farm parameter change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal liquid-farm-parameter-change <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Liquid Farm Parameter Change Proposal",
  "description": "Change liquid farm parameters",
  "changes": [
    {
      "liquid_farm_id": "1",
      "min_bid_amount": "10000000",
      "fee_rate": "0.001"
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
			var proposal types.LiquidFarmParameterChangeProposal
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
