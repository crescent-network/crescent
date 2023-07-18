package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
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
				market.MakerFeeRate = utils.ParseDec("-1.1")
			},
			"minus maker fee rate must not exceed 1.0: -1.100000000000000000",
		},
		{
			"negative taker fee rate",
			func(market *types.Market) {
				market.TakerFeeRate = utils.ParseDec("-0.0015")
			},
			"taker fee rate must not be negative: -0.001500000000000000",
		},
		{
			"too high taker fee rate",
			func(market *types.Market) {
				market.TakerFeeRate = utils.ParseDec("1.1")
			},
			"taker fee rate must not exceed 1.0: 1.100000000000000000",
		},
		{
			"too high maker fee rate",
			func(market *types.Market) {
				market.MakerFeeRate = utils.ParseDec("1.1")
			},
			"maker fee rate must not exceed 1.0:1.100000000000000000",
		},
		{
			"minus maker fee rate higher than taker fee rate",
			func(market *types.Market) {
				market.MakerFeeRate = utils.ParseDec("-0.004")
				market.TakerFeeRate = utils.ParseDec("0.003")
			},
			"minus maker fee rate must not exceed 0.003000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			market := types.NewMarket(1, "ucre", "uusd", utils.ParseDec("-0.0015"), utils.ParseDec("0.003"))
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

func TestMarket(t *testing.T) {
	// Test DepositCoin
	market := types.NewMarket(1, "ucre", "uusd", utils.ParseDec("-0.0015"), utils.ParseDec("0.003"))
	testutil.AssertEqual(t, utils.ParseDecCoin("1000000uusd"), market.DepositCoin(true, sdk.NewDec(1000000)))
	testutil.AssertEqual(t, utils.ParseDecCoin("1000000ucre"), market.DepositCoin(false, sdk.NewDec(1000000)))

	// Test DeductTakerFee
	deducted, fee := market.DeductTakerFee(sdk.NewDec(123456789), false)
	testutil.AssertEqual(t, utils.ParseDec("123086418.633"), deducted)
	testutil.AssertEqual(t, utils.ParseDec("370370.367"), fee)
	deducted, fee = market.DeductTakerFee(sdk.NewDec(123456789), true)
	testutil.AssertEqual(t, utils.ParseDec("123271603.8165"), deducted)
	testutil.AssertEqual(t, utils.ParseDec("185185.1835"), fee)

	r := rand.New(rand.NewSource(1))
	for i := 0; i < 50; i++ {
		amt := utils.RandomDec(r, sdk.NewDec(10), sdk.NewDec(100000000))
		deducted, fee = market.DeductTakerFee(amt, false)
		testutil.AssertEqual(t, amt, deducted.Add(fee))
		deducted, fee = market.DeductTakerFee(amt, true)
		testutil.AssertEqual(t, amt, deducted.Add(fee))
	}

	payDenom, receiveDenom := market.PayReceiveDenoms(true)
	require.Equal(t, "uusd", payDenom)
	require.Equal(t, "ucre", receiveDenom)
	payDenom, receiveDenom = market.PayReceiveDenoms(false)
	require.Equal(t, "ucre", payDenom)
	require.Equal(t, "uusd", receiveDenom)
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
