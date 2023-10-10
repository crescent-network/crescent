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
	DefaultFees        = "default_fees"
	MaxOrderPriceRatio = "max_order_price_ratio"
)

func GenFees(r *rand.Rand) types.Fees {
	takerFeeRate := utils.RandomDec(r, utils.ParseDec("0"), utils.ParseDec("0.01"))
	makerFeeRate := utils.RandomDec(r, utils.ParseDec("0"), utils.ParseDec("0.005"))
	orderSourceFeeRatio := sdk.NewDecWithPrec(r.Int63n(101), 2) // 0%, 1%, 2%, ..., 99%, 100%
	return types.NewFees(makerFeeRate, takerFeeRate, orderSourceFeeRatio)
}

func GenAmountLimits(r *rand.Rand) types.AmountLimits {
	min := utils.RandomInt(r, sdk.NewInt(1), sdk.NewInt(10000))
	max := utils.RandomInt(r, sdk.NewInt(10000), sdk.NewIntWithDecimal(1, 30))
	return types.NewAmountLimits(min, max)
}

func GenMaxOrderPriceRatio(r *rand.Rand) sdk.Dec {
	return utils.RandomDec(r, utils.ParseDec("0.1"), utils.ParseDec("0.5"))
}

// RandomizedGenState generates a random GenesisState for the module.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, DefaultFees, &genesis.Params.DefaultFees, simState.Rand,
		func(r *rand.Rand) { genesis.Params.DefaultFees = GenFees(r) },
	)
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxOrderPriceRatio, &genesis.Params.MaxOrderPriceRatio, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MaxOrderPriceRatio = GenMaxOrderPriceRatio(r) },
	)

	bz, _ := json.MarshalIndent(genesis, "", " ")
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
