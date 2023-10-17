package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/cremath"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestNextOrderTick(t *testing.T) {
	for i, tc := range []struct {
		// args
		isBuy                      bool
		liquidity                  sdk.Int
		currentSqrtPrice           cremath.BigDec
		minOrderQty, minOrderQuote sdk.Int

		// result
		tick  int32
		valid bool
	}{
		{ // 0
			true, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(10000),
			sdk.NewInt(10000),
			-200, true,
		},
		{ // 1
			true, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			-200, true,
		},
		{ // 2
			true, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			-200, true,
		},
		{ // 3
			false, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(10000),
			sdk.NewInt(10000),
			30, true,
		},
		{ // 4
			false, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			20, true,
		},
		{ // 5
			false, sdk.NewInt(10000000), utils.ParseBigDec("1"),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			30, true,
		},
		{ // 6
			true, sdk.NewInt(2385592), utils.ParseBigDec("0.5").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			-50300, true,
		},
		{ // 7
			true, sdk.NewInt(2385592), utils.ParseBigDec("0.5").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			-50600, true,
		},
		{ // 8
			false, sdk.NewInt(2385592), utils.ParseBigDec("0.5").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			-49700, true,
		},
		{ // 9
			false, sdk.NewInt(2385592), utils.ParseBigDec("0.5").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			-49400, true,
		},
		{ // 10
			true, sdk.NewInt(2385592), utils.ParseBigDec("5").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			39080, true,
		},
		{ // 11
			true, sdk.NewInt(2385592), utils.ParseBigDec("5").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			39810, true,
		},
		{ // 12
			false, sdk.NewInt(2385592), utils.ParseBigDec("5").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			40960, true,
		},
		{ // 13
			false, sdk.NewInt(2385592), utils.ParseBigDec("5").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			40190, true,
		},
		{ // 14
			true, sdk.NewInt(1000), utils.ParseBigDec("100").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			0, false,
		},
		{ // 15
			true, sdk.NewInt(1000), utils.ParseBigDec("100").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			-9520, true,
		},
		{ // 16
			false, sdk.NewInt(1000), utils.ParseBigDec("0.01").SqrtMut(),
			sdk.NewInt(1),
			sdk.NewInt(10000),
			1060, true,
		},
		{ // 17
			false, sdk.NewInt(1000), utils.ParseBigDec("0.01").SqrtMut(),
			sdk.NewInt(10000),
			sdk.NewInt(1),
			0, false,
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tickSpacing := uint32(10)
			tick, valid := types.NextOrderTick(
				tc.isBuy, tc.liquidity, tc.currentSqrtPrice,
				tc.minOrderQty, tc.minOrderQuote, tickSpacing)
			require.Equal(t, tc.valid, valid)
			if tc.valid {
				require.Equal(t, tc.tick, tick)

				orderPrice := exchangetypes.PriceAtTick(tick)
				orderSqrtPrice := cremath.NewBigDecFromDec(orderPrice).SqrtMut()
				var qty sdk.Int
				if tc.isBuy {
					qty = cremath.NewBigDecFromInt(
						types.Amount1DeltaRounding(
							tc.currentSqrtPrice, orderSqrtPrice, tc.liquidity, false)).
						QuoTruncateMut(cremath.NewBigDecFromDec(orderPrice)).TruncateInt()
				} else {
					qty = types.Amount0DeltaRounding(
						tc.currentSqrtPrice, orderSqrtPrice, tc.liquidity, false)
				}
				require.True(
					t, qty.GTE(tc.minOrderQty), "expected %s >= %s", qty, tc.minOrderQty)
				quote := orderPrice.MulInt(qty).Ceil().TruncateInt()
				require.True(
					t, quote.GTE(tc.minOrderQuote), "expected %s >= %s", quote, tc.minOrderQuote)
			}
		})
	}
}
