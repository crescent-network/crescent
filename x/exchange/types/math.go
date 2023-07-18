package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DepositAmount(isBuy bool, price sdk.Dec, qty sdk.Dec) sdk.Dec {
	if isBuy {
		return price.Mul(qty)
	}
	return qty
}

func QuoteAmount(isBuy bool, price, qty sdk.Dec) sdk.Dec {
	if isBuy {
		return price.Mul(qty)
	}
	return price.MulTruncate(qty)
}
