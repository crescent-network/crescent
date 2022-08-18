package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the marketmaker module
	ModuleName = "marketmaker"

	// RouterKey is the message router key for the marketmaker module
	RouterKey = ModuleName

	// StoreKey is the default store key for the marketmaker module
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the marketmaker module
	QuerierRoute = ModuleName
)

// keys for farming store prefixes
var (
	MarketMakerKeyPrefix              = []byte{0xc0}
	MarketMakerIndexByPairIdKeyPrefix = []byte{0xc1}
	DepositKeyPrefix                  = []byte{0xc2}

	IncentiveKeyPrefix = []byte{0xc5}
)

// GetMarketMakerKey returns a key for a market maker record.
func GetMarketMakerKey(mmAddr sdk.AccAddress, pairId uint64) []byte {
	return append(append(MarketMakerKeyPrefix, address.MustLengthPrefix(mmAddr)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetMarketMakerIndexByPairIdKey returns a key for a market maker record.
func GetMarketMakerIndexByPairIdKey(pairId uint64, mmAddress sdk.AccAddress) []byte {
	return append(append(MarketMakerIndexByPairIdKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), mmAddress...)
}

// GetDepositKey returns a key for a market maker record.
func GetDepositKey(mmAddr sdk.AccAddress, pairId uint64) []byte {
	return append(append(DepositKeyPrefix, address.MustLengthPrefix(mmAddr)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetIncentiveKey returns kv indexing key of the incentive
func GetIncentiveKey(mmAddr sdk.AccAddress) []byte {
	return append(IncentiveKeyPrefix, mmAddr...)
}

// GetMarketMakerByAddrPrefix returns a key prefix used to iterate
// market makers by a address.
func GetMarketMakerByAddrPrefix(mmAddr sdk.AccAddress) []byte {
	return append(MarketMakerKeyPrefix, address.MustLengthPrefix(mmAddr)...)
}

// GetMarketMakerByPairIdPrefix returns a key prefix used to iterate
// market makers by a pair id.
func GetMarketMakerByPairIdPrefix(pairId uint64) []byte {
	return append(MarketMakerIndexByPairIdKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// ParseMarketMakerIndexByPairIdKey parses a market maker index by pair id key.
func ParseMarketMakerIndexByPairIdKey(key []byte) (pairId uint64, mmAddr sdk.AccAddress) {
	if !bytes.HasPrefix(key, MarketMakerIndexByPairIdKeyPrefix) {
		panic("key does not have proper prefix")
	}
	pairId = sdk.BigEndianToUint64(key[1:9])
	mmAddr = key[9:]
	return
}

// ParseDepositKey parses a deposit key.
func ParseDepositKey(key []byte) (mmAddr sdk.AccAddress, pairId uint64) {
	if !bytes.HasPrefix(key, DepositKeyPrefix) {
		panic("key does not have proper prefix")
	}
	addrLen := key[1]
	mmAddr = key[2 : 2+addrLen]
	pairId = sdk.BigEndianToUint64(key[2+addrLen:])
	return
}
