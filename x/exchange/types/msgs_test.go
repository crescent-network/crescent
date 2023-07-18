package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestMsgCreateMarket(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCreateMarket)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCreateMarket) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCreateMarket) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid base denom",
			func(msg *types.MsgCreateMarket) {
				msg.BaseDenom = "invaliddenom!"
			},
			"invalid base denom: invalid denom: invaliddenom!: invalid request",
		},
		{
			"invalid quote denom",
			func(msg *types.MsgCreateMarket) {
				msg.QuoteDenom = "invaliddenom!"
			},
			"invalid quote denom: invalid denom: invaliddenom!: invalid request",
		},
		{
			"same base denom and quote denom",
			func(msg *types.MsgCreateMarket) {
				msg.BaseDenom = "ucre"
				msg.QuoteDenom = "ucre"
			},
			"base denom and quote denom must not be same: ucre: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCreateMarket(senderAddr, "ucre", "uusd")
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCreateMarket, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"too high price",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("50000000000000000000000000000000000000000")
			},
			"price is higher than the max price; 50000000000000000000000000000000000000000.000000000000000000 > 10000000000000000000000000000000000000000.000000000000000000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Quantity = sdk.NewDec(0)
			},
			"quantity must be positive: 0.000000000000000000: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Quantity = sdk.NewDec(-1000000)
			},
			"quantity must be positive: -1000000.000000000000000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgPlaceLimitOrder(
				senderAddr, 1, true, utils.ParseDec("12.345"), sdk.NewDec(1000000), time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgPlaceLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceBatchLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceBatchLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceBatchLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"too high price",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Price = utils.ParseDec("50000000000000000000000000000000000000000")
			},
			"price is higher than the max price; 50000000000000000000000000000000000000000.000000000000000000 > 10000000000000000000000000000000000000000.000000000000000000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Quantity = sdk.NewDec(0)
			},
			"quantity must be positive: 0.000000000000000000: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Quantity = sdk.NewDec(-1000000)
			},
			"quantity must be positive: -1000000.000000000000000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceBatchLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgPlaceBatchLimitOrder(
				senderAddr, 1, true, utils.ParseDec("12.345"), sdk.NewDec(1000000), time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgPlaceBatchLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceMMLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceMMLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceMMLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"too high price",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("50000000000000000000000000000000000000000")
			},
			"price is higher than the max price; 50000000000000000000000000000000000000000.000000000000000000 > 10000000000000000000000000000000000000000.000000000000000000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Quantity = sdk.NewDec(0)
			},
			"quantity must be positive: 0.000000000000000000: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Quantity = sdk.NewDec(-1000000)
			},
			"quantity must be positive: -1000000.000000000000000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceMMLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgPlaceMMLimitOrder(
				senderAddr, 1, true, utils.ParseDec("12.345"), sdk.NewDec(1000000), time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgPlaceMMLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceMMBatchLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceMMBatchLimitOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero price",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Price = utils.ParseDec("0")
			},
			"price is lower than the min price; 0.000000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"negative price",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Price = utils.ParseDec("-12.345")
			},
			"price is lower than the min price; -12.345000000000000000 < 0.000000000000010000: invalid request",
		},
		{
			"too high price",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Price = utils.ParseDec("50000000000000000000000000000000000000000")
			},
			"price is higher than the max price; 50000000000000000000000000000000000000000.000000000000000000 > 10000000000000000000000000000000000000000.000000000000000000: invalid request",
		},
		{
			"invalid price tick",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Price = utils.ParseDec("12.34567")
			},
			"invalid price tick: 12.345670000000000000: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Quantity = sdk.NewDec(0)
			},
			"quantity must be positive: 0.000000000000000000: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Quantity = sdk.NewDec(-1000000)
			},
			"quantity must be positive: -1000000.000000000000000000: invalid request",
		},
		{
			"zero lifespan",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Lifespan = 0
			},
			"",
		},
		{
			"negative lifespan",
			func(msg *types.MsgPlaceMMBatchLimitOrder) {
				msg.Lifespan = -time.Hour
			},
			"lifespan must not be negative: -1h0m0s: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgPlaceMMBatchLimitOrder(
				senderAddr, 1, true, utils.ParseDec("12.345"), sdk.NewDec(1000000), time.Hour)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgPlaceMMBatchLimitOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgPlaceMarketOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgPlaceMarketOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgPlaceMarketOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero quantity",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Quantity = sdk.NewDec(0)
			},
			"quantity must be positive: 0.000000000000000000: invalid request",
		},
		{
			"negative quantity",
			func(msg *types.MsgPlaceMarketOrder) {
				msg.Quantity = sdk.NewDec(-1000000)
			},
			"quantity must be positive: -1000000.000000000000000000: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgPlaceMarketOrder(
				senderAddr, 1, true, sdk.NewDec(1000000))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgPlaceMarketOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCancelOrder(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCancelOrder)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCancelOrder) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCancelOrder) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid order id",
			func(msg *types.MsgCancelOrder) {
				msg.OrderId = 0
			},
			"order id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCancelOrder(senderAddr, 1)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCancelOrder, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgCancelAllOrders(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgCancelAllOrders)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgCancelAllOrders) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgCancelAllOrders) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"invalid market id",
			func(msg *types.MsgCancelAllOrders) {
				msg.MarketId = 0
			},
			"market id must not be 0: invalid request",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgCancelAllOrders(senderAddr, 1)
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgCancelAllOrders, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestMsgSwapExactAmountIn(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(msg *types.MsgSwapExactAmountIn)
		expectedErr string
	}{
		{
			"valid",
			func(msg *types.MsgSwapExactAmountIn) {},
			"",
		},
		{
			"invalid sender",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Sender = "invalidaddr"
			},
			"invalid sender address: decoding bech32 failed: invalid separator index -1: invalid address",
		},
		{
			"empty routes",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Routes = []uint64{}
			},
			"routes must not be empty: invalid request",
		},
		{
			"invalid market id",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Routes = []uint64{0, 1}
			},
			"market id must not be 0: invalid request",
		},
		{
			"zero input",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Input = utils.ParseDecCoin("0ucre")
			},
			"input must be positive: 0.000000000000000000ucre: invalid coins",
		},
		{
			"negative input",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.Input = sdk.DecCoin{Denom: "ucre", Amount: sdk.NewDec(-1000000)}
			},
			"invalid input: decimal coin -1000000.000000000000000000ucre amount cannot be negative: invalid coins",
		},
		{
			"zero min output",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.MinOutput = utils.ParseDecCoin("0uusd")
			},
			"",
		},
		{
			"negative min output",
			func(msg *types.MsgSwapExactAmountIn) {
				msg.MinOutput = sdk.DecCoin{Denom: "uusd", Amount: sdk.NewDec(-5000000)}
			},
			"invalid min output: decimal coin -5000000.000000000000000000uusd amount cannot be negative: invalid coins",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			senderAddr := utils.TestAddress(1)
			msg := types.NewMsgSwapExactAmountIn(
				senderAddr, []uint64{1, 2, 3},
				utils.ParseDecCoin("1000000ucre"), utils.ParseDecCoin("5000000uusd"))
			require.NoError(t, msg.ValidateBasic())
			require.Equal(t, types.RouterKey, msg.Route())
			require.Equal(t, []sdk.AccAddress{senderAddr}, msg.GetSigners())
			require.Equal(t, types.TypeMsgSwapExactAmountIn, msg.Type())
			tc.malleate(msg)
			err := msg.ValidateBasic()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
