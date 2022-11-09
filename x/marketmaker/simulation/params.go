package simulation

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/tendermint/tendermint/libs/json"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

// DONTCOVER

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyDepositAmount),
			func(r *rand.Rand) string {
				bz, err := GenDepositAmount(r).MarshalJSON()
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncentiveBudgetAddress),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenIncentiveBudgetAddress(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyCommon),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(GenCommon(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyIncentivePairs),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(GenIncentivePairs(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
	}
}
