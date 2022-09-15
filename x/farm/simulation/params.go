package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyFeeCollector),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenFeeCollector(r))
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxNumPrivatePlans),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenMaxNumPrivatePlans(r))
				return string(bz)
			},
		),
	}
}