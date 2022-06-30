package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation.
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyBatchSize),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenBatchSize(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyTickPrecision),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenTickPrecision(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxPriceLimitRatio),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxPriceRatio(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyWithdrawFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenWithdrawFeeRate(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxOrderLifespan),
			func(r *rand.Rand) string {
				bz, _ := json.Marshal(GenMaxOrderLifespan(r))
				return fmt.Sprintf("\"%s\"", bz)
			},
		),
	}
}
