package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Simulation parameter constants
const (
	Fees               = "fees"
	MaxOrderPriceRatio = "max_order_price_ratio"
)

func GenFees(r *rand.Rand) types.Fees {
	takerFeeRate := utils.RandomDec(r, utils.ParseDec("0"), utils.ParseDec("0.01"))
	var makerFeeRate sdk.Dec
	if r.Float64() <= 0 { // 50% chance
		// positive maker fee rate
		makerFeeRate = utils.RandomDec(r, utils.ParseDec("0"), utils.ParseDec("0.01"))
	} else {
		// negative maker fee rate
		makerFeeRate = utils.RandomDec(r, takerFeeRate.Neg(), utils.ParseDec("0"))
	}
	return types.Fees{
		MarketCreationFee:   types.DefaultFees.MarketCreationFee,
		DefaultMakerFeeRate: makerFeeRate,
		DefaultTakerFeeRate: takerFeeRate,
	}
}

func GenMaxOrderPriceRatio(r *rand.Rand) sdk.Dec {
	return utils.RandomDec(r, utils.ParseDec("0.1"), utils.ParseDec("0.5"))
}

// RandomizedGenState generates a random GenesisState for the module.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, Fees, &genesis.Params.Fees, simState.Rand,
		func(r *rand.Rand) { genesis.Params.Fees = GenFees(r) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxOrderPriceRatio, &genesis.Params.MaxOrderPriceRatio, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxOrderPriceRatio = GenMaxOrderPriceRatio(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
