package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
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
	PoolKeyPrefix          = []byte{0x42} // poolId => Pool
	PoolIndexKeyPrefix     = []byte{0x43} // marketId + poolId => nil
	PositionKeyPrefix      = []byte{0x44} // positionId => Position
	PositionIndexKeyPrefix = []byte{0x45} // poolId + owner + lowerTick + upperTick => positionId
	TickInfoKeyPrefix      = []byte{0x46} // poolId + tick => TickInfo
)

func GetPoolKey(poolId uint64) []byte {
	return append(PoolKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

func GetPoolIndexKey(marketId string, poolId uint64) []byte {
	key := append(PoolIndexKeyPrefix, utils.LengthPrefixString(marketId)...)
	key = append(key, sdk.Uint64ToBigEndian(poolId)...)
	return key
}

func GetPoolsByMarketIdKeyPrefix(marketId string) []byte {
	return append(PoolIndexKeyPrefix, utils.LengthPrefixString(marketId)...)
}

func GetPositionKey(positionId uint64) []byte {
	return append(PositionKeyPrefix, sdk.Uint64ToBigEndian(positionId)...)
}

func GetPositionIndexKey(poolId uint64, ownerAddr sdk.AccAddress, lowerTick, upperTick int32) []byte {
	key := append(PositionIndexKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
	key = append(key, address.MustLengthPrefix(ownerAddr)...)
	key = append(key, tickToBytes(lowerTick)...)
	key = append(key, tickToBytes(upperTick)...)
	return key
}

func GetTickInfoKey(poolId uint64, tick int32) []byte {
	return append(append(TickInfoKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), tickToBytes(tick)...)
}

func GetTickInfoKeyPrefix(poolId uint64) []byte {
	return append(TickInfoKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

func ParsePoolIndexKey(key []byte) (marketId string, poolId uint64) {
	marketIdLen := key[1]
	marketId = string(key[2 : 2+marketIdLen])
	poolId = sdk.BigEndianToUint64(key[2+marketIdLen:])
	return
}

func ParseTickInfoKey(key []byte) (poolId uint64, tick int32) {
	poolId = sdk.BigEndianToUint64(key[1:9])
	tick = bytesToTick(key[9:])
	return
}

func tickToBytes(tick int32) []byte {
	bz := make([]byte, 5)
	if tick >= 0 {
		bz[0] = 1
	}
	binary.BigEndian.PutUint32(bz[1:], uint32(tick))
	return bz
}

func bytesToTick(bz []byte) int32 {
	return int32(binary.BigEndian.Uint32(bz[1:]))
}