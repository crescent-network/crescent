package types

import (
	"encoding/binary"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	TickPrecision = 4
	MinTick = -1260000
	MaxTick = 3600000
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

func TickToBytes(tick int32) []byte {
	bz := make([]byte, 5)
	if tick >= 0 {
		bz[0] = 1
	}
	binary.BigEndian.PutUint32(bz[1:], uint32(tick))
	return bz
}

func BytesToTick(bz []byte) int32 {
	return int32(binary.BigEndian.Uint32(bz[1:]))
}