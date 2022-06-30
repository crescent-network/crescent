package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// GetQueryCmd returns the cli query commands for the module.
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewQueryAirdropsCmd(),
		NewQueryAirdropCmd(),
		NewQueryClaimRecordCmd(),
	)

	return cmd
}

func NewQueryAirdropsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrops",
		Args:  cobra.NoArgs,
		Short: "Query for all airdrops",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all airdrops.

Example:
$ %s query %s airdrops
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryAirdropsRequest{
				Pagination: pageReq,
			}

			resp, err := queryClient.Airdrops(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewQueryAirdropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrop [airdrop-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for the specific airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for the specific airdrop.

Example:
$ %s query %s airdrop 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			airdropId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAirdropRequest{
				AirdropId: airdropId,
			}

			resp, err := queryClient.Airdrop(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewQueryClaimRecordCmd() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "claim-record [airdrop-id] [address]",
		Args:  cobra.ExactArgs(2),
		Short: "Query the claim record for an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the claim record for an account.
This contains an address' initial claimable amounts and its completed conditions.

Example:
$ %s query %s claim-record 1 %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, types.ModuleName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			airdropId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			recipient, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.ClaimRecord(
				cmd.Context(),
				&types.QueryClaimRecordRequest{
					AirdropId: airdropId,
					Recipient: recipient.String(),
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
