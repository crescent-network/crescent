package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
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

// MustMarshalPool returns the pool bytes.
// It throws panic if it fails.
func MustMarshalPool(cdc codec.BinaryCodec, pool Pool) []byte {
	return cdc.MustMarshal(&pool)
}

// MustUnmarshalPool return the unmarshaled pool from bytes.
// It throws panic if it fails.
func MustUnmarshalPool(cdc codec.BinaryCodec, value []byte) Pool {
	pool, err := UnmarshalPool(cdc, value)
	if err != nil {
		panic(err)
	}

	return pool
}

// UnmarshalPool returns the pool from bytes.
func UnmarshalPool(cdc codec.BinaryCodec, value []byte) (pool Pool, err error) {
	err = cdc.Unmarshal(value, &pool)
	return pool, err
}

// MustMarshalDepositRequest returns the DepositRequest bytes. Panics if fails.
func MustMarshalDepositRequest(cdc codec.BinaryCodec, msg DepositRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalDepositMsgState returns the DepositRequest from bytes.
func UnmarshalDepositRequest(cdc codec.BinaryCodec, value []byte) (msg DepositRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshalDepositRequest returns the DepositRequest from bytes.
// It throws panic if it fails.
func MustUnmarshalDepositRequest(cdc codec.BinaryCodec, value []byte) DepositRequest {
	msg, err := UnmarshalDepositRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}

// MustMarshaWithdrawRequest returns the WithdrawRequest bytes.
// It throws panic if it fails.
func MustMarshaWithdrawRequest(cdc codec.BinaryCodec, msg WithdrawRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalWithdrawRequest returns the WithdrawRequest from bytes.
func UnmarshalWithdrawRequest(cdc codec.BinaryCodec, value []byte) (msg WithdrawRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshaWithdrawRequest returns the WithdrawRequest from bytes.
// It throws panic if it fails.
func MustUnmarshaWithdrawRequest(cdc codec.BinaryCodec, value []byte) WithdrawRequest {
	msg, err := UnmarshalWithdrawRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}

// MustMarshaSwapRequest returns the SwapRequest bytes.
// It throws panic if it fails.
func MustMarshaSwapRequest(cdc codec.BinaryCodec, msg SwapRequest) []byte {
	return cdc.MustMarshal(&msg)
}

// UnmarshalSwapRequest returns the SwapRequest from bytes.
func UnmarshalSwapRequest(cdc codec.BinaryCodec, value []byte) (msg SwapRequest, err error) {
	err = cdc.Unmarshal(value, &msg)
	return msg, err
}

// MustUnmarshaSwapRequest returns the SwapRequest from bytes.
// It throws panic if it fails.
func MustUnmarshaSwapRequest(cdc codec.BinaryCodec, value []byte) SwapRequest {
	msg, err := UnmarshalSwapRequest(cdc, value)
	if err != nil {
		panic(err)
	}
	return msg
}
