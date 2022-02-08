package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBatchSize),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", genBatchSize(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyTickPrecision),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", genTickPrecision(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxPriceLimitRatio),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", genMaxPriceLimitRatio(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyWithdrawFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", genWithdrawFeeRate(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxOrderLifespan),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(genMaxOrderLifespan(r))
				return fmt.Sprintf("\"%s\"", bz)
			},
		),
	}
}
