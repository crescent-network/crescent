package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DepositAmount(isBuy bool, price sdk.Dec, qty sdk.Int) sdk.Int {
	if isBuy {
		return price.MulInt(qty).Ceil().TruncateInt()
	}
	return qty
}

func QuoteAmount(isBuy bool, price sdk.Dec, qty sdk.Int) sdk.Int {
	if isBuy {
		return price.MulInt(qty).Ceil().TruncateInt()
	}
	return price.MulInt(qty).TruncateInt()
}
