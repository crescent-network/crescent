package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/crescent-network/crescent/x/farming/types"
)

var (
	poolCoinDenomRegexp = regexp.MustCompile(`^pool([1-9]\d*)$`)
)

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

// ParsePoolCoinDenom parses a pool coin denom and returns the pool id.
func ParsePoolCoinDenom(denom string) (poolId uint64, err error) {
	chunks := poolCoinDenomRegexp.FindStringSubmatch(denom)
	if len(chunks) == 0 {
		return 0, fmt.Errorf("%s is not a pool coin denom", denom)
	}
	poolId, err = strconv.ParseUint(chunks[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse pool id: %w", err)
	}
	return poolId, nil
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
