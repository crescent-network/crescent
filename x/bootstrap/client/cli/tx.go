package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// GetTxCmd returns a root CLI command handler for all x/bootstrap transaction commands.
func GetTxCmd() *cobra.Command {
	bootstrapTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bootstrap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	bootstrapTxCmd.AddCommand(
		NewLimitOrderCmd(),
		//NewModifyOrderCmd(),
		// TODO: add tx functions
	)

	return bootstrapTxCmd
}

// TODO: update along msg struct
func NewLimitOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limit-order [pool-id] [direction] [offer-coin] [demand-coin-denom] [price] [amount]",
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse pool id: %w", err)
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

			//amt, ok := sdk.NewIntFromString(args[5])
			//if !ok {
			//	return fmt.Errorf("invalid amount: %s", args[5])
			//}

			msg := types.NewMsgLimitOrder(
				clientCtx.GetFromAddress(),
				poolId,
				dir,
				offerCoin,
				//demandCoinDenom,
				price,
				//amt,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// TODO: add modify order

// GetCmdSubmitBootstrapProposal implements the inclusion/exclusion/rejection/distribution for market maker command handler.
func GetCmdSubmitBootstrapProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap-proposal [proposal-file] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a bootstrap proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a bootstrap proposal along with an initial deposit. You can submit this governance proposal
to create a bootstrap pool, and place initial orders. The proposal details must be supplied via a JSON file. A JSON file to add request proposal is 
provided below.

Example:
$ %s tx gov submit-proposal bootstrap-proposal <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Bootstrap Proposal",
  "description": "TBD",
  "TBD": "TBD",
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

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			//proposal, err := ParseBootstrapProposal(clientCtx.Codec, args[0])
			//if err != nil {
			//	return err
			//}

			//content := types.NewBootstrapProposal(
			//	proposal.Title,
			//	proposal.Description,
			//	proposal.Inclusions,
			//	proposal.Exclusions,
			//	proposal.Rejections,
			//	proposal.Distributions,
			//)

			from := clientCtx.GetFromAddress()

			msg, err := gov.NewMsgSubmitProposal(nil, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
