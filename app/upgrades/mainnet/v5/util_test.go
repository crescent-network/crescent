package v5_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v5 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v5"
	utils "github.com/crescent-network/crescent/v5/types"
)

func TestAdjustPriceToTickSpacing(t *testing.T) {
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
			price := v5.AdjustPriceToTickSpacing(tc.price, tickSpacing, tc.roundUp)
			require.Equal(t, tc.expected.String(), price.String())
		})
	}
}
