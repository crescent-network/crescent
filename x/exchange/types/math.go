package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DepositAmount(isBuy bool, price, qty sdk.Dec) sdk.Int {
	if isBuy {
		return price.Mul(qty).Ceil().TruncateInt()
	}
	return qty.Ceil().TruncateInt()
}

func QuoteAmount(isBuy bool, price, qty sdk.Dec) sdk.Dec {
	if isBuy {
		return price.Mul(qty)
	}
	return price.MulTruncate(qty)
}
