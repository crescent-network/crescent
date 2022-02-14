package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	claimtypes "github.com/cosmosquad-labs/squad/x/claim/types"
	liqtypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
	liqstakingtypes "github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func PrepareGenesisCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis [network-type] [chain-id]",
		Args:  cobra.ExactArgs(2),
		Short: "Prepare a genesis file with initial setup",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Prepare a genesis file with initial setup.

The initial setup includes initial params for squad modules

Example:
$ %s prepare-genesis mainnet squad-1
$ %s prepare-genesis testnet squad-1

The genesis output file is at $HOME/.squad/config/genesis.json
`,
				version.AppName,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)

			serverCfg := serverCtx.Config

			// Read genesis file
			genFile := serverCfg.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// Parse genesis params depending on the network type
			networkType := args[0]
			genParams := ParseGenesisParams(networkType)
			if genParams == nil {
				return fmt.Errorf("you must choose between mainnet (m) or testnet (t): %s", args[0])
			}

			// Prepare genesis
			chainID := args[1]
			appState, genDoc, err = PrepareGenesis(clientCtx, appState, genDoc, genParams, chainID)
			if err != nil {
				return fmt.Errorf("failed to prepare genesis %w", err)
			}

			if err := mbm.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("failed to validate genesis file: %w", err)
			}

			// Marshal and save the app state
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			// Export the genesis state to a file
			if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return fmt.Errorf("failed to export genesis file %w", err)
			}

			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func PrepareGenesis(
	clientCtx client.Context,
	appState map[string]json.RawMessage,
	genDoc *tmtypes.GenesisDoc,
	genParams *GenesisParams,
	chainID string,
) (map[string]json.RawMessage, *tmtypes.GenesisDoc, error) {
	cdc := clientCtx.Codec

	genDoc.ChainID = chainID
	genDoc.GenesisTime = genParams.GenesisTime
	genDoc.ConsensusParams = genParams.ConsensusParams

	// Gov module genesis
	govGenState := govtypes.DefaultGenesisState()
	govGenState.DepositParams = genParams.GovParams.DepositParams
	govGenState.TallyParams = genParams.GovParams.TallyParams
	govGenState.VotingParams = genParams.GovParams.VotingParams
	govGenStateBz := cdc.MustMarshalJSON(govGenState)
	appState[govtypes.ModuleName] = govGenStateBz

	// Claim module genesis
	// claimGenState := claimtypes.DefaultGenesis()
	// claimGenState = genParams.ClaimGenesisState
	// claimGenStateBz := cdc.MustMarshalJSON(claimGenState)
	// appState[claimtypes.ModuleName] = claimGenStateBz

	return appState, genDoc, nil
}

type GenesisParams struct {
	GenesisTime     time.Time
	ChainId         string
	ConsensusParams *tmproto.ConsensusParams

	StakingParams       stakingtypes.Params
	GovParams           govtypes.Params
	LiquidityParams     liqtypes.Params
	LiquidStakingParams liqstakingtypes.Params
	ClaimParams         claimtypes.Params

	ClaimGenesisState claimtypes.GenesisState
}

func TestnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}
	genParams.GenesisTime = time.Now()

	// Set claim records
	genParams.ClaimParams = claimtypes.DefaultParams()
	genParams.ClaimGenesisState.ClaimRecords = []claimtypes.ClaimRecord{
		{
			Address:               "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v", // validator
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("airdrop", sdk.NewInt(500_000_000))),
			DepositActionClaimed:  false,
			SwapActionClaimed:     false,
			FarmingActionClaimed:  false,
		},
		{
			Address:               "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu", // user1
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("airdrop", sdk.NewInt(200_000_000))),
			DepositActionClaimed:  false,
			SwapActionClaimed:     false,
			FarmingActionClaimed:  false,
		},
		{
			Address:               "cosmos185fflsvwrz0cx46w6qada7mdy92m6kx4gqx0ny", // user2
			InitialClaimableCoins: sdk.NewCoins(sdk.NewCoin("airdrop", sdk.NewInt(900_000_000))),
			DepositActionClaimed:  false,
			SwapActionClaimed:     false,
			FarmingActionClaimed:  false,
		},
	}

	return genParams
}

func MainnetGenesisParams() *GenesisParams {
	genParams := &GenesisParams{}

	// TODO: not implemented yet

	return genParams
}

// ParseGenesisParams return GenesisParams based on the network type.
func ParseGenesisParams(networkType string) *GenesisParams {
	switch strings.ToLower(networkType) {
	case "t", "testnet":
		return TestnetGenesisParams()
	case "m", "mainnet":
		return MainnetGenesisParams()
	default:
		return nil
	}
}
