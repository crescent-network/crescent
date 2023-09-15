package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	oneBigInt = big.NewInt(1)
	tenTo18   = big.NewInt(1e18)
)

func DecSqrt(d sdk.Dec) sdk.Dec {
	if d.IsNegative() {
		panic("square root of negative number")
	}
	bi := d.BigInt()            // bi = d * 10^18
	bi.Mul(bi, tenTo18)         // bi = bi * 10^18
	r := new(big.Int).Sqrt(bi)  // r = sqrt(bi)
	c := new(big.Int).Mul(r, r) // c = r * r
	if c.Cmp(bi) == -1 {        // if c < bi
		r.Add(r, oneBigInt)
	}
	return sdk.NewDecFromBigIntWithPrec(r, sdk.Precision)
}
