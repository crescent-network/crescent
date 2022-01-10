package simulation

// DONTCOVER

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/x/farming/types"
)

// Simulation parameter constants.
const (
	PrivatePlanCreationFee = "private_plan_creation_fee"
	NextEpochDays          = "next_epoch_days"
	FarmingFeeCollector    = "farming_fee_collector"
	CurrentEpochDays       = "current_epoch_days"
)

// GenPrivatePlanCreationFee return randomized private plan creation fee.
func GenPrivatePlanCreationFee(r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 0, 100_000_000))))
}

// GenNextEpochDays return default next epoch days.
func GenNextEpochDays(r *rand.Rand) uint32 {
	return uint32(simulation.RandIntBetween(r, int(types.DefaultNextEpochDays), 10))
}

// GenCurrentEpochDays returns current epoch days.
func GenCurrentEpochDays(r *rand.Rand) uint32 {
	return uint32(simulation.RandIntBetween(r, int(types.DefaultCurrentEpochDays), 10))
}

// GenFarmingFeeCollector returns default farming fee collector.
func GenFarmingFeeCollector(r *rand.Rand) string {
	return types.DefaultFarmingFeeCollector.String()
}

// RandomizedGenState generates a random GenesisState for farming.
func RandomizedGenState(simState *module.SimulationState) {
	var privatePlanCreationFee sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, PrivatePlanCreationFee, &privatePlanCreationFee, simState.Rand,
		func(r *rand.Rand) { privatePlanCreationFee = GenPrivatePlanCreationFee(r) },
	)

	var nextEpochDays uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, NextEpochDays, &nextEpochDays, simState.Rand,
		func(r *rand.Rand) { nextEpochDays = GenNextEpochDays(r) },
	)

	var feeCollector string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, FarmingFeeCollector, &feeCollector, simState.Rand,
		func(r *rand.Rand) { feeCollector = GenFarmingFeeCollector(r) },
	)

	var currentEpochDays uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, CurrentEpochDays, &currentEpochDays, simState.Rand,
		func(r *rand.Rand) { currentEpochDays = GenCurrentEpochDays(r) },
	)

	farmingGenesis := types.GenesisState{
		Params: types.Params{
			PrivatePlanCreationFee: privatePlanCreationFee,
			NextEpochDays:          nextEpochDays,
			FarmingFeeCollector:    feeCollector,
		},
		CurrentEpochDays: currentEpochDays,
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&farmingGenesis)
}
