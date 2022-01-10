package types

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: remove
func TickToIndex(tick sdk.Dec, prec int) int {
	b := tick.BigInt()
	l := len(b.Text(10)) - 1
	d := int64(l - prec)
	if d > 0 {
		q := big.NewInt(10)
		q.Exp(q, big.NewInt(d), nil)
		b.Quo(b, q)
	}
	p := int(math.Pow10(prec))
	b.Sub(b, big.NewInt(int64(p)))
	return (l-prec)*9*p + int(b.Int64())
}

// TODO: remove
func TickFromIndex(i, prec int) sdk.Dec {
	p := int(math.Pow10(prec))
	l := i/(9*p) + prec
	t := big.NewInt(int64(p + i%(p*9)))
	if l > prec {
		m := big.NewInt(10)
		m.Exp(m, big.NewInt(int64(l-prec)), nil)
		t.Mul(t, m)
	}
	return sdk.NewDecFromBigIntWithPrec(t, sdk.Precision)
}

// char returns the characteristic(integral part) of
// log10(x * pow(10, sdk.Precision)).
func char(x sdk.Dec) int {
	if x.IsZero() {
		panic("cannot calculate log10 for 0")
	}
	return len(x.BigInt().Text(10)) - 1
}

// pow10 returns pow(10, n - sdk.Precision).
func pow10(n int) sdk.Dec {
	x := big.NewInt(10)
	x.Exp(x, big.NewInt(int64(n)), nil)
	return sdk.NewDecFromBigIntWithPrec(x, sdk.Precision)
}

// isPow10 returns whether x is a power of 10 or not.
func isPow10(x sdk.Dec) bool {
	b := x.BigInt()
	if b.Sign() <= 0 {
		return false
	}
	ten := big.NewInt(10)
	if b.Cmp(ten) == 0 {
		return true
	}
	zero := big.NewInt(0)
	m := new(big.Int)
	for b.Cmp(ten) >= 0 {
		b.DivMod(b, ten, m)
		if m.Cmp(zero) != 0 {
			return false
		}
	}
	return b.Cmp(big.NewInt(1)) == 0
}

// PriceToTick returns the highest price tick under(or equal to) the price.
func PriceToTick(price sdk.Dec, prec int) sdk.Dec {
	b := price.BigInt()
	l := char(price)
	d := int64(l - prec)
	if d > 0 {
		p := big.NewInt(10)
		p.Exp(p, big.NewInt(d), nil)
		b.Quo(b, p).Mul(b, p)
	}
	return sdk.NewDecFromBigIntWithPrec(b, sdk.Precision)
}

// UpTick returns the next lowest price tick above the price.
// UpTick guarantees that the price is already fit in ticks.
func UpTick(price sdk.Dec, prec int) sdk.Dec {
	l := char(price)
	return price.Add(pow10(l - prec))
}

// DownTick returns the next highest price tick under the price.
// DownTick guarantees that the price is already fit in ticks.
// DownTick doesn't check if the price is the lowest price tick.
func DownTick(price sdk.Dec, prec int) sdk.Dec {
	l := char(price)
	var d sdk.Dec
	if isPow10(price) {
		d = pow10(l - prec - 1)
	} else {
		d = pow10(l - prec)
	}
	return price.Sub(d)
}

// LowestTick returns the lowest possible price tick.
func LowestTick(prec int) sdk.Dec {
	return sdk.NewDecWithPrec(1, int64(sdk.Precision-prec))
}
