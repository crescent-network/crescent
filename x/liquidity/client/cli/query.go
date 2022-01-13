package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// GetQueryCmd returns the cli query commands for this module
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
		QueryPools(),
		QueryPool(),
		QueryPairs(),
		QueryPair(),
		// QueryDepositRequests(),
		// QueryDepositRequest(),
		// QueryWithdrawRequests(),
		// QueryWithdrawRequest(),
		// QuerySwapRequests(),
		// QuerySwapRequest(),
	)

	return cmd
}

func QueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current liquidity parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as liquidity parameters.
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

			resp, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&resp.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Args:  cobra.NoArgs,
		Short: "Query for all liquidity pools",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all existing liquidity pools on a network.
Example:
$ %s query %s pools
$ %s query %s pools --pair-id=[pair-id]
$ %s query %s pools --x-denom=[denom]
$ %s query %s pools --y-denom=[denom]
$ %s query %s pools --x-denom=[denom] --y-denom=[denom]
`,
				version.AppName, types.ModuleName,
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			var res *types.QueryPoolsResponse

			foundArg := false
			queryClient := types.NewQueryClient(clientCtx)

			pairIdStr, _ := cmd.Flags().GetString(FlagPairId)
			if pairIdStr != "" {
				foundArg = true
				pairId, err := strconv.ParseUint(pairIdStr, 10, 64)
				if err != nil {
					return err
				}

				if pairId != 0 {
					foundArg = true
					res, err = queryClient.PoolsByPair(
						context.Background(),
						&types.QueryPoolsByPairRequest{
							PairId: pairId,
						})
					if err != nil {
						return err
					}
				}
			}

			if !foundArg {
				xDenom, _ := cmd.Flags().GetString(FlagXDenom)
				yDenom, _ := cmd.Flags().GetString(FlagYDenom)

				res, err = queryClient.Pools(
					context.Background(),
					&types.QueryPoolsRequest{
						XDenom:     xDenom,
						YDenom:     yDenom,
						Pagination: pageReq,
					})
				if err != nil {
					return err
				}
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(flagSetPools())
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of the liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the liquidity pool
Example:
$ %s query %s pool 1
$ %s query %s pool --pool-coin-denom=[denom]
$ %s query %s pool --reserve-acc=[address]
`,
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

			var res *types.QueryPoolResponse

			foundArg := false
			queryClient := types.NewQueryClient(clientCtx)

			poolCoinDenom, _ := cmd.Flags().GetString(FlagPoolCoinDenom)
			if poolCoinDenom != "" {
				foundArg = true
				res, err = queryClient.PoolByPoolCoinDenom(
					context.Background(),
					&types.QueryPoolByPoolCoinDenomRequest{
						PoolCoinDenom: poolCoinDenom,
					})
				if err != nil {
					return err
				}
			}

			reserveAcc, _ := cmd.Flags().GetString(FlagReserveAcc)
			if !foundArg && reserveAcc != "" {
				foundArg = true
				res, err = queryClient.PoolByReserveAcc(
					context.Background(),
					&types.QueryPoolByReserveAccRequest{
						ReserveAcc: reserveAcc,
					})
				if err != nil {
					return err
				}
			}

			if !foundArg && len(args) > 0 {
				poolID, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return err
				}

				if poolID != 0 {
					foundArg = true
					res, err = queryClient.Pool(
						context.Background(),
						&types.QueryPoolRequest{
							PoolId: poolID,
						},
					)
					if err != nil {
						return err
					}
				}
			}

			if !foundArg {
				return fmt.Errorf("provide the pool-id argument or --%s or --%s flag", FlagPoolCoinDenom, FlagReserveAcc)
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(flagSetPool())
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryPairs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pairs",
		Args:  cobra.NoArgs,
		Short: "Query for all pairs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all existing pairs on a network.
Example:
$ %s query %s pairs
$ %s query %s pairs --x-denom=[denom]
$ %s query %s pairs --y-denom=[denom]
$ %s query %s pairs --x-denom=[denom] --y-denom=[denom]
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			xDenom, _ := cmd.Flags().GetString(FlagXDenom)
			yDenom, _ := cmd.Flags().GetString(FlagYDenom)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pairs(
				context.Background(),
				&types.QueryPairsRequest{
					XDenom:     xDenom,
					YDenom:     yDenom,
					Pagination: pageReq,
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().AddFlagSet(flagSetPairs())
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pair [pair-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query details of the pair",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the pair.
Example:
$ %s query %s pairs
$ %s query %s pairs --x-denom=[denom]
$ %s query %s pairs --y-denom=[denom]
$ %s query %s pairs --x-denom=[denom] --y-denom=[denom]
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

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pair(
				context.Background(),
				&types.QueryPairRequest{
					PairId: pairId,
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

func QueryDepositRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-requests [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for all deposit requests",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all deposit requests.
Example:
$ %s query %s deposit-requests
`,
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
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DepositRequests(
				context.Background(),
				&types.QueryDepositRequestsRequest{
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

func QueryDepositRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-request [pool-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of the specific deposit request",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the specific deposit request.
Example:
$ %s query %s deposit-requests 1 1
`,
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
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DepositRequest(
				context.Background(),
				&types.QueryDepositRequestRequest{
					PoolId: poolId,
					Id:     id,
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
