package amm_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

func TestMatchableAmount(t *testing.T) {
	order1 := newOrder(amm.Buy, utils.ParseDec("1.0"), sdk.NewInt(10000))
	for _, tc := range []struct {
		order    amm.Order
		price    sdk.Dec
		expected sdk.Int
	}{
		{order1, utils.ParseDec("1"), sdk.NewInt(10000)},
		{order1, utils.ParseDec("0.01"), sdk.NewInt(10000)},
		{order1, utils.ParseDec("100"), sdk.NewInt(100)},
		{order1, utils.ParseDec("100.1"), sdk.NewInt(99)},
		{order1, utils.ParseDec("9999"), sdk.NewInt(1)},
		{order1, utils.ParseDec("10001"), sdk.NewInt(0)},
	} {
		t.Run("", func(t *testing.T) {
			require.True(sdk.IntEq(t, tc.expected, amm.MatchableAmount(tc.order, tc.price)))
		})
	}
}
