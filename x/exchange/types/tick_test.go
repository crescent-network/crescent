package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestMinMaxTick(t *testing.T) {
	for i, tc := range []struct {
		prec     int
		min, max int32
	}{
		{0, -126, 360},
		{1, -1260, 3600},
		{2, -12600, 36000},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			min, max := types.MinMaxTick(tc.prec)
			require.Equal(t, tc.min, min)
			require.Equal(t, tc.max, max)
		})
	}
}

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
			tick, valid := types.ValidateTickPrice(tc.price, 4)
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
			price := types.PriceAtTick(tc.tick, 4)
			require.Equal(t, tc.price.String(), price.String())
		})
	}
}

func TestTickBytes(t *testing.T) {
	for tick := int32(-100); tick <= 100; tick++ {
		bz := types.TickToBytes(tick)
		tick2 := types.BytesToTick(bz)
		require.Equal(t, tick, tick2)
	}
}
