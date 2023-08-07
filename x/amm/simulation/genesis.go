package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// Simulation parameter constants
const (
	DefaultTickSpacing        = "default_tick_spacing"
	MaxNumPrivateFarmingPlans = "max_num_private_farming_plans"
	MaxFarmingBlockTime       = "max_farming_block_time"
)

func GenDefaultTickSpacing(r *rand.Rand) uint32 {
	// Exclude tick spacing 1 temporarily.
	return types.AllowedTickSpacings[1+r.Intn(len(types.AllowedTickSpacings)-1)]
}

func GenMaxNumPrivateFarmingPlans(r *rand.Rand) uint32 {
	return 1 + r.Uint32()%10
}

func GenMaxFarmingBlockTime(r *rand.Rand) time.Duration {
	return time.Duration(1+r.Intn(60)) * time.Second
}

// RandomizedGenState generates a random GenesisState for the module.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, DefaultTickSpacing, &genesis.Params.DefaultTickSpacing, simState.Rand,
		func(r *rand.Rand) { genesis.Params.DefaultTickSpacing = GenDefaultTickSpacing(r) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxNumPrivateFarmingPlans, &genesis.Params.MaxNumPrivateFarmingPlans, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxNumPrivateFarmingPlans = GenMaxNumPrivateFarmingPlans(r) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxFarmingBlockTime, &genesis.Params.MaxFarmingBlockTime, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxFarmingBlockTime = GenMaxFarmingBlockTime(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
