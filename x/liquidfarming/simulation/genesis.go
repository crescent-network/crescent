package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// Simulation parameter constants.
const (
	RewardsAuctionDuration = "rewards_auction_duration"
)

// GenRewardsAuctionDuration returns randomized rewards auction duration.
func GenRewardsAuctionDuration(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 24)) * time.Hour
}

// RandomizedGenState generates a random GenesisState.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, RewardsAuctionDuration, &genesis.Params.RewardsAuctionDuration, simState.Rand,
		func(r *rand.Rand) { genesis.Params.RewardsAuctionDuration = GenRewardsAuctionDuration(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated farm parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
