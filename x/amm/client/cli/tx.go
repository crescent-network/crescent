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

	"github.com/crescent-network/crescent/v5/x/amm/types"
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
		NewCreatePoolCmd(),
		NewAddLiquidityCmd(),
		NewRemoveLiquidityCmd(),
		NewCollectCmd(),
		NewCreatePrivateFarmingPlanCmd(),
	)

	return cmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [market-id] [price]",
		Args:  cobra.ExactArgs(2),
		Short: "Create a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a pool.

Example:
$ %s tx %s create-pool 1 10 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			marketId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid market id: %w", err)
			}
			price, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}
			msg := types.NewMsgCreatePool(clientCtx.GetFromAddress(), marketId, price)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewAddLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [pool-id] [lower-price] [upper-price] [desired-amount]",
		Args:  cobra.ExactArgs(4),
		Short: "Add liquidity to a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add liquidity to a pool.

Example:
$ %s tx %s add-liquidity 1 9.5 10.5 1000000ucre,10000000uusd --from mykey
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
				return fmt.Errorf("invalid pool id: %w", err)
			}
			lowerPrice, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid lower price: %w", err)
			}
			upperPrice, err := sdk.NewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid upper price: %w", err)
			}
			desiredAmt, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return fmt.Errorf("invalid desired amount: %w", err)
			}
			msg := types.NewMsgAddLiquidity(
				clientCtx.GetFromAddress(), poolId, lowerPrice, upperPrice, desiredAmt)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewRemoveLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [position-id] [liquidity]",
		Args:  cobra.ExactArgs(2),
		Short: "Remove liquidity from a pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove liquidity from a pool.

Example:
$ %s tx %s remove-liquidity 1 10000000000000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			liquidity, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid liquidity: %w", err)
			}
			msg := types.NewMsgRemoveLiquidity(
				clientCtx.GetFromAddress(), positionId, liquidity)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCollectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect [position-id] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Collect fees and farming rewards accrued in the position",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Collect fees and farming rewards accrued in the position.

Example:
$ %s tx %s collect 1 100000ucre,1000000uusd --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			positionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid position id: %w", err)
			}
			amt, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}
			msg := types.NewMsgCollect(
				clientCtx.GetFromAddress(), positionId, amt)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCreatePrivateFarmingPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-farming-plan [description] [start-time] [end-time] [reward-allocations...]",
		Args:  cobra.MinimumNArgs(4),
		Short: "Create a new private farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new private farming plan.
The newly created plan's farming pool address is automatically generated and
will have no balances in the account initially.
Manually send enough reward coins to the generated farming pool address to make
sure that the rewards allocation happens.
The plan's termination address is set to the plan creator.

[description]: a brief description of the plan
[start-time]: the time at which the plan begins, in RFC3339 format
[end-time]: the time at which the plan ends, in RFC3339 format
[reward-allocations...]: whitespace-separated list of the reward allocations

A reward allocation is specified in the following format: <pool-id>:<rewards_per_day>

Example:
$ %s tx %s create-private-farming-plan "New Farming Plan" 2023-01-01T00:00:00Z 2024-01-01T00:00:00Z 1:1000000stake,500000uatom 2:500000stake --from mykey
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
				poolIdStr, rewardsPerDayStr, found := strings.Cut(arg, ":")
				if !found {
					return fmt.Errorf("invalid reward allocation: %s", arg)
				}
				poolId, err := strconv.ParseUint(poolIdStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
				}
				rewardsPerDay, err := sdk.ParseCoinsNormalized(rewardsPerDayStr)
				if err != nil {
					return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
				}
				rewardAllocs = append(rewardAllocs, types.NewRewardAllocation(poolId, rewardsPerDay))
			}
			msg := types.NewMsgCreatePrivateFarmingPlan(
				clientCtx.GetFromAddress(), description, rewardAllocs, startTime, endTime)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
