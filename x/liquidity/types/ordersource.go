package types

import (
	"fmt"
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
		amount = amount.Add(ticks.Ticks[i].Orders.OpenAmount())
	}
	return amount
}

func (ticks *OrderBookTicks) AmountLTE(price sdk.Dec) sdk.Int {
	i, _ := ticks.FindPrice(price)
	amount := sdk.ZeroInt()
	for ; i < len(ticks.Ticks); i++ {
		amount = amount.Add(ticks.Ticks[i].Orders.OpenAmount())
	}
	return amount
}

func (ticks *OrderBookTicks) Orders(price sdk.Dec) Orders {
	i, exact := ticks.FindPrice(price)
	if !exact {
		return nil
	}
	return ticks.Ticks[i].Orders
}

func (ticks *OrderBookTicks) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	tick = UpTick(price, ticks.TickPrecision)
	found = true
	return
}

func (ticks *OrderBookTicks) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
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

func (ticks *OrderBookTicks) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, _ := ticks.FindPrice(price)
	if i == 0 {
		return
	}
	for i--; i >= 0; i-- {
		if ticks.Ticks[i].Orders.OpenAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks *OrderBookTicks) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	i, exact := ticks.FindPrice(price)
	if !exact {
		i--
	}
	if i >= len(ticks.Ticks)-1 {
		return
	}
	for i++; i < len(ticks.Ticks); i++ {
		if ticks.Ticks[i].Orders.OpenAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks *OrderBookTicks) HighestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	for i := range ticks.Ticks {
		if ticks.Ticks[i].Orders.OpenAmount().IsPositive() {
			return ticks.Ticks[i].Price, true
		}
	}
	return
}

