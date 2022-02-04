package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Redelegation struct {
	Delegator    sdk.AccAddress
	SrcValidator sdk.ValAddress
	DstValidator sdk.ValAddress
	Amount       sdk.Int
}

// DivideByWeight divide the input value by the ratio of the param weight of the liquid validator and return it with crumb
func DivideByWeight(vs LiquidValidators, input sdk.Int, whitelistedValMap WhitelistedValMap) (outputs []sdk.Int, crumb sdk.Int) {
	totalWeight := vs.TotalWeight(whitelistedValMap)
	totalShares := sdk.ZeroInt()
	sharePerWeight := input.ToDec().QuoTruncate(totalWeight.ToDec())
	for _, val := range vs {
		weightedShare := sharePerWeight.MulInt(val.GetWeight(whitelistedValMap)).TruncateInt()
		totalShares = totalShares.Add(weightedShare)
		outputs = append(outputs, weightedShare)
	}
	return outputs, input.Sub(totalShares)
}
