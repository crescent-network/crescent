package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/crescent-network/crescent/v2/x/mint/types"
)

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	mintGenesis := types.DefaultGenesisState()
	mintGenesis.Params.InflationSchedules = []types.InflationSchedule{
		{
			StartTime: simState.GenTimestamp,
			EndTime:   simState.GenTimestamp.Add(time.Hour * 24 * 365), // 1 year
			Amount:    sdk.NewInt(300000000000000),
		},
		{
			StartTime: simState.GenTimestamp.Add(time.Hour * 24 * 365),
			EndTime:   simState.GenTimestamp.Add(time.Hour * 24 * 365 * 2),
			Amount:    sdk.NewInt(200000000000000),
		},
	}

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated mint parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
