package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/crescent-network/crescent/v5/x/marketmaker/types"
)

// DONTCOVER

// Simulation parameter constants.
const (
	IncentiveBudgetAddress = "incentive_budget_address"
	DepositAmount          = "deposit_amount"
	Common                 = "common"
	IncentivePairs         = "incentive_pairs"
)

// GenIncentiveBudgetAddress return default incentive budget address.
func GenIncentiveBudgetAddress(r *rand.Rand) string {
	return types.DefaultIncentiveBudgetAddress.String()
}

// GenDepositAmount return randomized market maker application deposit.
func GenDepositAmount(r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 0, 100_000_000))))
}

// GenCommon return default common.
func GenCommon(r *rand.Rand) types.Common {
	return types.DefaultCommon
}

// GenIncentivePairs return randomized incentive pairs.
func GenIncentivePairs(r *rand.Rand) []types.IncentivePair {
	if simulation.RandIntBetween(r, 1, 3) == 1 {
		return []types.IncentivePair{
			{
				PairId:          uint64(1),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
			{
				PairId:          uint64(2),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
			{
				PairId:          uint64(3),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
		}
	} else {
		return []types.IncentivePair{
			{
				PairId:          uint64(2),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
			{
				PairId:          uint64(3),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
			{
				PairId:          uint64(4),
				UpdateTime:      time.Time{},
				IncentiveWeight: sdk.ZeroDec(),
				MaxSpread:       sdk.ZeroDec(),
				MinWidth:        sdk.ZeroDec(),
				MinDepth:        sdk.ZeroInt(),
			},
		}
	}
}

// RandomizedGenState generates a random GenesisState for marketmaker.
func RandomizedGenState(simState *module.SimulationState) {
	var depositAmount sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, DepositAmount, &depositAmount, simState.Rand,
		func(r *rand.Rand) { depositAmount = GenDepositAmount(r) },
	)

	var incentiveBudgetAddress string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, IncentiveBudgetAddress, &incentiveBudgetAddress, simState.Rand,
		func(r *rand.Rand) { incentiveBudgetAddress = GenIncentiveBudgetAddress(r) },
	)

	var common types.Common
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Common, &common, simState.Rand,
		func(r *rand.Rand) { common = GenCommon(r) },
	)

	var incentivePairs []types.IncentivePair
	simState.AppParams.GetOrGenerate(
		simState.Cdc, IncentivePairs, &incentivePairs, simState.Rand,
		func(r *rand.Rand) { incentivePairs = GenIncentivePairs(r) },
	)

	genesis := &types.GenesisState{
		Params: types.Params{
			IncentiveBudgetAddress: incentiveBudgetAddress,
			DepositAmount:          depositAmount,
			IncentivePairs:         incentivePairs,
		},
	}

	bz, _ := json.MarshalIndent(&genesis, "", " ")
	fmt.Printf("Selected randomly generated marketmaker parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
