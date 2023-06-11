package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyDefaultTickSpacing),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenDefaultTickSpacing(r))
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxNumPrivateFarmingPlans),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenMaxNumPrivateFarmingPlans(r))
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxFarmingBlockTime),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenMaxFarmingBlockTime(r))
				return string(bz)
			},
		),
	}
}
