package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmosquad-labs/squad/x/claim/types"
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
		QueryParams(),
		QueryClaimRecord(),
	)

	return cmd
}

func QueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current claim parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as claim parameters.
Example:
$ %s query %s params
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryClaimRecord() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "claim-record [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the claim record for an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the claim record for an account.
This contains an address' initial claimable amounts and its completed actions.

Example:
$ %s query %s claim-record %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, types.ModuleName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			recipient, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			resp, err := queryClient.ClaimRecord(
				context.Background(),
				&types.QueryClaimRecordRequest{
					Address: recipient.String(),
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
