package amm_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

const defTickPrec = amm.TickPrecision(3)

func TestPriceToTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		expected sdk.Dec
	}{
		{squad.ParseDec("0.000000000000099999"), squad.ParseDec("0.00000000000009999")},
		{squad.ParseDec("1.999999999999999999"), squad.ParseDec("1.999")},
		{squad.ParseDec("99.999999999999999999"), squad.ParseDec("99.99")},
		{squad.ParseDec("100.999999999999999999"), squad.ParseDec("100.9")},
		{squad.ParseDec("9999.999999999999999999"), squad.ParseDec("9999")},
		{squad.ParseDec("10019"), squad.ParseDec("10010")},
		{squad.ParseDec("1000100005"), squad.ParseDec("1000000000")},
	} {
		require.True(sdk.DecEq(t, tc.expected, amm.PriceToDownTick(tc.price, 3)))
	}
}

func TestTick(t *testing.T) {
	for _, tc := range []struct {
		i        int
		prec     int
		expected sdk.Dec
	}{
		{0, 3, sdk.NewDecWithPrec(1, int64(sdk.Precision-defTickPrec))},
		{1, 3, squad.ParseDec("0.000000000000001001")},
		{8999, 3, squad.ParseDec("0.000000000000009999")},
		{9000, 3, squad.ParseDec("0.000000000000010000")},
		{9001, 3, squad.ParseDec("0.000000000000010010")},
		{17999, 3, squad.ParseDec("0.000000000000099990")},
		{18000, 3, squad.ParseDec("0.000000000000100000")},
		{135000, 3, sdk.NewDec(1)},
		{135001, 3, squad.ParseDec("1.001")},
	} {
		t.Run("", func(t *testing.T) {
			res := amm.TickFromIndex(tc.i, tc.prec)
			require.True(sdk.DecEq(t, tc.expected, res))
			require.Equal(t, tc.i, amm.TickToIndex(res, tc.prec))
		})
	}
}

func TestUpTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{squad.ParseDec("1000000000000000000"), 3, squad.ParseDec("1001000000000000000")},
		{squad.ParseDec("1000"), 3, squad.ParseDec("1001")},
		{squad.ParseDec("999.9"), 3, squad.ParseDec("1000")},
		{squad.ParseDec("999.0"), 3, squad.ParseDec("999.1")},
		{squad.ParseDec("1.100"), 3, squad.ParseDec("1.101")},
		{squad.ParseDec("1.000"), 3, squad.ParseDec("1.001")},
		{squad.ParseDec("0.9999"), 3, squad.ParseDec("1.000")},
		{squad.ParseDec("0.1000"), 3, squad.ParseDec("0.1001")},
		{squad.ParseDec("0.09999"), 3, squad.ParseDec("0.1000")},
		{squad.ParseDec("0.09997"), 3, squad.ParseDec("0.09998")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, amm.UpTick(tc.price, tc.prec)))
		})
	}
}

func TestDownTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{squad.ParseDec("1000000000000000000"), 3, squad.ParseDec("999900000000000000")},
		{squad.ParseDec("10010"), 3, squad.ParseDec("10000")},
		{squad.ParseDec("100.0"), 3, squad.ParseDec("99.99")},
		{squad.ParseDec("99.99"), 3, squad.ParseDec("99.98")},
		{squad.ParseDec("1.000"), 3, squad.ParseDec("0.9999")},
		{squad.ParseDec("0.9990"), 3, squad.ParseDec("0.9989")},
		{squad.ParseDec("0.9999"), 3, squad.ParseDec("0.9998")},
		{squad.ParseDec("0.1"), 3, squad.ParseDec("0.09999")},
		{squad.ParseDec("0.00000000000001000"), 3, squad.ParseDec("0.000000000000009999")},
		{squad.ParseDec("0.000000000000001001"), 3, squad.ParseDec("0.000000000000001000")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, amm.DownTick(tc.price, tc.prec)))
		})
	}
}

