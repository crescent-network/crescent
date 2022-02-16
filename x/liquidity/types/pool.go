package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
)

var _ amm.OrderSource = (*BasicPoolOrderSource)(nil)

// PoolReserveAddress returns a unique pool reserve account address for each pool.
func PoolReserveAddress(poolId uint64) sdk.AccAddress {
	return farmingtypes.DeriveAddress(
		AddressType,
		ModuleName,
		strings.Join([]string{PoolReserveAddressPrefix, strconv.FormatUint(poolId, 10)}, ModuleAddressNameSplitter),
	)
}

// NewPool returns a new pool object.
func NewPool(id, pairId uint64) Pool {
	return Pool{
		Id:                    id,
		PairId:                pairId,
		ReserveAddress:        PoolReserveAddress(id).String(),
		PoolCoinDenom:         PoolCoinDenom(id),
		LastDepositRequestId:  0,
		LastWithdrawRequestId: 0,
		Disabled:              false,
	}
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

func (pool Pool) GetReserveAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(pool.ReserveAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// Validate validates Pool for genesis.
func (pool Pool) Validate() error {
	if pool.Id == 0 {
		return fmt.Errorf("pool id must not be 0")
	}
	if pool.PairId == 0 {
		return fmt.Errorf("pair id must not be 0")
	}
	if _, err := sdk.AccAddressFromBech32(pool.ReserveAddress); err != nil {
		return fmt.Errorf("invalid reserve address %s: %w", pool.ReserveAddress, err)
	}
	if err := sdk.ValidateDenom(pool.PoolCoinDenom); err != nil {
		return fmt.Errorf("invalid pool coin denom: %w", err)
	}
	return nil
}

// BasicPoolOrderSource is the order source for a pool which implements
// amm.OrderSource.
type BasicPoolOrderSource struct {
	amm.Pool
	PoolId                        uint64
	PoolReserveAddress            sdk.AccAddress
	BaseCoinDenom, QuoteCoinDenom string
}

// NewBasicPoolOrderSource returns a new BasicPoolOrderSource.
func NewBasicPoolOrderSource(
	pool amm.Pool, poolId uint64, reserveAddr sdk.AccAddress, baseCoinDenom, quoteCoinDenom string) *BasicPoolOrderSource {
	return &BasicPoolOrderSource{
		Pool:               pool,
		PoolId:             poolId,
		PoolReserveAddress: reserveAddr,
		BaseCoinDenom:      baseCoinDenom,
		QuoteCoinDenom:     quoteCoinDenom,
	}
}

func (os *BasicPoolOrderSource) BuyOrdersOver(price sdk.Dec) []amm.Order {
	// TODO: use providable x amount?
	amt := os.BuyAmountOver(price)
	if amt.IsZero() {
		return nil
	}
	quoteCoin := sdk.NewCoin(os.QuoteCoinDenom, amm.OfferCoinAmount(amm.Buy, price, amt))
	return []amm.Order{NewPoolOrder(os.PoolId, os.PoolReserveAddress, amm.Buy, price, amt, quoteCoin, os.BaseCoinDenom)}
}

func (os *BasicPoolOrderSource) SellOrdersUnder(price sdk.Dec) []amm.Order {
	amt := os.SellAmountUnder(price)
	if amt.IsZero() {
		return nil
	}
	baseCoin := sdk.NewCoin(os.BaseCoinDenom, amt)
	return []amm.Order{NewPoolOrder(os.PoolId, os.PoolReserveAddress, amm.Sell, price, amt, baseCoin, os.QuoteCoinDenom)}
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
