package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

// activeVals containing ValidatorStatusActive which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
// TODO: check minimum top vali power for prevent unbonded
func (k Keeper) Rebalancing(ctx sdk.Context, moduleAcc sdk.AccAddress, liquidVals types.LiquidValidators, threshold sdk.Dec) (rebalancedLiquidVals types.LiquidValidators) {
	totalLiquidTokens := liquidVals.TotalLiquidTokens()
	totalWeight := liquidVals.TotalWeight()
	liquidValsMap := liquidVals.Map()
	//for _, val := range liquidVals {
	//	totalLiquidTokens = totalLiquidTokens.Add(val.LiquidTokens)
	//	if val.Status == types.ValidatorStatusActive {
	//		totalWeight = totalWeight.Add(val.Weight)
	//	}
	//}

	var targetWeight sdk.Int
	targetMap := map[string]sdk.Int{}
	for _, val := range liquidVals {
		if val.Status == types.ValidatorStatusActive {
			targetWeight = val.Weight
		} else if val.Status == types.ValidatorStatusDelisting ||
			val.Status == types.ValidatorStatusDelisted {
			targetWeight = sdk.ZeroInt()
		} else {
			//val.Status == types.ValidatorStatusUnspecified
			targetWeight = sdk.ZeroInt()
		}
		targetMap[val.OperatorAddress] = totalLiquidTokens.Mul(targetWeight).Quo(totalWeight)
	}
	fmt.Println(targetMap)

	for i := 0; i < len(liquidVals); i++ {
		maxVal, minVal, amountNeeded := liquidVals.MinMaxGap(targetMap)
		if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold.TruncateInt())) {
			break
		}
		addedVal := 0
		subtractedVal := 0
		liquidValsMap[maxVal.OperatorAddress].LiquidTokens = liquidValsMap[maxVal.OperatorAddress].LiquidTokens.Add(amountNeeded)
		liquidValsMap[minVal.OperatorAddress].LiquidTokens = liquidValsMap[minVal.OperatorAddress].LiquidTokens.Sub(amountNeeded)
		for idx := range liquidVals {
			if liquidVals[idx].OperatorAddress == maxVal.OperatorAddress {
				addedVal = idx
				liquidVals[idx].LiquidTokens = liquidVals[idx].LiquidTokens.Add(amountNeeded)
			}
			if liquidVals[idx].OperatorAddress == minVal.OperatorAddress {
				subtractedVal = idx
				liquidVals[idx].LiquidTokens = liquidVals[idx].LiquidTokens.Sub(amountNeeded)
			}
		}
		fmt.Println("[rebalancing]", liquidValsMap[maxVal.OperatorAddress].GetOperator().String(), "-->", liquidValsMap[minVal.OperatorAddress].GetOperator().String(), amountNeeded)
		fmt.Println("[rebalancing2]", liquidVals[addedVal].OperatorAddress, "-->", liquidVals[subtractedVal].OperatorAddress, amountNeeded)
	}
	fmt.Println(liquidVals)

	return liquidVals
}
