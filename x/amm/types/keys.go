package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "amm"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	LastPoolIdKey          = []byte{0x40}
	LastPositionIdKey      = []byte{0x41}
	PoolKeyPrefix          = []byte{0x42}
	PositionKeyPrefix      = []byte{0x43}
	PositionIndexKeyPrefix = []byte{0x44}
	TickInfoKeyPrefix      = []byte{0x45}
)

func GetPoolKey(poolId uint64) []byte {
	return append(PoolKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

func GetPositionKey(positionId uint64) []byte {
	return append(PositionKeyPrefix, sdk.Uint64ToBigEndian(positionId)...)
}

func GetPositionIndexKey(poolId uint64, ownerAddr sdk.AccAddress, lowerTick, upperTick int32) []byte {
	key := append(PositionIndexKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
	key = append(key, address.MustLengthPrefix(ownerAddr)...)
	key = append(key, uint32ToBigEndian(uint32(lowerTick))...)
	key = append(key, uint32ToBigEndian(uint32(upperTick))...)
	return key
}

func GetTickInfoKey(poolId uint64, tick int32) []byte {
	return append(append(TickInfoKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), tickToBytes(tick)...)
}

func uint32ToBigEndian(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

func tickToBytes(tick int32) []byte {
	var sign byte
	if tick > 0 {
		sign = 1
	}
	return append([]byte{sign}, uint32ToBigEndian(uint32(tick))...)
}
