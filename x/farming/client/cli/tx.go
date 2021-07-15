package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
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
		NewClaimCmd(),
	)

	return farmingTxCmd
}

func NewCreateFixedAmountPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-fixed-plan",
		Aliases: []string{"cf"},
		Args:    cobra.ExactArgs(2),
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
			epochDays := uint32(1)
			epochAmount := sdk.Coins{}

			msg := types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr,
				stakingCoinWeights,
				startTime,
				endTime,
				epochDays,
				epochAmount,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewCreateRatioPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-ratio-plan",
		Aliases: []string{"cr"},
		Args:    cobra.ExactArgs(2),
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
			epochDays := uint32(1)
			epochRatio := sdk.Dec{}

			msg := types.NewMsgCreateRatioPlan(
				farmingPoolAddr,
				stakingCoinWeights,
				startTime,
				endTime,
				epochDays,
				epochRatio,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewStakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake",
		Args:  cobra.ExactArgs(2),
		Short: "stake coins to the farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`stake coins to the farming plan.
Example:
$ %s tx %s stake --from mykey
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
			planID := uint64(1)
			farmer := sdk.AccAddress{}
			stakingCoins := sdk.Coins{}

			msg := types.NewMsgStake(
				planID,
				farmer,
				stakingCoins,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewUnstakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake",
		Args:  cobra.ExactArgs(2),
		Short: "unstake coins from the farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`unstake coins from the farming plan.
Example:
$ %s tx %s unstake --from mykey
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
			planID := uint64(1)
			farmer := sdk.AccAddress{}
			stakingCoins := sdk.Coins{}

			msg := types.NewMsgUnstake(
				planID,
				farmer,
				stakingCoins,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Args:  cobra.ExactArgs(2),
		Short: "claim farming rewards from the farming plan",
		Long: strings.TrimSpace(
			fmt.Sprintf(`claim farming rewards from the farming plan.
Example:
$ %s tx %s claim --from mykey
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
			planID := uint64(1)
			farmer := sdk.AccAddress{}

			msg := types.NewMsgClaim(
				planID,
				farmer,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
