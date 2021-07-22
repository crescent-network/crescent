package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

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
		Use:     "create-fixed-plan",
		Aliases: []string{"cf"},
		Args:    cobra.ExactArgs(0),
		Short:   "create fixed amount farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create fixed amount farming plan.
Example:
$ %s tx %s create-fixed-plan --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			planCreator := clientCtx.GetFromAddress()

			fmt.Println("planCreator: ", planCreator)

			// TODO: replace dummy data
			farmingPoolAddr := sdk.AccAddress{}
			stakingCoinWeights := sdk.DecCoins{}
			startTime := time.Time{}
			endTime := time.Time{}
			epochAmount := sdk.Coins{}

			msg := types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr,
				stakingCoinWeights,
				startTime,
				endTime,
				epochAmount,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCreateRatioPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-ratio-plan",
		Aliases: []string{"cr"},
		Args:    cobra.ExactArgs(0),
		Short:   "create ratio farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create ratio farming plan.
Example:
$ %s tx %s create-ratio-plan --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			planCreator := clientCtx.GetFromAddress()

			fmt.Println("planCreator: ", planCreator)

			// TODO: replace dummy data
			farmingPoolAddr := sdk.AccAddress{}
			stakingCoinWeights := sdk.DecCoins{}
			startTime := time.Time{}
			endTime := time.Time{}
			epochRatio := sdk.Dec{}

			msg := types.NewMsgCreateRatioPlan(
				farmingPoolAddr,
				stakingCoinWeights,
				startTime,
				endTime,
				epochRatio,
			)

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
