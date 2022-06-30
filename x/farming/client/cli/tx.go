package cli

// DONTCOVER
// client is excluded from test coverage in MVP version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// GetTxCmd returns a root CLI command handler for all x/farming transaction commands.
func GetTxCmd() *cobra.Command {
	farmingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Farming transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	farmingTxCmd.AddCommand(
		NewCreateFixedAmountPlanCmd(),
		NewStakeCmd(),
		NewUnstakeCmd(),
		NewHarvestCmd(),
		NewRemovePlanCmd(),
	)
	if keeper.EnableRatioPlan {
		farmingTxCmd.AddCommand(NewCreateRatioPlanCmd())
	}
	if keeper.EnableAdvanceEpoch {
		farmingTxCmd.AddCommand(NewAdvanceEpochCmd())
	}

	return farmingTxCmd
}

// NewCreateFixedAmountPlanCmd implements the create a fixed amount plan command handler.
func NewCreateFixedAmountPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-fixed-plan [plan-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Create private fixed amount farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create private fixed amount farming plan.
The plan details must be provided through a JSON file. 
		
Example:
$ %s tx %s create-private-fixed-plan <path/to/plan.json> --from mykey 

Where plan.json contains:

{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "pool1",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-08-06T09:00:00Z",
  "end_time": "2022-08-13T09:00:00Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "1"
    }
  ]
}

Description for the parameters:

