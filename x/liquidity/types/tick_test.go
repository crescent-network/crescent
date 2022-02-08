package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

var tickPrec = int(types.DefaultTickPrecision)

func TestPriceToTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Dec
	}{
		{parseDec("0.000000000000099999"), parseDec("0.00000000000009999")},
		{parseDec("1.999999999999999999"), parseDec("1.999")},
		{parseDec("99.999999999999999999"), parseDec("99.99")},
		{parseDec("100.999999999999999999"), parseDec("100.9")},
		{parseDec("9999.999999999999999999"), parseDec("9999")},
		{parseDec("10019"), parseDec("10010")},
		{parseDec("1000100005"), parseDec("1000000000")},
	} {
		require.True(sdk.DecEq(t, tc.expected, types.PriceToTick(tc.price, tickPrec)))
	}
}

func TestTick(t *testing.T) {
	for _, tc := range []struct {
		i        int
		prec     int
		expected sdk.Dec
	}{
		{0, tickPrec, sdk.NewDecWithPrec(1, int64(sdk.Precision-tickPrec))},
		{1, tickPrec, parseDec("0.000000000000001001")},
		{8999, tickPrec, parseDec("0.000000000000009999")},
		{9000, tickPrec, parseDec("0.000000000000010000")},
		{9001, tickPrec, parseDec("0.000000000000010010")},
		{17999, tickPrec, parseDec("0.000000000000099990")},
		{18000, tickPrec, parseDec("0.000000000000100000")},
		{135000, tickPrec, sdk.NewDec(1)},
		{135001, tickPrec, parseDec("1.001")},
	} {
		t.Run("", func(t *testing.T) {
			res := types.TickFromIndex(tc.i, tc.prec)
			require.True(sdk.DecEq(t, tc.expected, res))
			require.Equal(t, tc.i, types.TickToIndex(res, tc.prec))
		})
	}
}

func TestUpTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{parseDec("1000000000000000000"), tickPrec, parseDec("1001000000000000000")},
		{parseDec("1000"), tickPrec, parseDec("1001")},
		{parseDec("999.9"), tickPrec, parseDec("1000")},
		{parseDec("999.0"), tickPrec, parseDec("999.1")},
		{parseDec("1.100"), tickPrec, parseDec("1.101")},
		{parseDec("1.000"), tickPrec, parseDec("1.001")},
		{parseDec("0.9999"), tickPrec, parseDec("1.000")},
		{parseDec("0.1000"), tickPrec, parseDec("0.1001")},
		{parseDec("0.09999"), tickPrec, parseDec("0.1000")},
		{parseDec("0.09997"), tickPrec, parseDec("0.09998")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, types.UpTick(tc.price, tc.prec)))
		})
	}
}

func TestDownTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{parseDec("1000000000000000000"), tickPrec, parseDec("999900000000000000")},
		{parseDec("10010"), tickPrec, parseDec("10000")},
		{parseDec("100.0"), tickPrec, parseDec("99.99")},
		{parseDec("99.99"), tickPrec, parseDec("99.98")},
		{parseDec("1.000"), tickPrec, parseDec("0.9999")},
		{parseDec("0.9990"), tickPrec, parseDec("0.9989")},
		{parseDec("0.9999"), tickPrec, parseDec("0.9998")},
		{parseDec("0.1"), tickPrec, parseDec("0.09999")},
		{parseDec("0.00000000000001000"), tickPrec, parseDec("0.000000000000009999")},
		{parseDec("0.000000000000001001"), tickPrec, parseDec("0.000000000000001000")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, types.DownTick(tc.price, tc.prec)))
		})
	}
}

func TestLowestTick(t *testing.T) {
	for _, tc := range []struct {
		prec     int
		expected sdk.Dec
	}{
		{0, sdk.NewDecWithPrec(1, 18)},
		{tickPrec, sdk.NewDecWithPrec(1, 15)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, types.LowestTick(tc.prec)))
		})
	}
}

func TestPriceToUpTick(t *testing.T) {
	for _, tc := range []struct {
		price sdk.Dec
		prec int
		expected sdk.Dec
	}{
		{parseDec("1.0015"), tickPrec, parseDec("1.002")},
		{parseDec("100"), tickPrec, parseDec("100")},
		{parseDec("100.01"), tickPrec, parseDec("100.1")},
		{parseDec("100.099"), tickPrec, parseDec("100.1")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, types.PriceToUpTick(tc.price, tc.prec)))
		})
	}
}

func TestRoundTickIndex(t *testing.T) {
	for _, tc := range []struct {
		i        int
		expected int
	}{
		{0, 0},
		{1, 2},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 6},
		{6, 6},
		{7, 8},
		{8, 8},
		{9, 10},
		{10, 10},
	} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.expected, types.RoundTickIndex(tc.i))
		})
	}
}

func TestRoundPrice(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{parseDec("0.000000000000001000"), tickPrec, parseDec("0.000000000000001000")},
		{parseDec("0.000000000000010000"), tickPrec, parseDec("0.000000000000010000")},
		{parseDec("0.000000000000010005"), tickPrec, parseDec("0.000000000000010000")},
		{parseDec("0.000000000000010015"), tickPrec, parseDec("0.000000000000010020")},
		{parseDec("0.000000000000010025"), tickPrec, parseDec("0.000000000000010020")},
		{parseDec("0.000000000000010035"), tickPrec, parseDec("0.000000000000010040")},
		{parseDec("0.000000000000010045"), tickPrec, parseDec("0.000000000000010040")},
		{parseDec("1.0005"), tickPrec, parseDec("1.0")},
		{parseDec("1.0015"), tickPrec, parseDec("1.002")},
		{parseDec("1.0025"), tickPrec, parseDec("1.002")},
		{parseDec("1.0035"), tickPrec, parseDec("1.004")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, types.RoundPrice(tc.price, tc.prec)))
		})
	}
}
