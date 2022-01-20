package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
)

var (
	_ PoolI = (*PoolInfo)(nil)
	// TODO: add RangedPoolInfo for v2
)

// NewPool returns a new pool object.
func NewPool(id, pairId uint64) Pool {
	return Pool{
		Id:                    id,
		PairId:                pairId,
		ReserveAddress:        PoolReserveAcc(id).String(),
		PoolCoinDenom:         PoolCoinDenom(id),
		LastDepositRequestId:  0,
		LastWithdrawRequestId: 0,
		Disabled:              false,
	}
}

func (pool Pool) GetReserveAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pool.ReserveAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// PoolReserveAcc returns a unique pool reserve account address for each pool.
// TODO: rename to PoolReserveAddr
func PoolReserveAcc(poolId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		AddressType,
		ModuleName,
		strings.Join([]string{PoolReserveAccPrefix, strconv.FormatUint(poolId, 10)}, ModuleAddrNameSplitter),
	)
}

// PoolCoinDenom returns a unique pool coin denom for a pool.
func PoolCoinDenom(poolId uint64) string {
	return fmt.Sprintf("pool%d", poolId)
}

// ParsePoolCoinDenom trims pool prefix from the pool coin denom and returns pool id.
func ParsePoolCoinDenom(denom string) uint64 {
	if !strings.HasPrefix(denom, "pool") {
		return 0
	}

	poolId, err := strconv.ParseUint(strings.TrimPrefix(denom, "pool"), 10, 64)
	if err != nil {
		return 0
	}

	return poolId
}

type PoolI interface {
	Balance() (rx, ry sdk.Int)
	PoolCoinSupply() sdk.Int
	Price() sdk.Dec
}

type PoolInfo struct {
	RX, RY sdk.Int
	PS     sdk.Int
}

func NewPoolInfo(rx, ry, ps sdk.Int) PoolInfo {
	return PoolInfo{
		RX: rx,
		RY: ry,
		PS: ps,
	}
}

func (info PoolInfo) Balance() (rx, ry sdk.Int) {
	return info.RX, info.RY
}

func (info PoolInfo) PoolCoinSupply() sdk.Int {
	return info.PS
}

func (info PoolInfo) Price() sdk.Dec {
	if info.RX.IsZero() || info.RY.IsZero() {
		panic("pool price is not defined for a depleted pool")
	}
	return info.RX.ToDec().Quo(info.RY.ToDec())
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

// MustUnmarshalPool return the unmarshalled pool from bytes.
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
