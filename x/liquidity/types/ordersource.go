package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ OrderSource = (*OrderBookTicks)(nil)
	_ OrderSource = (*PoolOrderSource)(nil)
	_ OrderSource = (*MergedOrderSources)(nil)
)

// OrderSource defines a source of orders which can be an order book or
// a pool.
type OrderSource interface {
	AmountGTE(price sdk.Dec) sdk.Int
	AmountLTE(price sdk.Dec) sdk.Int
	Orders(price sdk.Dec) Orders
	UpTick(price sdk.Dec) (tick sdk.Dec, found bool)
	DownTick(price sdk.Dec) (tick sdk.Dec, found bool)
	UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool)
	DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool)
	HighestTick() (tick sdk.Dec, found bool)
	LowestTick() (tick sdk.Dec, found bool)
}

type OrderBookTicks struct {
	Ticks         []*OrderBookTick
	TickPrecision int
}

func NewOrderBookTicks(prec int) *OrderBookTicks {
	return &OrderBookTicks{
		TickPrecision: prec,
	}
}

func (ticks *OrderBookTicks) FindPrice(price sdk.Dec) (i int, exact bool) {
	i = sort.Search(len(ticks.Ticks), func(i int) bool {
		return ticks.Ticks[i].Price.LTE(price)
	})
	if i < len(ticks.Ticks) && ticks.Ticks[i].Price.Equal(price) {
		exact = true
	}
	return
}

func (ticks *OrderBookTicks) AddOrder(order Order) {
	i, exact := ticks.FindPrice(order.GetPrice())
	if exact {
		ticks.Ticks[i].Orders = append(ticks.Ticks[i].Orders, order)
	} else {
		if i < len(ticks.Ticks) {
			// Insert a new order book tick at index i.
			ticks.Ticks = append(ticks.Ticks[:i], append([]*OrderBookTick{NewOrderBookTick(order)}, ticks.Ticks[i:]...)...)
		} else {
			// Append a new order group at the end.
			ticks.Ticks = append(ticks.Ticks, NewOrderBookTick(order))
		}
	}
}

func (ticks *OrderBookTicks) AddOrders(orders ...Order) {
	for _, order := range orders {
		ticks.AddOrder(order)
	}
}

func (ticks *OrderBookTicks) AllOrders() []Order {
	var orders []Order
	for _, tick := range ticks.Ticks {
		orders = append(orders, tick.Orders...)
	}
	return orders
}

func (ticks *OrderBookTicks) AmountGTE(price sdk.Dec) sdk.Int {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	amount := sdk.ZeroInt()
	for ; i >= 0; i-- {
		amount = amount.Add(ticks.Ticks[i].Orders.OpenBaseCoinAmount())
	}
	return amount
}

func (ticks OrderBookTicks) AmountLTE(price sdk.Dec) sdk.Int {
	i, _ := ticks.FindPrice(price)
	amount := sdk.ZeroInt()
	for ; i < len(ticks.Ticks); i++ {
		amount = amount.Add(ticks.Ticks[i].Orders.OpenBaseCoinAmount())
	}
	return amount
}

func (ticks OrderBookTicks) Orders(price sdk.Dec) Orders {
	i, exact := ticks.FindPrice(price)
	if !exact {
		return nil
	}
	return ticks.Ticks[i].Orders
}

func (ticks OrderBookTicks) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	tick = UpTick(price, ticks.TickPrecision)
	found = true
	return
}

func (ticks OrderBookTicks) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	if i >= len(ticks.Ticks)-1 {
		return
	}
	tick = DownTick(price, ticks.TickPrecision)
	found = true
	return
}

