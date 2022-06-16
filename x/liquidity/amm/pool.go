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
	Balances() (rx, ry sdk.Int)
	WithBalances(rx, ry sdk.Int) Pool
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

// Balances returns the balances of the pool.
func (pool *BasicPool) Balances() (rx, ry sdk.Int) {
	return pool.rx, pool.ry
}

func (pool *BasicPool) WithBalances(rx, ry sdk.Int) Pool {
	return NewBasicPool(rx, ry, pool.ps)
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
	dx := pool.rx.ToDec().Sub(price.MulInt(pool.ry)).TruncateInt()
	if dx.IsZero() {
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

func (pool *BasicPool) BuyAmountTo(price sdk.Dec) (amt sdk.Int) {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	sqrtRx := utils.DecApproxSqrt(pool.rx.ToDec())
	sqrtRy := utils.DecApproxSqrt(pool.ry.ToDec())
	sqrtPrice := utils.DecApproxSqrt(price)
	dx := pool.rx.ToDec().Sub(sqrtPrice.Mul(sqrtRx.Mul(sqrtRy))) // dx = rx - sqrt(P * rx * ry)
	if dx.IsZero() {
		return sdk.ZeroInt()
	}
	utils.SafeMath(func() {
		amt = dx.QuoTruncate(price).TruncateInt() // dy = dx / P
		if amt.GT(MaxCoinAmount) {
			amt = MaxCoinAmount
		}
	}, func() {
		amt = MaxCoinAmount
	})
	return
}

func (pool *BasicPool) SellAmountTo(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	sqrtRx := utils.DecApproxSqrt(pool.rx.ToDec())
	sqrtRy := utils.DecApproxSqrt(pool.ry.ToDec())
	sqrtPrice := utils.DecApproxSqrt(price)
	// dy = ry - sqrt(rx * ry / P)
	return pool.ry.ToDec().Sub(sqrtRx.Mul(sqrtRy).Quo(sqrtPrice)).TruncateInt()
}

type RangedPool struct {
	rx, ry             sdk.Int
	ps                 sdk.Int
	transX, transY     sdk.Dec
	xComp, yComp       sdk.Dec
	minPrice, maxPrice *sdk.Dec
}

// NewRangedPool returns a new RangedPool.
func NewRangedPool(rx, ry, ps sdk.Int, transX, transY sdk.Dec, minPrice, maxPrice *sdk.Dec) *RangedPool {
	return &RangedPool{
		rx:       rx,
		ry:       ry,
		ps:       ps,
		transX:   transX,
		transY:   transY,
		xComp:    rx.ToDec().Add(transX),
		yComp:    ry.ToDec().Add(transY),
		minPrice: minPrice,
		maxPrice: maxPrice,
	}
}

// CreateRangedPool creates new RangedPool from given inputs, while validating
// the inputs and using only needed amount of x/y coins(the rest should be refunded).
func CreateRangedPool(x, y sdk.Int, initialPrice sdk.Dec, minPrice, maxPrice *sdk.Dec) (pool *RangedPool, err error) {
	ax, ay, transX, transY, err := createRangedPool(x, y, initialPrice, minPrice, maxPrice)
	if err != nil {
		return nil, err
	}
	return NewRangedPool(ax, ay, InitialPoolCoinSupply(ax, ay), transX, transY, minPrice, maxPrice), nil
}

func ValidateRangedPoolParams(initialPrice sdk.Dec, minPrice, maxPrice *sdk.Dec) error {
	if !initialPrice.IsPositive() {
		return fmt.Errorf("initial price must be positive: %s", initialPrice)
	}
	if minPrice == nil && maxPrice == nil {
		return fmt.Errorf("min price and max price must not be nil at the same time")
	}
	if minPrice != nil && !minPrice.IsPositive() {
		return fmt.Errorf("min price must be positive: %s", minPrice)
	}
	if maxPrice != nil && !maxPrice.IsPositive() {
		return fmt.Errorf("max price must be positive: %s", maxPrice)
	}
	if minPrice != nil && maxPrice != nil && !maxPrice.GT(*minPrice) {
		return fmt.Errorf("max price must be greater than min price")
	}
	return nil
}

func createRangedPool(x, y sdk.Int, initialPrice sdk.Dec, minPrice, maxPrice *sdk.Dec) (ax, ay sdk.Int, transX, transY sdk.Dec, err error) {
	if !x.IsPositive() && !y.IsPositive() {
		err = fmt.Errorf("either x or y must be positive")
		return
	}
	if err = ValidateRangedPoolParams(initialPrice, minPrice, maxPrice); err != nil {
		return
	}
	sqrt := utils.DecApproxSqrt
	// P = initialPrice, M = minPrice, L = maxPrice
	switch {
	case minPrice != nil && initialPrice.LTE(*minPrice):
		if y.IsZero() {
			err = fmt.Errorf("y amount must be positive")
			return
		}
		ax = zeroInt
		ay = y
		if maxPrice == nil {
			// transX = ay * M
			transX = ay.ToDec().Mul(*minPrice)
			// transY = 0
			transY = zeroDec
			return
		}
		sqrtM := sqrt(*minPrice)
		sqrtL := sqrt(*maxPrice)
		// sqrtK = sqrt(k) = ay / (1/sqrt(M) - 1/sqrt(L))
		sqrtK := ay.ToDec().Quo(inv(sqrtM).Sub(inv(sqrtL)))
		// transX = sqrt(k) * sqrt(M)
		transX = sqrtK.Mul(sqrtM)
		// transY = sqrt(k) / sqrt(L)
		transY = sqrtK.Quo(sqrtL)
		return
	case maxPrice != nil && initialPrice.GTE(*maxPrice):
		if x.IsZero() {
			err = fmt.Errorf("x amount must be positive")
			return
		}
		ax = x
		ay = zeroInt
		if minPrice == nil {
			// transX = 0
			transX = zeroDec
			// transY = ax / L
			transY = ax.ToDec().Quo(*maxPrice)
			return
		}
		sqrtM := sqrt(*minPrice)
		sqrtL := sqrt(*maxPrice)
		// sqrtK = sqrt(K) = ax / (sqrt(L) - sqrt(M))
		sqrtK := ax.ToDec().Quo(sqrtL.Sub(sqrtM))
		// transX = sqrt(k) * sqrt(M)
		transX = sqrtK.Mul(sqrtM)
		// transY = sqrt(k) / sqrt(L)
		transY = sqrtK.Quo(sqrtL)
		return
	}
	if x.IsZero() || y.IsZero() {
		err = fmt.Errorf("x and y amount must be positve")
		return
	}
	// Assume that we can accept all x
	ax = x
	switch {
	case minPrice == nil:
		sqrtPL := sqrt(initialPrice.Mul(*maxPrice))
		// ay = ax * (1/P - 1/sqrt(P * L))
		ay = ax.ToDec().Mul(inv(initialPrice).Sub(inv(sqrtPL))).Ceil().TruncateInt()
		if ay.LTE(y) {
			// transX = 0
			transX = zeroDec
			// transY = ax / sqrt(P * L)
			transY = ax.ToDec().Quo(sqrtPL)
			return
		}
	case maxPrice == nil:
		// ay = ax / (P - sqrt(P * M))
		ay = ax.ToDec().Quo(
			initialPrice.Sub(sqrt(initialPrice.Mul(*minPrice))),
		).Ceil().TruncateInt()
		if ay.LTE(y) {
			// transX = ax / (sqrt(P / M) - 1)
			transX = ax.ToDec().Quo(sqrt(initialPrice.Quo(*minPrice)).Sub(oneDec))
			// transY = 0
			transY = zeroDec
			return
		}
	default:
		sqrtP := sqrt(initialPrice)
		sqrtM := sqrt(*minPrice)
		sqrtL := sqrt(*maxPrice)
		// sqrtK = sqrt(k) = ax / (sqrt(P) - sqrt(M))
		sqrtK := ax.ToDec().Quo(sqrtP.Sub(sqrtM))
		// ay = sqrt(k) * (1/sqrt(P) - 1/sqrt(L))
		ay = sqrtK.Mul(
			inv(sqrtP).Sub(inv(sqrtL)),
		).Ceil().TruncateInt()
		if ay.LTE(y) {
			// transX = ax / (sqrt(P / M) - 1)
			transX = ax.ToDec().Quo(sqrt(initialPrice.Quo(*minPrice)).Sub(oneDec))
			// transY = sqrt(k) / sqrt(L)
			transY = sqrtK.Quo(sqrtL)
			return
		}
	}
	// We accept all y
	ay = y
	switch {
	case minPrice == nil:
		// ax = ay / (1/P - 1/sqrt(P * L))
		ax = ay.ToDec().Quo(
			inv(initialPrice).Sub(inv(sqrt(initialPrice.Mul(*maxPrice)))),
		).Ceil().TruncateInt()
		if ax.GT(x) {
			err = fmt.Errorf("invalid pool")
			return
		}
		// transX = 0
		transX = zeroDec
		// transY = ay / (sqrt(L / P) - 1)
		transY = ay.ToDec().Quo(sqrt(maxPrice.Quo(initialPrice)).Sub(oneDec))
	case maxPrice == nil:
		sqrtPM := sqrt(initialPrice.Mul(*minPrice))
		// ax = ay * (P - sqrt(P * M))
		ax = ay.ToDec().Mul(initialPrice.Sub(sqrtPM)).Ceil().TruncateInt()
		if ax.GT(x) {
			err = fmt.Errorf("invalid pool")
			return
		}
		// transX = ay * sqrt(P * M)
		transX = ay.ToDec().Mul(sqrtPM)
		// transY = 0
		transY = zeroDec
	default:
		sqrtP := sqrt(initialPrice)
		sqrtM := sqrt(*minPrice)
		// sqrtK = sqrt(k) = ay / (1/sqrt(P) - 1/sqrt(L))
		sqrtK := ay.ToDec().Quo(inv(sqrtP).Sub(inv(sqrt(*maxPrice))))
		// ax = sqrt(k) * (sqrt(P) - sqrt(M))
		ax = sqrtK.Mul(sqrtP.Sub(sqrtM)).Ceil().TruncateInt()
		if ax.GT(x) {
			err = fmt.Errorf("invalid pool")
			return
		}
		// transX = sqrt(k) * sqrt(M)
		transX = sqrtK.Mul(sqrtM)
		// transY = ay / (sqrt(L / P) - 1)
		transY = ay.ToDec().Quo(sqrt(maxPrice.Quo(initialPrice)).Sub(oneDec))
	}
	return
}

// Balances returns the balances of the pool.
func (pool *RangedPool) Balances() (rx, ry sdk.Int) {
	return pool.rx, pool.ry
}

func (pool *RangedPool) WithBalances(rx, ry sdk.Int) Pool {
	return NewRangedPool(rx, ry, pool.ps, pool.transX, pool.transY, pool.minPrice, pool.maxPrice)
}

// PoolCoinSupply returns the pool coin supply.
func (pool *RangedPool) PoolCoinSupply() sdk.Int {
	return pool.ps
}

func (pool *RangedPool) Translation() (transX, transY sdk.Dec) {
	return pool.transX, pool.transY
}

// Price returns the pool price.
func (pool *RangedPool) Price() sdk.Dec {
	if pool.rx.IsZero() && pool.ry.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return pool.xComp.Quo(pool.yComp) // (rx + transX) / (ry + transY)
}

// IsDepleted returns whether the pool is depleted or not.
func (pool *RangedPool) IsDepleted() bool {
	return pool.ps.IsZero() || (pool.rx.IsZero() && pool.ry.IsZero())
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
func (pool *RangedPool) BuyAmountOver(price sdk.Dec, _ bool) (amt sdk.Int) {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.minPrice != nil && price.LT(*pool.minPrice) {
		price = *pool.minPrice
	}
	// dx = (rx + transX) - P * (ry + transY)
	dx := pool.xComp.Sub(price.Mul(pool.yComp))
	if dx.IsZero() {
		return sdk.ZeroInt()
	}
	utils.SafeMath(func() {
		amt = dx.QuoTruncate(price).TruncateInt() // dy = dx / P
		if amt.GT(MaxCoinAmount) {
			amt = MaxCoinAmount
		}
	}, func() {
		amt = MaxCoinAmount
	})
	return
}

func (pool *RangedPool) SellAmountUnder(price sdk.Dec, _ bool) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.maxPrice != nil && price.GT(*pool.maxPrice) {
		price = *pool.maxPrice
	}
	// dy = (ry + transY) - (rx + transX) / P
	return pool.yComp.Sub(pool.xComp.QuoRoundUp(price)).TruncateInt()
}

func (pool *RangedPool) BuyAmountTo(price sdk.Dec) (amt sdk.Int) {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.minPrice != nil && price.LT(*pool.minPrice) {
		price = *pool.minPrice
	}
	sqrtXComp := utils.DecApproxSqrt(pool.xComp)
	sqrtYComp := utils.DecApproxSqrt(pool.yComp)
	sqrtPrice := utils.DecApproxSqrt(price)
	// dx = rx - (sqrt(P * (rx + transX) * (ry + transY)) - transX)
	dx := pool.rx.ToDec().Sub(sqrtPrice.Mul(sqrtXComp.Mul(sqrtYComp)).Sub(pool.transX))
	if dx.IsZero() {
		return sdk.ZeroInt()
	}
	utils.SafeMath(func() {
		amt = dx.QuoTruncate(price).TruncateInt() // dy = dx / P
		if amt.GT(MaxCoinAmount) {
			amt = MaxCoinAmount
		}
	}, func() {
		amt = MaxCoinAmount
	})
	return
}

func (pool *RangedPool) SellAmountTo(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	if pool.maxPrice != nil && price.GT(*pool.maxPrice) {
		price = *pool.maxPrice
	}
	sqrtXComp := utils.DecApproxSqrt(pool.xComp)
	sqrtYComp := utils.DecApproxSqrt(pool.yComp)
	sqrtPrice := utils.DecApproxSqrt(price)
	// dy = ry - (sqrt((x + transX) * (y + transY) / P) - b)
	return pool.ry.ToDec().Sub(sqrtXComp.Mul(sqrtYComp).QuoRoundUp(sqrtPrice).Sub(pool.transY)).TruncateInt()
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

func PoolBuyOrders(pool Pool, orderer Orderer, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	tmpPool := pool
	var orders []Order
	placeOrder := func(price sdk.Dec, amt sdk.Int) {
		orders = append(orders, orderer.Order(Buy, price, amt))
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

func PoolSellOrders(pool Pool, orderer Orderer, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	tmpPool := pool
	var orders []Order
	placeOrder := func(price sdk.Dec, amt sdk.Int) {
		orders = append(orders, orderer.Order(Sell, price, amt))
		rx, ry := tmpPool.Balances()
		rx = rx.Add(price.MulInt(amt).TruncateInt()) // quote coin truncation
		ry = ry.Sub(amt)
		tmpPool = pool.WithBalances(rx, ry)
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

func PoolOrders(pool Pool, orderer Orderer, lowestPrice, highestPrice sdk.Dec, tickPrec int) []Order {
	return append(
		PoolBuyOrders(pool, orderer, lowestPrice, highestPrice, tickPrec),
		PoolSellOrders(pool, orderer, lowestPrice, highestPrice, tickPrec)...)
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
