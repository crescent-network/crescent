package amm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Pool = (*BasicPool)(nil)

type Pool interface {
	OrderView
	Price() sdk.Dec
	IsDepleted() bool
	Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int)
	Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int)
}

type BasicPool struct {
	rx, ry sdk.Dec
	ps     sdk.Dec
}

func NewBasicPool(rx, ry, ps sdk.Int) *BasicPool {
	return &BasicPool{
		rx: rx.ToDec(),
		ry: ry.ToDec(),
		ps: ps.ToDec(),
	}
}

func (pool *BasicPool) Price() sdk.Dec {
	return pool.rx.Quo(pool.ry)
}

func (pool *BasicPool) IsDepleted() bool {
	return pool.ps.IsZero() || pool.rx.IsZero() || pool.ry.IsZero()
}

func (pool *BasicPool) Deposit(x, y sdk.Int) (ax, ay, pc sdk.Int) {
	// Calculate accepted amount and minting amount.
	// Note that we take as many coins as possible(by ceiling numbers)
	// from depositor and mint as little coins as possible.
	// pc = min(ps * (x / rx), ps * (y / ry))
	pc = sdk.MinDec(
		pool.ps.MulTruncate(x.ToDec().QuoTruncate(pool.rx)),
		pool.ps.MulTruncate(y.ToDec().QuoTruncate(pool.ry)),
	).TruncateInt()

	mintProportion := pc.ToDec().Quo(pool.ps)             // pc / ps
	ax = pool.rx.Mul(mintProportion).Ceil().TruncateInt() // rx * mintProportion
	ay = pool.ry.Mul(mintProportion).Ceil().TruncateInt() // ry * mintProportion
	return
}

func (pool *BasicPool) Withdraw(pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	if pc.ToDec().Equal(pool.ps) {
		// Redeeming the last pool coin.
		x = pool.rx.TruncateInt()
		y = pool.ry.TruncateInt()
		return
	}

	proportion := pc.ToDec().QuoTruncate(pool.ps)                             // pc / ps
	multiplier := sdk.OneDec().Sub(feeRate)                                   // 1 - feeRate
	x = pool.rx.MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // rx * proportion * multiplier
	y = pool.ry.MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // ry * proportion * multiplier
	return
}

func (pool *BasicPool) HighestBuyPrice() (price sdk.Dec, found bool) {
	// The highest buy price is actually a bit lower than pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

func (pool *BasicPool) LowestSellPrice() (price sdk.Dec, found bool) {
	// The lowest sell price is actually a bit higher than the pool price,
	// but it's not important for our matching logic.
	return pool.Price(), true
}

func (pool *BasicPool) BuyAmountOver(price sdk.Dec) sdk.Int {
	if price.GTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.rx.QuoTruncate(price).Sub(pool.ry).TruncateInt()
}

func (pool *BasicPool) SellAmountUnder(price sdk.Dec) sdk.Int {
	if price.LTE(pool.Price()) {
		return sdk.ZeroInt()
	}
	return pool.ry.Sub(pool.rx.QuoRoundUp(price)).TruncateInt()
}
