package types

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func divMod(x, y int) (q, r int) {
	r = (x%y + y) % y
	q = (x - r) / y
	return
}

func PriceAtTick(tick int32, prec int) sdk.Dec {
	pow10 := int(math.Pow10(prec))
	q, r := divMod(int(tick), 9*pow10)
	//if q < prec-sdk.Precision {
	//	panic("price underflow")
	//}
	return sdk.NewDecFromIntWithPrec(
		sdk.NewIntWithDecimal(int64(pow10+r), sdk.Precision+q-prec), sdk.Precision)
}

func TickAtPrice(price sdk.Dec, prec int) int32 {
	b := price.BigInt()
	c := int32(len(b.Text(10)) - 1) // characteristic of b
	ten := big.NewInt(10)
	i := int32(b.Quo(b, ten.Exp(ten, big.NewInt(int64(c-int32(prec))), nil)).Int64())
	pow10 := int32(math.Pow10(prec))
	return (i - pow10) + 9*pow10*(c-sdk.Precision)
}
