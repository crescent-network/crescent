package cremath_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
)

type decimalTestSuite struct {
	suite.Suite
}

func TestDecimalTestSuite(t *testing.T) {
	suite.Run(t, new(decimalTestSuite))
}

// assertMutResult given expected value after applying a math operation, a start value,
// mutative and non-mutative results with start values, asserts that mutation are only applied
// to the mutative versions. Also, asserts that both results match the expected value.
func (s *decimalTestSuite) assertMutResult(expectedResult, startValue, mutativeResult, nonMutativeResult, mutativeStartValue, nonMutativeStartValue cremath.BigDec) {
	// assert both results are as expected.
	s.Require().Equal(expectedResult, mutativeResult)
	s.Require().Equal(expectedResult, nonMutativeResult)

	// assert that mutative method mutated the receiver
	s.Require().Equal(mutativeStartValue, expectedResult)
	// assert that non-mutative method did not mutate the receiver
	s.Require().Equal(nonMutativeStartValue, startValue)
}

func (s *decimalTestSuite) TestAddMut() {
	toAdd := cremath.MustNewBigDecFromStr("10")
	tests := map[string]struct {
		startValue        cremath.BigDec
		expectedMutResult cremath.BigDec
	}{
		"0":  {cremath.NewBigDec(0), cremath.NewBigDec(10)},
		"1":  {cremath.NewBigDec(1), cremath.NewBigDec(11)},
		"10": {cremath.NewBigDec(10), cremath.NewBigDec(20)},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			startMut := tc.startValue.Clone()
			startNonMut := tc.startValue.Clone()

			resultMut := startMut.AddMut(toAdd)
			resultNonMut := startNonMut.Add(toAdd)

			s.assertMutResult(tc.expectedMutResult, tc.startValue, resultMut, resultNonMut, startMut, startNonMut)
		})
	}
}

func (s *decimalTestSuite) TestQuoMut() {
	quoBy := cremath.MustNewBigDecFromStr("2")
	tests := map[string]struct {
		startValue        cremath.BigDec
		expectedMutResult cremath.BigDec
	}{
		"0":  {cremath.NewBigDec(0), cremath.NewBigDec(0)},
		"1":  {cremath.NewBigDec(1), cremath.MustNewBigDecFromStr("0.5")},
		"10": {cremath.NewBigDec(10), cremath.NewBigDec(5)},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			startMut := tc.startValue.Clone()
			startNonMut := tc.startValue.Clone()

			resultMut := startMut.QuoMut(quoBy)
			resultNonMut := startNonMut.Quo(quoBy)

			s.assertMutResult(tc.expectedMutResult, tc.startValue, resultMut, resultNonMut, startMut, startNonMut)
		})
	}
}

// create a decimal from a decimal string (ex. "1234.5678")
func (s *decimalTestSuite) MustNewDecFromStr(str string) (d cremath.BigDec) {
	d, err := cremath.NewBigDecFromStr(str)
	s.Require().NoError(err)

	return d
}

