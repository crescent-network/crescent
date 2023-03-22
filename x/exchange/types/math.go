package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ZeroInt = sdk.ZeroInt()
)

func MinInt(a, b sdk.Int) sdk.Int {
	if a.LT(b) {
		return a
	}
	return b
}

func OfferAmount(isBuy bool, price sdk.Dec, qty sdk.Int) sdk.Int {
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
