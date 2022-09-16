package simulation

// DONTCOVER

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// Simulation parameter constants.
const (
	LiquidFarms = "liquid_farms"
)

// GenLiquidFarms return randomized liquid farm list.
func GenLiquidFarms(r *rand.Rand) []types.LiquidFarm {
	return []types.LiquidFarm{} // TODO: not implemented yet
}

// RandomizedGenState generates a random GenesisState.
func RandomizedGenState(simState *module.SimulationState) {
	var liquidFarms []types.LiquidFarm
	simState.AppParams.GetOrGenerate(
		simState.Cdc, LiquidFarms, &liquidFarms, simState.Rand,
		func(r *rand.Rand) { liquidFarms = GenLiquidFarms(r) },
	)

	genState := types.GenesisState{
		Params: types.Params{
			LiquidFarms: liquidFarms,
		},
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genState)
}