func (s *decimalTestSuite) TestNewDecFromStr() {
	largeBigInt, success := new(big.Int).SetString("3144605511029693144278234343371835", 10)
	s.Require().True(success)

	tests := []struct {
		decimalStr string
		expErr     bool
		exp        cremath.BigDec
	}{
		{"", true, cremath.BigDec{}},
		{"0.-75", true, cremath.BigDec{}},
		{"0", false, cremath.NewBigDec(0)},
		{"1", false, cremath.NewBigDec(1)},
		{"1.1", false, cremath.NewBigDecWithPrec(11, 1)},
		{"0.75", false, cremath.NewBigDecWithPrec(75, 2)},
		{"0.8", false, cremath.NewBigDecWithPrec(8, 1)},
		{"0.11111", false, cremath.NewBigDecWithPrec(11111, 5)},
		{"314460551102969.31442782343433718353144278234343371835", true, cremath.NewBigDec(3141203149163817869)},
		{
			"314460551102969314427823434337.18357180924882313501835718092488231350",
			true, cremath.NewBigDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{
			"314460551102969314427823434337.1835",
			false, cremath.NewBigDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{".", true, cremath.BigDec{}},
		{".0", true, cremath.NewBigDec(0)},
		{"1.", true, cremath.NewBigDec(1)},
		{"foobar", true, cremath.BigDec{}},
		{"0.foobar", true, cremath.BigDec{}},
		{"0.foobar.", true, cremath.BigDec{}},
		{"179769313486231590772930519078902473361797697894230657273430081157732675805500963132708477322407536021120113879871393357658789768814416622492847430639474124377767893424865485276302219601246094119453082952085005768838150682342462881473913110540827237163350510684586298239947245938479716304835356329624224137216", true, cremath.BigDec{}},
	}

	for tcIndex, tc := range tests {
		res, err := cremath.NewBigDecFromStr(tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			s.Require().True(res.Equal(tc.exp), "equality was incorrect, res %v, exp %v, tc %v", res, tc.exp, tcIndex)
		}

		// negative tc
		res, err = cremath.NewBigDecFromStr("-" + tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			exp := tc.exp.Mul(cremath.NewBigDec(-1))
			s.Require().True(res.Equal(exp), "equality was incorrect, res %v, exp %v, tc %v", res, exp, tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestDecString() {
	tests := []struct {
		d    cremath.BigDec
		want string
	}{
		{cremath.NewBigDec(0), "0.000000000000000000000000000000000000"},
		{cremath.NewBigDec(1), "1.000000000000000000000000000000000000"},
		{cremath.NewBigDec(10), "10.000000000000000000000000000000000000"},
		{cremath.NewBigDec(12340), "12340.000000000000000000000000000000000000"},
		{cremath.NewBigDecWithPrec(12340, 4), "1.234000000000000000000000000000000000"},
		{cremath.NewBigDecWithPrec(12340, 5), "0.123400000000000000000000000000000000"},
		{cremath.NewBigDecWithPrec(12340, 8), "0.000123400000000000000000000000000000"},
		{cremath.NewBigDecWithPrec(1009009009009009009, 17), "10.090090090090090090000000000000000000"},
		{cremath.MustNewBigDecFromStr("10.090090090090090090090090090090090090"), "10.090090090090090090090090090090090090"},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, tc.d.String(), "bad String(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestSdkDec() {
	tests := []struct {
		d        cremath.BigDec
		want     sdk.Dec
		expPanic bool
	}{
		{cremath.NewBigDec(0), sdk.MustNewDecFromStr("0.000000000000000000"), false},
		{cremath.NewBigDec(1), sdk.MustNewDecFromStr("1.000000000000000000"), false},
		{cremath.NewBigDec(10), sdk.MustNewDecFromStr("10.000000000000000000"), false},
		{cremath.NewBigDec(12340), sdk.MustNewDecFromStr("12340.000000000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 4), sdk.MustNewDecFromStr("1.234000000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 5), sdk.MustNewDecFromStr("0.123400000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 8), sdk.MustNewDecFromStr("0.000123400000000000"), false},
		{cremath.NewBigDecWithPrec(1009009009009009009, 17), sdk.MustNewDecFromStr("10.090090090090090090"), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { tc.d.Dec() })
		} else {
			value := tc.d.Dec()
			s.Require().Equal(tc.want, value, "bad SdkDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestSdkDecRoundUp() {
	tests := []struct {
		d        cremath.BigDec
		want     sdk.Dec
		expPanic bool
	}{
		{cremath.NewBigDec(0), sdk.MustNewDecFromStr("0.000000000000000000"), false},
		{cremath.NewBigDec(1), sdk.MustNewDecFromStr("1.000000000000000000"), false},
		{cremath.NewBigDec(10), sdk.MustNewDecFromStr("10.000000000000000000"), false},
		{cremath.NewBigDec(12340), sdk.MustNewDecFromStr("12340.000000000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 4), sdk.MustNewDecFromStr("1.234000000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 5), sdk.MustNewDecFromStr("0.123400000000000000"), false},
		{cremath.NewBigDecWithPrec(12340, 8), sdk.MustNewDecFromStr("0.000123400000000000"), false},
		{cremath.NewBigDecWithPrec(1009009009009009009, 17), sdk.MustNewDecFromStr("10.090090090090090090"), false},
		{cremath.NewBigDecWithPrec(1009009009009009009, 19), sdk.MustNewDecFromStr("0.100900900900900901"), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { tc.d.DecRoundUp() })
		} else {
			value := tc.d.DecRoundUp()
			s.Require().Equal(tc.want, value, "bad SdkDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestBigDecFromSdkDec() {
	tests := []struct {
		d        sdk.Dec
		want     cremath.BigDec
		expPanic bool
	}{
		{sdk.MustNewDecFromStr("0.000000000000000000"), cremath.NewBigDec(0), false},
		{sdk.MustNewDecFromStr("1.000000000000000000"), cremath.NewBigDec(1), false},
		{sdk.MustNewDecFromStr("10.000000000000000000"), cremath.NewBigDec(10), false},
		{sdk.MustNewDecFromStr("12340.000000000000000000"), cremath.NewBigDec(12340), false},
		{sdk.MustNewDecFromStr("1.234000000000000000"), cremath.NewBigDecWithPrec(12340, 4), false},
		{sdk.MustNewDecFromStr("0.123400000000000000"), cremath.NewBigDecWithPrec(12340, 5), false},
		{sdk.MustNewDecFromStr("0.000123400000000000"), cremath.NewBigDecWithPrec(12340, 8), false},
		{sdk.MustNewDecFromStr("10.090090090090090090"), cremath.NewBigDecWithPrec(1009009009009009009, 17), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { cremath.NewBigDecFromDec(tc.d) })
		} else {
			value := cremath.NewBigDecFromDec(tc.d)
			s.Require().Equal(tc.want, value, "bad cremath.NewBigDecFromDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestBigDecFromSdkInt() {
	tests := []struct {
		i        sdk.Int
		want     cremath.BigDec
		expPanic bool
	}{
		{sdk.ZeroInt(), cremath.NewBigDec(0), false},
		{sdk.OneInt(), cremath.NewBigDec(1), false},
		{sdk.NewInt(10), cremath.NewBigDec(10), false},
		{sdk.NewInt(10090090090090090), cremath.NewBigDecWithPrec(10090090090090090, 0), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { cremath.NewBigDecFromInt(tc.i) })
		} else {
			value := cremath.NewBigDecFromInt(tc.i)
			s.Require().Equal(tc.want, value, "bad cremath.NewBigDecFromDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestEqualities() {
	tests := []struct {
		d1, d2     cremath.BigDec
		gt, lt, eq bool
	}{
		{cremath.NewBigDec(0), cremath.NewBigDec(0), false, false, true},
		{cremath.NewBigDecWithPrec(0, 2), cremath.NewBigDecWithPrec(0, 4), false, false, true},
		{cremath.NewBigDecWithPrec(100, 0), cremath.NewBigDecWithPrec(100, 0), false, false, true},
		{cremath.NewBigDecWithPrec(-100, 0), cremath.NewBigDecWithPrec(-100, 0), false, false, true},
		{cremath.NewBigDecWithPrec(-1, 1), cremath.NewBigDecWithPrec(-1, 1), false, false, true},
		{cremath.NewBigDecWithPrec(3333, 3), cremath.NewBigDecWithPrec(3333, 3), false, false, true},

		{cremath.NewBigDecWithPrec(0, 0), cremath.NewBigDecWithPrec(3333, 3), false, true, false},
		{cremath.NewBigDecWithPrec(0, 0), cremath.NewBigDecWithPrec(100, 0), false, true, false},
		{cremath.NewBigDecWithPrec(-1, 0), cremath.NewBigDecWithPrec(3333, 3), false, true, false},
		{cremath.NewBigDecWithPrec(-1, 0), cremath.NewBigDecWithPrec(100, 0), false, true, false},
		{cremath.NewBigDecWithPrec(1111, 3), cremath.NewBigDecWithPrec(100, 0), false, true, false},
		{cremath.NewBigDecWithPrec(1111, 3), cremath.NewBigDecWithPrec(3333, 3), false, true, false},
		{cremath.NewBigDecWithPrec(-3333, 3), cremath.NewBigDecWithPrec(-1111, 3), false, true, false},

		{cremath.NewBigDecWithPrec(3333, 3), cremath.NewBigDecWithPrec(0, 0), true, false, false},
		{cremath.NewBigDecWithPrec(100, 0), cremath.NewBigDecWithPrec(0, 0), true, false, false},
		{cremath.NewBigDecWithPrec(3333, 3), cremath.NewBigDecWithPrec(-1, 0), true, false, false},
		{cremath.NewBigDecWithPrec(100, 0), cremath.NewBigDecWithPrec(-1, 0), true, false, false},
		{cremath.NewBigDecWithPrec(100, 0), cremath.NewBigDecWithPrec(1111, 3), true, false, false},
		{cremath.NewBigDecWithPrec(3333, 3), cremath.NewBigDecWithPrec(1111, 3), true, false, false},
		{cremath.NewBigDecWithPrec(-1111, 3), cremath.NewBigDecWithPrec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestArithmetic() {
	tests := []struct {
		d1, d2                                cremath.BigDec
		expMul, expMulTruncate, expMulRoundUp cremath.BigDec
		expQuo, expQuoRoundUp, expQuoTruncate cremath.BigDec
		expAdd, expSub                        cremath.BigDec
	}{
		{cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0)},
		{cremath.NewBigDec(1), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(0), cremath.NewBigDec(1), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(1), cremath.NewBigDec(-1)},
		{cremath.NewBigDec(0), cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(-1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(-1), cremath.NewBigDec(-1)},

		{cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(2), cremath.NewBigDec(0)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(-2), cremath.NewBigDec(0)},
		{cremath.NewBigDec(1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(2)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(-2)},

		{
			cremath.NewBigDec(3), cremath.NewBigDec(7), cremath.NewBigDec(21), cremath.NewBigDec(21), cremath.NewBigDec(21),
			cremath.MustNewBigDecFromStr("0.428571428571428571428571428571428571"), cremath.MustNewBigDecFromStr("0.428571428571428571428571428571428572"), cremath.MustNewBigDecFromStr("0.428571428571428571428571428571428571"),
			cremath.NewBigDec(10), cremath.NewBigDec(-4),
		},
		{
			cremath.NewBigDec(2), cremath.NewBigDec(4), cremath.NewBigDec(8), cremath.NewBigDec(8), cremath.NewBigDec(8), cremath.NewBigDecWithPrec(5, 1), cremath.NewBigDecWithPrec(5, 1), cremath.NewBigDecWithPrec(5, 1),
			cremath.NewBigDec(6), cremath.NewBigDec(-2),
		},

		{cremath.NewBigDec(100), cremath.NewBigDec(100), cremath.NewBigDec(10000), cremath.NewBigDec(10000), cremath.NewBigDec(10000), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(200), cremath.NewBigDec(0)},

		{
			cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDecWithPrec(225, 2), cremath.NewBigDecWithPrec(225, 2), cremath.NewBigDecWithPrec(225, 2),
			cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(3), cremath.NewBigDec(0),
		},
		{
			cremath.NewBigDecWithPrec(3333, 4), cremath.NewBigDecWithPrec(333, 4), cremath.NewBigDecWithPrec(1109889, 8), cremath.NewBigDecWithPrec(1109889, 8), cremath.NewBigDecWithPrec(1109889, 8),
			cremath.MustNewBigDecFromStr("10.009009009009009009009009009009009009"), cremath.MustNewBigDecFromStr("10.009009009009009009009009009009009010"), cremath.MustNewBigDecFromStr("10.009009009009009009009009009009009009"),
			cremath.NewBigDecWithPrec(3666, 4), cremath.NewBigDecWithPrec(3, 1),
		},
	}

	for tcIndex, tc := range tests {
		tc := tc
		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		resMulTruncate := tc.d1.MulTruncate(tc.d2)
		resMulRoundUp := tc.d1.MulRoundUp(tc.d2)
		s.Require().True(tc.expAdd.Equal(resAdd), "exp %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		s.Require().True(tc.expSub.Equal(resSub), "exp %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		s.Require().True(tc.expMul.Equal(resMul), "exp %v, res %v, tc %d", tc.expMul, resMul, tcIndex)
		s.Require().True(tc.expMulTruncate.Equal(resMulTruncate), "exp %v, res %v, tc %d", tc.expMulTruncate, resMulTruncate, tcIndex)
		s.Require().True(tc.expMulRoundUp.Equal(resMulRoundUp), "exp %v, res %v, tc %d", tc.expMulRoundUp, resMulRoundUp, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			s.Require().Panics(func() { tc.d1.Quo(tc.d2) })
		} else {
			resQuo := tc.d1.Quo(tc.d2)
			s.Require().True(tc.expQuo.Equal(resQuo), "exp %v, res %v, tc %d", tc.expQuo.String(), resQuo.String(), tcIndex)

			resQuoRoundUp := tc.d1.QuoRoundUp(tc.d2)
			s.Require().True(tc.expQuoRoundUp.Equal(resQuoRoundUp), "exp %v, res %v, tc %d",
				tc.expQuoRoundUp.String(), resQuoRoundUp.String(), tcIndex)

			resQuoTruncate := tc.d1.QuoTruncate(tc.d2)
			s.Require().True(tc.expQuoTruncate.Equal(resQuoTruncate), "exp %v, res %v, tc %d",
				tc.expQuoTruncate.String(), resQuoTruncate.String(), tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestMulRoundUp_RoundingAtPrecisionEnd() {
	var (
		a                = cremath.MustNewBigDecFromStr("0.000000000000000000000000000000000009")
		b                = cremath.MustNewBigDecFromStr("0.000000000000000000000000000000000009")
		expectedRoundUp  = cremath.MustNewBigDecFromStr("0.000000000000000000000000000000000001")
		expectedTruncate = cremath.MustNewBigDecFromStr("0.000000000000000000000000000000000000")
	)

	actualRoundUp := a.MulRoundUp(b)
	s.Require().Equal(expectedRoundUp.String(), actualRoundUp.String(), "exp %v, res %v", expectedRoundUp, actualRoundUp)

	actualTruncate := a.MulTruncate(b)
	s.Require().Equal(expectedTruncate.String(), actualTruncate.String(), "exp %v, res %v", expectedTruncate, actualTruncate)
}

func (s *decimalTestSuite) TestStringOverflow() {
	// two random 64 bit primes
	dec1, err := cremath.NewBigDecFromStr("51643150036226787134389711697696177267")
	s.Require().NoError(err)
	dec2, err := cremath.NewBigDecFromStr("-31798496660535729618459429845579852627")
	s.Require().NoError(err)
	dec3 := dec1.Add(dec2)
	s.Require().Equal(
		"19844653375691057515930281852116324640.000000000000000000000000000000000000",
		dec3.String(),
	)
}

func (s *decimalTestSuite) TestDecMulInt() {
	tests := []struct {
		sdkDec cremath.BigDec
		sdkInt sdk.Int
		want   cremath.BigDec
	}{
		{cremath.NewBigDec(10), sdk.NewInt(2), cremath.NewBigDec(20)},
		{cremath.NewBigDec(1000000), sdk.NewInt(100), cremath.NewBigDec(100000000)},
		{cremath.NewBigDecWithPrec(1, 1), sdk.NewInt(10), cremath.NewBigDec(1)},
		{cremath.NewBigDecWithPrec(1, 5), sdk.NewInt(20), cremath.NewBigDecWithPrec(2, 4)},
	}
	for i, tc := range tests {
		got := tc.sdkDec.MulInt(tc.sdkInt)
		s.Require().Equal(tc.want, got, "Incorrect result on test case %d", i)
	}
}

func (s *decimalTestSuite) TestDecCeil() {
	testCases := []struct {
		input    cremath.BigDec
		expected cremath.BigDec
	}{
		{cremath.MustNewBigDecFromStr("0.001"), cremath.NewBigDec(1)},   // 0.001 => 1.0
		{cremath.MustNewBigDecFromStr("-0.001"), cremath.ZeroBigDec()},  // -0.001 => 0.0
		{cremath.ZeroBigDec(), cremath.ZeroBigDec()},                    // 0.0 => 0.0
		{cremath.MustNewBigDecFromStr("0.9"), cremath.NewBigDec(1)},     // 0.9 => 1.0
		{cremath.MustNewBigDecFromStr("4.001"), cremath.NewBigDec(5)},   // 4.001 => 5.0
		{cremath.MustNewBigDecFromStr("-4.001"), cremath.NewBigDec(-4)}, // -4.001 => -4.0
		{cremath.MustNewBigDecFromStr("4.7"), cremath.NewBigDec(5)},     // 4.7 => 5.0
		{cremath.MustNewBigDecFromStr("-4.7"), cremath.NewBigDec(-4)},   // -4.7 => -4.0
	}

	for i, tc := range testCases {
		res := tc.input.Ceil()
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestSqrt() {
	testCases := []struct {
		input    cremath.BigDec
		expected cremath.BigDec
	}{
		{cremath.OneBigDec(), cremath.OneBigDec()},                                                                        // 1.0 => 1.0
		{cremath.NewBigDecWithPrec(25, 2), cremath.NewBigDecWithPrec(5, 1)},                                               // 0.25 => 0.5
		{cremath.NewBigDecWithPrec(4, 2), cremath.NewBigDecWithPrec(2, 1)},                                                // 0.09 => 0.3
		{cremath.NewBigDecFromInt(sdk.NewInt(9)), cremath.NewBigDecFromInt(sdk.NewInt(3))},                                // 9 => 3
		{cremath.NewBigDecFromInt(sdk.NewInt(2)), cremath.MustNewBigDecFromStr("1.414213562373095048801688724209698079")}, // 2 => 1.414213562373095048801688724209698079
	}

	for i, tc := range testCases {
		res := tc.input.Sqrt()
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestSqrt_MutativeAndNonMutative() {
	start := cremath.NewBigDec(400)

	sqrt := start.Sqrt()
	s.Require().Equal(cremath.NewBigDec(20), sqrt)
	s.Require().Equal(cremath.NewBigDec(400), start)

	sqrt = start.SqrtMut()
	s.Require().Equal(cremath.NewBigDec(20), sqrt)
	s.Require().Equal(cremath.NewBigDec(20), start)
}

func (s *decimalTestSuite) TestDecEncoding() {
	testCases := []struct {
		input   cremath.BigDec
		rawBz   string
		jsonStr string
		yamlStr string
	}{
		{
			cremath.NewBigDec(0), "30",
			"\"0.000000000000000000000000000000000000\"",
			"\"0.000000000000000000000000000000000000\"\n",
		},
		{
			cremath.NewBigDecWithPrec(4, 2),
			"3430303030303030303030303030303030303030303030303030303030303030303030",
			"\"0.040000000000000000000000000000000000\"",
			"\"0.040000000000000000000000000000000000\"\n",
		},
		{
			cremath.NewBigDecWithPrec(-4, 2),
			"2D3430303030303030303030303030303030303030303030303030303030303030303030",
			"\"-0.040000000000000000000000000000000000\"",
			"\"-0.040000000000000000000000000000000000\"\n",
		},
		{
			cremath.MustNewBigDecFromStr("1.414213562373095048801688724209698079"),
			"31343134323133353632333733303935303438383031363838373234323039363938303739",
			"\"1.414213562373095048801688724209698079\"",
			"\"1.414213562373095048801688724209698079\"\n",
		},
		{
			cremath.MustNewBigDecFromStr("-1.414213562373095048801688724209698079"),
			"2D31343134323133353632333733303935303438383031363838373234323039363938303739",
			"\"-1.414213562373095048801688724209698079\"",
			"\"-1.414213562373095048801688724209698079\"\n",
		},
	}

	for _, tc := range testCases {
		bz, err := tc.input.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(tc.rawBz, fmt.Sprintf("%X", bz))

		var other cremath.BigDec
		s.Require().NoError((&other).Unmarshal(bz))
		s.Require().True(tc.input.Equal(other))

		bz, err = json.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.jsonStr, string(bz))
		s.Require().NoError(json.Unmarshal(bz, &other))
		s.Require().True(tc.input.Equal(other))

		bz, err = yaml.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.yamlStr, string(bz))
	}
}

// Showcase that different orders of operations causes different results.
func (s *decimalTestSuite) TestOperationOrders() {
	n1 := cremath.NewBigDec(10)
	n2 := cremath.NewBigDec(1000000010)
	s.Require().Equal(n1.Mul(n2).Quo(n2), cremath.NewBigDec(10))
	s.Require().NotEqual(n1.Mul(n2).Quo(n2), n1.Quo(n2).Mul(n2))
}

func BenchmarkMarshalTo(b *testing.B) {
	b.ReportAllocs()
	bis := []struct {
		in   cremath.BigDec
		want []byte
	}{
		{
			cremath.NewBigDec(1e8), []byte{
				0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			},
		},
		{cremath.NewBigDec(0), []byte{0x30}},
	}
	data := make([]byte, 100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, bi := range bis {
			if n, err := bi.in.MarshalTo(data); err != nil {
				b.Fatal(err)
			} else {
				if !bytes.Equal(data[:n], bi.want) {
					b.Fatalf("Mismatch\nGot:  % x\nWant: % x\n", data[:n], bi.want)
				}
			}
		}
	}
}

func (s *decimalTestSuite) TestClone() {
	tests := map[string]struct {
		startValue cremath.BigDec
	}{
		"1.1": {
			startValue: cremath.MustNewBigDecFromStr("1.1"),
		},
		"-3": {
			startValue: cremath.MustNewBigDecFromStr("-3"),
		},
		"0": {
			startValue: cremath.MustNewBigDecFromStr("-3"),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {

			copy := tc.startValue.Clone()

			s.Require().Equal(tc.startValue, copy)

			copy.MulMut(cremath.NewBigDec(2))
			// copy and startValue do not share internals.
			s.Require().NotEqual(tc.startValue, copy)
		})
	}
}

// TestMul_Mutation tests that MulMut mutates the receiver
// while Mut is not.
func (s *decimalTestSuite) TestMul_Mutation() {

	mulBy := cremath.MustNewBigDecFromStr("2")

	tests := map[string]struct {
		startValue        cremath.BigDec
		expectedMulResult cremath.BigDec
	}{
		"1.1": {
			startValue:        cremath.MustNewBigDecFromStr("1.1"),
			expectedMulResult: cremath.MustNewBigDecFromStr("2.2"),
		},
		"-3": {
			startValue:        cremath.MustNewBigDecFromStr("-3"),
			expectedMulResult: cremath.MustNewBigDecFromStr("-6"),
		},
		"0": {
			startValue:        cremath.ZeroBigDec(),
			expectedMulResult: cremath.ZeroBigDec(),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			startMut := tc.startValue.Clone()
			startNonMut := tc.startValue.Clone()

			resultMut := startMut.MulMut(mulBy)
			resultNonMut := startNonMut.Mul(mulBy)

			s.assertMutResult(tc.expectedMulResult, tc.startValue, resultMut, resultNonMut, startMut, startNonMut)
		})
	}
}

// TestPower_Mutation tests that PowerMut mutates the receiver
// while PowerInteger is not.
func (s *decimalTestSuite) TestPower_Mutation() {

	exponent := uint64(2)

	tests := map[string]struct {
		startValue     cremath.BigDec
		expectedResult cremath.BigDec
	}{
		"1": {
			startValue:     cremath.OneBigDec(),
			expectedResult: cremath.OneBigDec(),
		},
		"-3": {
			startValue:     cremath.MustNewBigDecFromStr("-3"),
			expectedResult: cremath.MustNewBigDecFromStr("9"),
		},
		"0": {
			startValue:     cremath.ZeroBigDec(),
			expectedResult: cremath.ZeroBigDec(),
		},
		"4": {
			startValue:     cremath.MustNewBigDecFromStr("4.5"),
			expectedResult: cremath.MustNewBigDecFromStr("20.25"),
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			startMut := tc.startValue.Clone()
			startNonMut := tc.startValue.Clone()

			resultMut := startMut.PowerMut(exponent)
			resultNonMut := startNonMut.Power(exponent)

			s.assertMutResult(tc.expectedResult, tc.startValue, resultMut, resultNonMut, startMut, startNonMut)
		})
	}
}

func (s *decimalTestSuite) TestQuoRoundUp_MutativeAndNonMutative() {
	fmt.Println(cremath.NewBigDec(1).QuoRoundUp(cremath.NewBigDec(-1)))
	fmt.Println(cremath.NewBigDec(1).QuoRoundUpMut(cremath.NewBigDec(-1)))

	tests := []struct {
		d1, d2, expQuoRoundUpMut cremath.BigDec
	}{
		{cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0)},
		{cremath.NewBigDec(1), cremath.NewBigDec(0), cremath.NewBigDec(0)},
		{cremath.NewBigDec(0), cremath.NewBigDec(1), cremath.NewBigDec(0)},
		{cremath.NewBigDec(0), cremath.NewBigDec(-1), cremath.NewBigDec(0)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(0)},

		{cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(1), cremath.NewBigDec(-1), cremath.NewBigDec(-1)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(1), cremath.NewBigDec(-1)},

		{
			cremath.NewBigDec(3), cremath.NewBigDec(7), cremath.MustNewBigDecFromStr("0.428571428571428571428571428571428572"),
		},
		{
			cremath.NewBigDec(2), cremath.NewBigDec(4), cremath.NewBigDecWithPrec(5, 1),
		},

		{cremath.NewBigDec(100), cremath.NewBigDec(100), cremath.NewBigDec(1)},

		{
			cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDec(1),
		},
		{
			cremath.NewBigDecWithPrec(3333, 4), cremath.NewBigDecWithPrec(333, 4), cremath.MustNewBigDecFromStr("10.009009009009009009009009009009009010"),
		},
	}

	for tcIndex, tc := range tests {
		tc := tc
		name := "testcase_" + fmt.Sprint(tcIndex)
		s.Run(name, func() {
			ConditionalPanic(s.T(), tc.d2.IsZero(), func() {
				copy := tc.d1.Clone()

				nonMutResult := copy.QuoRoundUp(tc.d2)

				// Return is as expected
				s.Require().Equal(tc.expQuoRoundUpMut, nonMutResult, "exp %v, res %v, tc %d", tc.expQuoRoundUpMut.String(), tc.d1.String(), tcIndex)

				// Receiver is not mutated
				s.Require().Equal(tc.d1, copy, "exp %v, res %v, tc %d", tc.expQuoRoundUpMut.String(), tc.d1.String(), tcIndex)

				// Receiver is mutated.
				tc.d1.QuoRoundUpMut(tc.d2)

				// Make sure d1 equals to expected
				s.Require().True(tc.expQuoRoundUpMut.Equal(tc.d1), "exp %v, res %v, tc %d", tc.expQuoRoundUpMut.String(), tc.d1.String(), tcIndex)
			})
		})
	}
}

func (s *decimalTestSuite) TestQuoTruncate_MutativeAndNonMutative() {
	tests := []struct {
		d1, d2, expQuoTruncateMut cremath.BigDec
	}{
		{cremath.NewBigDec(0), cremath.NewBigDec(0), cremath.NewBigDec(0)},
		{cremath.NewBigDec(1), cremath.NewBigDec(0), cremath.NewBigDec(0)},
		{cremath.NewBigDec(0), cremath.NewBigDec(1), cremath.NewBigDec(0)},
		{cremath.NewBigDec(0), cremath.NewBigDec(-1), cremath.NewBigDec(0)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(0), cremath.NewBigDec(0)},

		{cremath.NewBigDec(1), cremath.NewBigDec(1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(-1), cremath.NewBigDec(1)},
		{cremath.NewBigDec(1), cremath.NewBigDec(-1), cremath.NewBigDec(-1)},
		{cremath.NewBigDec(-1), cremath.NewBigDec(1), cremath.NewBigDec(-1)},

		{
			cremath.NewBigDec(3), cremath.NewBigDec(7), cremath.MustNewBigDecFromStr("0.428571428571428571428571428571428571"),
		},
		{
			cremath.NewBigDec(2), cremath.NewBigDec(4), cremath.NewBigDecWithPrec(5, 1),
		},

		{cremath.NewBigDec(100), cremath.NewBigDec(100), cremath.NewBigDec(1)},

		{
			cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDecWithPrec(15, 1), cremath.NewBigDec(1),
		},
		{
			cremath.NewBigDecWithPrec(3333, 4), cremath.NewBigDecWithPrec(333, 4), cremath.MustNewBigDecFromStr("10.009009009009009009009009009009009009"),
		},
	}

	for tcIndex, tc := range tests {
		tc := tc

		name := "testcase_" + fmt.Sprint(tcIndex)
		s.Run(name, func() {
			ConditionalPanic(s.T(), tc.d2.IsZero(), func() {
				copy := tc.d1.Clone()

				nonMutResult := copy.QuoTruncate(tc.d2)

				// Return is as expected
				s.Require().Equal(tc.expQuoTruncateMut, nonMutResult, "exp %v, res %v, tc %d", tc.expQuoTruncateMut.String(), tc.d1.String(), tcIndex)

				// Receiver is not mutated
				s.Require().Equal(tc.d1, copy, "exp %v, res %v, tc %d", tc.expQuoTruncateMut.String(), tc.d1.String(), tcIndex)

				// Receiver is mutated.
				tc.d1.QuoTruncateMut(tc.d2)

				// Make sure d1 equals to expected
				s.Require().True(tc.expQuoTruncateMut.Equal(tc.d1), "exp %v, res %v, tc %d", tc.expQuoTruncateMut.String(), tc.d1.String(), tcIndex)
			})
		})
	}
}

func ConditionalPanic(t *testing.T, expectPanic bool, f func()) {
	if expectPanic {
		require.Panics(t, f)
	} else {
		require.NotPanics(t, f)
	}
}
