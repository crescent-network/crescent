package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func newUserMemOrder(
	orderId uint64, isBuy bool, price, qty, openQty sdk.Dec) *types.MemOrder {
	return types.NewUserMemOrder(
		types.NewOrder(orderId, types.OrderTypeLimit, utils.TestAddress(1), 1, isBuy,
			price, qty, 1, openQty, types.DepositAmount(isBuy, price, openQty),
			utils.ParseTime("2023-06-01T00:00:00Z")))
}

func newOrderSourceMemOrder(
	isBuy bool, price, qty sdk.Dec, source types.OrderSource) *types.MemOrder {
	return types.NewOrderSourceMemOrder(
		utils.TestAddress(1), isBuy, price, qty, source)
}

func TestMemOrderBookSide_AddOrder(t *testing.T) {
	// Buy order book side
	obs := types.NewMemOrderBookSide(true)
	require.Panics(t, func() {
		obs.AddOrder(newUserMemOrder(
			1, false, utils.ParseDec("12.3"), sdk.NewDec(100_000000), sdk.NewDec(80_000000)))
	})
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		tick := 9000 + r.Int31n(30)
		obs.AddOrder(newUserMemOrder(
			uint64(i+1), true, types.PriceAtTick(tick),
			sdk.NewDec(100_000000), sdk.NewDec(90_000000)))
	}
	for i, level := range obs.Levels() {
		if i+1 < len(obs.Levels()) {
			require.True(t, level.Price().GT(obs.Levels()[i+1].Price()))
		}
		for _, order := range level.Orders() {
			require.Equal(t, level.Price().String(), order.Price().String())
		}
	}

	// Sell order book side
	obs = types.NewMemOrderBookSide(false)
	require.Panics(t, func() {
		obs.AddOrder(newUserMemOrder(
			1, true, utils.ParseDec("12.3"), sdk.NewDec(100_000000), sdk.NewDec(80_000000)))
	})
	for i := 0; i < 100; i++ {
		tick := 9000 + r.Int31n(30)
		obs.AddOrder(newUserMemOrder(
			uint64(i+1), false, types.PriceAtTick(tick),
			sdk.NewDec(100_000000), sdk.NewDec(90_000000)))
	}
	for i, level := range obs.Levels() {
		if i+1 < len(obs.Levels()) {
			require.True(t, level.Price().LT(obs.Levels()[i+1].Price()))
		}
		for _, order := range level.Orders() {
			require.Equal(t, level.Price().String(), order.Price().String())
		}
	}
}

func TestMemOrder_HasPriorityOver(t *testing.T) {
	source1 := types.NewMockOrderSource("source1")
	source2 := types.NewMockOrderSource("source2")
	price := utils.ParseDec("12.345")

	order1 := newUserMemOrder(1, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000))
	order2 := newUserMemOrder(2, false, price, sdk.NewDec(110_000000), sdk.NewDec(80_000000))
	// order2 quantity > order1 quantity
	require.True(t, order2.HasPriorityOver(order1))

	order1 = newUserMemOrder(1, false, price, sdk.NewDec(100_000000), sdk.NewDec(80_000000))
	order2 = newUserMemOrder(2, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000))
	// lower order id has priority
	require.True(t, order1.HasPriorityOver(order2))

	order1 = newUserMemOrder(1, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000))
	order2 = newOrderSourceMemOrder(false, price, sdk.NewDec(100_000000), source1)
	// OrderSourceMemOrder has priority over UserMemOrder
	require.True(t, order2.HasPriorityOver(order1))

	order1 = newUserMemOrder(1, false, price, sdk.NewDec(110_000000), sdk.NewDec(90_000000))
	order2 = newOrderSourceMemOrder(false, price, sdk.NewDec(100_000000), source1)
	// but if the quantity of UserMemOrder is higher, it takes priority
	require.True(t, order1.HasPriorityOver(order2))

	order1 = newOrderSourceMemOrder(false, price, sdk.NewDec(100_000000), source1)
	order2 = newOrderSourceMemOrder(false, price, sdk.NewDec(100_000000), source2)
	// lexicographical source name priority
	require.True(t, order1.HasPriorityOver(order2))
}

func TestGroupMemOrdersByMsgHeight(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	market := types.NewMarket(
		1, "ucre", "uusd", types.DefaultFees.DefaultMakerFeeRate, types.DefaultFees.DefaultTakerFeeRate)
	deadline := utils.ParseTime("2023-06-01T00:00:00Z")
	orderSourceOrdererAddr := utils.TestAddress(100)
	source := types.NewMockOrderSource("source")
	for i := 0; i < 100; i++ {
		var orders []*types.MemOrder
		hasOrderSourceOrders := false
		for j := 0; j < 100; j++ {
			price := utils.RandomDec(r, utils.ParseDec("9"), utils.ParseDec("11"))
			qty := utils.RandomDec(r, sdk.NewDec(100_000000), sdk.NewDec(1000_000000))
			deposit := types.DepositAmount(true, price, qty)
			if r.Float64() <= 0.3 { // 30% chance
				hasOrderSourceOrders = true
				orders = append(orders, types.NewOrderSourceMemOrder(
					orderSourceOrdererAddr, true, price, qty, source))
			} else {
				msgHeight := int64(r.Intn(20))
				order := types.NewOrder(
					uint64(j+1), types.OrderTypeLimit, utils.TestAddress(1), market.Id,
					true, price, qty, msgHeight, qty, deposit, deadline)
				orders = append(orders, types.NewUserMemOrder(order))
			}
		}
		groups := types.GroupMemOrdersByMsgHeight(orders)
		require.NotEmpty(t, groups)
		if hasOrderSourceOrders {
			require.EqualValues(t, -1, groups[0].MsgHeight())
		}
		for j := 0; j < len(groups); j++ {
			if j+1 < len(groups) {
				require.Less(t, groups[j].MsgHeight(), groups[j+1].MsgHeight())
			}
			for _, order := range groups[j].Orders() {
				if order.Type() == types.UserMemOrder {
					require.EqualValues(t, groups[j].MsgHeight(), order.Order().MsgHeight)
				} else {
					require.EqualValues(t, -1, groups[j].MsgHeight())
				}
			}
		}
	}
}
