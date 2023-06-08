package types

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	TickPrecision = 4
)

func ValidateTickPrice(price sdk.Dec) (tick int32, valid bool) {
	b := price.BigInt()
	c := int32(len(b.Text(10)) - 1) // characteristic of b
	ten := big.NewInt(10)
	q, r := b.QuoRem(b, ten.Exp(ten, big.NewInt(int64(c-TickPrecision)), nil), new(big.Int))
	i := int32(q.Int64())
	pow10 := int32(math.Pow10(TickPrecision))
	tick = (i - pow10) + 9*pow10*(c-sdk.Precision)
	if r.Sign() == 0 {
		valid = true
	}
	return
}

func PriceAtTick(tick int32) sdk.Dec {
	pow10 := int(math.Pow10(TickPrecision))
	q, r := utils.DivMod(int(tick), 9*pow10)
	//if q < prec-sdk.Precision {
	//	panic("price underflow")
	//}
	return sdk.NewDecFromIntWithPrec(
		sdk.NewIntWithDecimal(int64(pow10+r), sdk.Precision+q-TickPrecision), sdk.Precision)
}

func TickAtPrice(price sdk.Dec) int32 {
	tick, _ := ValidateTickPrice(price)
	return tick
}

// RoundTick returns rounded tick using banker's rounding.
func RoundTick(tick int32) int32 {
	return (tick + tick%2) / 2 * 2
}

// RoundPrice returns rounded tick price using banker's rounding.
func RoundPrice(price sdk.Dec) sdk.Dec {
	tick, valid := ValidateTickPrice(price)
	if valid {
		return price
	}
	return PriceAtTick(RoundTick(tick))
}