func (ticks OrderBookTicks) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	for i--; i >= 0; i-- {
		if ticks.Ticks[i].Orders.OpenBaseCoinAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks OrderBookTicks) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	if i >= len(ticks.Ticks)-1 {
		return
	}
	for i++; i < len(ticks.Ticks); i++ {
		if ticks.Ticks[i].Orders.OpenBaseCoinAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks OrderBookTicks) HighestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	for i := range ticks.Ticks {
		if ticks.Ticks[i].Orders.OpenBaseCoinAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks OrderBookTicks) LowestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	for i := len(ticks.Ticks) - 1; i >= 0; i-- {
		if ticks.Ticks[i].Orders.OpenBaseCoinAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

type OrderBookTick struct {
	Price  sdk.Dec
	Orders Orders
}

func NewOrderBookTick(order Order) *OrderBookTick {
	return &OrderBookTick{
		Price:  order.GetPrice(),
		Orders: Orders{order},
	}
}

type PoolOrderSource struct {
	PoolId          uint64
	ReserveAddress  sdk.AccAddress
	RX, RY          sdk.Int
	PoolPrice       sdk.Dec
	Direction       SwapDirection
	TickPrecision   int
	buyAmountCache  map[string]sdk.Int // map(price => buyAmountOnTick)
	sellAmountCache map[string]sdk.Int // map(price => sellAmountOnTick)
}

func NewPoolOrderSource(pool PoolI, poolId uint64, reserveAddr sdk.AccAddress, dir SwapDirection, prec int) OrderSource {
	rx, ry := pool.Balance()
	return &PoolOrderSource{
		PoolId:          poolId,
		ReserveAddress:  reserveAddr,
		RX:              rx,
		RY:              ry,
		PoolPrice:       pool.Price(),
		Direction:       dir,
		TickPrecision:   prec,
		sellAmountCache: map[string]sdk.Int{},
		buyAmountCache:  map[string]sdk.Int{},
	}
}

func (os PoolOrderSource) BuyAmountOnTick(price sdk.Dec) sdk.Int {
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	priceStr := price.String()
	res, ok := os.buyAmountCache[priceStr]
	if !ok {
		upPrice := UpTick(price, os.TickPrecision)                             // P'
		res = upPrice.Quo(price).Sub(sdk.OneDec()).MulInt(os.RY).TruncateInt() // (P'/P - 1) * RY
		os.buyAmountCache[priceStr] = res
	}
	return res
}

func (os PoolOrderSource) SellAmountOnTick(price sdk.Dec) sdk.Int {
	if price.LTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	priceStr := price.String()
	res, ok := os.sellAmountCache[priceStr]
	if !ok {
		downPrice := DownTick(price, os.TickPrecision) // P'
		rx := os.RX.ToDec()
		res = rx.QuoTruncate(downPrice).Sub(rx.QuoTruncate(price)).TruncateInt() // RX/P' - RX/P
		os.sellAmountCache[priceStr] = res
	}
	return res
}

func (os PoolOrderSource) AmountGTE(price sdk.Dec) sdk.Int {
	amount := sdk.ZeroInt()
	var found bool
	switch os.Direction {
	case SwapDirectionBuy:
		for price.LT(os.PoolPrice) {
			ba := os.BuyAmountOnTick(price)
			amount = amount.Add(ba)
			price, found = os.UpTickWithOrders(price)
			if !found {
				break
			}
		}
	case SwapDirectionSell:
		for {
			// If price <= poolPrice, then sell amount at price would be 0,
			// so it'll leave the result amount unchanged.
			// After that, price would become one tick higher than poolPrice,
			// and the calculation will be continued until there's no more
			// ticks left.
			// We could do an additional optimization that checks
			// if price <= poolPrice, but SellAmountOnTick is cached anyway
			// and doing such optimization doesn't have much benefit.
			// Same applies to the buy side of AmountLTE.
			sa := os.SellAmountOnTick(price)
			amount = amount.Add(sa)
			price, found = os.UpTickWithOrders(price)
			if !found {
				break
			}
		}
	}
	return amount
}

func (os PoolOrderSource) AmountLTE(price sdk.Dec) sdk.Int {
	amount := sdk.ZeroInt()
	var found bool
	switch os.Direction {
	case SwapDirectionBuy:
		for {
			ba := os.BuyAmountOnTick(price)
			amount = amount.Add(ba)
			price, found = os.DownTickWithOrders(price)
			if !found {
				break
			}
		}
	case SwapDirectionSell:
		for price.GT(os.PoolPrice) {
			sa := os.SellAmountOnTick(price)
			amount = amount.Add(sa)
			price, found = os.DownTickWithOrders(price)
			if !found {
				break
			}
		}
	}
	return amount
}

func (os PoolOrderSource) Orders(price sdk.Dec) Orders {
	switch os.Direction {
	case SwapDirectionBuy:
		return Orders{NewPoolOrder(os.PoolId, os.ReserveAddress, SwapDirectionBuy, price, os.BuyAmountOnTick(price))}
	case SwapDirectionSell:
		return Orders{NewPoolOrder(os.PoolId, os.ReserveAddress, SwapDirectionSell, price, os.SellAmountOnTick(price))}
	}
	return nil
}

func (os PoolOrderSource) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = UpTick(price, os.TickPrecision)
		found = tick.LT(os.PoolPrice)
	case SwapDirectionSell:
		tick = UpTick(price, os.TickPrecision)
		if tick.GT(os.PoolPrice) {
			found = os.SellAmountOnTick(tick).IsPositive()
		} else {
			found = true
		}
	}
	return
}

