package types

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func PriceToBytes(d sdk.Dec) []byte {
	if !d.IsPositive() {
		panic("non-positive price")
	}
	bz := make([]byte, 32)
	d.BigInt().FillBytes(bz)
	return bz
}

func BytesToPrice(bz []byte) sdk.Dec {
	if len(bz) != 32 {
		panic(fmt.Sprintf("wrong length of bytes: %d; must be 32", len(bz)))
	}
	b := new(big.Int).SetBytes(bz)
	return sdk.NewDecFromBigIntWithPrec(b, sdk.Precision)
}
