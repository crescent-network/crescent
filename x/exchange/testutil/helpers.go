package testutil

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
)

func PlaceSpotMarketOrder(
	t *testing.T, ctx sdk.Context, k keeper.Keeper,
	marketId string, ordererAddr sdk.AccAddress, isBuy bool, qty sdk.Int) (execQty, execQuote sdk.Int) {
	var err error
	execQty, execQuote, err = k.PlaceSpotMarketOrder(ctx, marketId, ordererAddr, isBuy, qty)
	require.NoError(t, err)
	return
}
