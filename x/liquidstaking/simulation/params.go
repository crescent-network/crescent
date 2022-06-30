package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{

		simulation.NewSimParamChange(types.ModuleName, string(types.KeyWhitelistedValidators),
			func(r *rand.Rand) string {
				bz, err := json.Marshal(genWhitelistedValidator(r))
				if err != nil {
					panic(err)
				}
				return string(bz)
			},
		),

		simulation.NewSimParamChange(types.ModuleName, string(types.KeyLiquidBondDenom),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", genLiquidBondDenom(r))
			},
		),

		simulation.NewSimParamChange(types.ModuleName, string(types.KeyUnstakeFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", genUnstakeFeeRate(r).String())
			},
		),

		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMinLiquidStakingAmount),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", genMinLiquidStakingAmount(r))
			},
		),
	}
}
