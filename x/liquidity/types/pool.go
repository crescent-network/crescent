package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ PoolI = (*PoolInfo)(nil)
	// TODO: add RangedPoolInfo for v2
)

func (pool Pool) GetReserveAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pool.ReserveAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

type PoolI interface {
	Balance() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
}

type PoolInfo struct {
	rx, ry sdk.Int
	ps     sdk.Int
}

func NewPoolInfo(rx, ry, ps sdk.Int) PoolInfo {
	return PoolInfo{
		rx: rx,
		ry: ry,
		ps: ps,
	}
}

func (info PoolInfo) Balance() (rx, ry sdk.Int) {
	return info.rx, info.ry
}

func (info PoolInfo) PoolCoinSupply() sdk.Int {
	return info.ps
}

func (info PoolInfo) Price() sdk.Dec {
	if info.rx.IsZero() || info.ry.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return info.rx.ToDec().Quo(info.ry.ToDec())
}

func IsDepletedPool(pool PoolI) bool {
	ps := pool.PoolCoinSupply()
	if ps.IsZero() {
		return true
	}
	rx, ry := pool.Balance()
	if rx.IsZero() || ry.IsZero() {
		return true
	}
	return false
}

// DepositToPool returns accepted x amount, accepted y amount and
// minted pool coin amount.
func DepositToPool(pool PoolI, x, y sdk.Int) (ax, ay, pc sdk.Int) {
	// Calculate accepted amount and minting amount.
	// Note that we take as many coins as possible(by ceiling numbers)
	// from depositor and mint as little coins as possible.
	rx, ry := pool.Balance()
	ps := pool.PoolCoinSupply().ToDec()
	// pc = min(ps * (x / rx), ps * (y / ry))
	pc = sdk.MinDec(
		ps.MulTruncate(x.ToDec().QuoTruncate(rx.ToDec())),
		ps.MulTruncate(y.ToDec().QuoTruncate(ry.ToDec())),
	).TruncateInt()

	mintProportion := pc.ToDec().Quo(ps)                     // pc / ps
	ax = rx.ToDec().Mul(mintProportion).Ceil().TruncateInt() // rx * mintProportion
	ay = ry.ToDec().Mul(mintProportion).Ceil().TruncateInt() // ry * mintProportion

	return
}

func WithdrawFromPool(pool PoolI, pc sdk.Int, feeRate sdk.Dec) (x, y sdk.Int) {
	rx, ry := pool.Balance()
	ps := pool.PoolCoinSupply()

	// Redeeming the last pool coin
	if pc.Equal(ps) {
		x = rx
		y = ry
		return
	}

	proportion := pc.ToDec().QuoTruncate(ps.ToDec())                             // pc / ps
	multiplier := sdk.OneDec().Sub(feeRate)                                      // 1 - feeRate
	x = rx.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // rx * proportion * multiplier
	y = ry.ToDec().MulTruncate(proportion).MulTruncate(multiplier).TruncateInt() // ry * proportion * multiplier

	return
}
