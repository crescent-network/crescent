package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/tendermint/farming/x/farming/types"
)

// GetQueryCmd returns a root CLI command handler for all x/farming query commands.
func GetQueryCmd() *cobra.Command {
	farmingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the farming module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	farmingQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryPlans(),
		GetCmdQueryPlan(),
		GetCmdQueryStakings(),
		GetCmdQueryStaking(),
		GetCmdQueryRewards(),
	)

	return farmingQueryCmd
}

func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current farming parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as farming parameters.

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

func GetCmdQueryPlans() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Args:  cobra.NoArgs,
		Short: "Query for all plans",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all farming plans on a network.

Example:
$ %s query %s plans
$ %s query %s plans --plan-type private
$ %s query %s plans --farming-pool-addr %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
$ %s query %s plans --reward-pool-addr %s1gshap5099dwjdlxk2ym9z8u40jtkm7hvux45pze8em08fwarww6qc0tvl0
$ %s query %s plans --termination-addr %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
$ %s query %s plans --staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			planType, _ := cmd.Flags().GetString(FlagPlanType)
			farmingPoolAddr, _ := cmd.Flags().GetString(FlagFarmingPoolAddr)
			rewardPoolAddr, _ := cmd.Flags().GetString(FlagRewardPoolAddr)
			terminationAddr, _ := cmd.Flags().GetString(FlagTerminationAddr)
			stakingCoinDenom, _ := cmd.Flags().GetString(FlagStakingCoinDenom)

			var resp *types.QueryPlansResponse

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			if planType != "" {
				var pType types.PlanType
				if planType == "public" {
					pType = types.PlanTypePublic
				} else if planType == "private" {
					pType = types.PlanTypePrivate
				} else {
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "plan type must be either public or private")
				}

				resp, err = queryClient.Plans(cmd.Context(), &types.QueryPlansRequest{
					Type:               pType.String(),
					FarmingPoolAddress: farmingPoolAddr,
					RewardPoolAddress:  rewardPoolAddr,
					TerminationAddress: terminationAddr,
					StakingCoinDenom:   stakingCoinDenom,
					Pagination:         pageReq,
				})
				if err != nil {
					return err
				}
			} else {
				resp, err = queryClient.Plans(cmd.Context(), &types.QueryPlansRequest{
					FarmingPoolAddress: farmingPoolAddr,
					RewardPoolAddress:  rewardPoolAddr,
					TerminationAddress: terminationAddr,
					StakingCoinDenom:   stakingCoinDenom,
					Pagination:         pageReq,
				})
				if err != nil {
					return err
				}
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().AddFlagSet(flagSetPlans())
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "plans")

	return cmd
}

func GetCmdQueryPlan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [plan-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about a specific plan.

Example:
$ %s query %s plan
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
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "plan-id %s is not valid", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)

			resp, err := queryClient.Plan(cmd.Context(), &types.QueryPlanRequest{
				PlanId: planId,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryStakings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stakings [optional flags]",
		Args:  cobra.NoArgs,
		Short: "Query for all stakings",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all farming stakings on a network.

Example:
$ %s query %s stakings
$ %s query %s stakings --farmer-addr %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
$ %s query %s stakings --staking-coin-denom poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			farmerAddr, _ := cmd.Flags().GetString(FlagFarmerAddr)
			stakingCoinDenom, _ := cmd.Flags().GetString(FlagStakingCoinDenom)

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := queryClient.Stakings(cmd.Context(), &types.QueryStakingsRequest{
				Farmer:           farmerAddr,
				StakingCoinDenom: stakingCoinDenom,
				Pagination:       pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().AddFlagSet(flagSetStaking())
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "stakings")

	return cmd
}

func GetCmdQueryStaking() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "staking [staking-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query a specific staking",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about a specific plan.

Example:
$ %s query %s staking 1
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			stakingId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "staking-id %s is not valid", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)

			resp, err := queryClient.Staking(cmd.Context(), &types.QueryStakingRequest{
				StakingId: stakingId,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards",
		Args:  cobra.NoArgs,
		Short: "Query for all rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query rewards that are accumulated on a network.

Example:
$ %s query %s rewards
$ %s query %s rewards --farmer-addr %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
$ %s query %s rewards --staking-coin-denom uatom
$ %s query %s rewards --staking-coin-denom uatom --farmer-addr %s1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
				version.AppName, types.ModuleName, sdk.Bech32MainPrefix,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			farmerAddr, _ := cmd.Flags().GetString(FlagFarmerAddr)
			stakingCoinDenom, _ := cmd.Flags().GetString(FlagStakingCoinDenom)

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			resp, err := queryClient.Rewards(cmd.Context(), &types.QueryRewardsRequest{
				Farmer:           farmerAddr,
				StakingCoinDenom: stakingCoinDenom,
				Pagination:       pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	cmd.Flags().AddFlagSet(flagSetRewards())
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "rewards")

	return cmd
}
