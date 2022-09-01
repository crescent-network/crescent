package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func TestMMOrderTicks(t *testing.T) {
	require.Equal(t,
		[]types.MMOrderTick{
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("105"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("104.45"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("103.89"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("103.34"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("102.78"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("102.23"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("101.67"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("101.12"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("100.56"), Amount: sdk.NewInt(100000)},
			{OfferCoinAmount: sdk.NewInt(100000), Price: utils.ParseDec("100"), Amount: sdk.NewInt(100000)},
		},
		types.MMOrderTicks(
			types.OrderDirectionSell, utils.ParseDec("100"), utils.ParseDec("105"),
			sdk.NewInt(1000000), types.DefaultMaxNumMarketMakingOrderTicks, 4),
	)

	require.Equal(t,
		[]types.MMOrderTick{
			{
				OfferCoinAmount: sdk.NewInt(5402),
				Price:           utils.ParseDec("100.02"),
				Amount:          sdk.NewInt(54),
			},
			{
				OfferCoinAmount: sdk.NewInt(5502),
				Price:           utils.ParseDec("100.03"),
				Amount:          sdk.NewInt(55),
			},
		},
		types.MMOrderTicks(
			types.OrderDirectionBuy, utils.ParseDec("100.02"), utils.ParseDec("100.03"),
			sdk.NewInt(109), types.DefaultMaxNumMarketMakingOrderTicks, 4),
	)
}
