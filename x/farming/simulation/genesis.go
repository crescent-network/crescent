package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/tendermint/farming/x/farming/types"
)

// Simulation parameter constants
const (
	PrivatePlanCreationFee = "private_plan_creation_fee"
)

// GenPrivatePlanCreationFee return default PrivatePlanCreationFee
func GenPrivatePlanCreationFee(r *rand.Rand) sdk.Coins {
	// TODO: randomize private plan creation fee
	return types.DefaultPrivatePlanCreationFee
}

// RandomizedGenState generates a random GenesisState for farming
func RandomizedGenState(simState *module.SimulationState) {
	var privatePlanCreationFee sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, PrivatePlanCreationFee, &privatePlanCreationFee, simState.Rand,
		func(r *rand.Rand) { privatePlanCreationFee = GenPrivatePlanCreationFee(r) },
	)

	farmingGenesis := types.GenesisState{
		Params: types.Params{
			PrivatePlanCreationFee: privatePlanCreationFee,
		},
	}

	bz, _ := json.MarshalIndent(&farmingGenesis, "", " ")
	fmt.Printf("Selected randomly generated farming parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&farmingGenesis)
}
