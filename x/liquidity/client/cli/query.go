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
		QueryDepositRequests(),
		QueryDepositRequest(),
		QueryWithdrawRequests(),
		QueryWithdrawRequest(),
		QuerySwapRequests(),
		QuerySwapRequest(),
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

			queryClient := types.NewQueryClient(clientCtx)

			pairIdStr, _ := cmd.Flags().GetString(FlagPairId)
			xDenom, _ := cmd.Flags().GetString(FlagXCoinDenom)
			yDenom, _ := cmd.Flags().GetString(FlagYCoinDenom)

			if excConditions(pairIdStr != "", xDenom != "" || yDenom != "") {
				return fmt.Errorf("invalid request")
			}

			switch {
			case pairIdStr != "":
				pairId, err := strconv.ParseUint(pairIdStr, 10, 64)
				if err != nil {
					return err
				}

				if pairId != 0 {
					res, err = queryClient.PoolsByPair(
						context.Background(),
						&types.QueryPoolsByPairRequest{
							PairId: pairId,
						})
					if err != nil {
						return err
					}
				}

			case xDenom != "" || yDenom != "":
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

			default:
				res, err = queryClient.Pools(
					context.Background(),
					&types.QueryPoolsRequest{
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
		Args:  cobra.MaximumNArgs(1),
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

			var poolId *uint64
			if len(args) > 0 {
				id, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("parse pool id: %w", err)
				}
				poolId = &id
			}
			poolCoinDenom, _ := cmd.Flags().GetString(FlagPoolCoinDenom)
			reserveAcc, _ := cmd.Flags().GetString(FlagReserveAcc)

			if !excConditions(poolId != nil, poolCoinDenom != "", reserveAcc != "") {
				return fmt.Errorf("invalid request")
			}

			queryClient := types.NewQueryClient(clientCtx)
			var res *types.QueryPoolResponse
			switch {
			case poolId != nil:
				res, err = queryClient.Pool(
					context.Background(),
					&types.QueryPoolRequest{
						PoolId: *poolId,
					},
				)
				if err != nil {
					return err
				}
			case poolCoinDenom != "":
				res, err = queryClient.PoolByPoolCoinDenom(
					context.Background(),
					&types.QueryPoolByPoolCoinDenomRequest{
						PoolCoinDenom: poolCoinDenom,
					})
				if err != nil {
					return err
				}
			case reserveAcc != "":
				res, err = queryClient.PoolByReserveAcc(
					context.Background(),
					&types.QueryPoolByReserveAccRequest{
						ReserveAcc: reserveAcc,
					})
				if err != nil {
					return err
				}
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

			xDenom, _ := cmd.Flags().GetString(FlagXCoinDenom)
			yDenom, _ := cmd.Flags().GetString(FlagYCoinDenom)

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
$ %s query %s deposit-requests 1
`,
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DepositRequests(
				context.Background(),
				&types.QueryDepositRequestsRequest{
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

func QueryWithdrawRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-requests [pool-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for all withdraw requests",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all withdraw requests.
Example:
$ %s query %s withdraw-requests 1
`,
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

			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.WithdrawRequests(
				context.Background(),
				&types.QueryWithdrawRequestsRequest{
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

func QueryWithdrawRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-request [pool-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of the specific withdraw request",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the specific withdraw request.
Example:
$ %s query %s withdraw-requests 1 1
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

			res, err := queryClient.WithdrawRequest(
				context.Background(),
				&types.QueryWithdrawRequestRequest{
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

func QuerySwapRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-requests [pair-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for all swap requests",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all swap requests.
Example:
$ %s query %s swap-requests
`,
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

			pairId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SwapRequests(
				context.Background(),
				&types.QuerySwapRequestsRequest{
					PairId:     pairId,
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

func QuerySwapRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-request [pair-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query details of the specific swap request",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details of the specific swap request.
Example:
$ %s query %s swap-requests 1 1
`,
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

			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SwapRequest(
				context.Background(),
				&types.QuerySwapRequestRequest{
					PairId: pairId,
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
