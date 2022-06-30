package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyPrivatePlanCreationFee),
			func(r *rand.Rand) string {
				bz, err := GenPrivatePlanCreationFee(r).MarshalJSON()
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyNextEpochDays),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenNextEpochDays(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyFarmingFeeCollector),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenFarmingFeeCollector(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxNumPrivatePlans),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenMaxNumPrivatePlans(r))
			},
		),
	}
}
