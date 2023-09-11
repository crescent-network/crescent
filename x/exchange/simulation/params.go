package simulation

import (
	"encoding/json"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyFees),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenFees(r))
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxOrderPriceRatio),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenMaxOrderPriceRatio(r))
				return string(bz)
			},
		),
	}
}
