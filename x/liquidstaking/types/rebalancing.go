package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Redelegation struct {
	Delegator    sdk.AccAddress
	SrcValidator LiquidValidator
	DstValidator LiquidValidator
	Amount       sdk.Int
	Last         bool
}

// DivideByWeight divide the input value by the ratio of the param weight of the liquid validator and return it with crumb
// which is may occur while dividing according to the weight of active liquid validators by decimal error.
func DivideByWeight(activeVals ActiveLiquidValidators, input sdk.Int, whitelistedValMap WhitelistedValMap) (outputs []sdk.Int, crumb sdk.Int) {
	totalWeight := activeVals.TotalWeight(whitelistedValMap)
	if !totalWeight.IsPositive() {
		return []sdk.Int{}, sdk.ZeroInt()
	}
	totalShares := sdk.ZeroInt()
	sharePerWeight := input.ToDec().QuoTruncate(totalWeight.ToDec())
	for _, val := range activeVals {
		weightedShare := sharePerWeight.MulInt(val.GetWeight(whitelistedValMap, true)).TruncateInt()
		totalShares = totalShares.Add(weightedShare)
		outputs = append(outputs, weightedShare)
	}
	return outputs, input.Sub(totalShares)
}

// DivideByCurrentWeight divide the input value by the ratio of the weight of the liquid validator's liquid token and return it with crumb
// which is may occur while dividing according to the weight of liquid validators by decimal error.
func DivideByCurrentWeight(avs ActiveLiquidValidators, input sdk.Dec, totalLiquidTokens sdk.Int, liquidTokenMap map[string]sdk.Int) (outputs []sdk.Dec, crumb sdk.Dec) {
	if !totalLiquidTokens.IsPositive() {
		return []sdk.Dec{}, sdk.ZeroDec()
	}
	totalOutput := sdk.ZeroDec()
	unitInput := input.QuoTruncate(totalLiquidTokens.ToDec())
	for _, val := range avs {
		output := unitInput.MulTruncate(liquidTokenMap[val.OperatorAddress].ToDec()).TruncateDec()
		totalOutput = totalOutput.Add(output)
		outputs = append(outputs, output)
	}
	return outputs, input.Sub(totalOutput)
}
