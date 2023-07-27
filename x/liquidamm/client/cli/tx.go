package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
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
		Use:   "mint-share [public-position-id] [desired-amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Mint public position share for auto compounding rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint public position share for auto compounding rewards. 

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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}
			desiredAmt, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid desired amount: %w", err)
			}
			msg := types.NewMsgMintShare(clientCtx.GetFromAddress(), publicPositionId, desiredAmt)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewBurnShareCmd implements the burn share command handler.
func NewBurnShareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-share [public-position-id] [share]",
		Args:  cobra.ExactArgs(2),
		Short: "Burn public position share",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn public position share.

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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}
			share, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid share: %w", err)
			}
			msg := types.NewMsgBurnShare(clientCtx.GetFromAddress(), publicPositionId, share)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewPlaceBidCmd implements the place bid command handler.
func NewPlaceBidCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-bid [public-position-id] [auction-id] [share]",
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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
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
				clientCtx.GetFromAddress(), publicPositionId, auctionId, share)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCmdSubmitPublicPositionCreateProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-position-create [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public position create proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public public position create proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal public-position-create <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Public Position Create Proposal",
  "description": "Let's create a new public position",
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
			var proposal types.PublicPositionCreateProposal
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

func NewCmdSubmitPublicPositionParameterChangeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-position-parameter-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public position parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public public position parameter change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal public-position-parameter-change <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Public Position Parameter Change Proposal",
  "description": "Change public position parameters",
  "changes": [
    {
      "public_position_id": "1",
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
			var proposal types.PublicPositionParameterChangeProposal
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
