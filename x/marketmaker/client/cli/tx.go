package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

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
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

// GetTxCmd returns a root CLI command handler for all x/marketmaker transaction commands.
func GetTxCmd() *cobra.Command {
	marketmakerTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "MarketMaker transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	marketmakerTxCmd.AddCommand(
		NewApplyMarketMaker(),
		NewClaimIncentives(),
	)

	return marketmakerTxCmd
}

// NewApplyMarketMaker implements apply market maker command handler.
func NewApplyMarketMaker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [pool-ids]",
		Args:  cobra.ExactArgs(1),
		Short: "Apply to be a market maker",
		Long: strings.TrimSpace(
			fmt.Sprintf(`
Apply to be a market maker for a number of pairs. The deposit amount defined in params is required to deposit, and the amount is expected to be refunded when you are either included or rejected by the community (through a governance proposal).

Example:
$ %s tx %s apply 1 --from mykey
$ %s tx %s apply 1,2 --from mykey
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

			farmer := clientCtx.GetFromAddress()
			pairIds := []uint64{}
			pairIdsStr := strings.Split(args[0], ",")

			for _, i := range pairIdsStr {
				pairId, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return fmt.Errorf("parse pair id: %w", err)
				}
				pairIds = append(pairIds, pairId)
			}

			msg := types.NewMsgApplyMarketMaker(farmer, pairIds)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewClaimIncentives implements the remove plan handler.
func NewClaimIncentives() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Args:  cobra.ExactArgs(0),
		Short: "Claim all claimable incentives",
		Long: fmt.Sprintf(`
Claim all market making incentives distributed through governance

Example:
$ %s tx %s claim --from mykey`,
			version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress()

			msg := types.NewMsgClaimIncentives(creator)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSubmitMarketMakerProposal implements the inclusion/exclusion/rejection/distribution for market maker command handler.
func GetCmdSubmitMarketMakerProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-maker-proposal [proposal-file] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a market maker proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a market maker proposal along with an initial deposit. You can submit this governance proposal
to inclusion, exclusion, rejection, and incentive distribution for market maker. The proposal details must be supplied via a JSON file. A JSON file to add request proposal is 
provided below.

Example:
$ %s tx gov submit-proposal market-maker-proposal <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Market Maker Proposal",
  "description": "Are you ready to market making?",
  "inclusions": [
    {
      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
      "pair_id": "1"
    }
  ],
  "exclusions": [],
  "rejections": [
    {
      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
      "pair_id": "2"
    }
  ],
  "distributions": [
    {
      "address": "cosmos1vqac3p8fl4kez7ehjz8eltugd2fm67pckpl7pn",
      "pair_id": "1",
      "amount": [
        {
          "denom": "stake",
          "amount": "100000000"
        }
      ]
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

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			proposal, err := ParseMarketMakerProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			content := types.NewMarketMakerProposal(
				proposal.Title,
				proposal.Description,
				proposal.Inclusions,
				proposal.Exclusions,
				proposal.Rejections,
				proposal.Distributions,
			)

			from := clientCtx.GetFromAddress()

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
