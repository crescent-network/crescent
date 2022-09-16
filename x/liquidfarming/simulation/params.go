package simulation

// DONTCOVER

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyLiquidFarms),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(GenLiquidFarms(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
