package amm

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TickPrecision represents a tick precision.
type TickPrecision int

func (prec TickPrecision) PriceToDownTick(price sdk.Dec) sdk.Dec {
	return PriceToDownTick(price, int(prec))
}

func (prec TickPrecision) PriceToUpTick(price sdk.Dec) sdk.Dec {
	return PriceToUpTick(price, int(prec))
}

func (prec TickPrecision) UpTick(price sdk.Dec) sdk.Dec {
	return UpTick(price, int(prec))
}

func (prec TickPrecision) DownTick(price sdk.Dec) sdk.Dec {
	return DownTick(price, int(prec))
}

func (prec TickPrecision) HighestTick() sdk.Dec {
	return HighestTick(int(prec))
}

func (prec TickPrecision) LowestTick() sdk.Dec {
	return LowestTick(int(prec))
}

func (prec TickPrecision) TickToIndex(tick sdk.Dec) int {
	return TickToIndex(tick, int(prec))
}

func (prec TickPrecision) TickFromIndex(i int) sdk.Dec {
	return TickFromIndex(i, int(prec))
}

func (prec TickPrecision) RoundPrice(price sdk.Dec) sdk.Dec {
	return RoundPrice(price, int(prec))
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

// PriceToDownTick returns the highest price tick under(or equal to) the price.
func PriceToDownTick(price sdk.Dec, prec int) sdk.Dec {
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

// PriceToUpTick returns the lowest price tick greater or equal than
// the price.
func PriceToUpTick(price sdk.Dec, prec int) sdk.Dec {
	tick := PriceToDownTick(price, prec)
	if !tick.Equal(price) {
		return UpTick(tick, prec)
	}
	return tick
}

// UpTick returns the next lowest price tick above the price.
func UpTick(price sdk.Dec, prec int) sdk.Dec {
	tick := PriceToDownTick(price, prec)
	if tick.Equal(price) {
		l := char(price)
		return price.Add(pow10(l - prec))
	}
	l := char(tick)
	return tick.Add(pow10(l - prec))
}

// DownTick returns the next highest price tick under the price.
// DownTick doesn't check if the price is the lowest price tick.
func DownTick(price sdk.Dec, prec int) sdk.Dec {
	tick := PriceToDownTick(price, prec)
	if tick.Equal(price) {
		l := char(price)
		var d sdk.Dec
		if isPow10(price) {
			d = pow10(l - prec - 1)
		} else {
			d = pow10(l - prec)
		}
		return price.Sub(d)
	}
	return tick
}

// HighestTick returns the highest possible price tick.
func HighestTick(prec int) sdk.Dec {
	i := new(big.Int).SetBits([]big.Word{0, 0, 0, 0, 0x1000000000000000})
	return PriceToDownTick(sdk.NewDecFromBigIntWithPrec(i, sdk.Precision), prec)
}

// LowestTick returns the lowest possible price tick.
func LowestTick(prec int) sdk.Dec {
	return sdk.NewDecWithPrec(1, int64(sdk.Precision-prec))
}

// TickToIndex returns a tick index for given price.
// Tick index 0 means the lowest possible price fit in ticks.
func TickToIndex(price sdk.Dec, prec int) int {
	b := price.BigInt()
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

// TickFromIndex returns a price for given tick index.
// See TickToIndex for more details about tick indices.
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

// RoundTickIndex returns rounded tick index using banker's rounding.
func RoundTickIndex(i int) int {
	return (i + 1) / 2 * 2
}

// RoundPrice returns rounded price using banker's rounding.
func RoundPrice(price sdk.Dec, prec int) sdk.Dec {
	tick := PriceToDownTick(price, prec)
	if price.Equal(tick) {
		return price
	}
	return TickFromIndex(RoundTickIndex(TickToIndex(tick, prec)), prec)
}

// Ticks returns ticks around basePrice with maximum number of numTicks*2+1.
// If basePrice fits in tick system, then the number of ticks will be numTicks*2+1.
// Else, the number will be numTicks*2.
func Ticks(basePrice sdk.Dec, numTicks, tickPrec int) []sdk.Dec {
	prec := TickPrecision(tickPrec)
	baseTick := prec.PriceToDownTick(basePrice)
	i := prec.TickToIndex(baseTick)
	highestTick := prec.TickFromIndex(i + numTicks)
	var lowestTick sdk.Dec
	if baseTick.Equal(basePrice) {
		lowestTick = prec.TickFromIndex(i - numTicks)
	} else {
		lowestTick = prec.TickFromIndex(i - numTicks + 1)
	}
	var ticks []sdk.Dec
	for tick := lowestTick; tick.LTE(highestTick); tick = prec.UpTick(tick) {
		ticks = append(ticks, tick)
	}
	sortTicks(ticks)
	return ticks
}

// EvenTicks returns ticks around basePrice with maximum number of numTicks*2+1.
// The gap is the same across all ticks.
// If basePrice fits in tick system, then the number of ticks will be numTicks*2+1.
// Else, the number will be numTicks*2.
func EvenTicks(basePrice sdk.Dec, numTicks, tickPrec int) []sdk.Dec {
	prec := TickPrecision(tickPrec)
	baseTick := prec.PriceToDownTick(basePrice)
	highestTick := prec.TickFromIndex(prec.TickToIndex(baseTick) + numTicks)
	highestGap := highestTick.Sub(prec.DownTick(highestTick))
	currentGap := prec.UpTick(baseTick).Sub(baseTick)
	var (
		start         sdk.Dec
		numTotalTicks int
	)
	if !currentGap.Equal(highestGap) {
		start = pow10(char(highestTick)).Sub(highestGap)
		numTotalTicks = 2 * numTicks
	} else {
		start = baseTick
		if baseTick.Equal(basePrice) {
			numTotalTicks = 2*numTicks + 1
		} else {
			numTotalTicks = 2 * numTicks
		}
	}
	highestTick = start.Add(highestGap.MulInt64(int64(numTicks)))
	var ticks []sdk.Dec
	for i, tick := 0, highestTick; i < numTotalTicks; i, tick = i+1, tick.Sub(highestGap) {
		ticks = append(ticks, tick)
	}
	sortTicks(ticks)
	return ticks
}
