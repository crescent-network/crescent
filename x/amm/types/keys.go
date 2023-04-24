package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "amm"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	TStoreKey = "transient_amm"

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	LastPoolIdKey                      = []byte{0x40}
	LastPositionIdKey                  = []byte{0x41}
	PoolKeyPrefix                      = []byte{0x42} // poolId => Pool
	PoolStateKeyPrefix                 = []byte{0x43} // poolId => PoolState
	PoolByReserveAddressIndexKeyPrefix = []byte{0x44} // reserveAddress => poolId
	PoolsByMarketIndexKeyPrefix        = []byte{0x45} // marketId + poolId => nil
	PositionKeyPrefix                  = []byte{0x46} // positionId => Position
	PositionIndexKeyPrefix             = []byte{0x47} // poolId + owner + lowerTick + upperTick => positionId
	TickInfoKeyPrefix                  = []byte{0x48} // poolId + tick => TickInfo
)

func GetPoolKey(poolId uint64) []byte {
	return utils.Key(PoolKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

func GetPoolStateKey(poolId uint64) []byte {
	return utils.Key(PoolStateKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

func GetPoolByReserveAddressIndexKey(reserveAddr sdk.AccAddress) []byte {
	return utils.Key(PoolByReserveAddressIndexKeyPrefix, reserveAddr)
}

func GetPoolsByMarketIndexKey(marketId uint64, poolId uint64) []byte {
	return utils.Key(
		PoolsByMarketIndexKeyPrefix,
		sdk.Uint64ToBigEndian(marketId),
		sdk.Uint64ToBigEndian(poolId))
}

func GetPoolsByMarketIndexKeyPrefix(marketId uint64) []byte {
	return utils.Key(PoolsByMarketIndexKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetPositionKey(positionId uint64) []byte {
	return utils.Key(PositionKeyPrefix, sdk.Uint64ToBigEndian(positionId))
}

func GetPositionIndexKey(poolId uint64, ownerAddr sdk.AccAddress, lowerTick, upperTick int32) []byte {
	return utils.Key(
		PositionIndexKeyPrefix,
		sdk.Uint64ToBigEndian(poolId),
		address.MustLengthPrefix(ownerAddr),
		exchangetypes.TickToBytes(lowerTick),
		exchangetypes.TickToBytes(upperTick))
}

func GetTickInfoKey(poolId uint64, tick int32) []byte {
	return utils.Key(
		TickInfoKeyPrefix,
		sdk.Uint64ToBigEndian(poolId),
		exchangetypes.TickToBytes(tick))
}

func GetTickInfoKeyPrefix(poolId uint64) []byte {
	return utils.Key(TickInfoKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

func ParsePoolsByMarketIndexKey(key []byte) (marketId, poolId uint64) {
	marketId = sdk.BigEndianToUint64(key[1:9])
	poolId = sdk.BigEndianToUint64(key[9:17])
	return
}

func ParseTickInfoKey(key []byte) (poolId uint64, tick int32) {
	poolId = sdk.BigEndianToUint64(key[1:9])
	tick = exchangetypes.BytesToTick(key[9:])
	return
}
