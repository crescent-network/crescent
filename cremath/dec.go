package cremath

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BigDec struct {
	i *big.Int
}

const (
	// Precision is the number of decimal places
	Precision = 36

	// DecimalPrecisionBits is the number of bits required to represent
	// the above precision
	// Ceiling[Log2[10^Precision - 1]]
	DecimalPrecisionBits = 120

	// decimalTruncateBits is the minimum number of bits removed
	// by a truncate operation. It is equal to
	// Floor[Log2[10^Precision - 1]].
	decimalTruncateBits = DecimalPrecisionBits - 1

	maxDecBitLen = 256 + decimalTruncateBits
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil)
	sdkPrecisionReuse    = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdk.Precision), nil)
	fivePrecision        = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	precisionMultipliers []*big.Int
	zeroInt              = big.NewInt(0)
	oneInt               = big.NewInt(1)
	tenInt               = big.NewInt(10)
)

// Decimal errors
var (
	ErrEmptyDecimalStr      = errors.New("decimal string cannot be empty")
	ErrInvalidDecimalLength = errors.New("invalid decimal length")
	ErrInvalidDecimalStr    = errors.New("invalid decimal string")
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]*big.Int, Precision+1)
	for i := 0; i <= Precision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}
}

func precisionInt() *big.Int {
	return new(big.Int).Set(precisionReuse)
}

func ZeroBigDec() BigDec { return BigDec{new(big.Int).Set(zeroInt)} }
func OneBigDec() BigDec  { return BigDec{precisionInt()} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	zerosToAdd := Precision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil)
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	return precisionMultipliers[prec]
}

// NewBigDec creates a new BigDec from integer assuming whole number
func NewBigDec(i int64) BigDec {
	return NewBigDecWithPrec(i, 0)
}

// NewBigDecWithPrec creates a new BigDec from integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewBigDecWithPrec(i, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(big.NewInt(i), precisionMultiplier(prec)),
	}
}

func NewBigDecFromDec(d sdk.Dec) BigDec {
	return NewBigDecFromBigIntWithPrec(d.BigInt(), sdk.Precision)
}

func (d BigDec) Dec() sdk.Dec {
	// Truncate any additional decimal values that exist due to BigDec's additional precision
	// This relies on big.Int's QuoMut function doing floor division
	intRepresentation := new(big.Int).Quo(d.BigInt(), precisionMultiplier(sdk.Precision))

	// convert int representation back to SDK Dec precision
	truncatedDec := sdk.NewDecFromBigIntWithPrec(intRepresentation, sdk.Precision)

	return truncatedDec
}

func (d BigDec) DecRoundUp() sdk.Dec {
	return sdk.NewDecFromBigIntWithPrec(chopPrecisionAndRoundUpDec(d.i), sdk.Precision)
}

// NewBigDecFromBigInt creates a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewBigDecFromBigInt(i *big.Int) BigDec {
	return NewBigDecFromBigIntWithPrec(i, 0)
}

// NewBigDecFromBigIntWithPrec creates a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewBigDecFromBigIntWithPrec(i *big.Int, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(i, precisionMultiplier(prec)),
	}
}

// NewBigDecFromInt creates a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewBigDecFromInt(i sdk.Int) BigDec {
	return NewBigDecFromIntWithPrec(i, 0)
}

// NewBigDecFromIntWithPrec creates a new BigDec from big integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewBigDecFromIntWithPrec(i sdk.Int, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(i.BigInt(), precisionMultiplier(prec)),
	}
}

// NewBigDecFromStr creates a decimal from an input decimal string.
// valid must come in the form:
//
//	(-) whole integers (.) decimal integers
//
// examples of acceptable input include:
//
//	-123.456
//	456.7890
//	345
//	-456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewBigDecFromStr(str string) (BigDec, error) {
	if len(str) == 0 {
		return BigDec{}, ErrEmptyDecimalStr
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return BigDec{}, ErrEmptyDecimalStr
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return BigDec{}, ErrInvalidDecimalLength
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return BigDec{}, ErrInvalidDecimalStr
	}

	if lenDecs > Precision {
		return BigDec{}, fmt.Errorf("value '%s' exceeds max precision by %d decimal places: max precision %d", str, Precision-lenDecs, Precision)
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr += zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return BigDec{}, fmt.Errorf("failed to set decimal string with base 10: %s", combinedStr)
	}
	if combined.BitLen() > maxDecBitLen {
		return BigDec{}, fmt.Errorf("decimal '%s' out of range; bitLen: got %d, max %d", str, combined.BitLen(), maxDecBitLen)
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}

	return BigDec{combined}, nil
}

