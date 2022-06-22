package amm_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/amm"
)

func TestBasicPool(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 1000; i++ {
		rx, ry := sdk.NewInt(1+r.Int63n(100000000)), sdk.NewInt(1+r.Int63n(100000000))
		pool := amm.NewBasicPool(rx, ry, sdk.Int{})

		highest, found := pool.HighestBuyPrice()
		require.True(t, found)
		require.True(sdk.DecEq(t, pool.Price(), highest))
		lowest, found := pool.LowestSellPrice()
		require.True(t, found)
		require.True(sdk.DecEq(t, pool.Price(), lowest))
	}
}

func TestBasicPool_Price(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rx, ry int64   // reserve balance
		ps     int64   // pool coin supply
		p      sdk.Dec // expected pool price
	}{
		{
			name: "normal pool",
			ps:   10000,
			rx:   20000,
			ry:   100,
			p:    sdk.NewDec(200),
		},
		{
			name: "decimal rounding",
			ps:   10000,
			rx:   200,
			ry:   300,
			p:    sdk.MustNewDecFromStr("0.666666666666666667"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.NewBasicPool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			require.True(sdk.DecEq(t, tc.p, pool.Price()))
		})
	}

	// panicking cases
	for _, tc := range []struct {
		rx, ry int64
		ps     int64
	}{
		{
			rx: 0,
			ry: 1000,
			ps: 1000,
		},
		{
			rx: 1000,
			ry: 0,
			ps: 1000,
		},
	} {
		t.Run("panics", func(t *testing.T) {
			require.Panics(t, func() {
				pool := amm.NewBasicPool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
				pool.Price()
			})
		})
	}
}

func TestBasicPool_IsDepleted(t *testing.T) {
	for _, tc := range []struct {
		name       string
		rx, ry     int64 // reserve balance
		ps         int64 // pool coin supply
		isDepleted bool
	}{
		{
			name:       "empty pool",
			rx:         0,
			ry:         0,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "depleted, with some coins from outside",
			rx:         100,
			ry:         0,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "depleted, with some coins from outside #2",
			rx:         100,
			ry:         100,
			ps:         0,
			isDepleted: true,
		},
		{
			name:       "normal pool",
			rx:         10000,
			ry:         10000,
			ps:         10000,
			isDepleted: false,
		},
		{
			name:       "not depleted, but reserve coins are gone",
			rx:         0,
			ry:         10000,
			ps:         10000,
			isDepleted: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.NewBasicPool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			require.Equal(t, tc.isDepleted, pool.IsDepleted())
		})
	}
}

func TestBasicPool_Deposit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rx, ry int64 // reserve balance
		ps     int64 // pool coin supply
		x, y   int64 // depositing coin amount
		ax, ay int64 // expected accepted coin amount
		pc     int64 // expected minted pool coin amount
	}{
		{
			name: "ideal deposit",
			rx:   2000,
			ry:   100,
			ps:   10000,
			x:    200,
			y:    10,
			ax:   200,
			ay:   10,
			pc:   1000,
		},
		{
			name: "unbalanced deposit",
			rx:   2000,
			ry:   100,
			ps:   10000,
			x:    100,
			y:    2000,
			ax:   100,
			ay:   5,
			pc:   500,
		},
		{
			name: "decimal truncation",
			rx:   222,
			ry:   333,
			ps:   333,
			x:    100,
			y:    100,
			ax:   66,
			ay:   99,
			pc:   99,
		},
		{
			name: "decimal truncation #2",
			rx:   200,
			ry:   300,
			ps:   333,
			x:    80,
			y:    80,
			ax:   53,
			ay:   80,
			pc:   88,
		},
		{
			name: "zero minting amount",
			ps:   100,
			rx:   10000,
			ry:   10000,
			x:    99,
			y:    99,
			ax:   0,
			ay:   0,
			pc:   0,
		},
		{
			name: "tiny minting amount",
			rx:   10000,
			ry:   10000,
			ps:   100,
			x:    100,
			y:    100,
			ax:   100,
			ay:   100,
			pc:   1,
		},
		{
			name: "tiny minting amount #2",
			rx:   10000,
			ry:   10000,
			ps:   100,
			x:    199,
			y:    199,
			ax:   100,
			ay:   100,
			pc:   1,
		},
		{
			name: "zero minting amount",
			rx:   10000,
			ry:   10000,
			ps:   999,
			x:    10,
			y:    10,
			ax:   0,
			ay:   0,
			pc:   0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := amm.NewBasicPool(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps))
			ax, ay, pc := amm.Deposit(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps), sdk.NewInt(tc.x), sdk.NewInt(tc.y))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.ax), ax))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.ay), ay))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.pc), pc))
			// Additional assertions
			if !pool.IsDepleted() {
				require.True(t, (ax.Int64()*tc.ps) >= (pc.Int64()*tc.rx)) // (ax / rx) > (pc / ps)
				require.True(t, (ay.Int64()*tc.ps) >= (pc.Int64()*tc.ry)) // (ay / ry) > (pc / ps)
			}
		})
	}
}

