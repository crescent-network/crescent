package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//AddStakingTargetMap is make add staking target map for one-way rebalancing, it can be called recursively.
func AddStakingTargetMap(activeVals LiquidValidators, addStakingAmt sdk.Int) map[string]sdk.Int {
	targetMap := make(map[string]sdk.Int)
	if addStakingAmt.IsNil() || !addStakingAmt.IsPositive() || activeVals.Len() == 0 {
		return targetMap
	}
	totalLiquidTokens := activeVals.TotalLiquidTokens()
	totalWeight := activeVals.TotalWeight()
	ToBeTotalLiquidTokens := totalLiquidTokens.Add(addStakingAmt)
	existOverWeightedVal := false

	sharePerWeight := ToBeTotalLiquidTokens.Quo(totalWeight)
	crumb := ToBeTotalLiquidTokens.Sub(sharePerWeight.Mul(totalWeight))

	i := 0
	for _, val := range activeVals {
		weightedShare := val.Weight.Mul(sharePerWeight)
		if val.LiquidTokens.GT(weightedShare) {
			existOverWeightedVal = true
		} else {
			activeVals[i] = val
			i++
			targetMap[val.OperatorAddress] = weightedShare.Sub(val.LiquidTokens)
		}
	}
	// remove overWeightedVals for recursive call
	activeVals = activeVals[:i]

	if !existOverWeightedVal {
		if v, ok := targetMap[activeVals[0].OperatorAddress]; ok {
			targetMap[activeVals[0].OperatorAddress] = v.Add(crumb)
		} else {
			targetMap[activeVals[0].OperatorAddress] = crumb
		}
		return targetMap
	} else {
		fmt.Println("[AddStakingTargetMap] recursive call for", activeVals, addStakingAmt, totalLiquidTokens, ToBeTotalLiquidTokens, totalWeight, sharePerWeight, crumb)
		return AddStakingTargetMap(activeVals, addStakingAmt)
	}
}
