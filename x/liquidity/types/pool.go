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
	_ OrderSource = (*PoolOrderSource)(nil)
)

// NewPool returns a new pool object.
func NewPool(id, pairId uint64, xCoinDenom, yCoinDenom string) Pool {
	return Pool{
		Id:                    id,
		PairId:                pairId,
		XCoinDenom:            xCoinDenom,
		YCoinDenom:            yCoinDenom,
		ReserveAddress:        PoolReserveAcc(id).String(),
		PoolCoinDenom:         PoolCoinDenom(id),
		LastDepositRequestId:  0,
		LastWithdrawRequestId: 0,
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

type PoolOrderSource struct {
	ReserveAddress sdk.AccAddress
	RX, RY         sdk.Int
	PoolPrice      sdk.Dec
	Direction      SwapDirection
	TickPrecision  int
	pxCache        map[string]sdk.Int // map(price => providableXOnTick)
	pyCache        map[string]sdk.Int // map(price => providableYOnTick)
}

func NewPoolOrderSource(pool PoolI, reserveAddr sdk.AccAddress, dir SwapDirection, prec int) OrderSource {
	rx, ry := pool.Balance()
	return &PoolOrderSource{
		ReserveAddress: reserveAddr,
		RX:             rx,
		RY:             ry,
		PoolPrice:      pool.Price(),
		Direction:      dir,
		TickPrecision:  prec,
		pxCache:        map[string]sdk.Int{},
		pyCache:        map[string]sdk.Int{},
	}
}

func (os PoolOrderSource) ProvidableX(price sdk.Dec) sdk.Int {
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	return os.RX.ToDec().Sub(price.MulInt(os.RY)).TruncateInt()
}

func (os PoolOrderSource) ProvidableY(price sdk.Dec) sdk.Int {
	if price.LTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	return price.MulInt(os.RY).Sub(os.RX.ToDec()).Quo(price).TruncateInt()
}

func (os PoolOrderSource) ProvidableXOnTick(price sdk.Dec) sdk.Int {
	if price.GTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	s := price.String()
	px, ok := os.pxCache[s]
	if !ok {
		px = os.ProvidableX(price).Sub(os.ProvidableX(UpTick(price, os.TickPrecision)))
		os.pxCache[s] = px
	}
	return px
}

func (os PoolOrderSource) ProvidableYOnTick(price sdk.Dec) sdk.Int {
	if price.LTE(os.PoolPrice) {
		return sdk.ZeroInt()
	}
	s := price.String()
	py, ok := os.pyCache[s]
	if !ok {
		py = os.ProvidableY(price).Sub(os.ProvidableY(DownTick(price, os.TickPrecision)))
		os.pyCache[s] = py
	}
	return py
}

func (os PoolOrderSource) AmountGTE(price sdk.Dec) sdk.Int {
	amount := sdk.ZeroInt()
	switch os.Direction {
	case SwapDirectionBuy:
		for price.LT(os.PoolPrice) {
			px := os.ProvidableXOnTick(price)
			if px.IsZero() { // TODO: will it happen?
				break
			}
			amount = amount.Add(px)
			price = UpTick(price, os.TickPrecision)
		}
	case SwapDirectionSell:
		for price.GT(os.PoolPrice) {
			py := os.ProvidableYOnTick(price)
			if py.IsZero() {
				break
			}
			amount = amount.Add(py)
			price = UpTick(price, os.TickPrecision)
		}
	}
	return amount
}

func (os PoolOrderSource) AmountLTE(price sdk.Dec) sdk.Int {
	amount := sdk.ZeroInt()
	switch os.Direction {
	case SwapDirectionBuy:
		for price.LT(os.PoolPrice) {
			px := os.ProvidableXOnTick(price)
			if px.IsZero() {
				break
			}
			amount = amount.Add(px)
			price = DownTick(price, os.TickPrecision)
		}
	case SwapDirectionSell:
		for price.GT(os.PoolPrice) {
			py := os.ProvidableYOnTick(price)
			if py.IsZero() { // TODO: will it happen?
				break
			}
			amount = amount.Add(py)
			price = DownTick(price, os.TickPrecision)
		}
	}
	return amount
}

func (os PoolOrderSource) Orders(price sdk.Dec) Orders {
	switch os.Direction {
	case SwapDirectionBuy:
		return Orders{NewPoolOrder(os.ReserveAddress, SwapDirectionBuy, price, os.ProvidableXOnTick(price))}
	case SwapDirectionSell:
		return Orders{NewPoolOrder(os.ReserveAddress, SwapDirectionSell, price, os.ProvidableYOnTick(price))}
	}
	return nil
}

func (os PoolOrderSource) UpTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = UpTick(price, os.TickPrecision)
		if tick.GTE(os.PoolPrice) {
			return
		}
		found = true
	case SwapDirectionSell:
		tick = UpTick(price, os.TickPrecision)
		if tick.LTE(os.PoolPrice) {
			return
		}
		py := os.ProvidableYOnTick(price)
		if py.IsZero() {
			return
		}
		found = true
	}
	return
}

func (os PoolOrderSource) DownTick(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = DownTick(price, os.TickPrecision)
		if tick.GTE(os.PoolPrice) {
			return
		}
		px := os.ProvidableXOnTick(price)
		if px.IsZero() {
			return
		}
		found = true
	case SwapDirectionSell:
		tick = DownTick(price, os.TickPrecision)
		if tick.LTE(os.PoolPrice) {
			return
		}
		found = true
	}
	return
}

func (os PoolOrderSource) UpTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		price = UpTick(price, os.TickPrecision)
		if price.GTE(os.PoolPrice) {
			return
		}
		px := os.ProvidableXOnTick(price)
		if px.IsZero() {
			return
		}
		found = true
	case SwapDirectionSell:
		price = UpTick(price, os.TickPrecision)
		if price.LTE(os.PoolPrice) {
			return
		}
		py := os.ProvidableYOnTick(price)
		if py.IsZero() {
			return
		}
		found = true
	}
	return
}

func (os PoolOrderSource) DownTickWithOrders(price sdk.Dec) (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		price = DownTick(price, os.TickPrecision)
		if price.GTE(os.PoolPrice) {
			return
		}
		px := os.ProvidableXOnTick(price)
		if px.IsZero() {
			return
		}
		found = true
	case SwapDirectionSell:
		price = DownTick(price, os.TickPrecision)
		if price.LTE(os.PoolPrice) {
			return
		}
		py := os.ProvidableYOnTick(price)
		if py.IsZero() {
			return
		}
		found = true
	}
	return
}

func (os PoolOrderSource) HighestTick() (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		tick = PriceToTick(os.PoolPrice, os.TickPrecision)
		if os.PoolPrice.Equal(tick) {
			tick = DownTick(tick, os.TickPrecision)
		}
		found = true
	case SwapDirectionSell:
		// TODO: is it possible to calculate?
		panic("not implemented")
	}
	return
}

func (os PoolOrderSource) LowestTick() (tick sdk.Dec, found bool) {
	switch os.Direction {
	case SwapDirectionBuy:
		// TODO: is it possible to calculate?
		panic("not implemented")
	case SwapDirectionSell:
		tick = UpTick(PriceToTick(os.PoolPrice, os.TickPrecision), os.TickPrecision)
		found = true
	}
	return
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
