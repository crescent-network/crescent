package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmosquad-labs/squad/x/mint/types"
)

//// Simulation parameter constants
//const (
//	Inflation          = "inflation"
//	BlockTimeThreshold = "inflation_rate_change"
//	InflationMax       = "inflation_max"
//	InflationMin       = "inflation_min"
//	GoalBonded         = "goal_bonded"
//)
//
//// GenInflation randomized Inflation
//func GenInflation(r *rand.Rand) sdk.Dec {
//	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
//}
//
//// GenBlockTimeThreshold randomized BlockTimeThreshold
//func GenBlockTimeThreshold(r *rand.Rand) sdk.Dec {
//	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
//}
//
//// GenInflationMax randomized InflationMax
//func GenInflationMax(r *rand.Rand) sdk.Dec {
//	return sdk.NewDecWithPrec(20, 2)
//}
//
//// GenInflationMin randomized InflationMin
//func GenInflationMin(r *rand.Rand) sdk.Dec {
//	return sdk.NewDecWithPrec(7, 2)
//}
//
//// GenGoalBonded randomized GoalBonded
//func GenGoalBonded(r *rand.Rand) sdk.Dec {
//	return sdk.NewDecWithPrec(67, 2)
//}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	mintGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}

	// minter
	//var inflation sdk.Dec
	//simState.AppParams.GetOrGenerate(
	//	simState.Cdc, Inflation, &inflation, simState.Rand,
	//	func(r *rand.Rand) { inflation = GenInflation(r) },
	//)

	// params
	//var BlockTimeThreshold sdk.Dec
	//simState.AppParams.GetOrGenerate(
	//	simState.Cdc, BlockTimeThreshold, &BlockTimeThreshold, simState.Rand,
	//	func(r *rand.Rand) { BlockTimeThreshold = GenBlockTimeThreshold(r) },
	//)

	//mintDenom := sdk.DefaultBondDenom
	//params := types.NewParams(mintDenom, BlockTimeThreshold)

	//mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&mintGenesis)
}
