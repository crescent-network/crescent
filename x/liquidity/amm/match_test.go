package amm_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

func TestFindMatchPrice_Rounding(t *testing.T) {
	basePrice := squad.ParseDec("0.9990")

	for i := 0; i < 50; i++ {
		ob := amm.NewOrderBook(
			newOrder(amm.Buy, defTickPrec.UpTick(defTickPrec.UpTick(basePrice)), sdk.NewInt(80)),
			newOrder(amm.Sell, defTickPrec.UpTick(basePrice), sdk.NewInt(20)),
			newOrder(amm.Buy, basePrice, sdk.NewInt(10)), newOrder(amm.Sell, basePrice, sdk.NewInt(10)),
			newOrder(amm.Sell, defTickPrec.DownTick(basePrice), sdk.NewInt(70)),
		)
		matchPrice, found := amm.FindMatchPrice(ob, int(defTickPrec))
		require.True(t, found)
		require.True(sdk.DecEq(t,
			defTickPrec.RoundPrice(basePrice.Add(defTickPrec.UpTick(basePrice)).QuoInt64(2)),
			matchPrice))

		basePrice = defTickPrec.UpTick(basePrice)
	}
}
