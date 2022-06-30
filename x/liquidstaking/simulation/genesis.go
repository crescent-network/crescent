package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// DONTCOVER

// Simulation parameter constants
const (
	unstakeFeeRate         = "unstake_fee_rate"
	liquidBondDenom        = "liquid_bond_denom"
	minLiquidStakingAmount = "min_liquid_staking_amount"
	whitelistedValidator   = "whiteliqted_validator"
)

func genUnstakeFeeRate(r *rand.Rand) sdk.Dec {
	return simtypes.RandomDecAmount(r, sdk.NewDecWithPrec(1, 2))
}

func genLiquidBondDenom(r *rand.Rand) string {
	return types.DefaultLiquidBondDenom
}

func genMinLiquidStakingAmount(r *rand.Rand) sdk.Int {
	return sdk.NewInt(int64(simtypes.RandIntBetween(r, 0, 10000000)))
}

func genTargetWeight(r *rand.Rand) sdk.Int {
	return sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 20)))
}

// genWhitelistedValidator returns randomized whitelisted validators.
func genWhitelistedValidator(r *rand.Rand) []types.WhitelistedValidator {
	return []types.WhitelistedValidator{}
}

// RandomizedGenState generates a random GenesisState for liquidstaking.
func RandomizedGenState(simState *module.SimulationState) {
	genesis := types.DefaultGenesisState()

	simState.AppParams.GetOrGenerate(
		simState.Cdc, unstakeFeeRate, &genesis.Params.UnstakeFeeRate, simState.Rand,
		func(r *rand.Rand) { genesis.Params.UnstakeFeeRate = genUnstakeFeeRate(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, liquidBondDenom, &genesis.Params.LiquidBondDenom, simState.Rand,
		func(r *rand.Rand) { genesis.Params.LiquidBondDenom = genLiquidBondDenom(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, minLiquidStakingAmount, &genesis.Params.MinLiquidStakingAmount, simState.Rand,
		func(r *rand.Rand) { genesis.Params.MinLiquidStakingAmount = genMinLiquidStakingAmount(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, whitelistedValidator, &genesis.Params.WhitelistedValidators, simState.Rand,
		func(r *rand.Rand) { genesis.Params.WhitelistedValidators = genWhitelistedValidator(r) },
	)

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated liquidstaking parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
