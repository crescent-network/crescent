package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/crescent-network/crescent/v2/x/claim/types"
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
	)

	return cmd
}

func NewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [airdrop-id] [condition-type]",
		Args:  cobra.ExactArgs(2),
		Short: "Claim the claimable amount with a condition type",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim your claimable amount with a condition type. 
Full allocation can be claimed by completing all tasks in core network activities. 
There are 4 different tasks (condition types) and you must complete the task before claiming the amount. 
Reference the spec docs to understand the mechanism. 

Example:
$ %s tx %s claim 1 deposit --from mykey
$ %s tx %s claim 1 swap --from mykey
$ %s tx %s claim 1 liquidstake --from mykey
$ %s tx %s claim 1 vote --from mykey
`,
				version.AppName, types.ModuleName,
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
