package types

import (
	"encoding/binary"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func PriceAtTick(tick int32, prec int) sdk.Dec {
	pow10 := int(math.Pow10(prec))
	q, r := utils.DivMod(int(tick), 9*pow10)
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