func MustNewBigDecFromStr(s string) BigDec {
	dec, err := NewBigDecFromStr(s)
	if err != nil {
		panic(err)
	}
	return dec
}

// Clone performs a deep copy of the receiver
// and returns the new result.
func (d BigDec) Clone() BigDec {
	return BigDec{new(big.Int).Set(d.i)}
}

func (d BigDec) IsNil() bool          { return d.i == nil }                    // is decimal nil
func (d BigDec) IsZero() bool         { return (d.i).Sign() == 0 }             // is equal to zero
func (d BigDec) IsNegative() bool     { return (d.i).Sign() == -1 }            // is negative
func (d BigDec) IsPositive() bool     { return (d.i).Sign() == 1 }             // is positive
func (d BigDec) Equal(d2 BigDec) bool { return (d.i).Cmp(d2.i) == 0 }          // equal decimals
func (d BigDec) GT(d2 BigDec) bool    { return (d.i).Cmp(d2.i) > 0 }           // greater than
func (d BigDec) GTE(d2 BigDec) bool   { return (d.i).Cmp(d2.i) >= 0 }          // greater than or equal
func (d BigDec) LT(d2 BigDec) bool    { return (d.i).Cmp(d2.i) < 0 }           // less than
func (d BigDec) LTE(d2 BigDec) bool   { return (d.i).Cmp(d2.i) <= 0 }          // less than or equal
func (d BigDec) Neg() BigDec          { return BigDec{new(big.Int).Neg(d.i)} } // reverse the decimal sign
func (d BigDec) Abs() BigDec          { return BigDec{new(big.Int).Abs(d.i)} } // absolute value

// BigInt returns a copy of the underlying big.Int.
func (d BigDec) BigInt() *big.Int {
	if d.i == nil {
		return nil
	}
	return new(big.Int).Set(d.i)
}

func (d BigDec) Add(d2 BigDec) BigDec {
	return d.Clone().AddMut(d2)
}

