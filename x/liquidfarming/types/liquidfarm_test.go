package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func TestShareDenom(t *testing.T) {
	for i, tc := range []struct {
		denom      string
		expectsErr bool
	}{
		{"lfshare1", false},
		{"lfshare10", false},
		{"lfshare18446744073709551615", false},
		{"lfshare18446744073709551616", true},
		{"lfshareabc", true},
		{"lfshare01", true},
		{"lfshare-10", true},
		{"lfshare+10", true},
		{"ucre", true},
		{"denom1", true},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			liquidFarmId, err := types.ParseShareDenom(tc.denom)
			if tc.expectsErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.denom, types.ShareDenom(liquidFarmId))
			}
		})
	}
}

func TestLiquidFarm_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(liquidFarm *types.LiquidFarm)
		expectedErr string
	}{
		{
			"valid",
			func(liquidFarm *types.LiquidFarm) {},
			"",
		},
		{
			"invalid liquid farm id",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid pool id",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.PoolId = 0
			},
			"pool id must not be 0",
		},
		{
			"lower tick >= upper tick",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.LowerTick = 100
				liquidFarm.UpperTick = 100
			},
			"lower tick must be lower than upper tick",
		},
		{
			"lower tick > upper tick",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.LowerTick = 200
				liquidFarm.UpperTick = 100
			},
			"lower tick must be lower than upper tick",
		},
		{
			"invalid bid reserve address",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.BidReserveAddress = "invalidaddr"
			},
			"invalid bid reserve address decoding bech32 failed: invalid separator index -1",
		},
		{
			"negative fee rate",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.FeeRate = sdk.NewDec(-1)
			},
			"fee rate must not be negative: -1.000000000000000000",
		},
		{
			"too high fee rate",
			func(liquidFarm *types.LiquidFarm) {
				liquidFarm.FeeRate = sdk.NewDec(2)
			},
			"fee rate must not be greater than 1: 2.000000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			liquidFarm := types.NewLiquidFarm(
				1, 2, -100, 100, sdk.NewInt(10000), utils.ParseDec("0.003"))
			tc.malleate(&liquidFarm)
			err := liquidFarm.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestCalculateMintedShareAmount(t *testing.T) {
	for _, tc := range []struct {
		name           string
		shareSupply    sdk.Int
		totalLiquidity sdk.Int
		addedLiquidity sdk.Int
		expected       sdk.Int
	}{
		{
			name:           "initial minting",
			shareSupply:    sdk.ZeroInt(),
			totalLiquidity: sdk.ZeroInt(),
			addedLiquidity: sdk.NewInt(1_000_00_000),
			expected:       sdk.NewInt(1_000_00_000),
		},
		{
			name:           "normal",
			shareSupply:    sdk.NewInt(1_000_000_000),
			totalLiquidity: sdk.NewInt(1_000_000_000),
			addedLiquidity: sdk.NewInt(250_000_000),
			expected:       sdk.NewInt(250_000_000),
		},
		{
			name:           "rewards are auto compounded",
			shareSupply:    sdk.NewInt(1_000_000_000),
			totalLiquidity: sdk.NewInt(1_100_000_000),
			addedLiquidity: sdk.NewInt(100_000_000),
			expected:       sdk.NewInt(90_909_090),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mintingAmt := types.CalculateMintedShareAmount(
				tc.addedLiquidity, tc.totalLiquidity, tc.shareSupply)
			require.Equal(t, tc.expected, mintingAmt)
		})
	}
}

func TestCalculateRemovedLiquidity(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		shareSupply            sdk.Int
		totalLiquidity         sdk.Int
		burnedShare            sdk.Int
		prevWinningBidShareAmt sdk.Int
		expectedAmt            sdk.Int
	}{
		{
			name:                   "burn all",
			shareSupply:            sdk.NewInt(100_000_000),
			totalLiquidity:         sdk.NewInt(100_000_000),
			burnedShare:            sdk.NewInt(100_000_000),
			prevWinningBidShareAmt: sdk.ZeroInt(),
			expectedAmt:            sdk.NewInt(100_000_000),
		},
		{
			name:                   "burning small amount #1: no previous winning bid",
			shareSupply:            sdk.NewInt(100_000_000),
			totalLiquidity:         sdk.NewInt(100_000_000),
			burnedShare:            sdk.NewInt(1),
			prevWinningBidShareAmt: sdk.ZeroInt(),
			expectedAmt:            sdk.NewInt(1),
		},
		{
			name:                   "burning small amount #2: with previous winning bid",
			shareSupply:            sdk.NewInt(100_000_000),
			totalLiquidity:         sdk.NewInt(100_000_100),
			burnedShare:            sdk.NewInt(1),
			prevWinningBidShareAmt: sdk.NewInt(100),
			expectedAmt:            sdk.NewInt(1),
		},
		{
			name:                   "rewards are auto compounded #1: no previous winning bid",
			shareSupply:            sdk.NewInt(1_000_000_000),
			totalLiquidity:         sdk.NewInt(1_100_000_000),
			burnedShare:            sdk.NewInt(100_000_000),
			prevWinningBidShareAmt: sdk.ZeroInt(),
			expectedAmt:            sdk.NewInt(110_000_000),
		},
		{
			name:                   "rewards are auto compounded #1: with previous winning bid",
			shareSupply:            sdk.NewInt(1_000_000_000),
			totalLiquidity:         sdk.NewInt(1_100_000_000),
			burnedShare:            sdk.NewInt(100_000_000),
			prevWinningBidShareAmt: sdk.NewInt(100_000),
			expectedAmt:            sdk.NewInt(109_990_000),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unfarmingAmt := types.CalculateRemovedLiquidity(
				tc.burnedShare, tc.shareSupply, tc.totalLiquidity, tc.prevWinningBidShareAmt)
			require.Equal(t, tc.expectedAmt, unfarmingAmt)
		})
	}
}

func TestDeductFees(t *testing.T) {
	for _, tc := range []struct {
		name            string
		feeRate         sdk.Dec
		rewards         sdk.Coins
		deductedRewards sdk.Coins
		fees            sdk.Coins
	}{
		{
			name:            "zero fee rate",
			feeRate:         sdk.ZeroDec(),
			rewards:         utils.ParseCoins("100denom1"),
			deductedRewards: utils.ParseCoins("100denom1"),
			fees:            nil,
		},
		{
			name:            "fee rate - 10%",
			feeRate:         sdk.MustNewDecFromStr("0.1"),
			rewards:         utils.ParseCoins("100denom1"),
			deductedRewards: utils.ParseCoins("90denom1"),
			fees:            utils.ParseCoins("10denom1"),
		},
		{
			name:            "fee rate - 6.666666666666%",
			feeRate:         sdk.MustNewDecFromStr("0.066666666666666"),
			rewards:         utils.ParseCoins("100000denom1"),
			deductedRewards: utils.ParseCoins("93333denom1"),
			fees:            utils.ParseCoins("6667denom1"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deductedRewards, fees := types.DeductFees(tc.rewards, tc.feeRate)
			require.Equal(t, tc.deductedRewards, deductedRewards)
			require.Equal(t, tc.fees, fees)
		})
	}
}
