package testutil

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func PlaceSpotMarketOrder(
	t *testing.T, ctx sdk.Context, k keeper.Keeper,
	ordererAddr sdk.AccAddress, marketId string, isBuy bool, qty sdk.Int) (order types.SpotOrder, execQuote sdk.Int) {
	var err error
	order, execQuote, err = k.PlaceSpotMarketOrder(ctx, ordererAddr, marketId, isBuy, qty)
	require.NoError(t, err)
	return
}
