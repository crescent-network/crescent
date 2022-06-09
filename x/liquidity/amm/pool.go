package amm

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/types"
)

var (
	_ Pool = (*BasicPool)(nil)
)

// Pool is the interface of a pool.
// It also satisfies OrderView interface.
type Pool interface {
	NewOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) Order
	Balances() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
	IsDepleted() bool
	Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int)
	Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int)
	BuyAmount(price sdk.Dec) sdk.Int
	SellAmount(price sdk.Dec) sdk.Int
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

func (pool *BasicPool) NewOrder(dir OrderDirection, price sdk.Dec, amt sdk.Int) Order {
	return NewPoolOrder(0, nil, dir, price, amt)
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

	utils.SafeMath(func() {
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
	}, func() {
		ax, ay, pc = sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroInt()
	})

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

	utils.SafeMath(func() {
		proportion := pc.ToDec().QuoTruncate(pool.ps.ToDec())                             // pc / ps
		multiplier := sdk.OneDec().Sub(feeRate)                                           // 1 - feeRate
		x = pool.rx.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(rx * proportion * multiplier)
		y = pool.ry.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // floor(ry * proportion * multiplier)
	}, func() {
		x, y = sdk.ZeroInt(), sdk.ZeroInt()
	})

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

func (pool *BasicPool) BuyAmount(price sdk.Dec) (amt sdk.Int) {
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

func (pool *BasicPool) SellAmount(price sdk.Dec) sdk.Int {
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
	for {
		rx, ry := tmpPool.Balances()
		if !rx.IsPositive() {
			break
		}
		tick := PriceToDownTick(rx.ToDec().QuoInt(ry.Add(MinCoinAmount)), tickPrec) // TODO: generalize
		if tick.LT(lowestPrice) {
			break
		}
		placeOrder(tick, tmpPool.BuyAmount(tick))
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
	for {
		rx, ry := tmpPool.Balances()
		if !ry.GT(MinCoinAmount) {
			break
		}
		tick := PriceToUpTick(sdk.MaxDec(
			rx.Add(sdk.OneInt()).ToDec().QuoInt(ry),
			rx.ToDec().QuoInt(ry.Sub(MinCoinAmount)),
		), tickPrec) // TODO: generalize
		if tick.GT(highestPrice) {
			break
		}
		placeOrder(tick, tmpPool.SellAmount(tick))
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
