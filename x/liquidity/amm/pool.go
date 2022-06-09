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
	Id() uint64
	Balances() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
	IsDepleted() bool
	Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int)
	Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int)
	Orders(lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order
}

// BasicPool is the basic pool type.
type BasicPool struct {
	id uint64
	// rx and ry are the pool's reserve balance of each x/y coin.
	// In perspective of a pair, x coin is the quote coin and
	// y coin is the base coin.
	rx, ry sdk.Int
	// ps is the pool's pool coin supply.
	ps sdk.Int
}

// NewBasicPool returns a new BasicPool.
// It is OK to pass an empty sdk.Int to ps when ps is not going to be used.
func NewBasicPool(id uint64, rx, ry, ps sdk.Int) *BasicPool {
	return &BasicPool{
		id: id,
		rx: rx,
		ry: ry,
		ps: ps,
	}
}

func (pool *BasicPool) Id() uint64 {
	return pool.id
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

func (pool *BasicPool) Orders(lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	return append(
		pool.BuyOrders(lowestPrice, highestPrice, tickPrec),
		pool.SellOrders(lowestPrice, highestPrice, tickPrec)...)
}

func (pool *BasicPool) BuyOrders(lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	rx, ry := pool.Balances()
	var orders []Order
	if pool.Price().GT(highestPrice) {
		rxSqrt, err := rx.ToDec().ApproxSqrt()
		if err != nil {
			panic(err) // TODO: prevent panic
		}
		rySqrt, err := ry.ToDec().ApproxSqrt()
		if err != nil {
			panic(err)
		}
		highestPriceSqrt, err := highestPrice.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		dx := rx.ToDec().Sub(highestPriceSqrt.Mul(rxSqrt.Mul(rySqrt))) // dx = rx - sqrt(P * rx * ry)
		dy := dx.QuoTruncate(highestPrice).TruncateInt()               // dy = dx / P
		orders = append(orders, NewPoolOrder(pool.id, Buy, highestPrice, dy))
		rx = rx.Sub(highestPrice.MulInt(dy).Ceil().TruncateInt()) // buy side quote coin ceiling
		ry = ry.Add(dy)
	}
	for rx.IsPositive() {
		tick := PriceToDownTick(rx.ToDec().QuoInt(ry.Add(MinCoinAmount)), tickPrec)
		if tick.LT(lowestPrice) {
			break
		}
		dx := rx.ToDec().Sub(tick.MulInt(ry))    // dx = rx - P * ry
		dy := dx.QuoTruncate(tick).TruncateInt() // dy = dx / P
		orders = append(orders, NewPoolOrder(pool.id, Buy, tick, dy))
		rx = rx.Sub(tick.MulInt(dy).Ceil().TruncateInt())
		ry = ry.Add(dy)
	}
	return orders
}

func (pool *BasicPool) SellOrders(lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	rx, ry := pool.Balances()
	var orders []Order
	if pool.Price().LT(lowestPrice) {
		rxSqrt, err := rx.ToDec().ApproxSqrt()
		if err != nil {
			panic(err) // TODO: prevent panic
		}
		rySqrt, err := ry.ToDec().ApproxSqrt()
		if err != nil {
			panic(err)
		}
		lowestPriceSqrt, err := lowestPrice.ApproxSqrt()
		if err != nil {
			panic(err)
		}
		dy := ry.ToDec().Sub(rxSqrt.Mul(rySqrt).Quo(lowestPriceSqrt)).TruncateInt() // dy = ry - sqrt(rx * ry / P)
		orders = append(orders, NewPoolOrder(pool.id, Sell, lowestPrice, dy))
		rx = rx.Add(lowestPrice.MulInt(dy).TruncateInt()) // sell side quote coin truncation
		ry = ry.Sub(dy)
	}
	for ry.GT(MinCoinAmount) {
		tick := PriceToUpTick(sdk.MaxDec(
			rx.Add(sdk.OneInt()).ToDec().QuoInt(ry),
			rx.ToDec().QuoInt(ry.Sub(MinCoinAmount)),
		), tickPrec)
		if tick.GT(highestPrice) {
			break
		}
		dy := ry.ToDec().Sub(rx.ToDec().QuoRoundUp(tick)).TruncateInt() // dy = ry - rx / P
		orders = append(orders, NewPoolOrder(pool.id, Sell, tick, dy))
		rx = rx.Add(tick.MulInt(dy).TruncateInt())
		ry = ry.Sub(dy)
	}
	return orders
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
