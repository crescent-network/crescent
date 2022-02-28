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
	Price() sdk.Dec
	IsDepleted() bool
	Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int)
	Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int)
}

// BasicPool is the basic pool type.
type BasicPool struct {
	// rx and ry are the pool's reserve balance of each x/y coin.
	// In perspective of a pair, x coin is the quote coin and
	// y coin is the base coin.
	rx, ry sdk.Dec
	// ps is the pool's pool coin supply.
	ps sdk.Dec
}

// NewBasicPool returns a new BasicPool.
// It is OK to pass an empty sdk.Int to ps when ps is not going to be used.
func NewBasicPool(rx, ry, ps sdk.Int) *BasicPool {
	return &BasicPool{
		rx: rx.ToDec(),
		ry: ry.ToDec(),
		ps: ps.ToDec(),
	}
}

// Price returns the pool price.
func (pool *BasicPool) Price() sdk.Dec {
	if pool.rx.IsZero() || pool.ry.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return pool.rx.Quo(pool.ry)
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

	// pc = floor(ps * min(x / rx, y / ry))
	pc = pool.ps.MulTruncate(sdk.MinDec(
		x.ToDec().QuoTruncate(pool.rx),
		y.ToDec().QuoTruncate(pool.ry),
	)).TruncateInt()

	mintProportion := pc.ToDec().Quo(pool.ps)             // pc / ps
	ax = pool.rx.Mul(mintProportion).Ceil().TruncateInt() // ceil(rx * mintProportion)
	ay = pool.ry.Mul(mintProportion).Ceil().TruncateInt() // ceil(ry * mintProportion)
	return
}

// Withdraw returns withdrawn x and y coin amount when someone withdraws
// pc pool coin.
// Withdraw also takes care of the fee rate.
func (pool *BasicPool) Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	if pc.ToDec().Equal(pool.ps) {
		// Redeeming the last pool coin - give all remaining rx and ry.
		x = pool.rx.TruncateInt()
		y = pool.ry.TruncateInt()
		return
	}

	proportion := pc.ToDec().QuoTruncate(pool.ps)                             // pc / ps
	multiplier := sdk.OneDec().Sub(feeRate)                                   // 1 - feeRate
	x = pool.rx.MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(rx * proportion * multiplier)
	y = pool.ry.MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(ry * proportion * multiplier)
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
	return pool.rx.QuoTruncate(price).Sub(pool.ry).TruncateInt()
}

// SellAmountUnder returns the amount of sell orders for price less or equal
// than given price.
func (pool *BasicPool) SellAmountUnder(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.ry.Sub(pool.rx.QuoRoundUp(price)).TruncateInt()
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
