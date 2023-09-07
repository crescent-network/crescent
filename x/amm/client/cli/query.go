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
		NewQueryAllPositionsCmd(),
		NewQueryPositionCmd(),
		NewQueryPositionAssetsCmd(),
		NewQueryAddLiquiditySimulationCmd(),
		NewQueryRemoveLiquiditySimulationCmd(),
		NewQueryCollectibleCoinsCmd(),
		NewQueryAllTickInfosCmd(),
		NewQueryTickInfoCmd(),
		NewQueryAllFarmingPlansCmd(),
		NewQueryFarmingPlanCmd(),
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
	const flagMarketId = "market-id"
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
			marketId, err := cmd.Flags().GetUint64(flagMarketId)
			if err != nil {
				return fmt.Errorf("invalid market id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			res, err := queryClient.AllPools(cmd.Context(), &types.QueryAllPoolsRequest{
				MarketId:   marketId,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	cmd.Flags().Uint64(flagMarketId, 0, "Query pool by market ID")
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

func NewQueryAllPositionsCmd() *cobra.Command {
	const (
		flagPoolId = "pool-id"
		flagOwner  = "owner"
	)
	cmd := &cobra.Command{
		Use:   "positions",
		Args:  cobra.NoArgs,
		Short: "Query all positions",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all positions.

Example:
$ %s query %s positions
$ %s query %s positions --owner=cre1...
$ %s query %s positions --pool-id=1
$ %s query %s positions --pool-id=1 --owner=cre1...
`,
				version.AppName, types.ModuleName,
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
			poolId, err := cmd.Flags().GetUint64(flagPoolId)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}
			owner, _ := cmd.Flags().GetString(flagOwner)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AllPositions(cmd.Context(), &types.QueryAllPositionsRequest{
				PoolId:     poolId,
				Owner:      owner,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	cmd.Flags().Uint64(flagPoolId, 0, "Filter positions by pool ID")
	cmd.Flags().String(flagOwner, "", "Filter positions by owner")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "positions")
	return cmd
}

func NewQueryPositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "position [position-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific position.

Example:
$ %s query %s position 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Position(cmd.Context(), &types.QueryPositionRequest{
				PositionId: positionId,
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

func NewQueryPositionAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "position-assets [position-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a position's underlying assets",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a position's underlying assets.
Example:
$ %s query %s position-assets 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PositionAssets(cmd.Context(), &types.QueryPositionAssetsRequest{
				PositionId: positionId,
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

func NewQueryAddLiquiditySimulationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity-simulation [pool-id] [lower-price] [upper-price] [desired-amount]",
		Args:  cobra.ExactArgs(4),
		Short: "Simulate liquidity addition",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Simulate liquidity addition.

Example:
$ %s query %s add-liquidity-simulation 1 0.9 1.1 1000000ucre,1000000uusd
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
			lowerPrice := args[1]
			upperPrice := args[2]
			desiredAmt := args[4]
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AddLiquiditySimulation(cmd.Context(), &types.QueryAddLiquiditySimulationRequest{
				PoolId:        poolId,
				LowerPrice:    lowerPrice,
				UpperPrice:    upperPrice,
				DesiredAmount: desiredAmt,
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

func NewQueryRemoveLiquiditySimulationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity-simulation [position-id] [liquidity]",
		Args:  cobra.ExactArgs(2),
		Short: "Simulate liquidity removal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Simulate liquidity removal.

Example:
$ %s query %s remove-liquidity-simulation 1 20000000
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			liquidity := args[1]
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RemoveLiquiditySimulation(cmd.Context(), &types.QueryRemoveLiquiditySimulationRequest{
				PositionId: positionId,
				Liquidity:  liquidity,
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

func NewQueryCollectibleCoinsCmd() *cobra.Command {
	const (
		flagOwner      = "owner"
		flagPositionId = "position-id"
	)
	cmd := &cobra.Command{
		Use:   "collectible-coins",
		Args:  cobra.NoArgs,
		Short: "Query collectible coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query collectible coins.
Exactly one of both owner and position id flags must be specified.

Example:
$ %s query %s collectible-coins --owner=cre1...
$ %s query %s collectible-coins --position-id=1
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
			owner, _ := cmd.Flags().GetString(flagOwner)
			positionId, err := cmd.Flags().GetUint64(flagPositionId)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.CollectibleCoins(cmd.Context(), &types.QueryCollectibleCoinsRequest{
				Owner:      owner,
				PositionId: positionId,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	cmd.Flags().String(flagOwner, "", "Query the owner's collectible coins")
	cmd.Flags().Uint64(flagPositionId, 0, "Query the position's collectible coins")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewQueryAllTickInfosCmd() *cobra.Command {
	const (
		flagLowerTick = "lower-tick"
		flagUpperTick = "upper-tick"
	)
	cmd := &cobra.Command{
		Use:   "tick-infos [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all tick infos",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all tick infos.

Example:
$ %s query %s tick-infos 1
$ %s query %s tick-infos 1 --lower-tick=-1000
$ %s query %s tick-infos 1 --upper-tick=1000
$ %s query %s tick-infos 1 --lower-tick=-1000 --upper-tick=1000
`,
				version.AppName, types.ModuleName,
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
			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pool id: %w", err)
			}
			lowerTick, _ := cmd.Flags().GetString(flagLowerTick)
			upperTick, _ := cmd.Flags().GetString(flagUpperTick)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AllTickInfos(cmd.Context(), &types.QueryAllTickInfosRequest{
				PoolId:     poolId,
				LowerTick:  lowerTick,
				UpperTick:  upperTick,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	cmd.Flags().String(flagLowerTick, "", "Query tick infos above lower tick inclusive")
	cmd.Flags().String(flagUpperTick, "", "Query tick infos below upper tick inclusive")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "tick infos")
	return cmd
}

func NewQueryTickInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tick-info [pool-id] [tick]",
		Args:  cobra.ExactArgs(2),
		Short: "Query a specific tick info",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific tick info.

Example:
$ %s query %s tick-info 1 1000
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
			tick, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid tick: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TickInfo(cmd.Context(), &types.QueryTickInfoRequest{
				PoolId: poolId,
				Tick:   int32(tick),
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

func NewQueryFarmingPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "farming-plan [plan-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a specific farming plan.

Example:
$ %s query %s farming-plan 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			planId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid plan id: %w", err)
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.FarmingPlan(cmd.Context(), &types.QueryFarmingPlanRequest{
				PlanId: planId,
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
