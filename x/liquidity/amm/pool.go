package amm

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
)

var (
	_ Pool = (*BasicPool)(nil)
	_ Pool = (*RangedPool)(nil)
)

// Pool is the interface of a pool.
type Pool interface {
	NewOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) Order
	Balances() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
	IsDepleted() bool

	HighestBuyPrice() (sdk.Dec, bool)
	LowestSellPrice() (sdk.Dec, bool)
	BuyAmountOver(price sdk.Dec, inclusive bool) sdk.Int
	SellAmountUnder(price sdk.Dec, inclusive bool) sdk.Int
	BuyAmountTo(price sdk.Dec) sdk.Int
	SellAmountTo(price sdk.Dec) sdk.Int
}

// BasicPool is the basic pool type.
type BasicPool struct {
	// rx and ry are the pool's reserve balance of each x/y coin.
	// In perspective of a pair, x coin is the quote coin and
	// y coin is the base coin.
	rx, ry sdk.Int
	// ps is the pool's pool coin supply.
	ps sdk.Int
}

// NewBasicPool returns a new BasicPool.
// It is OK to pass an empty sdk.Int to ps when ps is not going to be used.
func NewBasicPool(rx, ry, ps sdk.Int) *BasicPool {
	return &BasicPool{
		rx: rx,
		ry: ry,
		ps: ps,
	}
}

// NewOrder returns new pool order.
// Users should override this method to return proper order type.
func (pool *BasicPool) NewOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) Order {
	return NewBaseOrder(dir, price, amt)
}

// Balances returns the balances of the pool.
func (pool *BasicPool) Balances() (rx, ry sdk.Int) {
	return pool.rx, pool.ry
}

// PoolCoinSupply returns the pool coin supply.
func (pool *BasicPool) PoolCoinSupply() sdk.Int {
	return pool.ps
}

// Price returns the pool price.
func (pool *BasicPool) Price() sdk.Dec {
	if pool.rx.IsZero() || pool.ry.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return pool.rx.ToDec().Quo(pool.ry.ToDec())
}

// IsDepleted returns whether the pool is depleted or not.
func (pool *BasicPool) IsDepleted() bool {
	return pool.ps.IsZero() || pool.rx.IsZero() || pool.ry.IsZero()
}

