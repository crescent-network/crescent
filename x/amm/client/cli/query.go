package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v5/x/amm/types"
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
		NewQueryAllPoolsCmd(),
		NewQueryPoolCmd(),
		NewQueryPositionsCmd(),
		NewQueryPoolPositionsCmd(),
		NewQueryAllFarmingPlansCmd(),
	)

	return cmd
}

// NewQueryParamsCmd implements the params query command.
func NewQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current amm parameters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the current amm parameters.

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

func NewQueryAllPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Args:  cobra.NoArgs,
		Short: "Query all pools",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all pools.

Example:
$ %s query %s pools
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
			res, err := queryClient.AllPools(cmd.Context(), &types.QueryAllPoolsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "pools")
	return cmd
}

func NewQueryPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific pool by its ID.

Example:
$ %s query %s pool 1
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
				return fmt.Errorf("invalid pool id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Pool(cmd.Context(), &types.QueryPoolRequest{
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

func NewQueryPositionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "positions [owner]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query positions",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all positions or query positions by the owner.
The owner argument is optional and if omitted, all positions will be returned.

Example:
$ %s query %s positions
$ %s query %s positions cre1...
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
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			if len(args) == 0 {
				res, err := queryClient.AllPositions(cmd.Context(), &types.QueryAllPositionsRequest{
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			} else {
				res, err := queryClient.Positions(cmd.Context(), &types.QueryPositionsRequest{
					Owner:      args[0],
					Pagination: pageReq,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			}
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "positions")
	return cmd
}

func NewQueryPoolPositionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-positions [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a pool's positions",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a pool's positions.

Example:
$ %s query %s pool-positions 1
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
				return fmt.Errorf("invalid pool id: %w", err)
			}
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PoolPositions(cmd.Context(), &types.QueryPoolPositionsRequest{
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
	flags.AddPaginationFlagsToCmd(cmd, "positions")
	return cmd
}

func NewQueryAllFarmingPlansCmd() *cobra.Command {
	const (
		flagIsPrivate    = "is-private"
		flagIsTerminated = "is-terminated"
	)
	cmd := &cobra.Command{
		Use:   "farming-plans",
		Args:  cobra.NoArgs,
		Short: "Query all farming plans",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all farming plans.

Example:
$ %s query %s farming-plans
$ %s query %s farming-plans --is-private=true
$ %s query %s farming-plans --is-terminated=false
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			isPrivate, _ := cmd.Flags().GetString(flagIsPrivate)
			isTerminated, _ := cmd.Flags().GetString(flagIsTerminated)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AllFarmingPlans(cmd.Context(), &types.QueryAllFarmingPlansRequest{
				IsPrivate:    isPrivate,
				IsTerminated: isTerminated,
				Pagination:   pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "farming-plans")
	cmd.Flags().String(flagIsPrivate, "", "Filter farming plans by is_private field (true|false)")
	cmd.Flags().String(flagIsTerminated, "", "Filter farming plans by is_terminated field (true|false)")
	return cmd
}