func (ticks *OrderBookTicks) LowestTick() (tick sdk.Dec, found bool) {
	if len(ticks.Ticks) == 0 {
		return
	}
	for i := len(ticks.Ticks) - 1; i >= 0; i-- {
		if ticks.Ticks[i].Orders.OpenAmount().IsPositive() {
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

type AmountCache struct {
	cache map[string]sdk.Dec
}

func NewAmountCache() *AmountCache {
	return &AmountCache{
		cache: map[string]sdk.Dec{},
	}
}

func (c *AmountCache) SetOrGet(price sdk.Dec, cb func() sdk.Dec) sdk.Dec {
	priceStr := price.String()
	amt, ok := c.cache[priceStr]
	if ok {
		return amt
	}
	amt = cb()
	c.cache[priceStr] = amt
	return amt
}

type PoolOrderSource struct {
	PoolId                     uint64
	ReserveAddress             sdk.AccAddress
	RX, RY                     sdk.Int
	PoolPrice                  sdk.Dec
	Direction                  SwapDirection
	TickPrecision              int
	highestTickWithOrders      sdk.Dec
	foundHighestTickWithOrders *bool
	lowestTickWithOrders       sdk.Dec
	foundLowestTickWithOrders  *bool
	buyAmtCache                *AmountCache
	sellAmtCache               *AmountCache
}

func NewPoolOrderSource(pool PoolI, poolId uint64, reserveAddr sdk.AccAddress, dir SwapDirection, prec int) *PoolOrderSource {
	rx, ry := pool.Balance()
	return &PoolOrderSource{
		PoolId:         poolId,
		ReserveAddress: reserveAddr,
		RX:             rx,
		RY:             ry,
		PoolPrice:      pool.Price(),
		Direction:      dir,
		TickPrecision:  prec,
		buyAmtCache:    NewAmountCache(),
		sellAmtCache:   NewAmountCache(),
	}
}

func (os *PoolOrderSource) ProvidableXAmountGTE(price sdk.Dec) sdk.Dec {
	if os.Direction == SwapDirectionSell {
		return sdk.ZeroDec()
	}
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroDec()
	}
	return os.RX.ToDec().Sub(price.Mul(os.RY.ToDec())) // TODO: use MulTruncate?
}

func (os *PoolOrderSource) BuyAmountGTE(price sdk.Dec) sdk.Dec {
	if os.Direction == SwapDirectionSell {
		return sdk.ZeroDec()
	}
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroDec()
	}
	return os.ProvidableXAmountGTE(price).QuoTruncate(price)
}

func (os *PoolOrderSource) SellAmountLTE(price sdk.Dec) sdk.Dec {
	if os.Direction == SwapDirectionBuy {
		return sdk.ZeroDec()
	}
	if price.LT(os.PoolPrice) {
		return sdk.ZeroDec()
	}
	return os.RY.ToDec().Sub(os.RX.ToDec().QuoRoundUp(price))
}

func (os *PoolOrderSource) BuyAmountOnTick(price sdk.Dec) sdk.Int {
	if os.Direction == SwapDirectionSell {
		return sdk.ZeroInt()
	}
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	upTick := UpTick(price, os.TickPrecision)
	upAmt := os.buyAmtCache.SetOrGet(upTick, func() sdk.Dec { return os.ProvidableXAmountGTE(upTick) })
	amt := os.buyAmtCache.SetOrGet(price, func() sdk.Dec { return os.ProvidableXAmountGTE(price) })
	return amt.Sub(upAmt).QuoTruncate(price).TruncateInt()
}

func (os *PoolOrderSource) SellAmountOnTick(price sdk.Dec) sdk.Int {
	if os.Direction == SwapDirectionBuy {
		return sdk.ZeroInt()
	}
	if price.LTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	downTick := DownTick(price, os.TickPrecision)
	downAmt := os.sellAmtCache.SetOrGet(downTick, func() sdk.Dec { return os.SellAmountLTE(downTick) })
	amt := os.sellAmtCache.SetOrGet(price, func() sdk.Dec { return os.SellAmountLTE(price) })
	return amt.Sub(downAmt).TruncateInt()
}

func (os *PoolOrderSource) AmountGTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	var found bool
	switch os.Direction {
	case SwapDirectionBuy:
		for tick := price; tick.LT(os.PoolPrice); {
			amt = amt.Add(os.BuyAmountOnTick(tick))
			tick, found = os.UpTickWithOrders(tick)
			if !found {
				break
			}
		}
	case SwapDirectionSell:
		for tick := price; ; {
			// If price <= poolPrice, then sell amount at price would be 0,
			// so it'll leave the result amount unchanged.
			// After that, price would become one tick higher than poolPrice,
			// and the calculation will be continued until there's no more
			// ticks left.
			// We could do an additional optimization that checks
			// if price <= poolPrice, but SellAmountOnTick is cached anyway
			// and doing such optimization doesn't have much benefit.
			// Same applies to the buy side of AmountLTE.
			amt = amt.Add(os.SellAmountOnTick(tick))
			tick, found = os.UpTickWithOrders(tick)
			if !found {
				break
			}
		}
	}
	return amt
}

func (os *PoolOrderSource) AmountLTE(price sdk.Dec) sdk.Int {
	amt := sdk.ZeroInt()
	var found bool
	switch os.Direction {
	case SwapDirectionBuy:
		for tick := price; ; {
			amt = amt.Add(os.BuyAmountOnTick(tick))
			tick, found = os.DownTickWithOrders(tick)
			if !found {
				break
			}
		}
	case SwapDirectionSell:
		for tick := price; tick.GT(os.PoolPrice); {
			amt = amt.Add(os.SellAmountOnTick(tick))
			tick, found = os.DownTickWithOrders(tick)
			if !found {
				break
			}
		}
	}
	return amt
}

func (os *PoolOrderSource) Orders(price sdk.Dec) Orders {
	switch os.Direction {
	case SwapDirectionBuy:
		return Orders{NewPoolOrder(os.PoolId, os.ReserveAddress, SwapDirectionBuy, price, os.BuyAmountOnTick(price))}
	case SwapDirectionSell:
		return Orders{NewPoolOrder(os.PoolId, os.ReserveAddress, SwapDirectionSell, price, os.SellAmountOnTick(price))}
	}
	return nil
}

func (os *PoolOrderSource) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = UpTick(price, os.TickPrecision)
		return tick, tick.LT(os.PoolPrice)
	case SwapDirectionSell:
		tick = UpTick(price, os.TickPrecision)
		if tick.LTE(os.PoolPrice) {
			return tick, true
		}
		return tick, tick.LTE(os.highestTick())
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
}