// HighestBuyPrice returns the highest buy price of the pool.
func (pool *BasicPool) HighestBuyPrice() (price sdk.Dec, found bool) {
	// The highest buy price is actually a bit lower than pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

// LowestSellPrice returns the lowest sell price of the pool.
func (pool *BasicPool) LowestSellPrice() (price sdk.Dec, found bool) {
	// The lowest sell price is actually a bit higher than the pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

func (pool *BasicPool) BuyAmountOver(price sdk.Dec, _ bool) (amt sdk.Int) {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	utils.SafeMath(func() {
		amt = pool.rx.ToDec().QuoTruncate(price).Sub(pool.ry.ToDec()).TruncateInt()
		if amt.GT(MaxCoinAmount) {
			amt = MaxCoinAmount
		}
	}, func() {
		amt = MaxCoinAmount
	})
	return
}

func (pool *BasicPool) SellAmountUnder(price sdk.Dec, _ bool) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.ry.ToDec().Sub(pool.rx.ToDec().QuoRoundUp(price)).TruncateInt()
}

func (pool *BasicPool) BuyAmountTo(price sdk.Dec) sdk.Int {
	rxSqrt, err := pool.rx.ToDec().ApproxSqrt()
	if err != nil {
		panic(err) // TODO: prevent panic
	}
	rySqrt, err := pool.ry.ToDec().ApproxSqrt()
	if err != nil {
		panic(err)
	}
	priceSqrt, err := price.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	// TODO: optimize formula
	dx := pool.rx.ToDec().Sub(priceSqrt.Mul(rxSqrt.Mul(rySqrt))) // dx = rx - sqrt(P * rx * ry)
	// TODO: possible overflow?
	return dx.QuoTruncate(price).TruncateInt() // dy = dx / P
}

func (pool *BasicPool) SellAmountTo(price sdk.Dec) sdk.Int {
	rxSqrt, err := pool.rx.ToDec().ApproxSqrt()
	if err != nil {
		panic(err) // TODO: prevent panic
	}
	rySqrt, err := pool.ry.ToDec().ApproxSqrt()
	if err != nil {
		panic(err)
	}
	priceSqrt, err := price.ApproxSqrt()
	if err != nil {
		panic(err)
	}
	return pool.ry.ToDec().Sub(rxSqrt.Mul(rySqrt).Quo(priceSqrt)).TruncateInt() // dy = ry - sqrt(rx * ry / P)
}

type RangedPool struct {
	rx, ry         sdk.Int
	ps             sdk.Int
	transX, transY *sdk.Dec
}

// NewRangedPool returns a new RangedPool.
func NewRangedPool(rx, ry, ps sdk.Int, transX, transY *sdk.Dec) *RangedPool {
	return &RangedPool{
		rx:     rx,
		ry:     ry,
		ps:     ps,
		transX: transX,
		transY: transY,
	}
}

// CreateRangedPool creates new RangedPool from given inputs, while validating
// the inputs and using only needed amount of x/y coins(the rest should be refunded).
func CreateRangedPool(x, y sdk.Int, initialPrice sdk.Dec, minPrice, maxPrice *sdk.Dec) (pool *RangedPool, err error) {
	if !x.IsPositive() && !y.IsPositive() {
		return nil, fmt.Errorf("either x or y must be positive")
	}
	if !initialPrice.IsPositive() {
		return nil, fmt.Errorf("initial price must be positive: %s", initialPrice)
	}
	if minPrice == nil && maxPrice == nil {
		return nil, fmt.Errorf("min price and max price must not be nil at the same time")
	}
	if minPrice != nil && !minPrice.IsPositive() {
		return nil, fmt.Errorf("min price must be positive: %s", minPrice)
	}
	if maxPrice != nil && !maxPrice.IsPositive() {
		return nil, fmt.Errorf("max price must be positive: %s", maxPrice)
	}
	if minPrice != nil && maxPrice != nil && !maxPrice.GT(*minPrice) {
		return nil, fmt.Errorf("max price must be greater than min price")
	}

	switch {
	case minPrice != nil && initialPrice.LTE(*minPrice):
		// single y asset pool
		if y.IsZero() {
			return nil, fmt.Errorf("y coin amount must not be 0 for single asset pool")
		}
		ax, ay := sdk.ZeroInt(), y
		return NewRangedPool(ax, ay, InitialPoolCoinSupply(ax, ay), minPrice, maxPrice), nil
	case maxPrice != nil && initialPrice.GTE(*maxPrice):
		// single x asset pool
		if x.IsZero() {
			return nil, fmt.Errorf("x coin amount must not be 0 for single asset pool")
		}
		ax, ay := x, sdk.ZeroInt()
		return NewRangedPool(ax, ay, InitialPoolCoinSupply(ax, ay), minPrice, maxPrice), nil
	}

	m := sdk.ZeroDec()
	if minPrice != nil {
		m = *minPrice
	}

	var ax, ay sdk.Int
	if maxPrice != nil {
		l := *maxPrice
		r := l.Sub(initialPrice).Quo(l.Mul(initialPrice.Sub(m))) // r = (transY-p)/(transY*(p-transX)))

		ay = x.ToDec().Mul(r).Ceil().TruncateInt() // ay = ceil(x * r)
		if ay.GT(y) {
			ax = y.ToDec().Quo(r).Ceil().TruncateInt() // ax = ceil(y / r)
			ay = y
		} else {
			ax = x
		}
	} else {
		r := initialPrice.Sub(m) // r = p - transX

		ay = x.ToDec().QuoRoundUp(r).Ceil().TruncateInt() // ay = ceil(x / r)
		if ay.GT(y) {
			ax = y.ToDec().Mul(r).Ceil().TruncateInt() // ax = ceil(y * r)
			ay = y
		} else {
			ax = x
		}
	}

	return NewRangedPool(ax, ay, InitialPoolCoinSupply(ax, ay), minPrice, maxPrice), nil
}

// Balances returns the balances of the pool.
func (pool *RangedPool) Balances() (rx, ry sdk.Int) {
	return pool.rx, pool.ry
}

// PoolCoinSupply returns the pool coin supply.
func (pool *RangedPool) PoolCoinSupply() sdk.Int {
	return pool.ps
}

// Price returns the pool price.
func (pool *RangedPool) Price() sdk.Dec {
	if pool.rx.ToDec().Add(a).IsZero() || pool.ry.ToDec().Add(b).IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return pool.rx.ToDec().Add(a).Quo(pool.ry.ToDec().Add(b)) // (rx + a) / (ry + b)
}

// IsDepleted returns whether the pool is depleted or not.
func (pool *RangedPool) IsDepleted() bool {
	return pool.ps.IsZero() || pool.rx.ToDec().Add(a).IsZero() || pool.ry.ToDec().Add(b).IsZero()
}

// HighestBuyPrice returns the highest buy price of the pool.
func (pool *RangedPool) HighestBuyPrice() (price sdk.Dec, found bool) {
	// The highest buy price is actually a bit lower than pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

// LowestSellPrice returns the lowest sell price of the pool.
func (pool *RangedPool) LowestSellPrice() (price sdk.Dec, found bool) {
	// The lowest sell price is actually a bit higher than the pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

// BuyAmountOver returns the amount of buy orders for price greater than
// or equal to given price.
func (pool *RangedPool) BuyAmountOver(price sdk.Dec) (amt sdk.Int) {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.transX != nil && price.LT(*pool.transX) {
		price = *pool.transX
	}
	a, b := pool.ab()
	utils.SafeMath(func() {
		amt = pool.rx.ToDec().Add(a).QuoTruncate(price).Sub(pool.ry.ToDec().Add(b)).TruncateInt()
		if amt.GT(MaxCoinAmount) {
			amt = MaxCoinAmount
		} // It also satisfies OrderView interface.
	}, func() {
		amt = MaxCoinAmount
	})
	return
}

func (pool *RangedPool) ProvidableYAmountUnder(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.transY != nil && price.GT(*pool.transY) {
		price = *pool.transY
	}
	a, b := pool.ab()
	return pool.ry.ToDec().Add(b).Sub(pool.rx.ToDec().Add(a).QuoRoundUp(price)).TruncateInt()
}

// Deposit returns accepted x and y coin amount and minted pool coin amount
// when someone deposits x and y coins.
func Deposit(rx, ry, ps, x, y sdk.Int) (ax, ay, pc sdk.Int) {
	// Calculate accepted amount and minting amount.
	// Note that we take as many coins as possible(by ceiling numbers)
	// from depositor and mint as little coins as possible.

	utils.SafeMath(func() {
		rx, ry := rx.ToDec(), ry.ToDec()
		ps := ps.ToDec()

		// pc = floor(ps * min(x / rx, y / ry))
		pc = ps.MulTruncate(sdk.MinDec(
			x.ToDec().QuoTruncate(rx),
			y.ToDec().QuoTruncate(ry),
		)).TruncateInt()

		mintProportion := pc.ToDec().Quo(ps)             // pc / ps
		ax = rx.Mul(mintProportion).Ceil().TruncateInt() // ceil(rx * mintProportion)
		ay = ry.Mul(mintProportion).Ceil().TruncateInt() // ceil(ry * mintProportion)
	}, func() {
		ax, ay, pc = sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	})

	return
}

// Withdraw returns withdrawn x and y coin amount when someone withdraws
// pc pool coin.
// Withdraw also takes care of the fee rate.
func Withdraw(rx, ry, ps, pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	if pc.Equal(ps) {
		// Redeeming the last pool coin - give all remaining rx and ry.
		x = rx
		y = ry
		return
	}

	utils.SafeMath(func() {
		proportion := pc.ToDec().QuoTruncate(ps.ToDec())                             // pc / ps
		multiplier := sdk.OneDec().Sub(feeRate)                                      // 1 - feeRate
		x = rx.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(rx * proportion * multiplier)
		y = ry.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(ry * proportion * multiplier)
	}, func() {
		x, y = sdk.ZeroInt(), sdk.ZeroInt()
	})

	return
}

func PoolBuyOrders(pool Pool, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	tmpPool := pool
	var orders []Order
	placeOrder := func(price sdk.Dec, amt sdk.Int) {
		orders = append(orders, pool.NewOrder(Buy, price, amt))
		rx, ry := tmpPool.Balances()
		rx = rx.Sub(price.MulInt(amt).Ceil().TruncateInt()) // quote coin ceiling
		ry = ry.Add(amt)
		tmpPool = NewBasicPool(rx, ry, sdk.Int{})
	}
	if pool.Price().GT(highestPrice) {
		placeOrder(highestPrice, tmpPool.BuyAmountTo(highestPrice))
	}
	tick := PriceToDownTick(tmpPool.Price(), tickPrec)
	priceMultiplier := sdk.OneDec().Sub(PoolOrderPriceDiffRatio)
	for tick.GTE(lowestPrice) {
		amt := tmpPool.BuyAmountOver(tick, true)
		if amt.LT(MinCoinAmount) {
			tick = DownTick(tick, tickPrec) // TODO: check if the tick is the lowest possible tick
			continue
		}
		placeOrder(tick, amt)
		rx, _ := tmpPool.Balances()
		if !rx.IsPositive() {
			break
		}
		tick = PriceToDownTick(tick.Mul(priceMultiplier), tickPrec)
	}
	return orders
}

func PoolSellOrders(pool Pool, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	tmpPool := pool
	var orders []Order
	placeOrder := func(price sdk.Dec, amt sdk.Int) {
		orders = append(orders, pool.NewOrder(Sell, price, amt))
		rx, ry := tmpPool.Balances()
		rx = rx.Add(price.MulInt(amt).TruncateInt()) // quote coin truncation
		ry = ry.Sub(amt)
		tmpPool = NewBasicPool(rx, ry, sdk.Int{})
	}
	if pool.Price().LT(lowestPrice) {
		placeOrder(lowestPrice, tmpPool.SellAmountTo(lowestPrice))
	}
	tick := PriceToUpTick(tmpPool.Price(), tickPrec)
	priceMultiplier := sdk.OneDec().Add(PoolOrderPriceDiffRatio)
	for tick.LTE(highestPrice) {
		amt := tmpPool.SellAmountUnder(tick, true)
		if amt.LT(MinCoinAmount) || tick.MulInt(amt).TruncateInt().IsZero() {
			tick = UpTick(tick, tickPrec)
			continue
		}
		placeOrder(tick, amt)
		_, ry := tmpPool.Balances()
		if !ry.GT(MinCoinAmount) {
			break
		}
		tick = PriceToUpTick(tick.Mul(priceMultiplier), tickPrec)
	}
	return orders
}

func PoolOrders(pool Pool, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	return append(
		PoolBuyOrders(pool, lowestPrice, highestPrice, tickPrec),
		PoolSellOrders(pool, lowestPrice, highestPrice, tickPrec)...)
}

// InitialPoolCoinSupply returns ideal initial pool coin minting amount.
func InitialPoolCoinSupply(x, y sdk.Int) sdk.Int {
	cx := len(x.BigInt().Text(10)) - 1 // characteristic of x
	cy := len(y.BigInt().Text(10)) - 1 // characteristic of y
	c := ((cx + 1) + (cy + 1) + 1) / 2 // ceil(((cx + 1) + (cy + 1)) / 2)
	res := big.NewInt(10)
	res.Exp(res, big.NewInt(int64(c)), nil) // 10^c
	return sdk.NewIntFromBigInt(res)
}
