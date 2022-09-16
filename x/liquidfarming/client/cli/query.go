package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// GetQueryCmd returns the cli query commands for the module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewQueryParamsCmd(),
		NewQueryLiquidFarmsCmd(),
		NewQueryLiquidFarmCmd(),
		NewQueryRewardsAuctionsCmd(),
		NewQueryRewardsAuctionCmd(),
		NewQueryBidsCmd(),
	)

	return cmd
}

func NewQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current liquidfarming parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as liquidfarming parameters.
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

func NewQueryLiquidFarmsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidfarms",
		Args:  cobra.NoArgs,
		Short: "Query for all liquidfarms",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all liquidfarms on a network.

Example:
$ %s query %s liquidfarms
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LiquidFarms(cmd.Context(), &types.QueryLiquidFarmsRequest{
				Pagination: pageReq,
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

func NewQueryLiquidFarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidfarm [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the specific liquidfarm",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the specific liquidfarm on a network.

Example:
$ %s query %s liquidfarm 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LiquidFarm(cmd.Context(), &types.QueryLiquidFarmRequest{
				PoolId: poolId,
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

func NewQueryRewardsAuctionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-auctions [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all rewards auctions for the liquidfarm",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards auctions for the liquidfarm on a network.

Example:
$ %s query %s rewards-auctions 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RewardsAuctions(cmd.Context(), &types.QueryRewardsAuctionsRequest{
				PoolId:     poolId,
				Pagination: pageReq,
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

func NewQueryRewardsAuctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-auction [pool-id] [auction-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query the specific reward auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the specific reward auction on a network.

Example:
$ %s query %s rewards-auction 1 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			auctionId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to auction pool id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RewardsAuction(cmd.Context(), &types.QueryRewardsAuctionRequest{
				PoolId:    poolId,
				AuctionId: auctionId,
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

func NewQueryBidsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bids [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all bids for the rewards auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all bids for the rewards auction on a network.

Example:
$ %s query %s bids
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse pool id: %w", err)
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Bids(cmd.Context(), &types.QueryBidsRequest{
				PoolId:     poolId,
				Pagination: pageReq,
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
