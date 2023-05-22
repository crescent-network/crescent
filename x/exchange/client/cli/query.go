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

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewQueryParamsCmd(),
		NewQueryAllMarketsCmd(),
		NewQueryMarketCmd(),
		NewQueryOrderCmd(),
		NewQueryBestSwapExactAmountInRoutesCmd(),
	)

	return cmd
}

// NewQueryParamsCmd implements the params query command.
func NewQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current exchange parameters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current exchange parameters.

Example:
$ %s query %s params
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
			resp, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(&resp.Params)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewQueryAllMarketsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markets",
		Args:  cobra.NoArgs,
		Short: "Query all markets",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all markets.

Example:
$ %s query %s markets
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
			res, err := queryClient.AllMarkets(cmd.Context(), &types.QueryAllMarketsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "markets")
	return cmd
}

func NewQueryMarketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market [market-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific market",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific market by its ID.

Example:
$ %s query %s market 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			marketId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid market id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Market(cmd.Context(), &types.QueryMarketRequest{
				MarketId: marketId,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewQueryOrderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order [order-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific order by its ID.

Example:
$ %s query %s order 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			orderId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid market id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Order(cmd.Context(), &types.QueryOrderRequest{
				OrderId: orderId,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewQueryBestSwapExactAmountInRoutesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "best-swap-exact-amount-in-routes [input] [output-denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query the best routes for a swap with exact amount in",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the best routes for a swap with exact amount in.

Example:
$ %s query %s best-swap-exact-amount-in-routes 1000000stake uatom
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			input, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid input: %w", err)
			}
			outputDenom := args[1]
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.BestSwapExactAmountInRoutes(cmd.Context(), &types.QueryBestSwapExactAmountInRoutesRequest{
				Input:       input,
				OutputDenom: outputDenom,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
