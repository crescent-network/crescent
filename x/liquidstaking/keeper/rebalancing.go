package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) UpdateLiquidValidators(ctx sdk.Context) {
	//liquidVals := k.GetAllLiquidValidators()
	// TODO: GET, SET, GETALL, Iterate LiquidValidators, indexing
	//for _, val := range liquidVals {
	//	if val.
	//}
	//k.stakingKeeper.GetLastTotalPower()
}

// activeVals containing ValidatorStatusActive which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
func (k Keeper) Rebalancing(ctx sdk.Context, moduleAcc sdk.AccAddress, activeVals types.LiquidValidators, threshold sdk.Dec) (rebalancedLiquidVals types.LiquidValidators) {
	totalLiquidTokens := sdk.ZeroInt()
	totalWeight := sdk.ZeroDec()
	for _, val := range activeVals {
		totalLiquidTokens = totalLiquidTokens.Add(val.LiquidTokens)
		if val.Status == 1 {
			totalWeight = totalWeight.Add(val.Weight)
		}
	}

	var targetWeight sdk.Dec
	targetMap := map[string]sdk.Int{}
	for _, val := range activeVals {
		if val.Status == 1 {
			targetWeight = val.Weight
		} else {
			targetWeight = sdk.ZeroDec()
		}
		targetMap[val.OperatorAddress] = totalLiquidTokens.ToDec().MulTruncate(targetWeight).QuoTruncate(totalWeight).TruncateInt()
	}
	fmt.Println(targetMap)

	for i := 0; i < len(activeVals); i++ {
		maxVal, minVal, amountNeeded := activeVals.MinMaxGap(targetMap)
		if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold.TruncateInt())) {
			break
		}
		for idx := range activeVals {
			if activeVals[idx].OperatorAddress == maxVal.OperatorAddress {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Add(amountNeeded)
			}
			if activeVals[idx].OperatorAddress == minVal.OperatorAddress {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Sub(amountNeeded)
			}
		}
	}
	fmt.Println(activeVals)

	return activeVals
}
