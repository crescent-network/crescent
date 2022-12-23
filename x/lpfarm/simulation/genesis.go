package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/lpfarm/types"
)

// Simulation parameter constants
const (
	FeeCollector       = "fee_collector"
	MaxNumPrivatePlans = "max_num_private_plans"
)

func GenFeeCollector(r *rand.Rand) string {
	return utils.TestAddress(r.Int()).String()
}

func GenMaxNumPrivatePlans(r *rand.Rand) uint32 {
	return uint32(5 + r.Intn(100))
}

// RandomizedGenState generates a random GenesisState for the module.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, FeeCollector, &genesis.Params.FeeCollector, simState.Rand,
		func(r *rand.Rand) { genesis.Params.FeeCollector = GenFeeCollector(r) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxNumPrivatePlans, &genesis.Params.MaxNumPrivatePlans, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxNumPrivatePlans = GenMaxNumPrivatePlans(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated lpfarm parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