func TestBasicPool_Withdraw(t *testing.T) {
	for _, tc := range []struct {
		name    string
		rx, ry  int64 // reserve balance
		ps      int64 // pool coin supply
		pc      int64 // redeeming pool coin amount
		feeRate sdk.Dec
		x, y    int64 // withdrawn coin amount
	}{
		{
			name:    "ideal withdraw",
			rx:      2000,
			ry:      100,
			ps:      10000,
			pc:      1000,
			feeRate: sdk.ZeroDec(),
			x:       200,
			y:       10,
		},
		{
			name:    "ideal withdraw - with fee",
			rx:      2000,
			ry:      100,
			ps:      10000,
			pc:      1000,
			feeRate: sdk.MustNewDecFromStr("0.003"),
			x:       199,
			y:       9,
		},
		{
			name:    "withdraw all",
			rx:      123,
			ry:      567,
			ps:      10,
			pc:      10,
			feeRate: sdk.MustNewDecFromStr("0.003"),
			x:       123,
			y:       567,
		},
		{
			name:    "advantageous for pool",
			rx:      100,
			ry:      100,
			ps:      10000,
			pc:      99,
			feeRate: sdk.ZeroDec(),
			x:       0,
			y:       0,
		},
		{
			name:    "advantageous for pool",
			rx:      10000,
			ry:      100,
			ps:      10000,
			pc:      99,
			feeRate: sdk.ZeroDec(),
			x:       99,
			y:       0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x, y := amm.Withdraw(sdk.NewInt(tc.rx), sdk.NewInt(tc.ry), sdk.NewInt(tc.ps), sdk.NewInt(tc.pc), tc.feeRate)
			require.True(sdk.IntEq(t, sdk.NewInt(tc.x), x))
			require.True(sdk.IntEq(t, sdk.NewInt(tc.y), y))
			// Additional assertions
			require.True(t, (tc.pc*tc.rx) >= (x.Int64()*tc.ps))
			require.True(t, (tc.pc*tc.ry) >= (y.Int64()*tc.ps))
		})
	}
}

func TestInitialPoolCoinSupply(t *testing.T) {
	for _, tc := range []struct {
		x, y sdk.Int
		ps   sdk.Int
	}{
		{sdk.NewInt(1000000), sdk.NewInt(1000000), sdk.NewInt(10000000)},
		{sdk.NewInt(1000000), sdk.NewInt(10000000), sdk.NewInt(100000000)},
		{sdk.NewInt(1000000), sdk.NewInt(100000000), sdk.NewInt(100000000)},
		{sdk.NewInt(10000000), sdk.NewInt(100000000), sdk.NewInt(1000000000)},
		{sdk.NewInt(999999), sdk.NewInt(9999999), sdk.NewInt(10000000)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.ps, amm.InitialPoolCoinSupply(tc.x, tc.y)))
		})
	}
}

func TestBasicPool_BuyAmountOverOverflow(t *testing.T) {
	n, _ := sdk.NewIntFromString("10000000000000000000000000000000000000000000")
	pool := amm.NewBasicPool(n, sdk.NewInt(1000), sdk.Int{})
	amt := pool.BuyAmountOver(defTickPrec.LowestTick(), true)
	require.True(sdk.IntEq(t, amm.MaxCoinAmount, amt))
}

func TestBasicPoolOrders(t *testing.T) {
	pool := amm.NewBasicPool(sdk.NewInt(862431695563), sdk.NewInt(37852851767), sdk.Int{})
	poolPrice := pool.Price()
	lowestPrice := poolPrice.Mul(sdk.NewDecWithPrec(9, 1))
	highestPrice := poolPrice.Mul(sdk.NewDecWithPrec(11, 1))
	require.Len(t, amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4), 375)
}

func BenchmarkBasicPoolOrders(b *testing.B) {
	pool := amm.NewBasicPool(sdk.NewInt(862431695563), sdk.NewInt(37852851767), sdk.Int{})
	poolPrice := pool.Price()
	lowestPrice := poolPrice.Mul(sdk.NewDecWithPrec(9, 1))
	highestPrice := poolPrice.Mul(sdk.NewDecWithPrec(11, 1))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		amm.PoolOrders(pool, amm.DefaultOrderer, lowestPrice, highestPrice, 4)
	}
}