[name]: specifies the name for the plan 
[staking_coin_weights]: specifies coin weights for the plan
[start_time]: specifies the time for the plan to start 
[end_time]: specifies the time for the plan to end
[epoch_amount]: specifies an amount to distribute for every epoch
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			plan, err := ParsePrivateFixedPlan(args[0])
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to parse %s file due to %v", args[0], err)
			}

			msg := types.NewMsgCreateFixedAmountPlan(
				plan.Name,
				clientCtx.GetFromAddress(),
				plan.StakingCoinWeights,
				plan.StartTime,
				plan.EndTime,
				plan.EpochAmount,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCreateRatioPlanCmd implements the create a ratio plan command handler.
func NewCreateRatioPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-ratio-plan [plan-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Create private ratio farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create private ratio farming plan.
The plan details must be provided through a JSON file. 
		
Example:
$ %s tx %s create-private-ratio-plan <path/to/plan.json> --from mykey 

Where plan.json contains:

{
  "name": "This plan intends to provide incentives for Cosmonauts!",
  "staking_coin_weights": [
    {
      "denom": "pool1",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-08-06T09:00:00Z",
  "end_time": "2022-08-13T09:00:00Z",
  "epoch_ratio": "1.000000000000000000"
}

Description for the parameters:

[name]: specifies the name for the plan 
[staking_coin_weights]: specifies coin weights for the plan
[start_time]: specifies the time for the plan to start 
[end_time]: specifies the time for the plan to end
[epoch_ratio]: specifies a ratio to distribute for every epoch. 1.000000000000000000 means to distribute all coins for an epoch
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			plan, err := ParsePrivateRatioPlan(args[0])
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to parse %s file due to %v", args[0], err)
			}

			msg := types.NewMsgCreateRatioPlan(
				plan.Name,
				clientCtx.GetFromAddress(),
				plan.StakingCoinWeights,
				plan.StartTime,
				plan.EndTime,
				plan.EpochRatio,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewStakeCmd implements the stake coin(s) command handler.
func NewStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Stake coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Stake coins. 
			
To get farming rewards, you must stake coins that are defined in available plans on a network. 

Example:
$ %s tx %s stake 1000pool1 --from mykey
$ %s tx %s stake 500pool1,500pool2 --from mykey
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			farmer := clientCtx.GetFromAddress()

			stakingCoins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgStake(farmer, stakingCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewUnstakeCmd implements the unstake coin(s) command handler.
func NewUnstakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Unstake coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unstake coins. 
			
Note that unstaking doesn't require any period and your accumulated rewards are automatically withdrawn to your wallet.

Example:
$ %s tx %s unstake 500pool1 --from mykey
$ %s tx %s unstake 500pool1,500pool2 --from mykey
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			farmer := clientCtx.GetFromAddress()

			unstakingCoins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgUnstake(farmer, unstakingCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewHarvestCmd implements the harvest rewards command handler.
func NewHarvestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harvest [staking-coin-denoms]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Harvest farming rewards",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Harvest farming rewards from the staking coin denoms that are defined in the availble plans.

Example:
$ %s tx %s harvest pool1 --from mykey
$ %s tx %s harvest pool1,pool2 --from mykey
$ %s tx %s harvest --all --from mykey
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			var denoms []string

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			farmer := clientCtx.GetFromAddress()

			switch len(args) {
			case 0:
				all, _ := cmd.Flags().GetBool(FlagAll)
				if !all {
					return fmt.Errorf("either staking-coin-denoms or --all flag must be specified")
				}

				// The transaction cannot be generated offline with --all flag since it
				// requires a query to get all the staking denoms.
				if clientCtx.Offline {
					return fmt.Errorf("cannot generate tx in offline mode with --all flag")
				}

				queryClient := types.NewQueryClient(clientCtx)
				resp, err := queryClient.Position(cmd.Context(), &types.QueryPositionRequest{
					Farmer: farmer.String(),
				})
				if err != nil {
					return err
				}
				for _, stakedCoin := range resp.StakedCoins {
					denoms = append(denoms, stakedCoin.Denom)
				}
				if len(denoms) == 0 {
					return fmt.Errorf("there is no staked coins")
				}
			case 1:
				denoms = strings.Split(args[0], ",")
			default:
				return fmt.Errorf("either staking-coin-denoms or --all flag must be specified")
			}

			msg := types.NewMsgHarvest(farmer, denoms)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetHarvest())
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewRemovePlanCmd implements the remove plan handler.
func NewRemovePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-plan [plan-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Remove a terminated private plan",
		Long: fmt.Sprintf(`Remove a terminated private plan and get the plan creation fee refunded.
Example:
$ %s tx %s remove-plan 1 --from mykey`,
			version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress()

			planId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parse plan id: %w", err)
			}

			msg := types.NewMsgRemovePlan(creator, planId)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewAdvanceEpochCmd implements the advance epoch by 1 command handler.
func NewAdvanceEpochCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advance-epoch",
		Args:  cobra.NoArgs,
		Short: "Advance epoch by 1 to simulate reward distribution",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Advance epoch by 1 to simulate reward distribution.
This message is available for testing purpose and it can only be enabled when you build the binary with "make install-testing" command. 

Example:
$ %s tx %s advance-epoch --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			requesterAcc := clientCtx.GetFromAddress()

			msg := types.NewMsgAdvanceEpoch(requesterAcc)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSubmitPublicPlanProposal implements the create/update/delete public farming plan command handler.
func GetCmdSubmitPublicPlanProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-farming-plan [proposal-file] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a a public farming plan along with an initial deposit. You can submit this governance proposal
to add, update, and delete farming plan. The proposal details must be supplied via a JSON file. A JSON file to add plan request proposal is 
provided below.

Example:
$ %s tx gov submit-proposal public-farming-plan <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "name": "Cosmos Hub Community Tax",
  "add_plan_requests": [
    {
      "farming_pool_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "termination_address": "cre1mzgucqnfr2l8cj5apvdpllhzt4zeuh2c5l33n3",
      "staking_coin_weights": [
        {
          "denom": "pool1",
          "amount": "0.800000000000000000"
        },
        {
          "denom": "stake",
          "amount": "0.100000000000000000"
        },
        {
          "denom": "uatom",
          "amount": "0.100000000000000000"
        }
      ],
      "start_time": "2021-08-06T09:00:00Z",
      "end_time": "2022-08-13T09:00:00Z",
      "epoch_amount": [
        {
          "denom": "uatom",
          "amount": "1"
        }
      ]
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

			proposal, err := ParsePublicPlanProposal(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			content := types.NewPublicPlanProposal(
				proposal.Title,
				proposal.Description,
				proposal.AddPlanRequests,
				proposal.ModifyPlanRequests,
				proposal.DeletePlanRequests,
			)

			from := clientCtx.GetFromAddress()

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
