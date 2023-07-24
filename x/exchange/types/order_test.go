package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestOrder_Validate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(order *types.Order)
		expectedErr string
	}{
		{
			"happy case",
			func(order *types.Order) {},
			"",
		},
		{
			"invalid order id",
			func(order *types.Order) {
				order.Id = 0
			},
			"id must not be 0",
		},
		{
			"invalid order type",
			func(order *types.Order) {
				order.Type = 100
			},
			"invalid order type: 100",
		},
		{
			"invalid orderer",
			func(order *types.Order) {
				order.Orderer = "invalidaddr"
			},
			"invalid orderer address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid market id",
			func(order *types.Order) {
				order.MarketId = 0
			},
			"market id must not be 0",
		},
		{
			"zero price",
			func(order *types.Order) {
				order.Price = sdk.ZeroDec()
			},
			"price must be positive: 0.000000000000000000",
		},
		{
			"negative price",
			func(order *types.Order) {
				order.Price = utils.ParseDec("-1")
			},
			"price must be positive: -1.000000000000000000",
		},
		{
			"invalid tick price",
			func(order *types.Order) {
				order.Price = utils.ParseDec("1.23456789")
			},
			"invalid tick price: 1.234567890000000000",
		},
		{
			"zero quantity",
			func(order *types.Order) {
				order.Quantity = sdk.ZeroDec()
			},
			"quantity must be positive: 0.000000000000000000",
		},
		{
			"negative quantity",
			func(order *types.Order) {
				order.Quantity = sdk.NewDec(-100_000000)
			},
			"quantity must be positive: -100000000.000000000000000000",
		},
		{
			"zero open quantity",
			func(order *types.Order) {
				order.OpenQuantity = sdk.ZeroDec()
			},
			"",
		},
		{
			"negative open quantity",
			func(order *types.Order) {
				order.OpenQuantity = sdk.NewDec(-100_000000)
			},
			"open quantity must not be negative: -100000000.000000000000000000",
		},
		{
			"open quantity > quantity",
			func(order *types.Order) {
				order.Quantity = sdk.NewDec(100_000000)
				order.OpenQuantity = sdk.NewDec(200_000000)
			},
			"open quantity must be smaller than quantity: 200000000.000000000000000000 > 100000000.000000000000000000",
		},
		{
			"zero remaining deposit",
			func(order *types.Order) {
				order.RemainingDeposit = sdk.ZeroDec()
			},
			"remaining deposit must be positive: 0.000000000000000000",
		},
		{
			"negative remaining deposit",
			func(order *types.Order) {
				order.RemainingDeposit = sdk.NewDec(-50_000000)
			},
			"remaining deposit must be positive: -50000000.000000000000000000",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			order := types.NewOrder(
				1, types.OrderTypeLimit, utils.TestAddress(1), 1, false,
				utils.ParseDec("2"), sdk.NewDec(100_000000), 100,
				sdk.NewDec(50_000000), sdk.NewDec(50_000000),
				utils.ParseTime("2023-06-01T00:00:00Z"))
			tc.malleate(&order)
			err := order.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestOrder_ExecutableQuantity(t *testing.T) {
	for i, tc := range []struct {
		isBuy            bool
		openQty          sdk.Dec
		remainingDeposit sdk.Dec
		executableQty    sdk.Dec
	}{
		{
			isBuy:            true,
			openQty:          sdk.NewDec(100_000000),
			remainingDeposit: sdk.NewDec(123_450000),
			executableQty:    sdk.NewDec(100_000000),
		},
		{
			isBuy:            true,
			openQty:          sdk.NewDec(100_000000),
			remainingDeposit: sdk.NewDec(100_000000),
			executableQty:    utils.ParseDec("81004455.245038477116241393"),
		},
		{
			isBuy:            true,
			openQty:          sdk.NewDec(50_000000),
			remainingDeposit: sdk.NewDec(100_000000),
			executableQty:    sdk.NewDec(50_000000),
		},
		{
			isBuy:            false,
			openQty:          sdk.NewDec(100_000000),
			remainingDeposit: sdk.NewDec(100_000000),
			executableQty:    sdk.NewDec(100_000000),
		},
		{
			isBuy:            false,
			openQty:          sdk.NewDec(90_000000),
			remainingDeposit: sdk.NewDec(100_000000),
			executableQty:    sdk.NewDec(90_000000),
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			order := types.NewOrder(
				1, types.OrderTypeLimit, utils.TestAddress(1), 1, tc.isBuy,
				utils.ParseDec("1.2345"), tc.openQty, 100,
				tc.openQty, tc.remainingDeposit, utils.ParseTime("2023-06-01T00:00:00Z"))
			executableQty := order.ExecutableQuantity()
			require.Equal(t, tc.executableQty, executableQty)
		})
	}
}
