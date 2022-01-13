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

	PairKeyPrefix            = []byte{0xa5}
	PairIndexKeyPrefix       = []byte{0xa6}
	PairLookupIndexKeyPrefix = []byte{0xa7}

	PoolKeyPrefix                  = []byte{0xab}
	PoolByReserveAccIndexKeyPrefix = []byte{0xac}
	PoolByPairIndexKeyPrefix       = []byte{0xad}

	DepositRequestKeyPrefix    = []byte{0xb0}
	WithdrawRequestKeyPrefix   = []byte{0xb1}
	SwapRequestKeyPrefix       = []byte{0xb2}
	CancelSwapRequestKeyPrefix = []byte{0xb3}
)

// GetPairKey returns the store key to retrieve pair object from the pair id.
func GetPairKey(pairId uint64) []byte {
	return append(PairKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// GetPairIndexKey returns the index key to get a pair by denoms.
func GetPairIndexKey(denomX, denomY string) []byte {
	return append(append(PairIndexKeyPrefix, LengthPrefixString(denomX)...), LengthPrefixString(denomY)...)
}

// GetPairLookupIndexKey returns the index key to lookup pairs with given denoms.
func GetPairLookupIndexKey(denomA, denomB string, pairId uint64) []byte {
	return append(append(append(PairLookupIndexKeyPrefix, LengthPrefixString(denomA)...), LengthPrefixString(denomB)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetPairByDenomKeyPrefix returns the single denom index key.
func GetPairByDenomKeyPrefix(denom string) []byte {
	return append(PairIndexKeyPrefix, LengthPrefixString(denom)...)
}

// GetPoolKey returns the store key to retrieve pool object from the pool id.
func GetPoolKey(poolId uint64) []byte {
	return append(PoolKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetPoolByReserveAccIndexKey returns the index key to retrieve the particular pool.
func GetPoolByReserveAccIndexKey(reserveAcc sdk.AccAddress) []byte {
	return append(PoolByReserveAccIndexKeyPrefix, address.MustLengthPrefix(reserveAcc)...)
}

// GetPoolsByPairIndexKey returns the index key to retrieve pool id that is used to iterate pools.
func GetPoolsByPairIndexKey(pairId, poolId uint64) []byte {
	return append(append(PoolByPairIndexKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), sdk.Uint64ToBigEndian(poolId)...)
}

// GetPoolsByPairIndexKeyPrefix returns the store key to retrieve pool id to iterate pools.
func GetPoolsByPairIndexKeyPrefix(pairId uint64) []byte {
	return append(PoolByPairIndexKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// GetDepositRequestKey returns the store key to retrieve deposit request object from the pool id and request id.
func GetDepositRequestKey(poolId, id uint64) []byte {
	return append(append(DepositRequestKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetWithdrawRequestKey returns the store key to retrieve withdraw request object from the pool id and request id.
func GetWithdrawRequestKey(poolId, id uint64) []byte {
	return append(append(WithdrawRequestKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetSwapRequestKey returns the store key to retrieve swap request object from the pair id and request id.
func GetSwapRequestKey(pairId, id uint64) []byte {
	return append(append(SwapRequestKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetCancelSwapRequestKey returns the store key to retrieve cancel swap request object from pair id and request id.
func GetCancelSwapRequestKey(pairId, id uint64) []byte {
	return append(append(CancelSwapRequestKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), sdk.Uint64ToBigEndian(id)...)
}

// ParsePairByDenomIndexKey parses a pair by denom index key.
func ParsePairByDenomIndexKey(key []byte) (denomB string, pairId uint64) {
	if !bytes.HasPrefix(key, PairLookupIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}

	denomALen := key[1]
	denomBLen := key[2+denomALen]
	denomB = string(key[3+denomALen : 3+denomALen+denomBLen])
	pairId = sdk.BigEndianToUint64(key[3+denomALen+denomBLen:])

	return
}

// ParsePoolsByPairIndexKey parses a pool id from the index key.
func ParsePoolsByPairIndexKey(key []byte) (poolId uint64) {
	if !bytes.HasPrefix(key, PoolByPairIndexKeyPrefix) {
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
