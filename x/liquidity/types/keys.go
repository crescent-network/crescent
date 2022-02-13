package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "liquidity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	LastPairIdKey = []byte{0xa0} // key for the latest pair id
	LastPoolIdKey = []byte{0xa1} // key for the latest pool id

	PairKeyPrefix               = []byte{0xa5}
	PairIndexKeyPrefix          = []byte{0xa6}
	PairsByDenomsIndexKeyPrefix = []byte{0xa7}

	PoolKeyPrefix                      = []byte{0xab}
	PoolByReserveAddressIndexKeyPrefix = []byte{0xac}
	PoolsByPairIndexKeyPrefix          = []byte{0xad}

	DepositRequestKeyPrefix  = []byte{0xb0}
	WithdrawRequestKeyPrefix = []byte{0xb1}
	OrderKeyPrefix           = []byte{0xb2}
)

// GetPairKey returns the store key to retrieve pair object from the pair id.
func GetPairKey(pairId uint64) []byte {
	return append(PairKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// GetPairIndexKey returns the index key to get a pair by denoms.
func GetPairIndexKey(baseCoinDenom, quoteCoinDenom string) []byte {
	return append(append(PairIndexKeyPrefix, LengthPrefixString(baseCoinDenom)...), LengthPrefixString(quoteCoinDenom)...)
}

// GetPairsByDenomsIndexKey returns the index key to lookup pairs with given denoms.
func GetPairsByDenomsIndexKey(denomA, denomB string, pairId uint64) []byte {
	return append(append(append(PairsByDenomsIndexKeyPrefix, LengthPrefixString(denomA)...), LengthPrefixString(denomB)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetPairsByDenomIndexKeyPrefix returns the index key prefix to lookup pairs with given denom.
func GetPairsByDenomIndexKeyPrefix(denomA string) []byte {
	return append(PairsByDenomsIndexKeyPrefix, LengthPrefixString(denomA)...)
}

// GetPairsByDenomsIndexKeyPrefix returns the index key prefix to lookup pairs with given denoms.
func GetPairsByDenomsIndexKeyPrefix(denomA, denomB string) []byte {
	return append(append(PairsByDenomsIndexKeyPrefix, LengthPrefixString(denomA)...), LengthPrefixString(denomB)...)
}

// GetPoolKey returns the store key to retrieve pool object from the pool id.
func GetPoolKey(poolId uint64) []byte {
	return append(PoolKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetPoolByReserveAddressIndexKey returns the index key to retrieve the particular pool.
func GetPoolByReserveAddressIndexKey(reserveAddr sdk.AccAddress) []byte {
	return append(PoolByReserveAddressIndexKeyPrefix, address.MustLengthPrefix(reserveAddr)...)
}

// GetPoolsByPairIndexKey returns the index key to retrieve pool id that is used to iterate pools.
func GetPoolsByPairIndexKey(pairId, poolId uint64) []byte {
	return append(append(PoolsByPairIndexKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), sdk.Uint64ToBigEndian(poolId)...)
}

// GetPoolsByPairIndexKeyPrefix returns the store key to retrieve pool id to iterate pools.
func GetPoolsByPairIndexKeyPrefix(pairId uint64) []byte {
	return append(PoolsByPairIndexKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// GetDepositRequestKey returns the store key to retrieve deposit request object from the pool id and request id.
func GetDepositRequestKey(poolId, id uint64) []byte {
	return append(append(DepositRequestKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetWithdrawRequestKey returns the store key to retrieve withdraw request object from the pool id and request id.
func GetWithdrawRequestKey(poolId, id uint64) []byte {
	return append(append(WithdrawRequestKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetOrderKey returns the store key to retrieve order object from the pair id and request id.
func GetOrderKey(pairId, id uint64) []byte {
	return append(append(OrderKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetOrdersByPairKeyPrefix returns the store key to iterate orders by pair.
func GetOrdersByPairKeyPrefix(pairId uint64) []byte {
	return append(OrderKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// ParsePairsByDenomsIndexKey parses a pair by denom index key.
func ParsePairsByDenomsIndexKey(key []byte) (denomA, denomB string, pairId uint64) {
	if !bytes.HasPrefix(key, PairsByDenomsIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}

	denomALen := key[1]
	denomA = string(key[2 : 2+denomALen])
	denomBLen := key[2+denomALen]
	denomB = string(key[3+denomALen : 3+denomALen+denomBLen])
	pairId = sdk.BigEndianToUint64(key[3+denomALen+denomBLen:])

	return
}

// ParsePoolsByPairIndexKey parses a pool id from the index key.
func ParsePoolsByPairIndexKey(key []byte) (poolId uint64) {
	if !bytes.HasPrefix(key, PoolsByPairIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}

	bytesLen := 8
	poolId = sdk.BigEndianToUint64(key[1+bytesLen:])
	return
}

// LengthPrefixString returns length-prefixed bytes representation
// of a string.
func LengthPrefixString(s string) []byte {
	bz := []byte(s)
	bzLen := len(bz)
	if bzLen == 0 {
		return bz
	}
	return append([]byte{byte(bzLen)}, bz...)
}
