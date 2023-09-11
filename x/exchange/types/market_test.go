package types_test

import (
	"testing"

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
			"too low maker fee rate",
			func(market *types.Market) {
				market.MakerFeeRate = utils.ParseDec("-0.0015")
			},
			"maker fee rate must be in range [0, 1]: -0.001500000000000000",
		},
		{
			"too high maker fee rate",
			func(market *types.Market) {
				market.MakerFeeRate = utils.ParseDec("1.1")
			},
			"maker fee rate must be in range [0, 1]: 1.100000000000000000",
		},
		{
			"too low taker fee rate",
			func(market *types.Market) {
				market.TakerFeeRate = utils.ParseDec("-0.0015")
			},
			"taker fee rate must be in range [0, 1]: -0.001500000000000000",
		},
		{
			"too high taker fee rate",
			func(market *types.Market) {
				market.TakerFeeRate = utils.ParseDec("1.1")
			},
			"taker fee rate must be in range [0, 1]: 1.100000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			market := types.NewMarket(
				1, "ucre", "uusd", utils.ParseDec("0.0015"), utils.ParseDec("0.003"), utils.ParseDec("0.5"))
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
