package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ types.OrderSource = mockOrderSource{}

type mockOrderSource struct{ sourceName string }

func (os mockOrderSource) Name() string { return os.sourceName }

func (mockOrderSource) GenerateOrders(sdk.Context, types.Market, types.CreateOrderFunc, types.GenerateOrdersOptions) {
}

func (mockOrderSource) AfterOrdersExecuted(sdk.Context, types.Market, []types.MemOrder) {}

func newMemOrder(
	orderId uint64, market types.Market, isBuy bool,
	price, qty, openQty sdk.Dec, source types.OrderSource) *types.MemOrder {
	return types.NewMemOrder(
		types.NewOrder(orderId, types.OrderTypeLimit, utils.TestAddress(1), market.Id, isBuy,
			price, qty, 1, openQty, types.DepositAmount(isBuy, price, openQty),
			utils.ParseTime("2023-06-01T00:00:00Z")),
		market, source)
}

func TestMemOrderBookSide_AddOrder(t *testing.T) {
	market := types.NewMarket(
		1, "ucre", "uusd", types.DefaultFees.DefaultMakerFeeRate, types.DefaultFees.DefaultTakerFeeRate)

	// Buy order book side
	obs := types.NewMemOrderBookSide(true)
	require.Panics(t, func() {
		obs.AddOrder(newMemOrder(
			1, market, false, utils.ParseDec("12.3"), sdk.NewDec(100_000000), sdk.NewDec(80_000000), nil))
	})
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		tick := 9000 + r.Int31n(30)
		obs.AddOrder(newMemOrder(
			uint64(i+1), market, true, types.PriceAtTick(tick),
			sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil))
	}
	for i, level := range obs.Levels {
		if i+1 < len(obs.Levels) {
			require.True(t, level.Price.GT(obs.Levels[i+1].Price))
		}
		for _, order := range level.Orders {
			require.Equal(t, level.Price.String(), order.Price.String())
		}
	}

	// Sell order book side
	obs = types.NewMemOrderBookSide(false)
	require.Panics(t, func() {
		obs.AddOrder(newMemOrder(
			1, market, true, utils.ParseDec("12.3"), sdk.NewDec(100_000000), sdk.NewDec(80_000000), nil))
	})
	for i := 0; i < 100; i++ {
		tick := 9000 + r.Int31n(30)
		obs.AddOrder(newMemOrder(
			uint64(i+1), market, false, types.PriceAtTick(tick),
			sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil))
	}
	for i, level := range obs.Levels {
		if i+1 < len(obs.Levels) {
			require.True(t, level.Price.LT(obs.Levels[i+1].Price))
		}
		for _, order := range level.Orders {
			require.Equal(t, level.Price.String(), order.Price.String())
		}
	}
}

func TestNewMemOrder(t *testing.T) {
	market := types.NewMarket(
		1, "ucre", "uusd", types.DefaultFees.DefaultMakerFeeRate, types.DefaultFees.DefaultTakerFeeRate)
	order := newMemOrder(1, market, true, utils.ParseDec("5.2"), sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	testutil.AssertEqual(t, utils.ParseDecCoin("0uusd"), order.Paid)
	require.Nil(t, order.Received)
	order = newMemOrder(1, market, false, utils.ParseDec("5.2"), sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	testutil.AssertEqual(t, utils.ParseDecCoin("0ucre"), order.Paid)
	require.Nil(t, order.Received)
}

func TestMemOrder_HasPriorityOver(t *testing.T) {
	market := types.NewMarket(
		1, "ucre", "uusd", types.DefaultFees.DefaultMakerFeeRate, types.DefaultFees.DefaultTakerFeeRate)
	source1 := mockOrderSource{"source1"}
	source2 := mockOrderSource{"source2"}
	price := utils.ParseDec("12.345")

	order1 := newMemOrder(1, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	order2 := newMemOrder(2, market, false, price, sdk.NewDec(110_000000), sdk.NewDec(80_000000), nil)
	// order2 quantity > order1 quantity
	require.True(t, order2.HasPriorityOver(order1))

	order1 = newMemOrder(1, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(80_000000), nil)
	order2 = newMemOrder(2, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	// lower order id has priority
	require.True(t, order1.HasPriorityOver(order2))

	order1 = newMemOrder(1, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	order2 = newMemOrder(2, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(80_000000), source1)
	// user orders has priority over orders from OrderSource
	require.True(t, order1.HasPriorityOver(order2))

	order1 = newMemOrder(1, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000), nil)
	order2 = newMemOrder(2, market, false, price, sdk.NewDec(110_000000), sdk.NewDec(80_000000), source1)
	// but if the quantity of the order from OrderSource is higher, it takes priority
	require.True(t, order2.HasPriorityOver(order1))

	order1 = newMemOrder(1, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(80_000000), source1)
	order2 = newMemOrder(2, market, false, price, sdk.NewDec(100_000000), sdk.NewDec(90_000000), source2)
	// lexicographical source name priority
	require.True(t, order1.HasPriorityOver(order2))
}

func TestGroupMemOrdersByMsgHeight(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	market := types.NewMarket(
		1, "ucre", "uusd", types.DefaultFees.DefaultMakerFeeRate, types.DefaultFees.DefaultTakerFeeRate)
	deadline := utils.ParseTime("2023-06-01T00:00:00Z")
	for i := 0; i < 100; i++ {
		var orders []*types.MemOrder
		hasZeroHeight := false
		for j := 0; j < 100; j++ {
			price := utils.RandomDec(r, utils.ParseDec("9"), utils.ParseDec("11"))
			qty := utils.RandomDec(r, sdk.NewDec(100_000000), sdk.NewDec(1000_000000))
			deposit := types.DepositAmount(true, price, qty)
			var msgHeight int64
			if r.Float64() <= 0.3 { // 30% chance
				msgHeight = 0
				hasZeroHeight = true
			} else {
				msgHeight = int64(r.Intn(20))
			}
			order := types.NewOrder(
				uint64(j+1), types.OrderTypeLimit, utils.TestAddress(1), market.Id,
				true, price, qty, msgHeight, qty, deposit, deadline)
			orders = append(orders, types.NewMemOrder(order, market, nil))
		}
		groups := types.GroupMemOrdersByMsgHeight(orders)
		require.NotEmpty(t, groups)
		if hasZeroHeight {
			require.EqualValues(t, 0, groups[len(groups)-1].MsgHeight)
		}
		for j := 0; j < len(groups); j++ {
			if j+1 < len(groups) {
				if groups[j+1].MsgHeight == 0 {
					require.EqualValues(t, len(groups)-1, j+1)
				} else {
					require.Less(t, groups[j].MsgHeight, groups[j+1].MsgHeight)
				}
			}
			for _, order := range groups[j].Orders {
				require.EqualValues(t, groups[j].MsgHeight, order.MsgHeight)
			}
		}
	}
}
