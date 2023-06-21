package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyEpochBlocks),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenEpochBlocks(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBudgets),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(InitBudgets(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