func (os *PoolOrderSource) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = DownTick(price, os.TickPrecision)
		if tick.GTE(os.PoolPrice) {
			return tick, true
		}
		return tick, tick.GTE(os.lowestTick())
	case SwapDirectionSell:
		tick = DownTick(price, os.TickPrecision)
		return tick, tick.GT(os.PoolPrice)
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
}

func (os *PoolOrderSource) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	var fn func(sdk.Dec) sdk.Int
	switch os.Direction {
	case SwapDirectionBuy:
		fn = os.BuyAmountOnTick
	case SwapDirectionSell:
		fn = os.SellAmountOnTick
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
	tick = price
	for {
		tick, found = os.UpTick(tick)
		if !found {
			break
		}
		if fn(tick).IsPositive() {
			return tick, true
		}
	}
	return sdk.Dec{}, false
}

func (os *PoolOrderSource) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	var fn func(sdk.Dec) sdk.Int
	switch os.Direction {
	case SwapDirectionBuy:
		fn = os.BuyAmountOnTick
	case SwapDirectionSell:
		fn = os.SellAmountOnTick
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
	tick = price
	for {
		tick, found = os.DownTick(tick)
		if !found {
			break
		}
		if fn(tick).IsPositive() {
			return tick, true
		}
	}
	return sdk.Dec{}, false
}

func (os *PoolOrderSource) highestTick() sdk.Dec {
	switch os.Direction {
	case SwapDirectionBuy:
		return DownTick(os.PoolPrice, os.TickPrecision)
	case SwapDirectionSell:
		return UpTick(os.RX.ToDec().Quo(MinCoinAmount.ToDec()), os.TickPrecision)
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
}

func (os *PoolOrderSource) lowestTick() sdk.Dec {
	switch os.Direction {
	case SwapDirectionBuy:
		return DownTick(MinCoinAmount.ToDec().Quo(os.RY.ToDec()), os.TickPrecision)
	case SwapDirectionSell:
		return UpTick(os.PoolPrice, os.TickPrecision)
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
}

func (os *PoolOrderSource) HighestTick() (tick sdk.Dec, found bool) {
	if os.foundHighestTickWithOrders != nil {
		return os.highestTickWithOrders, *os.foundHighestTickWithOrders
	}
	switch os.Direction {
	case SwapDirectionBuy:
		tick = DownTick(os.PoolPrice, os.TickPrecision)
		lowest := os.lowestTick()
		for ; tick.GTE(lowest); tick = DownTick(tick, os.TickPrecision) {
			if os.BuyAmountOnTick(tick).IsPositive() {
				found = true
				break
			}
		}
	case SwapDirectionSell:
		tick = os.highestTick()
		for ; tick.GT(os.PoolPrice); tick = DownTick(tick, os.TickPrecision) {
			if os.SellAmountOnTick(tick).IsPositive() {
				found = true
				break
			}
		}
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
	os.highestTickWithOrders = tick
	os.foundHighestTickWithOrders = &found
	return
}

func (os *PoolOrderSource) LowestTick() (tick sdk.Dec, found bool) {
	if os.foundLowestTickWithOrders != nil {
		return os.lowestTickWithOrders, *os.foundLowestTickWithOrders
	}
	switch os.Direction {
	case SwapDirectionBuy:
		tick = os.lowestTick()
		for ; tick.LT(os.PoolPrice); tick = UpTick(tick, os.TickPrecision) {
			if os.BuyAmountOnTick(tick).IsPositive() {
				found = true
				break
			}
		}
	case SwapDirectionSell:
		tick = UpTick(os.PoolPrice, os.TickPrecision)
		highest := os.highestTick()
		for ; tick.LTE(highest); tick = UpTick(tick, os.TickPrecision) {
			if os.SellAmountOnTick(tick).IsPositive() {
				found = true
				break
			}
		}
	default: // never happens
		panic(fmt.Sprintf("invalid direction: %s", os.Direction))
	}
	os.lowestTickWithOrders = tick
	os.foundLowestTickWithOrders = &found
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
