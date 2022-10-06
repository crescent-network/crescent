package simulation

// DONTCOVER

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
)

// Simulation parameter constants.
const (
	RewardsAuctionDuration = "rewards_auction_duration"
	LiquidFarms            = "liquid_farms"
)

// GenRewardsAuctionDuration returns randomized rewards auction duration.
func GenRewardsAuctionDuration(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 24)) * time.Hour
}

// GenLiquidFarms returns randomized liquid farm list.
func GenLiquidFarms(r *rand.Rand) []types.LiquidFarm {
	numLiquidFarms := r.Intn(5)
	liquidFarms := []types.LiquidFarm{}
	for i := 0; i < numLiquidFarms; i++ {
		liquidFarm := types.LiquidFarm{
			PoolId:        uint64(i + 1),
			MinFarmAmount: utils.RandomInt(r, sdk.ZeroInt(), sdk.NewInt(1_000_000)),
			MinBidAmount:  utils.RandomInt(r, sdk.ZeroInt(), sdk.NewInt(1_000_000)),
			FeeRate:       simulation.RandomDecAmount(r, sdk.NewDecWithPrec(1, 2)),
		}
		liquidFarms = append(liquidFarms, liquidFarm)
	}
	return liquidFarms
}

// RandomizedGenState generates a random GenesisState.
func RandomizedGenState(simState *module.SimulationState) {
	var rewardsAuctionDuration time.Duration
	simState.AppParams.GetOrGenerate(
		simState.Cdc, RewardsAuctionDuration, &rewardsAuctionDuration, simState.Rand,
		func(r *rand.Rand) { rewardsAuctionDuration = GenRewardsAuctionDuration(r) },
	)

	var liquidFarms []types.LiquidFarm
	simState.AppParams.GetOrGenerate(
		simState.Cdc, LiquidFarms, &liquidFarms, simState.Rand,
		func(r *rand.Rand) { liquidFarms = GenLiquidFarms(r) },
	)

	genState := types.GenesisState{
		Params: types.Params{
			RewardsAuctionDuration: rewardsAuctionDuration,
			LiquidFarms:            liquidFarms,
		},
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&genState)
}
