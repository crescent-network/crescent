package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// DONTCOVER

// Simulation parameter constants
const (
	EpochBlocks      = "epoch_blocks"
	LiquidValidators = "liquidValidators"
)

//// GenEpochBlocks returns randomized epoch blocks.
//func GenEpochBlocks(r *rand.Rand) uint32 {
//	return uint32(simtypes.RandIntBetween(r, int(types.DefaultEpochBlocks), 10))
//}

// GenLiquidValidators returns randomized liquidValidators.
func GenLiquidValidators(r *rand.Rand) []types.LiquidValidator {
	ranLiquidValidators := make([]types.LiquidValidator, 0)

	for i := 0; i < simtypes.RandIntBetween(r, 1, 3); i++ {
		liquidValidator := types.LiquidValidator{}
		ranLiquidValidators = append(ranLiquidValidators, liquidValidator)
	}

	return ranLiquidValidators
}

// RandomizedGenState generates a random GenesisState for liquidstaking.
func RandomizedGenState(simState *module.SimulationState) {
	var liquidValidators []types.LiquidValidator
	//simState.AppParams.GetOrGenerate(
	//	simState.Cdc, LiquidValidators, &liquidValidators, simState.Rand,
	//	func(r *rand.Rand) { liquidValidators = GenLiquidValidators(r) },
	//)

	liquidValidatorGenesis := types.GenesisState{
		Params:           types.DefaultParams(),
		LiquidValidators: liquidValidators,
	}

	bz, _ := json.MarshalIndent(&liquidValidatorGenesis, "", " ")
	fmt.Printf("Selected randomly generated liquidstaking parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&liquidValidatorGenesis)
}
