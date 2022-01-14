package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

// Simulation parameter constants.
const ()

// RandomizedGenState generates a random GenesisState for liquidity.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.GenesisState{
		Params: types.DefaultParams(),
	}

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated liquidity parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genesis)
}