func (os PoolOrderSource) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = DownTick(price, os.TickPrecision)
		if tick.LT(os.PoolPrice) {
			found = os.BuyAmountOnTick(tick).IsPositive()
		} else {
			found = true
		}
	case SwapDirectionSell:
		tick = DownTick(price, os.TickPrecision)
		found = tick.GT(os.PoolPrice)
	}
	return
}

func (os PoolOrderSource) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick, found = os.UpTick(price)
		if !found {
			break
		}
		for tick.LT(os.PoolPrice) {
			ba := os.BuyAmountOnTick(tick)
			if ba.IsPositive() {
				found = true
				break
			}
			tick, found = os.UpTick(tick)
			if !found {
				break
			}
		}
	case SwapDirectionSell:
		if price.LTE(os.PoolPrice) {
			return os.UpTick(os.PoolPrice)
		}
		return os.UpTick(price)
	}
	return
}

func (os PoolOrderSource) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		if price.GTE(os.PoolPrice) {
			return os.DownTick(os.PoolPrice)
		}
		return os.DownTick(price)
	case SwapDirectionSell:
		tick, found = os.DownTick(price)
		if !found {
			break
		}
		for tick.GT(os.PoolPrice) {
			sa := os.SellAmountOnTick(tick)
			if sa.IsPositive() {
				found = true
				break
			}
			tick, found = os.DownTick(tick)
			if !found {
				break
			}
		}
	}
	return
}

func (os PoolOrderSource) HighestTick() (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = PriceToTick(os.PoolPrice, os.TickPrecision)
		if os.PoolPrice.Equal(tick) {
			tick = DownTick(tick, os.TickPrecision)
		}
		found = true
	case SwapDirectionSell:
		// TODO: is it possible to calculate?
		panic("not implemented")
	}
	return
}

func (os PoolOrderSource) LowestTick() (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		// TODO: is it possible to calculate?
		panic("not implemented")
	case SwapDirectionSell:
		tick = UpTick(PriceToTick(os.PoolPrice, os.TickPrecision), os.TickPrecision)
		found = true
	}
	return
}

type MergedOrderSources struct {
	Sources []OrderSource
}

func MergeOrderSources(sources ...OrderSource) OrderSource {
	return &MergedOrderSources{Sources: sources}
}

func (os *MergedOrderSources) AmountGTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.Sources {
		amt = amt.Add(source.AmountGTE(price))
	}
	return amt
}

func (os *MergedOrderSources) AmountLTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	for _, source := range os.Sources {
		amt = amt.Add(source.AmountLTE(price))
	}
	return amt
}

func (os *MergedOrderSources) Orders(price sdk.Dec) Orders {
	var orders Orders
	for _, source := range os.Sources {
		orders = append(orders, source.Orders(price)...)
	}
	return orders
}

func (os *MergedOrderSources) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.UpTick(price)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.DownTick(price)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.UpTickWithOrders(price)
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.DownTickWithOrders(price)
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) HighestTick() (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.HighestTick()
		if f && (tick.IsNil() || t.GT(tick)) {
			tick = t
			found = true
		}
	}
	return
}

func (os *MergedOrderSources) LowestTick() (tick sdk.Dec, found bool) {
	for _, source := range os.Sources {
		t, f := source.LowestTick()
		if f && (tick.IsNil() || t.LT(tick)) {
			tick = t
			found = true
		}
	}
	return
}
