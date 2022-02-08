package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// Simulation parameter constants.
const (
	batchSize          = "batch_size"
	tickPrecision      = "tick_precision"
	maxPriceLimitRatio = "max_price_limit_ratio"
	withdrawFeeRate    = "withdraw_fee_rate"
	maxOrderLifespan   = "max_order_lifespan"
)

func genBatchSize(r *rand.Rand) uint32 {
	return uint32(r.Int31n(5) + 1)
}

func genTickPrecision(r *rand.Rand) uint32 {
	return uint32(r.Int31n(4))
}

func genMaxPriceLimitRatio(r *rand.Rand) sdk.Dec {
	return simtypes.RandomDecAmount(r, sdk.NewDecWithPrec(2, 1))
}

func genWithdrawFeeRate(r *rand.Rand) sdk.Dec {
	return simtypes.RandomDecAmount(r, sdk.NewDecWithPrec(1, 2))
}

func genMaxOrderLifespan(r *rand.Rand) time.Duration {
	return time.Duration(r.Int63n(int64(72 * time.Hour)))
}

// RandomizedGenState generates a random GenesisState for liquidity.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, batchSize, &genesis.Params.BatchSize, simState.Rand,
		func(r *rand.Rand) { genesis.Params.BatchSize = genBatchSize(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, tickPrecision, &genesis.Params.TickPrecision, simState.Rand,
		func(r *rand.Rand) { genesis.Params.TickPrecision = genTickPrecision(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, maxPriceLimitRatio, &genesis.Params.MaxPriceLimitRatio, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxPriceLimitRatio = genMaxPriceLimitRatio(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, withdrawFeeRate, &genesis.Params.WithdrawFeeRate, simState.Rand,
		func(r *rand.Rand) { genesis.Params.WithdrawFeeRate = genWithdrawFeeRate(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, maxOrderLifespan, &genesis.Params.MaxOrderLifespan, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxOrderLifespan = genMaxOrderLifespan(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated liquidity parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
