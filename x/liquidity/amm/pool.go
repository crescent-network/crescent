package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Pool        = (*BasicPool)(nil)
	_ OrderSource = (*MockPoolOrderSource)(nil)
)

// Pool is the interface of a pool.
// It also satisfies OrderView interface.
type Pool interface {
	OrderView
	Balances() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
	IsDepleted() bool
	Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int)
	Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int)
}

// BasicPool is the basic pool type.
type BasicPool struct {
	rx, ry sdk.Int
	ps     sdk.Int
}

// NewBasicPool returns a new BasicPool.
func NewBasicPool(rx, ry, ps sdk.Int) *BasicPool {
	return &BasicPool{
		rx: rx,
		ry: ry,
		ps: ps,
	}
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

// Deposit returns accepted x and y coin amount and minted pool coin amount
// when someone deposits x and y coins.
func (pool *BasicPool) Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int) {
	// Calculate accepted amount and minting amount.
	// Note that we take as many coins as possible(by ceiling numbers)
	// from depositor and mint as little coins as possible.

	rx, ry := pool.rx.ToDec(), pool.ry.ToDec()
	ps := pool.ps.ToDec()

	// pc = floor(ps * min(x / rx, y / ry))
	pc = ps.MulTruncate(sdk.MinDec(
		x.ToDec().QuoTruncate(rx),
		y.ToDec().QuoTruncate(ry),
	)).TruncateInt()

	mintProportion := pc.ToDec().Quo(ps)             // pc / ps
	ax = rx.Mul(mintProportion).Ceil().TruncateInt() // ceil(rx * mintProportion)
	ay = ry.Mul(mintProportion).Ceil().TruncateInt() // ceil(ry * mintProportion)
	return
}

// Withdraw returns withdrawn x and y coin amount when someone withdraws
// pc pool coin.
// Withdraw also takes care of the fee rate.
func (pool *BasicPool) Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	if pc.Equal(pool.ps) {
		// Redeeming the last pool coin - give all remaining rx and ry.
		x = pool.rx
		y = pool.ry
		return
	}

	proportion := pc.ToDec().QuoTruncate(pool.ps.ToDec())                             // pc / ps
	multiplier := sdk.OneDec().Sub(feeRate)                                           // 1 - feeRate
	x = pool.rx.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(rx * proportion * multiplier)
	y = pool.ry.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(ry * proportion * multiplier)
	return
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

// BuyAmountOver returns the amount of buy orders for price greater or equal
// than given price.
func (pool *BasicPool) BuyAmountOver(price sdk.Dec) sdk.Int {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.rx.ToDec().QuoTruncate(price).Sub(pool.ry.ToDec()).TruncateInt()
}

// SellAmountUnder returns the amount of sell orders for price less or equal
// than given price.
func (pool *BasicPool) SellAmountUnder(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.ry.ToDec().Sub(pool.rx.ToDec().QuoRoundUp(price)).TruncateInt()
}

// PoolsOrderBook assumes that lastPrice is on ticks.
func PoolsOrderBook(pools []Pool, lastPrice sdk.Dec, numTicks, tickPrec int) *OrderBook {
	prec := TickPrecision(tickPrec)
	i := prec.TickToIndex(lastPrice)
	highestTick := prec.TickFromIndex(i + numTicks)
	lowestTick := prec.TickFromIndex(i - numTicks)
	ob := NewOrderBook()
	for _, pool := range pools {
		poolPrice := pool.Price()
		if poolPrice.GT(lowestTick) { // Buy orders
			startTick := sdk.MinDec(prec.DownTick(poolPrice), highestTick)
			accAmt := sdk.ZeroInt()
			for tick := startTick; tick.GTE(lowestTick); tick = prec.DownTick(tick) {
				amt := pool.BuyAmountOver(tick).Sub(accAmt)
				ob.Add(NewBaseOrder(Buy, tick, amt, sdk.Coin{}, "denom"))
				accAmt = accAmt.Add(amt)
			}
		}
		if poolPrice.LT(highestTick) { // Sell orders
			startTick := sdk.MaxDec(prec.UpTick(poolPrice), lowestTick)
			accAmt := sdk.ZeroInt()
			for tick := startTick; tick.LTE(highestTick); tick = prec.UpTick(tick) {
				amt := pool.SellAmountUnder(tick).Sub(accAmt)
				ob.Add(NewBaseOrder(Sell, tick, amt, sdk.Coin{}, "denom"))
				accAmt = accAmt.Add(amt)
			}
		}
	}
	return ob
}

// MockPoolOrderSource demonstrates how to implement a pool OrderSource.
type MockPoolOrderSource struct {
	Pool
	baseCoinDenom, quoteCoinDenom string
}

// NewMockPoolOrderSource returns a new MockPoolOrderSource for testing.
func NewMockPoolOrderSource(pool Pool, baseCoinDenom, quoteCoinDenom string) *MockPoolOrderSource {
	return &MockPoolOrderSource{
		Pool:           pool,
		baseCoinDenom:  baseCoinDenom,
		quoteCoinDenom: quoteCoinDenom,
	}
}

// BuyOrdersOver returns buy orders for price greater or equal than given price.
func (os *MockPoolOrderSource) BuyOrdersOver(price sdk.Dec) []Order {
	amt := os.BuyAmountOver(price)
	if amt.IsZero() {
		return nil
	}
	quoteCoin := sdk.NewCoin(os.quoteCoinDenom, OfferCoinAmount(Buy, price, amt))
	return []Order{NewBaseOrder(Buy, price, amt, quoteCoin, os.baseCoinDenom)}
}

// SellOrdersUnder returns sell orders for price less or equal than given price.
func (os *MockPoolOrderSource) SellOrdersUnder(price sdk.Dec) []Order {
	amt := os.SellAmountUnder(price)
	if amt.IsZero() {
		return nil
	}
	baseCoin := sdk.NewCoin(os.baseCoinDenom, amt)
	return []Order{NewBaseOrder(Sell, price, amt, baseCoin, os.quoteCoinDenom)}
}