func TestHighestTick(t *testing.T) {
	for _, tc := range []struct {
		prec     int
		expected string
	}{
		{3, "133400000000000000000000000000000000000000000000000000000000000000000000000000"},
		{0, "100000000000000000000000000000000000000000000000000000000000000000000000000000"},
		{1, "130000000000000000000000000000000000000000000000000000000000000000000000000000"},
	} {
		t.Run("", func(t *testing.T) {
			i, ok := new(big.Int).SetString(tc.expected, 10)
			require.True(t, ok)
			tick := amm.HighestTick(tc.prec)
			require.True(sdk.DecEq(t, sdk.NewDecFromBigInt(i), tick))
			require.Panics(t, func() {
				amm.UpTick(tick, tc.prec)
			})
		})
	}
}

func TestLowestTick(t *testing.T) {
	for _, tc := range []struct {
		prec     int
		expected sdk.Dec
	}{
		{0, sdk.NewDecWithPrec(1, 18)},
		{3, sdk.NewDecWithPrec(1, 15)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, amm.LowestTick(tc.prec)))
		})
	}
}

func TestPriceToUpTick(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{squad.ParseDec("1.0015"), 3, squad.ParseDec("1.002")},
		{squad.ParseDec("100"), 3, squad.ParseDec("100")},
		{squad.ParseDec("100.01"), 3, squad.ParseDec("100.1")},
		{squad.ParseDec("100.099"), 3, squad.ParseDec("100.1")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, amm.PriceToUpTick(tc.price, tc.prec)))
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
			require.Equal(t, tc.expected, amm.RoundTickIndex(tc.i))
		})
	}
}

func TestRoundPrice(t *testing.T) {
	for _, tc := range []struct {
		price    sdk.Dec
		prec     int
		expected sdk.Dec
	}{
		{squad.ParseDec("0.000000000000001000"), 3, squad.ParseDec("0.000000000000001000")},
		{squad.ParseDec("0.000000000000010000"), 3, squad.ParseDec("0.000000000000010000")},
		{squad.ParseDec("0.000000000000010005"), 3, squad.ParseDec("0.000000000000010000")},
		{squad.ParseDec("0.000000000000010015"), 3, squad.ParseDec("0.000000000000010020")},
		{squad.ParseDec("0.000000000000010025"), 3, squad.ParseDec("0.000000000000010020")},
		{squad.ParseDec("0.000000000000010035"), 3, squad.ParseDec("0.000000000000010040")},
		{squad.ParseDec("0.000000000000010045"), 3, squad.ParseDec("0.000000000000010040")},
		{squad.ParseDec("1.0005"), 3, squad.ParseDec("1.0")},
		{squad.ParseDec("1.0015"), 3, squad.ParseDec("1.002")},
		{squad.ParseDec("1.0025"), 3, squad.ParseDec("1.002")},
		{squad.ParseDec("1.0035"), 3, squad.ParseDec("1.004")},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.DecEq(t, tc.expected, amm.RoundPrice(tc.price, tc.prec)))
		})
	}
}

func BenchmarkUpTick(b *testing.B) {
	b.Run("price fit in ticks", func(b *testing.B) {
		price := squad.ParseDec("0.9999")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			amm.UpTick(price, 3)
		}
	})
	b.Run("price not fit in ticks", func(b *testing.B) {
		price := squad.ParseDec("0.99995")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			amm.UpTick(price, 3)
		}
	})
}

func BenchmarkDownTick(b *testing.B) {
	b.Run("price fit in ticks", func(b *testing.B) {
		price := squad.ParseDec("0.9999")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			amm.DownTick(price, 3)
		}
	})
	b.Run("price not fit in ticks", func(b *testing.B) {
		price := squad.ParseDec("0.99995")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			amm.DownTick(price, 3)
		}
	})
	b.Run("price at edge", func(b *testing.B) {
		price := squad.ParseDec("1")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			amm.DownTick(price, 3)
		}
	})
}