// AddMut sets d to the sum d+d2 and returns d.
func (d BigDec) AddMut(d2 BigDec) BigDec {
	d.i.Add(d.i, d2.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) Sub(d2 BigDec) BigDec {
	return d.Clone().SubMut(d2)
}

// SubMut sets d to the difference d-d2 and returns d.
func (d BigDec) SubMut(d2 BigDec) BigDec {
	d.i.Sub(d.i, d2.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) Mul(d2 BigDec) BigDec {
	return d.Clone().MulMut(d2)
}

// MulMut sets d to the product d*d2 and returns d.
func (d BigDec) MulMut(d2 BigDec) BigDec {
	d.i.Mul(d.i, d2.i)
	d.i = chopPrecisionAndRound(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) MulTruncate(d2 BigDec) BigDec {
	return d.Clone().MulTruncateMut(d2)
}

func (d BigDec) MulTruncateMut(d2 BigDec) BigDec {
	d.i.Mul(d.i, d2.i)
	d.i = chopPrecisionAndTruncateMut(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) MulRoundUp(d2 BigDec) BigDec {
	return d.Clone().MulRoundUpMut(d2)
}

func (d BigDec) MulRoundUpMut(d2 BigDec) BigDec {
	d.i.Mul(d.i, d2.i)
	d.i = chopPrecisionAndRoundUpBigDec(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) MulInt(i sdk.Int) BigDec {
	return d.Clone().MulIntMut(i)
}

func (d BigDec) MulIntMut(i sdk.Int) BigDec {
	d.i.Mul(d.i, i.BigInt()) // XXX
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) MulInt64(i int64) BigDec {
	return d.Clone().MulInt64Mut(i)
}

// MulInt64Mut - multiplication with int64
func (d BigDec) MulInt64Mut(i int64) BigDec {
	d.i.Mul(d.i, big.NewInt(i))
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) Quo(d2 BigDec) BigDec {
	return d.Clone().QuoMut(d2)
}

// QuoMut sets d to the quotient d/d2 and returns z.
func (d BigDec) QuoMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, precisionReuse)
	d.i.Mul(d.i, precisionReuse)
	d.i.Quo(d.i, d2.i)
	d.i = chopPrecisionAndRound(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) QuoTruncate(d2 BigDec) BigDec {
	return d.Clone().QuoTruncateMut(d2)
}

func (d BigDec) QuoTruncateMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, precisionReuse)
	d.i.Mul(d.i, precisionReuse)
	d.i.Quo(d.i, d2.i)
	d.i = chopPrecisionAndTruncateMut(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) QuoRoundUp(d2 BigDec) BigDec {
	return d.Clone().QuoRoundUpMut(d2)
}

func (d BigDec) QuoRoundUpMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, precisionReuse)
	d.i.Mul(d.i, precisionReuse)
	d.i.Quo(d.i, d2.i)
	d.i = chopPrecisionAndRoundUpMut(d.i, precisionReuse)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

func (d BigDec) QuoInt(i sdk.Int) BigDec {
	return d.Clone().QuoIntMut(i)
}

func (d BigDec) QuoIntMut(i sdk.Int) BigDec {
	d.i.Quo(d.i, i.BigInt()) // XXX
	return d
}

func (d BigDec) QuoInt64(i int64) BigDec {
	return d.Clone().QuoInt64Mut(i)
}

// QuoInt64Mut - quotient with int64
func (d BigDec) QuoInt64Mut(i int64) BigDec {
	d.i.Quo(d.i, big.NewInt(i))
	return d
}

func (d BigDec) Power(power uint64) BigDec {
	return d.Clone().PowerMut(power)
}

// PowerMut returns the result of raising to a positive integer power
func (d BigDec) PowerMut(power uint64) BigDec {
	if power == 0 {
		return OneBigDec()
	}
	tmp := OneBigDec()
	for i := power; i > 1; {
		if i%2 != 0 {
			tmp.MulMut(d)
		}
		i /= 2
		d.MulMut(d)
	}
	return d.MulMut(tmp)
}

// is integer, e.g. decimals are zero
func (d BigDec) IsInteger() bool {
	return new(big.Int).Rem(d.i, precisionReuse).Sign() == 0
}

// format decimal state
func (d BigDec) Format(s fmt.State, verb rune) {
	_, err := s.Write([]byte(d.String()))
	if err != nil {
		panic(err)
	}
}

func (d BigDec) String() string {
	if d.i == nil {
		return d.i.String()
	}

	isNeg := d.IsNegative()

	if isNeg {
		d = d.Neg()
	}

	bzInt, err := d.i.MarshalText()
	if err != nil {
		return ""
	}
	inputSize := len(bzInt)

	var bzStr []byte

	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= Precision {
		bzStr = make([]byte, Precision+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < Precision-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(Precision-inputSize):], bzInt)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - Precision

		copy(bzStr, bzInt[:decPointPlace])                   // pre-decimal digits
		bzStr[decPointPlace] = byte('.')                     // decimal point
		copy(bzStr[decPointPlace+1:], bzInt[decPointPlace:]) // post-decimal digits
	}

	if isNeg {
		return "-" + string(bzStr)
	}

	return string(bzStr)
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

// chopPrecisionAndRoundUpBigDec removes a Precision amount of rightmost digits and rounds up.
// Non-mutative.
func chopPrecisionAndRoundUpBigDec(d *big.Int) *big.Int {
	return chopPrecisionAndRoundUpMut(d, precisionReuse)
}

// chopPrecisionAndRoundUpDec removes  DecPrecision amount of rightmost digits and rounds up.
// Non-mutative.
func chopPrecisionAndRoundUpDec(d *big.Int) *big.Int {
	copy := new(big.Int).Set(d)
	return chopPrecisionAndRoundUpMut(copy, sdkPrecisionReuse)
}

// chopPrecisionAndRoundUp removes a Precision amount of rightmost digits and rounds up.
// Mutates input d.
// Mutations occur:
// - By calling chopPrecisionAndTruncateMut.
// - Using input d directly in QuoRem.
func chopPrecisionAndRoundUpMut(d *big.Int, precisionReuse *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		// truncate since d is negative...
		d = chopPrecisionAndTruncateMut(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	_, rem := d.QuoRem(d, precisionReuse, big.NewInt(0))

	if rem.Sign() == 0 { // remainder is zero
		return d
	}

	return d.Add(d, oneInt)
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It does not mutate the input.
func chopPrecisionAndTruncate(d *big.Int) *big.Int {
	return new(big.Int).Quo(d, precisionReuse)
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It mutates the input.
func chopPrecisionAndTruncateMut(d *big.Int) *big.Int {
	return d.Quo(d, precisionReuse)
}

// TruncateInt truncates the decimals from the number and returns an Int
func (d BigDec) TruncateInt() sdk.Int {
	return sdk.NewIntFromBigInt(chopPrecisionAndTruncate(d.i))
}

// Truncate truncates the decimals from the number and returns a BigDec
func (d BigDec) Truncate() BigDec {
	return NewBigDecFromBigInt(chopPrecisionAndTruncate(d.i))
}

// Ceil returns the smallest interger value (as a decimal) that is greater than
// or equal to the given decimal.
func (d BigDec) Ceil() BigDec {
	tmp := new(big.Int).Set(d.i)

	quo, rem := tmp, big.NewInt(0)
	quo, rem = quo.QuoRem(tmp, precisionReuse, rem)

	// no need to round with a zero remainder regardless of sign
	if rem.Cmp(zeroInt) == 0 {
		return NewBigDecFromBigInt(quo)
	}

	if rem.Sign() == -1 {
		return NewBigDecFromBigInt(quo)
	}

	return NewBigDecFromBigInt(quo.Add(quo, oneInt))
}

// reuse nil values
var nilJSON []byte

func init() {
	empty := new(big.Int)
	bz, _ := empty.MarshalText()
	nilJSON, _ = json.Marshal(string(bz))
}

// MarshalJSON marshals the decimal
func (d BigDec) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return nilJSON, nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON defines custom decoding scheme
func (d *BigDec) UnmarshalJSON(bz []byte) error {
	if d.i == nil {
		d.i = new(big.Int)
	}
	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}
	newDec, err := NewBigDecFromStr(text)
	if err != nil {
		return err
	}
	d.i = newDec.i
	return nil
}

// MarshalYAML returns the YAML representation.
func (d BigDec) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// Marshal implements the gogo proto custom type interface.
func (d BigDec) Marshal() ([]byte, error) {
	if d.i == nil {
		d.i = new(big.Int)
	}
	return d.i.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (d BigDec) MarshalTo(data []byte) (n int, err error) {
	if d.i == nil {
		d.i = new(big.Int)
	}
	if d.i.Sign() == 0 {
		copy(data, []byte{0x30})
		return 1, nil
	}
	bz, err := d.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (d *BigDec) Unmarshal(data []byte) error {
	if len(data) == 0 {
		d.i = nil
		return nil
	}
	if d.i == nil {
		d.i = new(big.Int)
	}
	if err := d.i.UnmarshalText(data); err != nil {
		return err
	}
	if d.i.BitLen() > maxDecBitLen {
		return fmt.Errorf("decimal out of range; got: %d, max: %d", d.i.BitLen(), maxDecBitLen)
	}
	return nil
}

// Size implements the gogo proto custom type interface.
func (d BigDec) Size() int {
	bz, _ := d.Marshal()
	return len(bz)
}

// Override Amino binary serialization by proxying to protobuf.

func (d BigDec) MarshalAmino() ([]byte, error)   { return d.Marshal() }
func (d *BigDec) UnmarshalAmino(bz []byte) error { return d.Unmarshal(bz) }

// MinBigDec returns minimum decimal between two.
func MinBigDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// MaxBigDec returns maximum decimal between two.
func MaxBigDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

func (d BigDec) Sqrt() BigDec {
	return d.Clone().SqrtMut()
}

func (d BigDec) SqrtMut() BigDec {
	if d.IsNegative() {
		panic("square root of negative number")
	}
	d.i.Mul(d.i, precisionReuse)
	t := new(big.Int).Set(d.i)
	d.i.Sqrt(d.i)
	c := new(big.Int).Mul(d.i, d.i)
	if c.Cmp(t) == -1 {
		d.i.Add(d.i, oneInt)
	}
	return d
}
