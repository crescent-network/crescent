package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

// GetTxCmd returns the transaction commands for the module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreatePrivatePlanCmd(),
		NewFarmCmd(),
		NewUnfarmCmd(),
		NewHarvestCmd(),
	)

	return cmd
}

func NewCreatePrivatePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-plan [description] [start-time] [end-time] [reward-allocations...]",
		Args:  cobra.MinimumNArgs(4),
		Short: "Create a new private farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new private farming plan.

[description]: a brief description of the plan
[start-time]: the time at which the plan begins, in RFC3339 format
[end-time]: the time at which the plan ends, in RFC3339 format
[reward-allocations...]: whitespace-separated list of the reward allocations

A reward allocation is specified in following format:
<pair-id>:<rewards_per_day>

Example:
$ %s tx %s create-private-plan "New Farming Plan" 2022-01-01T00:00:00Z 2023-01-01T00:00:00Z 1:10000stake 2:5000stake,1000uatom --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			description := args[0]
			startTime, err := time.Parse(time.RFC3339, args[1])
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			endTime, err := time.Parse(time.RFC3339, args[2])
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			var rewardAllocs []types.RewardAllocation
			for _, arg := range args[3:] {
				pairIdStr, rewardsStr, found := strings.Cut(arg, ":")
				if !found {
					return fmt.Errorf("invalid reward allocation: %s", arg)
				}
				pairId, err := strconv.ParseUint(pairIdStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
				}
				rewards, err := sdk.ParseCoinsNormalized(rewardsStr)
				if err != nil {
					return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
				}
				rewardAllocs = append(rewardAllocs, types.NewRewardAllocation(pairId, rewards))
			}

			msg := types.NewMsgCreatePrivatePlan(
				clientCtx.GetFromAddress(), description, rewardAllocs, startTime, endTime)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewFarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "farm [coin]",
		Args:  cobra.ExactArgs(1),
		Short: "Start farming coin",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Start farming coin.

Example:
$ %s tx %s farm 1000000stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin: %w", err)
			}

			msg := types.NewMsgFarm(clientCtx.GetFromAddress(), coin)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUnfarmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfarm [coin]",
		Args:  cobra.ExactArgs(1),
		Short: "Unfarm farming coin",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfarm farming coin.

Example:
$ %s tx %s unfarm 1000000stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid coin: %w", err)
			}

			msg := types.NewMsgUnfarm(clientCtx.GetFromAddress(), coin)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewHarvestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harvest [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Harvest farming rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Harvest farming rewards.

Example:
$ %s tx %s harvest stake --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgHarvest(clientCtx.GetFromAddress(), args[0])

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
