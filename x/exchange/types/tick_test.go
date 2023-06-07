package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestValidateTickPrice(t *testing.T) {
	for i, tc := range []struct {
		price sdk.Dec
		tick  int32
		valid bool
	}{
		{utils.ParseDec("1.0000"), 0, true},
		{utils.ParseDec("9.9999"), 89999, true},
		{utils.ParseDec("9.99999"), 89999, false},
		{utils.ParseDec("1.23456"), 2345, false},
		{utils.ParseDec("0.005"), -230000, true},
		{utils.ParseDec("0.0050001"), -229999, true},
		{utils.ParseDec("0.00500001"), -230000, false},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tick, valid := types.ValidateTickPrice(tc.price)
			require.Equal(t, tc.valid, valid)
			require.Equal(t, tc.tick, tick)
		})
	}
}

func TestPriceAtTick(t *testing.T) {
	for i, tc := range []struct {
		tick  int32
		price sdk.Dec
	}{
		{0, sdk.NewDec(1)},
		{2345, utils.ParseDec("1.2345")},
		{-230000, utils.ParseDec("0.005")},
		{-1000000, utils.ParseDec("0.000000000009000000")},
		{1000000, utils.ParseDec("200000000000")},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			price := types.PriceAtTick(tc.tick)
			require.Equal(t, tc.price.String(), price.String())
		})
	}
}

func TestTickAtPrice(t *testing.T) {
	for i, tc := range []struct {
		price sdk.Dec
		tick  int32
	}{
		{utils.ParseDec("1"), 0},
		{utils.ParseDec("1.0001"), 1},
		{utils.ParseDec("1.2345"), 2345},
		{utils.ParseDec("1.23456789"), 2345},
		{utils.ParseDec("12345"), 362345},
		{utils.ParseDec("0.000000000009000000"), -1000000},
		{utils.ParseDec("0.000000000009000001"), -1000000},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tick := types.TickAtPrice(tc.price)
			require.Equal(t, tc.tick, tick)
		})
	}
}

func TestRoundTick(t *testing.T) {
	for i, tc := range []struct {
		tick     int32
		expected int32
	}{
		{-5, -6},
		{-4, -4},
		{-3, -4},
		{-2, -2},
		{-1, -2},
		{0, 0},
		{1, 2},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 6},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			res := types.RoundTick(tc.tick)
			require.Equal(t, tc.expected, res)
		})
	}
}
