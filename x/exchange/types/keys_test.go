package types_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestMarketKey(t *testing.T) {
	require.Equal(t, []byte{0x62, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetMarketKey(1000000))
}

func TestMarketStateKey(t *testing.T) {
	require.Equal(t, []byte{0x63, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetMarketStateKey(1000000))
}

func TestMarketByDenomsIndexKey(t *testing.T) {
	key := types.GetMarketByDenomsIndexKey("ucre", "uusd")
	require.Equal(t, []byte{0x64, 0x4, 0x75, 0x63, 0x72, 0x65, 0x75, 0x75, 0x73, 0x64}, key)
	baseDenom, quoteDenom := types.ParseMarketByDenomsIndexKey(key)
	require.Equal(t, "ucre", baseDenom)
	require.Equal(t, "uusd", quoteDenom)
}

func TestOrderKey(t *testing.T) {
	require.Equal(t, []byte{0x65, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40}, types.GetOrderKey(1000000))
}

func TestOrderBookOrderIndexKey(t *testing.T) {
	key := types.GetOrderBookOrderIndexKey(1000000, true, types.MaxPrice, 10000000)
	require.Equal(t, []byte{
		0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x0, 0x1, 0x0, 0x20, 0xf5, 0x80, 0xff, 0xff,
		0xff, 0xff, 0xff, 0x67, 0x69, 0x80,
	}, key)
	orderId := types.ParseOrderIdFromOrderBookOrderIndexKey(key)
	require.EqualValues(t, 10000000, orderId)
	price := types.ParsePriceFromOrderBookOrderIndexKey(key)
	utils.AssertEqual(t, types.MaxPrice, price)
	prefix := types.GetOrderBookSideIteratorPrefix(1000000, true)
	require.True(t, bytes.HasPrefix(key, prefix))
	prefix = types.GetOrdersByMarketIteratorPrefix(1000000)
	require.True(t, bytes.HasPrefix(key, prefix))

	key = types.GetOrderBookOrderIndexKey(1000000, false, types.MinPrice, 10000000)
	require.Equal(t, []byte{
		0x66, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x1, 0x0, 0xff, 0xec, 0xc6, 0x20, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x98, 0x96, 0x80,
	}, key)
	orderId = types.ParseOrderIdFromOrderBookOrderIndexKey(key)
	require.EqualValues(t, 10000000, orderId)
	price = types.ParsePriceFromOrderBookOrderIndexKey(key)
	utils.AssertEqual(t, types.MinPrice, price)
	prefix = types.GetOrderBookSideIteratorPrefix(1000000, false)
	require.True(t, bytes.HasPrefix(key, prefix))
	prefix = types.GetOrdersByMarketIteratorPrefix(1000000)
	require.True(t, bytes.HasPrefix(key, prefix))

	key1 := types.GetOrderBookOrderIndexKey(1, true, utils.ParseDec("0.8"), 1)
	key2 := types.GetOrderBookOrderIndexKey(1, true, utils.ParseDec("0.9"), 2)
	key3 := types.GetOrderBookOrderIndexKey(1, true, utils.ParseDec("1.1"), 3)
	require.Equal(t, -1, bytes.Compare(key1, key2))
	require.Equal(t, -1, bytes.Compare(key2, key3))
}

func TestOrdersByOrdererIndexKey(t *testing.T) {
	ordererAddr := utils.TestAddress(1000000)
	key := types.GetOrdersByOrdererIndexKey(ordererAddr, 1000000, 10000000)
	require.Equal(t, []byte{
		0x67, 0x14, 0x80, 0x89, 0x7a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x98, 0x96, 0x80,
	}, key)
	orderId := types.ParseOrderIdFromOrdersByOrdererIndexKey(key)
	require.EqualValues(t, 10000000, orderId)
	prefix := types.GetOrdersByOrdererIteratorPrefix(ordererAddr)
	require.True(t, bytes.HasPrefix(key, prefix))
	prefix = types.GetOrdersByOrdererAndMarketIteratorPrefix(ordererAddr, 1000000)
	require.True(t, bytes.HasPrefix(key, prefix))
}

func TestNumMMOrdersKey(t *testing.T) {
	ordererAddr := utils.TestAddress(1000000)
	key := types.GetNumMMOrdersKey(ordererAddr, 1000000)
	require.Equal(t, []byte{
		0x68, 0x14, 0x80, 0x89, 0x7a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x42, 0x40,
	}, key)
	ordererAddr2, marketId := types.ParseNumMMOrdersKey(key)
	require.Equal(t, ordererAddr, ordererAddr2)
	require.EqualValues(t, 1000000, marketId)
}

func TestPriceToBytes(t *testing.T) {
	maxPrice := types.MaxPrice
	for price := types.MinPrice; price.LT(maxPrice); price = price.MulInt64(10) {
		bz := types.PriceToBytes(price)
		price2 := types.BytesToPrice(bz)
		require.Equal(t, price, price2)
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
