package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

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
			liquidity, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid liquidity: %s", args[1])
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
		Use:   "create-private-farming-plan [description] [termination-address] [start-time] [end-time] [reward-allocations...]",
		Args:  cobra.MinimumNArgs(5),
		Short: "Create a new private farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new private farming plan.
The newly created plan's farming pool address is automatically generated and
will have no balances in the account initially.
Manually send enough reward coins to the generated farming pool address to make
sure that the rewards allocation happens.
The plan's termination address is set to the plan creator.

[description]: a brief description of the plan
[termination-address]: address where the remaining farming rewards in the
farming pool transferred when the plan is terminated
[start-time]: the time at which the plan begins, in RFC3339 format
[end-time]: the time at which the plan ends, in RFC3339 format
[reward-allocations...]: whitespace-separated list of the reward allocations

A reward allocation is specified in the following format: <pool_id>:<rewards_per_day>

Example:
$ %s tx %s create-private-farming-plan "New Farming Plan" cre1... \
    2023-01-01T00:00:00Z 2024-01-01T00:00:00Z \
    1:1000000stake,500000uatom 2:500000stake --from mykey
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
			termAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("invalid termination address: %w", err)
			}
			startTime, err := time.Parse(time.RFC3339, args[2])
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			endTime, err := time.Parse(time.RFC3339, args[3])
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			var rewardAllocs []types.FarmingRewardAllocation
			for _, arg := range args[4:] {
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
				rewardAllocs = append(rewardAllocs, types.NewFarmingRewardAllocation(poolId, rewardsPerDay))
			}
			msg := types.NewMsgCreatePrivateFarmingPlan(
				clientCtx.GetFromAddress(), description, termAddr, rewardAllocs, startTime, endTime)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdSubmitPoolParameterChangeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-parameter-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a pool parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a pool parameter change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal pool-parameter-change <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Pool parameter change",
  "description": "Change tick spacing",
  "changes": [
    {
      "pool_id": "1",
      "tick_spacing": 10
    },
    {
	  "pool_id": "2",
      "tick_spacing": 5
    }
  ]
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			depositStr, _ := cmd.Flags().GetString(cli.FlagDeposit)
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return fmt.Errorf("invalid deposit: %w", err)
			}
			var proposal types.PoolParameterChangeProposal
			bz, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read proposal: %w", err)
			}
			if err = clientCtx.Codec.UnmarshalJSON(bz, &proposal); err != nil {
				return fmt.Errorf("unmarshal proposal: %w", err)
			}
			msg, err := gov.NewMsgSubmitProposal(&proposal, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	return cmd
}

func NewCmdSubmitPublicFarmingPlanProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-farming-plan [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public farming plan proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public farming plan proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal public-farming-plan <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Public Farming Plan Proposal",
  "description": "Let's start farming",
  "create_requests": [
    {
      "description": "New Farming Plan",
      "farming_pool_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "termination_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "reward_allocations": [
        {
          "pool_id": "1",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "100000000"
            }
          ]
        },
        {
          "pool_id": "2",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "200000000"
            }
          ]
        }
      ],
      "start_time": "2023-01-01T00:00:00Z",
      "end_time": "2024-01-01T00:00:00Z"
    }
  ],
  "terminate_requests": [
    {
      "farming_plan_id": "1"
    },
    {
      "farming_plan_id": "2"
    }
  ]
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			depositStr, _ := cmd.Flags().GetString(cli.FlagDeposit)
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return fmt.Errorf("invalid deposit: %w", err)
			}
			var proposal types.PublicFarmingPlanProposal
			bz, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read proposal: %w", err)
			}
			if err = clientCtx.Codec.UnmarshalJSON(bz, &proposal); err != nil {
				return fmt.Errorf("unmarshal proposal: %w", err)
			}
			msg, err := gov.NewMsgSubmitProposal(&proposal, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	return cmd
}
