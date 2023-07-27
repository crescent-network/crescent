package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
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
		NewQueryPublicPositionsCmd(),
		NewQueryPublicPositionCmd(),
		NewQueryRewardsAuctionsCmd(),
		NewQueryRewardsAuctionCmd(),
		NewQueryBidsCmd(),
		NewQueryRewardsCmd(),
	)

	return cmd
}

func NewQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current liquidamm parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as liquidamm parameters.
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

func NewQueryPublicPositionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-positions",
		Args:  cobra.NoArgs,
		Short: "Query for all public positions",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all public positions on a network.

Example:
$ %s query %s public-positions
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
			res, err := queryClient.PublicPositions(cmd.Context(), &types.QueryPublicPositionsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "public-positions")

	return cmd
}

func NewQueryPublicPositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-position [public-position-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the specific public position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the specific public position on a network.

Example:
$ %s query %s public-position 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PublicPosition(cmd.Context(), &types.QueryPublicPositionRequest{
				PublicPositionId: publicPositionId,
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
		Use:   "rewards-auctions [public-position-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all rewards auctions for the public position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards auctions for the public position on a network.

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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}
			auctionStatus, _ := cmd.Flags().GetString(FlagRewardsAuctionStatus)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RewardsAuctions(cmd.Context(), &types.QueryRewardsAuctionsRequest{
				PublicPositionId: publicPositionId,
				Status:           auctionStatus,
				Pagination:       pageReq,
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
		Use:   "rewards-auction [public-position-id] [auction-id]",
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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}
			auctionId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RewardsAuction(cmd.Context(), &types.QueryRewardsAuctionRequest{
				PublicPositionId: publicPositionId,
				AuctionId:        auctionId,
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
		Use:   "bids [public-position-id] [auction-id]",
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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
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
				PublicPositionId: publicPositionId,
				AuctionId:        auctionId,
				Pagination:       pageReq,
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
		Use:   "rewards [public-position-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query accumulated rewards for public position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query accumulated rewards for public position.

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
			publicPositionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid public position id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Rewards(cmd.Context(), &types.QueryRewardsRequest{
				PublicPositionId: publicPositionId,
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
