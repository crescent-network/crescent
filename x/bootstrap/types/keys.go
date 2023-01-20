package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the bootstrap module
	ModuleName = "bootstrap"

	// RouterKey is the message router key for the bootstrap module
	RouterKey = ModuleName

	// StoreKey is the default store key for the bootstrap module
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the bootstrap module
	QuerierRoute = ModuleName
)

// keys for bootstrap store prefixes
var (
	BootstrapKeyPrefix              = []byte{0xc0}
	BootstrapIndexByPairIdKeyPrefix = []byte{0xc1}
	DepositKeyPrefix                = []byte{0xc2}

	IncentiveKeyPrefix = []byte{0xc5}
)

// GetBootstrapKey returns a key for a market maker record.
func GetBootstrapKey(mmAddr sdk.AccAddress, pairId uint64) []byte {
	return append(append(BootstrapKeyPrefix, address.MustLengthPrefix(mmAddr)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetBootstrapIndexByPairIdKey returns a key for a market maker record.
func GetBootstrapIndexByPairIdKey(pairId uint64, mmAddress sdk.AccAddress) []byte {
	return append(append(BootstrapIndexByPairIdKeyPrefix, sdk.Uint64ToBigEndian(pairId)...), mmAddress...)
}

// GetDepositKey returns a key for a market maker record.
func GetDepositKey(mmAddr sdk.AccAddress, pairId uint64) []byte {
	return append(append(DepositKeyPrefix, address.MustLengthPrefix(mmAddr)...), sdk.Uint64ToBigEndian(pairId)...)
}

// GetIncentiveKey returns kv indexing key of the incentive
func GetIncentiveKey(mmAddr sdk.AccAddress) []byte {
	return append(IncentiveKeyPrefix, mmAddr...)
}

// GetBootstrapByAddrPrefix returns a key prefix used to iterate
// market makers by a address.
func GetBootstrapByAddrPrefix(mmAddr sdk.AccAddress) []byte {
	return append(BootstrapKeyPrefix, address.MustLengthPrefix(mmAddr)...)
}

// GetBootstrapByPairIdPrefix returns a key prefix used to iterate
// market makers by a pair id.
func GetBootstrapByPairIdPrefix(pairId uint64) []byte {
	return append(BootstrapIndexByPairIdKeyPrefix, sdk.Uint64ToBigEndian(pairId)...)
}

// ParseBootstrapIndexByPairIdKey parses a market maker index by pair id key.
func ParseBootstrapIndexByPairIdKey(key []byte) (pairId uint64, mmAddr sdk.AccAddress) {
	if !bytes.HasPrefix(key, BootstrapIndexByPairIdKeyPrefix) {
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
