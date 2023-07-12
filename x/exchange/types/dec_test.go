package types_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestPriceToBytes(t *testing.T) {
	smallestDec := sdk.SmallestDec()
	maxPrice := types.MaxPrice
	for price := types.MinPrice; price.LT(maxPrice); price = price.MulInt64(10) {
		bz := types.PriceToBytes(price)
		price2 := types.BytesToPrice(bz)
		require.Equal(t, price, price2)
		lowerBz := types.PriceToBytes(price.Sub(smallestDec))
		upperBz := types.PriceToBytes(price.Add(smallestDec))
		require.Equal(t, -1, bytes.Compare(lowerBz, bz))
		require.Equal(t, -1, bytes.Compare(bz, upperBz))
	}
}

func BenchmarkPriceToBytes(b *testing.B) {
	// Maximum allowed by sdk.SortableDecBytes
	price := utils.ParseDec("1000000000000000000")
	b.Run("sdk.SortableDecBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sdk.SortableDecBytes(price)
		}
	})
	b.Run("PriceToBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			types.PriceToBytes(price)
		}
	})
}
