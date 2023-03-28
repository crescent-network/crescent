package types

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func divMod(x, y int) (q, r int) {
	r = (x%y + y) % y
	q = (x - r) / y
	return
}

func PriceAtTick(tick int32, prec int) (price sdk.Dec) {
	pow10 := int(math.Pow10(prec))
	q, r := divMod(int(tick), 9*pow10)
	//if q < prec-sdk.Precision {
	//	panic("price underflow")
	//}
	return sdk.NewDecFromIntWithPrec(
		sdk.NewIntWithDecimal(int64(pow10+r), sdk.Precision+q-prec), sdk.Precision)
}
