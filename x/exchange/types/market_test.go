package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestMarket_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(market *types.Market)
		expectedErr string
	}{
		{
			"valid",
			func(market *types.Market) {},
			"",
		},
		{
			"id is 0",
			func(market *types.Market) {
				market.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid base denom",
			func(market *types.Market) {
				market.BaseDenom = "invaliddenom!"
			},
			"invalid base denom: invalid denom: invaliddenom!",
		},
		{
			"invalid quote denom",
			func(market *types.Market) {
				market.QuoteDenom = "invaliddenom!"
			},
			"invalid quote denom: invalid denom: invaliddenom!",
		},
		{
			"same base denom and quote denom",
			func(market *types.Market) {
				market.BaseDenom = "ucre"
				market.QuoteDenom = "ucre"
			},
			"base denom and quote denom must not be same: ucre",
		},
		{
			"invalid escrow address",
			func(market *types.Market) {
				market.EscrowAddress = "invalidaddr"
			},
			"invalid escrow address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid fee collector",
			func(market *types.Market) {
				market.FeeCollector = "invalidaddr"
			},
			"invalid fee collector: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid fees",
			func(market *types.Market) {
				market.Fees.MakerFeeRate = utils.ParseDec("-0.0015")
			},
			"maker fee rate must be in range [0, 1]: -0.001500000000000000",
		},
		{
			"invalid order quantity limits",
			func(market *types.Market) {
				market.OrderQuantityLimits.Min = sdk.NewInt(0)
			},
			"invalid order quantity limits: the minimum value must be positive: 0",
		},
		{
			"invalid order quote limits",
			func(market *types.Market) {
				market.OrderQuoteLimits.Min = sdk.NewInt(0)
			},
			"invalid order quote limits: the minimum value must be positive: 0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			market := types.NewMarket(
				1, "ucre", "uusd",
				types.DefaultFees, types.DefaultOrderQuantityLimits, types.DefaultOrderQuoteLimits)
			tc.malleate(&market)
			err := market.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMarketState_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(marketState *types.MarketState)
		expectedErr string
	}{
		{
			"valid",
			func(marketState *types.MarketState) {},
			"",
		},
		{
			"negative last price",
			func(marketState *types.MarketState) {
				marketState.LastPrice = utils.ParseDecP("-12.345")
			},
			"last price must be positive: -12.345000000000000000",
		},
		{
			"invalid last price tick",
			func(marketState *types.MarketState) {
				marketState.LastPrice = utils.ParseDecP("12.34567")
			},
			"invalid last price tick: 12.345670000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marketState := types.NewMarketState(nil)
			tc.malleate(&marketState)
			err := marketState.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestOrderPriceLimit(t *testing.T) {
	for i, tc := range []struct {
		basePrice, maxOrderPriceRatio sdk.Dec
		minPrice, maxPrice            sdk.Dec
	}{
		{
			utils.ParseDec("1"), utils.ParseDec("0.1"),
			utils.ParseDec("0.9"), utils.ParseDec("1.1"),
		},
		{
			utils.ParseDec("5"), utils.ParseDec("0.1"),
			utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		},
		{
			utils.ParseDec("1.2345"), utils.ParseDec("0.1"),
			// Min price is rounded up and max price is rounded down
			utils.ParseDec("1.1111"), utils.ParseDec("1.3579"),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			minPrice, maxPrice := types.OrderPriceLimit(tc.basePrice, tc.maxOrderPriceRatio)
			require.True(sdk.DecEq(t, tc.minPrice, minPrice))
			require.True(sdk.DecEq(t, tc.maxPrice, maxPrice))
		})
	}
}
