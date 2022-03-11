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
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmosquad-labs/squad/x/claim/types"
	liqtypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
	lstypes "github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// GetTxCmd returns the transaction commands for the module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transaction subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewClaimCmd(),
		NewDepositAndClaimCmd(),
		NewSwapAndClaimCmd(),
		NewLiquidStakeAndClaimCmd(),
		NewVoteAndClaimCmd(),
	)

	return cmd
}

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [airdrop-id] [condition-type]",
		Args:  cobra.ExactArgs(2),
		Short: "Claim the claimable amount with a condition type",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim the claimable amount with a condition type. 
There are 3 different condition types. Reference the examples below.

Example:
$ %s tx %s claim 1 deposit --from mykey
$ %s tx %s claim 1 swap --from mykey
$ %s tx %s claim 1 farming --from mykey
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

			airdropId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			condType := NormalizeConditionType(args[1])
			if condType == types.ConditionTypeUnspecified {
				return fmt.Errorf("unknown condition type %s", args[0])
			}

			msg := types.NewMsgClaim(
				airdropId,
				clientCtx.GetFromAddress(),
				condType,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDepositAndClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-deposit-claim",
		Args:  cobra.NoArgs,
		Short: "test-deposit-claim",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Test
Example
$ %s tx %s test-deposit-claim --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			recipient := clientCtx.GetFromAddress()

			msgDeposit := liqtypes.NewMsgDeposit(
				recipient,
				1,
				sdk.NewCoins(
					sdk.NewCoin("uatom", sdk.NewInt(10_000_000)),
					sdk.NewCoin("uusd", sdk.NewInt(10_000_000)),
				),
			)

			msgClaim := types.NewMsgClaim(
				1,
				recipient,
				types.ConditionTypeDeposit,
			)

			msgs := []sdk.Msg{
				msgDeposit,
				msgClaim,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSwapAndClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-swap-claim",
		Args:  cobra.NoArgs,
		Short: "test-swap-claim",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Test
Example
$ %s tx %s test-swap-claim --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recipient := clientCtx.GetFromAddress()
			direction := liqtypes.OrderDirectionSell
			offerCoin := sdk.NewInt64Coin("uatom", 1000)
			demandCoinDenom := "uusd"
			price := sdk.MustNewDecFromStr("1.0")
			amount := sdk.NewInt(1000)
			orderLifespan := 10 * time.Second

			msgSwap := &liqtypes.MsgLimitOrder{
				Orderer:         recipient.String(),
				PairId:          1,
				Direction:       direction,
				OfferCoin:       offerCoin,
				DemandCoinDenom: demandCoinDenom,
				Price:           price,
				Amount:          amount,
				OrderLifespan:   orderLifespan,
			}

			msgClaim := types.NewMsgClaim(
				1,
				recipient,
				types.ConditionTypeSwap,
			)

			msgs := []sdk.Msg{
				msgSwap,
				msgClaim,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewLiquidStakeAndClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-liquidstake-claim",
		Args:  cobra.NoArgs,
		Short: "test-liquidstake-claim",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Test
Example
$ %s tx %s test-liquidstake-claim --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			recipient := clientCtx.GetFromAddress()

			msgLiquidStake := lstypes.NewMsgLiquidStake(
				recipient,
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			)

			msgClaim := types.NewMsgClaim(
				1,
				recipient,
				types.ConditionTypeLiquidStake,
			)

			msgs := []sdk.Msg{
				msgLiquidStake,
				msgClaim,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewVoteAndClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-vote-claim",
		Args:  cobra.NoArgs,
		Short: "test-vote-claim",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Test
Example
$ %s tx %s test-vote-claim --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			recipient := clientCtx.GetFromAddress()

			msgVote := govtypes.NewMsgVote(
				recipient,
				1,
				govtypes.OptionYes,
			)

			msgClaim := types.NewMsgClaim(
				1,
				recipient,
				types.ConditionTypeVote,
			)

			msgs := []sdk.Msg{
				msgVote,
				msgClaim,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
