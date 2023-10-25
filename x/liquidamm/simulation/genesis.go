package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// Simulation parameter constants.
const (
	RewardsAuctionDuration = "rewards_auction_duration"
)

// GenRewardsAuctionDuration returns randomized rewards auction duration.
func GenRewardsAuctionDuration(r *rand.Rand) time.Duration {
	return time.Duration(simtypes.RandIntBetween(r, 1, 8)) * time.Hour
}

// RandomizedGenState generates a random GenesisState.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, RewardsAuctionDuration, &genesis.Params.RewardsAuctionDuration, simState.Rand,
		func(r *rand.Rand) { genesis.Params.RewardsAuctionDuration = GenRewardsAuctionDuration(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
