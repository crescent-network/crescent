package types

import (
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
	LastBootstrapPoolIdKey = []byte{0xf1}

	BootstrapPoolKeyPrefix = []byte{0xf2}
	OrderKeyPrefix         = []byte{0xf3}
	OrderIndexKeyPrefix    = []byte{0xf4}

	// TODO:
	BootstrapPoolByReserveAddressIndexKeyPrefix = []byte{0xf5}
)

// GetBootstrapPoolKey returns a key for a bootstrap pool record.
func GetBootstrapPoolKey(id uint64) []byte {
	return append(BootstrapPoolKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

// GetOrderKey returns a key for a bootstrap order record.
func GetOrderKey(poolId, id uint64) []byte {
	return append(append(OrderKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetOrderIndexKeyByOrdererPrefix returns the index key prefix to iterate orders
// by an orderer.
func GetOrderIndexKeyByOrdererPrefix(orderer sdk.AccAddress) []byte {
	return append(OrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...)
}

// GetOrderIndexKeyByOrdererByPoolIdPrefix returns the index key prefix to iterate orders
// by an orderer.
func GetOrderIndexKeyByOrdererByPoolIdPrefix(orderer sdk.AccAddress, poolId uint64) []byte {
	return append(append(OrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...), sdk.Uint64ToBigEndian(poolId)...)
}

// GetOrderIndexKey returns the index key to map orders with an orderer.
func GetOrderIndexKey(orderer sdk.AccAddress, pairId, orderId uint64) []byte {
	return append(append(OrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...),
		sdk.Uint64ToBigEndian(pairId)...)
}

//// ParseBootstrapIndexByPairIdKey parses a market maker index by pair id key.
//func ParseBootstrapIndexByPairIdKey(key []byte) (pairId uint64, mmAddr sdk.AccAddress) {
//	if !bytes.HasPrefix(key, BootstrapIndexByPairIdKeyPrefix) {
//		panic("key does not have proper prefix")
//	}
//	pairId = sdk.BigEndianToUint64(key[1:9])
//	mmAddr = key[9:]
//	return
//}
//
//// ParseDepositKey parses a deposit key.
//func ParseDepositKey(key []byte) (mmAddr sdk.AccAddress, pairId uint64) {
//	if !bytes.HasPrefix(key, DepositKeyPrefix) {
//		panic("key does not have proper prefix")
//	}
//	addrLen := key[1]
//	mmAddr = key[2 : 2+addrLen]
//	pairId = sdk.BigEndianToUint64(key[2+addrLen:])
//	return
//}
