package types

import (
	"fmt"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const TickPrecision = 4

var p = int(math.Pow10(TickPrecision))

// char returns the characteristic(integral part) of
// log10(x).
func char(x *big.Int) int {
	if x.Sign() == 0 {
		panic("cannot calculate char for 0")
	}
	return len(x.Text(10)) - 1
}

// pow10 returns pow(10, n)
func pow10(n int) *big.Int {
	ten := big.NewInt(10)
	return ten.Exp(ten, big.NewInt(int64(n)), nil)
}

func ValidateTickPrice(price sdk.Dec) (tick int32, valid bool) {
	if !price.IsPositive() {
		panic("price must be positive")
	}
	b := price.BigInt()
	c := char(b)
	q, r := b.QuoRem(b, pow10(c-TickPrecision), new(big.Int))
	i := int32(q.Int64())
	tick = (i - int32(p)) + 9*int32(p)*(int32(c)-sdk.Precision)
	if r.Sign() == 0 {
		valid = true
	}
	return
}

func PriceAtTick(tick int32) sdk.Dec {
	q, r := utils.DivMod(int(tick), 9*p)
	if q < TickPrecision-sdk.Precision {
		panic(fmt.Sprintf("price underflow: %d", tick))
	}
	return sdk.NewDecFromIntWithPrec(
		sdk.NewIntWithDecimal(int64(p+r), sdk.Precision+q-TickPrecision), sdk.Precision)
}

func TickAtPrice(price sdk.Dec) int32 {
	tick, _ := ValidateTickPrice(price)
	return tick
}

func PriceIntervalAtTick(tick int32) sdk.Dec {
	q, _ := utils.DivMod(int(tick), 9*p)
	return sdk.NewDecFromIntWithPrec(sdk.NewIntWithDecimal(1, sdk.Precision+q-TickPrecision), sdk.Precision)
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
