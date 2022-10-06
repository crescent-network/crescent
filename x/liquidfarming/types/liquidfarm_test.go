package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
)

func TestLiquidFarm(t *testing.T) {
	liquidFarm := types.LiquidFarm{
		PoolId:        1,
		MinFarmAmount: sdk.ZeroInt(),
		MinBidAmount:  sdk.ZeroInt(),
		FeeRate:       sdk.ZeroDec(),
	}
	require.Equal(t, `fee_rate: "0.000000000000000000"
min_bid_amount: "0"
min_farm_amount: "0"
pool_id: "1"
`, liquidFarm.String())
}

func TestLiquidFarmCoinDenom(t *testing.T) {
	for _, tc := range []struct {
		denom      string
		expectsErr bool
	}{
		{"lf1", false},
		{"lf10", false},
		{"lf18446744073709551615", false},
		{"lf18446744073709551616", true},
		{"lfabc", true},
		{"lf01", true},
		{"lf-10", true},
		{"lf+10", true},
		{"ucre", true},
		{"denom1", true},
	} {
		t.Run("", func(t *testing.T) {
			poolId, err := types.ParseLiquidFarmCoinDenom(tc.denom)
			if tc.expectsErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.denom, types.LiquidFarmCoinDenom(poolId))
			}
		})
	}
}

func TestLiquidFarmReserveAddress(t *testing.T) {
	config := sdk.GetConfig()
	addrPrefix := config.GetBech32AccountAddrPrefix()

	for _, tc := range []struct {
		poolId   uint64
		expected string
	}{
		{1, addrPrefix + "1zyyf855slxure4c8dr06p00qjnkem95d2lgv8wgvry2rt437x6tsaf9tcf"},
		{2, addrPrefix + "1d2csu4ynxpuxll8wk72n9z98ytm649u78paj9efskjwrlc2wyhpq8h886j"},
	} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.expected, types.LiquidFarmReserveAddress(tc.poolId).String())
		})
	}
}

func TestCalculateLiquidFarmAmount(t *testing.T) {
	for _, tc := range []struct {
		name              string
		lfTotalSupplyAmt  sdk.Int
		lpTotalFarmingAmt sdk.Int
		newFarmingAmt     sdk.Int
		expectedAmt       sdk.Int
	}{
		{
			name:              "initial minting",
			lfTotalSupplyAmt:  sdk.ZeroInt(),
			lpTotalFarmingAmt: sdk.ZeroInt(),
			newFarmingAmt:     sdk.NewInt(1_000_00_000),
			expectedAmt:       sdk.NewInt(1_000_00_000),
		},
		{
			name:              "normal",
			lfTotalSupplyAmt:  sdk.NewInt(1_000_000_000),
			lpTotalFarmingAmt: sdk.NewInt(1_000_000_000),
			newFarmingAmt:     sdk.NewInt(250_000_000),
			expectedAmt:       sdk.NewInt(250_000_000),
		},
		{
			name:              "rewards are auto compounded",
			lfTotalSupplyAmt:  sdk.NewInt(1_000_000_000),
			lpTotalFarmingAmt: sdk.NewInt(1_100_000_000),
			newFarmingAmt:     sdk.NewInt(100_000_000),
			expectedAmt:       sdk.NewInt(90_909_090),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mintingAmt := types.CalculateLiquidFarmAmount(
				tc.lfTotalSupplyAmt,
				tc.lpTotalFarmingAmt,
				tc.newFarmingAmt,
			)
			require.Equal(t, tc.expectedAmt, mintingAmt)
		})
	}
}

func TestCalculateLiquidUnfarmAmount(t *testing.T) {
	for _, tc := range []struct {
		name               string
		lfTotalSupplyAmt   sdk.Int
		lpTotalFarmingAmt  sdk.Int
		unfarmingAmt       sdk.Int
		compoundingRewards sdk.Int
		expectedAmt        sdk.Int
	}{
		{
			name:               "unfarm all",
			lfTotalSupplyAmt:   sdk.NewInt(100_000_000),
			lpTotalFarmingAmt:  sdk.NewInt(100_000_000),
			unfarmingAmt:       sdk.NewInt(100_000_000),
			compoundingRewards: sdk.ZeroInt(),
			expectedAmt:        sdk.NewInt(100_000_000),
		},
		{
			name:               "unfarming small amount #1: no compounding rewards",
			lfTotalSupplyAmt:   sdk.NewInt(100_000_000),
			lpTotalFarmingAmt:  sdk.NewInt(100_000_000),
			unfarmingAmt:       sdk.NewInt(1),
			compoundingRewards: sdk.ZeroInt(),
			expectedAmt:        sdk.NewInt(1),
		},
		{
			name:               "unfarming small amount #2: with compounding rewards",
			lfTotalSupplyAmt:   sdk.NewInt(100_000_000),
			lpTotalFarmingAmt:  sdk.NewInt(100_000_100),
			unfarmingAmt:       sdk.NewInt(1),
			compoundingRewards: sdk.NewInt(100),
			expectedAmt:        sdk.NewInt(1),
		},
		{
			name:               "rewards are auto compounded #1: no compouding rewards",
			lfTotalSupplyAmt:   sdk.NewInt(1_000_000_000),
			lpTotalFarmingAmt:  sdk.NewInt(1_100_000_000),
			unfarmingAmt:       sdk.NewInt(100_000_000),
			compoundingRewards: sdk.ZeroInt(),
			expectedAmt:        sdk.NewInt(110_000_000),
		},
		{
			name:               "rewards are auto compounded #1: with compouding rewards",
			lfTotalSupplyAmt:   sdk.NewInt(1_000_000_000),
			lpTotalFarmingAmt:  sdk.NewInt(1_100_000_000),
			unfarmingAmt:       sdk.NewInt(100_000_000),
			compoundingRewards: sdk.NewInt(100_000),
			expectedAmt:        sdk.NewInt(109_990_000),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unfarmingAmt := types.CalculateLiquidUnfarmAmount(
				tc.lfTotalSupplyAmt,
				tc.lpTotalFarmingAmt,
				tc.unfarmingAmt,
				tc.compoundingRewards,
			)
			require.Equal(t, tc.expectedAmt, unfarmingAmt)
		})
	}
}

func TestDeductFees(t *testing.T) {
	for _, tc := range []struct {
		name     string
		feeRate  sdk.Dec
		rewards  sdk.Coins
		deducted sdk.Coins
	}{
		{
			name:     "zero fee rate",
			feeRate:  sdk.ZeroDec(),
			rewards:  utils.ParseCoins("100denom1"),
			deducted: utils.ParseCoins("100denom1"),
		},
		{
			name:     "fee rate - 10%",
			feeRate:  sdk.MustNewDecFromStr("0.1"),
			rewards:  utils.ParseCoins("100denom1"),
			deducted: utils.ParseCoins("90denom1"),
		},
		{
			name:     "fee rate - 6.666666666666%",
			feeRate:  sdk.MustNewDecFromStr("0.066666666666666"),
			rewards:  utils.ParseCoins("100000denom1"),
			deducted: utils.ParseCoins("93333denom1"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deducted, err := types.DeductFees(1, tc.rewards, tc.feeRate)
			require.NoError(t, err)
			require.Equal(t, tc.deducted, deducted)
		})
	}
}
