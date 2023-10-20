package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestSqrtPriceAtTick(t *testing.T) {
	for i, tc := range []struct {
		// args
		tick int32
		// result
		sqrtPrice cremath.BigDec
	}{
		{
			0,
			utils.ParseBigDec("1"),
		},
		{
			1,
			utils.ParseBigDec("1.000049998750062496094023416993798698"),
		},
		{
			-1,
			utils.ParseBigDec("0.999994999987499937499609372265604493"),
		},
		{
			types.MinTick,
			utils.ParseBigDec("0.0000001"),
		},
		{
			types.MaxTick,
			utils.ParseBigDec("1000000000000"),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			sqrtPrice := types.SqrtPriceAtTick(tc.tick)
			utils.AssertEqual(t, tc.sqrtPrice, sqrtPrice)
		})
	}
}

func TestAdjustTickToTickSpacing(t *testing.T) {
	tickSpacing := uint32(10)
	for i, tc := range []struct {
		price    sdk.Dec
		roundUp  bool
		expected sdk.Dec
	}{
		{utils.ParseDec("12345"), false, utils.ParseDec("12340")},
		{utils.ParseDec("12345"), true, utils.ParseDec("12350")},

		{utils.ParseDec("12.345"), false, utils.ParseDec("12.34")},
		{utils.ParseDec("12.345"), true, utils.ParseDec("12.35")},

		{utils.ParseDec("0.0012345"), false, utils.ParseDec("0.001234")},
		{utils.ParseDec("0.0012345"), true, utils.ParseDec("0.001235")},

		{utils.ParseDec("1.0001"), false, utils.ParseDec("1")},
		{utils.ParseDec("1.0001"), true, utils.ParseDec("1.001")},

		{utils.ParseDec("0.99999"), false, utils.ParseDec("0.9999")},
		{utils.ParseDec("0.99999"), true, utils.ParseDec("1")},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			price := exchangetypes.PriceAtTick(
				types.AdjustTickToTickSpacing(
					exchangetypes.TickAtPrice(tc.price), tickSpacing, tc.roundUp))
			require.Equal(t, tc.expected.String(), price.String())
		})
	}
}

func TestAdjustPriceToTickSpacing(t *testing.T) {
	for i, tc := range []struct {
		price    sdk.Dec
		roundUp  bool
		expected sdk.Dec
	}{
		{utils.ParseDec("12345"), true, utils.ParseDec("12350")},
		{utils.ParseDec("12345"), false, utils.ParseDec("12300")},
		{utils.ParseDec("12350.1"), true, utils.ParseDec("12400")},
		{utils.ParseDec("12350.1"), false, utils.ParseDec("12350")},
		{utils.ParseDec("12350"), true, utils.ParseDec("12350")},
		{utils.ParseDec("12350"), false, utils.ParseDec("12350")},
		{utils.ParseDec("0.00012345"), true, utils.ParseDec("0.00012350")},
		{utils.ParseDec("0.00012345"), false, utils.ParseDec("0.00012300")},
		{utils.ParseDec("0.000123501"), true, utils.ParseDec("0.00012400")},
		{utils.ParseDec("0.000123501"), false, utils.ParseDec("0.00012350")},
		{utils.ParseDec("0.00012350"), true, utils.ParseDec("0.00012350")},
		{utils.ParseDec("0.00012350"), false, utils.ParseDec("0.00012350")},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tick := types.AdjustPriceToTickSpacing(tc.price, 50, tc.roundUp)
			require.True(sdk.DecEq(t, tc.expected, exchangetypes.PriceAtTick(tick)))
		})
	}
}

func TestTickInfo_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(tickInfo *types.TickInfo)
		expectedErr string
	}{
		{
			"valid",
			func(tickInfo *types.TickInfo) {},
			"",
		},
		{
			"negative gross liquidity",
			func(tickInfo *types.TickInfo) {
				tickInfo.GrossLiquidity = sdk.NewInt(-1000_000000)
			},
			"gross liquidity must not be negative: -1000000000",
		},
		{
			"zero net liquidity",
			func(tickInfo *types.TickInfo) {
				tickInfo.NetLiquidity = sdk.NewInt(0)
			},
			"net liquidity must not be 0",
		},
		{
			"invalid fee growth outside",
			func(tickInfo *types.TickInfo) {
				tickInfo.FeeGrowthOutside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid fee growth outside: coin 0.000000000000000000ucre amount is not positive",
		},
		{
			"wrong fee growth outside coins number",
			func(tickInfo *types.TickInfo) {
				tickInfo.FeeGrowthOutside = utils.ParseDecCoins("0.0001ucre,0.0001uusd,0.0001uatom")
			},
			"number of coins in fee growth outside must not be higher than 2: 3",
		},
		{
			"invalid farming rewards growth outside",
			func(tickInfo *types.TickInfo) {
				tickInfo.FarmingRewardsGrowthOutside = sdk.DecCoins{sdk.NewInt64DecCoin("ucre", 0)}
			},
			"invalid farming rewards growth outside: coin 0.000000000000000000ucre amount is not positive",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tickInfo := types.TickInfo{
				GrossLiquidity:              sdk.NewInt(1000_000000),
				NetLiquidity:                sdk.NewInt(-1000_000000),
				FeeGrowthOutside:            utils.ParseDecCoins("0.0001ucre,0.0001uusd"),
				FarmingRewardsGrowthOutside: utils.ParseDecCoins("0.0001stake"),
			}
			tc.malleate(&tickInfo)
			err := tickInfo.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
