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
	LastBootstrapPoolIdKey = []byte{0xf1}
	BootstrapPoolKeyPrefix = []byte{0xf2}
	OrderKeyPrefix         = []byte{0xf3}
	OrderIndexKeyPrefix    = []byte{0xf4}
	LastOrderIdKeyPrefix   = []byte{0xf5}
)

// GetBootstrapPoolKey returns a key for a bootstrap pool record.
func GetBootstrapPoolKey(id uint64) []byte {
	return append(BootstrapPoolKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}

// GetOrderKey returns a key for a bootstrap order record.
func GetOrderKey(poolId, id uint64) []byte {
	return append(append(OrderKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(id)...)
}

// GetOrdersByPoolKeyPrefix returns the store key to iterate orders by pool.
func GetOrdersByPoolKeyPrefix(poolId uint64) []byte {
	return append(OrderKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetOrderIndexKeyByOrdererPrefix returns the index key prefix to iterate orders
// by an orderer.
func GetOrderIndexKeyByOrdererPrefix(orderer sdk.AccAddress) []byte {
	return append(OrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...)
}

// GetOrderIndexKey returns the index key to map orders with an orderer.
func GetOrderIndexKey(orderer sdk.AccAddress, poolId, orderId uint64) []byte {
	return append(append(append(OrderIndexKeyPrefix, address.MustLengthPrefix(orderer)...),
		sdk.Uint64ToBigEndian(poolId)...), sdk.Uint64ToBigEndian(orderId)...)
}

// ParseOrderIndexKey parses an order index key.
func ParseOrderIndexKey(key []byte) (orderer sdk.AccAddress, poolId, orderId uint64) {
	if !bytes.HasPrefix(key, OrderIndexKeyPrefix) {
		panic("key does not have proper prefix")
	}

	addrLen := key[1]
	orderer = key[2 : 2+addrLen]
	poolId = sdk.BigEndianToUint64(key[2+addrLen : 2+addrLen+8])
	orderId = sdk.BigEndianToUint64(key[2+addrLen+8:])
	return
}

// GetLastOrderIdIndexKey returns the index key to last order id of the pool.
func GetLastOrderIdIndexKey(poolId uint64) []byte {
	return append(LastOrderIdKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}
