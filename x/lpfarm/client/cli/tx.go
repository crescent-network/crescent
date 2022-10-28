package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
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
The newly created plan's farming pool address is automatically generated and
will have no balances in the account initially.
Manually send enough reward coins to the generated farming pool address to make
sure that the rewards allocation happens.
The plan's termination address is set to the plan creator.

[description]: a brief description of the plan
[start-time]: the time at which the plan begins, in RFC3339 format
[end-time]: the time at which the plan ends, in RFC3339 format
[reward-allocations...]: whitespace-separated list of the reward allocations

A reward allocation is specified in one of the following formats:
1. <denom>:<rewards_per_day>
2. pair<pair-id>:<rewards_per_day>

Example:
$ %s tx %s create-private-plan "New Farming Plan" 2022-01-01T00:00:00Z 2023-01-01T00:00:00Z pair1:10000stake,5000uatom pool2:5000stake --from mykey
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
				target, rewardsPerDayStr, found := strings.Cut(arg, ":")
				if !found {
					return fmt.Errorf("invalid reward allocation: %s", arg)
				}
				rewardsPerDay, err := sdk.ParseCoinsNormalized(rewardsPerDayStr)
				if err != nil {
					return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
				}
				var rewardAlloc types.RewardAllocation
				if strings.HasPrefix(target, "pair") {
					pairId, err := strconv.ParseUint(strings.TrimPrefix(target, "pair"), 10, 64)
					if err != nil {
						return fmt.Errorf("invalid reward allocation: %s: %w", arg, err)
					}
					rewardAlloc = types.NewPairRewardAllocation(pairId, rewardsPerDay)
				} else {
					rewardAlloc = types.NewDenomRewardAllocation(target, rewardsPerDay)
				}
				rewardAllocs = append(rewardAllocs, rewardAlloc)
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
$ %s tx %s farm 1000000pool1 --from mykey
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
$ %s tx %s unfarm 1000000pool1 --from mykey
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
$ %s tx %s harvest pool1 --from mykey
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

func NewCmdSubmitFarmingPlanProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "farming-plan [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a farming plan proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a farming plan proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal farming-plan <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Farming Plan Proposal",
  "description": "Let's start farming",
  "create_plan_requests": [
    {
      "description": "New Farming Plan",
      "farming_pool_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "termination_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "reward_allocations": [
        {
          "pair_id": "1",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "100000000"
            }
          ]
        },
        {
          "denom": "pool2",
          "rewards_per_day": [
            {
              "denom": "stake",
              "amount": "200000000"
            }
          ]
        }
      ],
      "start_time": "2022-01-01T00:00:00Z",
      "end_time": "2023-01-01T00:00:00Z"
    }
  ],
  "terminate_plan_requests": [
    {
      "plan_id": "1"
    },
    {
      "plan_id": "2"
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

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			content, err := ParseFarmingPlanProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			msg, err := gov.NewMsgSubmitProposal(&content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
