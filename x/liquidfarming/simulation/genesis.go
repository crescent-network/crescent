package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// Simulation parameter constants.
const (
	FeeCollector           = "fee_collector"
	RewardsAuctionDuration = "rewards_auction_duration"
	LiquidFarms            = "liquid_farms"
)

// GenFeeCollector returns randomized test account for fee collector.
func GenFeeCollector(r *rand.Rand) string {
	return utils.TestAddress(r.Int()).String()
}

// GenRewardsAuctionDuration returns randomized rewards auction duration.
func GenRewardsAuctionDuration(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 24)) * time.Hour
}

// GenLiquidFarms returns randomized liquid farm list.
func GenLiquidFarms(r *rand.Rand) []types.LiquidFarm {
	panic("not implemented")
	numLiquidFarms := r.Intn(5)
	liquidFarms := []types.LiquidFarm{}
	for i := 0; i < numLiquidFarms; i++ {
		liquidFarm := types.LiquidFarm{
			PoolId:       uint64(i + 1),
			MinBidAmount: utils.RandomInt(r, sdk.ZeroInt(), sdk.NewInt(1_000_000)),
			FeeRate:      simulation.RandomDecAmount(r, sdk.NewDecWithPrec(1, 2)),
		}
		liquidFarms = append(liquidFarms, liquidFarm)
	}
	return liquidFarms
}

// RandomizedGenState generates a random GenesisState.
func RandomizedGenState(simState *module.SimulationState) {
	panic("not implemented")
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, RewardsAuctionDuration, &genesis.Params.RewardsAuctionDuration, simState.Rand,
		func(r *rand.Rand) { genesis.Params.RewardsAuctionDuration = GenRewardsAuctionDuration(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated farm parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
