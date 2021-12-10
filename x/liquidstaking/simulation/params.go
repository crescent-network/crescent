package simulation

// DONTCOVER

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		//simulation.NewSimParamChange(types.ModuleName, string(types.KeyEpochBlocks),
		//	func(r *rand.Rand) string {
		//		return fmt.Sprintf("%d", GenEpochBlocks(r))
		//	},
		//),
		//simulation.NewSimParamChange(types.ModuleName, string(types.KeyBiquidStakings),
		//	func(r *rand.Rand) string {
		//		bz, err := json.Marshal(GenBiquidStakings(r))
		//		if err != nil {
		//			panic(err)
		//		}
		//		return string(bz)
		//	},
		//),
	}
}
