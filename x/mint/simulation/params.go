package simulation

// DONTCOVER

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

const (
	keyBlockTimeThreshold = "BlockTimeThreshold"
	keyInflationMax       = "InflationMax"
	keyInflationMin       = "InflationMin"
	keyGoalBonded         = "GoalBonded"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		//simulation.NewSimParamChange(types.ModuleName, keyBlockTimeThreshold,
		//	func(r *rand.Rand) string {
		//		return fmt.Sprintf("\"%s\"", GenBlockTimeThreshold(r))
		//	},
		//),
	}
}
