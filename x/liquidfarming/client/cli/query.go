package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// GetQueryCmd returns the cli query commands for the module
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
		NewQueryLiquidFarmsCmd(),
		NewQueryLiquidFarmCmd(),
		NewQueryRewardsAuctionsCmd(),
		NewQueryRewardsAuctionCmd(),
		NewQueryBidsCmd(),
		NewQueryRewardsCmd(),
		NewQueryExchangeRateCmd(),
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
	flags.AddPaginationFlagsToCmd(cmd, "liquidfarms")

	return cmd
}

func NewQueryLiquidFarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidfarm [liquid-farm-id]",
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

			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LiquidFarm(cmd.Context(), &types.QueryLiquidFarmRequest{
				LiquidFarmId: liquidFarmId,
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
		Use:   "rewards-auctions [liquid-farm-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all rewards auctions for the liquidfarm",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards auctions for the liquidfarm on a network.

Example:
$ %s query %s rewards-auctions 1
$ %s query %s rewards-auctions 1 --status=AUCTION_STATUS_SKIPPED
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			auctionStatus, _ := cmd.Flags().GetString(FlagRewardsAuctionStatus)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RewardsAuctions(cmd.Context(), &types.QueryRewardsAuctionsRequest{
				LiquidFarmId: liquidFarmId,
				Status:       auctionStatus,
				Pagination:   pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(flagSetRewardsAuctions())
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "auctions")

	return cmd
}

func NewQueryRewardsAuctionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards-auction [liquid-farm-id] [auction-id]",
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
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			auctionId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RewardsAuction(cmd.Context(), &types.QueryRewardsAuctionRequest{
				LiquidFarmId: liquidFarmId,
				AuctionId:    auctionId,
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
		Use:   "bids [liquid-farm-id] [auction-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query all bids for the rewards auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all bids for the rewards auction on a network.

Example:
$ %s query %s bids 1 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
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
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Bids(cmd.Context(), &types.QueryBidsRequest{
				LiquidFarmId: liquidFarmId,
				AuctionId:    auctionId,
				Pagination:   pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "bids")

	return cmd
}

func NewQueryRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [liquid-farm-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query accumulated rewards for liquid farm",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query accumulated rewards for liquid farm.

Example:
$ %s query %s rewards 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Rewards(cmd.Context(), &types.QueryRewardsRequest{
				LiquidFarmId: liquidFarmId,
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

func NewQueryExchangeRateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate [liquid-farm-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the exchange rate for liquid farm",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the exchange rate, such as mint rate and burn rate for liquid farm.

Example:
$ %s query %s exchange-rate 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			liquidFarmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid liquid farm id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ExchangeRate(cmd.Context(), &types.QueryExchangeRateRequest{
				LiquidFarmId: liquidFarmId,
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
