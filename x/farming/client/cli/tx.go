package cli

import (
	"fmt"
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

	"github.com/tendermint/farming/x/farming/types"
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
		NewCreateRatioPlanCmd(),
		NewStakeCmd(),
		NewUnstakeCmd(),
		NewHarvestCmd(),
	)

	return farmingTxCmd
}

func NewCreateFixedAmountPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-fixed-plan [plan-file]",
		Args:  cobra.ExactArgs(1),
		Short: "create private fixed amount farming plan",
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
      "denom": "uatom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-24T08:41:21.662422Z",
  "end_time": "2022-07-28T08:41:21.662422Z",
  "epoch_amount": [
    {
      "denom": "uatom",
      "amount": "1"
    }
  ]
}
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

			if err = msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateRatioPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-private-ratio-plan [plan-file]",
		Args:  cobra.ExactArgs(1),
		Short: "create private ratio farming plan",
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
      "denom": "uatom",
      "amount": "1.000000000000000000"
    }
  ],
  "start_time": "2021-07-15T08:41:21.662422Z",
  "end_time": "2022-07-16T08:41:21.662422Z",
  "epoch_ratio": "1.000000000000000000"
}
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

			if err = msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "stake coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`stake coins.
Example:
$ %s tx %s stake 1000uatom --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stakingCoins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgStake(clientCtx.GetFromAddress(), stakingCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUnstakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake",
		Args:  cobra.ExactArgs(1),
		Short: "unstake coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`unstake coins.
Example:
$ %s tx %s unstake 1000uatom --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			unstakingCoins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgUnstake(clientCtx.GetFromAddress(), unstakingCoins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewHarvestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "harvest",
		Args:  cobra.ExactArgs(0),
		Short: "harvest farming rewards from the farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`claim farming rewards from the farming plan.
Example:
$ %s tx %s harvest --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			farmer := clientCtx.GetFromAddress()

			stakingCoinDenoms := []string{"test"}

			msg := types.NewMsgHarvest(farmer, stakingCoinDenoms)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSubmitPublicPlanProposal implements a command handler for submitting a public farming plan transaction to create, update, delete plan.
func GetCmdSubmitPublicPlanProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-farming-plan [proposal-file] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a a public farming plan along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal public-farming-plan <path/to/proposal.json> --from=<key_or_address> --deposit=<deposit_amount>

Where proposal.json contains:

{
  "title": "Public Farming Plan",
  "description": "Are you ready to farm?",
  "name": "Cosmos Hub Community Tax",
  "add_request_proposals": [
    {
      "farming_pool_address": "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
      "termination_address": "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
      "staking_coin_weights": [
        {
          "denom": "PoolCoinDenom",
          "amount": "1.000000000000000000"
        }
      ],
      "start_time": "2021-07-15T08:41:21.662422Z",
      "end_time": "2022-07-16T08:41:21.662422Z",
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

			content, err := types.NewPublicPlanProposal(proposal.Title, proposal.Description,
				proposal.AddRequestProposals, proposal.UpdateRequestProposals, proposal.DeleteRequestProposals)
			if err != nil {
				return err
			}

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
