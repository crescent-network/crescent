package types_test

//
//import (
//	"testing"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/stretchr/testify/require"
//	"github.com/tendermint/tendermint/crypto"
//
//	"github.com/crescent-network/crescent/x/liquidity/types"
//)
//
//var testOrderer = sdk.AccAddress(crypto.AddressHash([]byte("orderer")))
//
//func newBuyOrder(price string, amt int64) types.Order {
//	return types.Order{
//		Orderer:         testOrderer,
//		Direction:       types.SwapDirectionXToY,
//		Price:           sdk.MustNewDecFromStr(price),
//		RemainingAmount: sdk.NewInt(amt),
//		ReceivedAmount:  sdk.ZeroInt(),
//	}
//}
//
//func TestOrderBook_Add(t *testing.T) {
//	var ob types.OrderBook
//	// Only calling `Add` will not modify the order book itself.
//	// To modify the order book, caller should assign it to the old
//	// variable, just like slices.
//	for i := 0; i < 10; i++ {
//		ob.Add(newBuyOrder("1", 1))
//	}
//	require.Len(t, ob, 0)
//
//	// Doing this 10 times to demonstrate how orders are
//	// grouped together based on their price.
//	for i := 0; i < 10; i++ {
//		price := sdk.OneDec()
//		// Add orders for 10 different prices.
//		for j := 0; j < 10; j++ {
//			ob = ob.Add(newBuyOrder(price.String(), 100))
//			price = price.Add(sdk.MustNewDecFromStr("0.1"))
//		}
//	}
//	require.Len(t, ob, 10)
//
//	// See if the orders are well grouped and sorted in descending order.
//	price := sdk.OneDec()
//	for i := 0; i < 10; i++ {
//		og := ob[len(ob)-i-1]
//		require.True(sdk.DecEq(t, price, og.Price))
//		require.Len(t, og.XToYOrders, 10)
//		require.Len(t, og.YToXOrders, 0)
//		price = price.Add(sdk.MustNewDecFromStr("0.1"))
//	}
//}
//
//func TestOrders_Match(t *testing.T) {
//	// TODO: write test
//}
