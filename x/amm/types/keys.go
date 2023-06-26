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
	LastPoolIdKey                      = []byte{0x40}
	LastPositionIdKey                  = []byte{0x41}
	PoolKeyPrefix                      = []byte{0x42} // poolId => Pool
	PoolStateKeyPrefix                 = []byte{0x43} // poolId => PoolState
	PoolByReserveAddressIndexKeyPrefix = []byte{0x44} // reserveAddress => poolId
	PoolByMarketIndexKeyPrefix         = []byte{0x45} // marketId => poolId
	PositionKeyPrefix                  = []byte{0x46} // positionId => Position
	PositionByParamsIndexKeyPrefix     = []byte{0x47} // poolId + owner + lowerTick + upperTick => positionId
	PositionsByPoolIndexKeyPrefix      = []byte{0x48} // poolId + positionId => nil
	TickInfoKeyPrefix                  = []byte{0x49} // poolId + tick => TickInfo
	LastFarmingPlanIdKey               = []byte{0x4a}
	FarmingPlanKeyPrefix               = []byte{0x4b} // planId => FarmingPlan
	NumPrivateFarmingPlansKey          = []byte{0x4c}
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

func GetPoolByMarketIndexKey(marketId uint64) []byte {
	return utils.Key(PoolByMarketIndexKeyPrefix, sdk.Uint64ToBigEndian(marketId))
}

func GetPositionKey(positionId uint64) []byte {
	return utils.Key(PositionKeyPrefix, sdk.Uint64ToBigEndian(positionId))
}

func GetPositionByParamsIndexKey(ownerAddr sdk.AccAddress, poolId uint64, lowerTick, upperTick int32) []byte {
	return utils.Key(
		PositionByParamsIndexKeyPrefix,
		address.MustLengthPrefix(ownerAddr),
		sdk.Uint64ToBigEndian(poolId),
		TickToBytes(lowerTick),
		TickToBytes(upperTick))
}

func GetPositionsByOwnerIteratorPrefix(ownerAddr sdk.AccAddress) []byte {
	return utils.Key(
		PositionByParamsIndexKeyPrefix,
		address.MustLengthPrefix(ownerAddr))
}

func GetPositionsByPoolAndOwnerIteratorPrefix(ownerAddr sdk.AccAddress, poolId uint64) []byte {
	return utils.Key(
		PositionByParamsIndexKeyPrefix,
		address.MustLengthPrefix(ownerAddr),
		sdk.Uint64ToBigEndian(poolId))
}

func GetPositionsByPoolIndexKey(poolId, positionId uint64) []byte {
	return utils.Key(
		PositionsByPoolIndexKeyPrefix,
		sdk.Uint64ToBigEndian(poolId),
		sdk.Uint64ToBigEndian(positionId))
}

func GetPositionsByPoolIteratorPrefix(poolId uint64) []byte {
	return utils.Key(PositionsByPoolIndexKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

func GetTickInfoKey(poolId uint64, tick int32) []byte {
	return utils.Key(
		TickInfoKeyPrefix,
		sdk.Uint64ToBigEndian(poolId),
		TickToBytes(tick))
}

func GetTickInfosByPoolIteratorPrefix(poolId uint64) []byte {
	return utils.Key(TickInfoKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

func GetFarmingPlanKey(planId uint64) []byte {
	return utils.Key(FarmingPlanKeyPrefix, sdk.Uint64ToBigEndian(planId))
}

func ParsePositionsByPoolIndexKey(key []byte) (poolId, positionId uint64) {
	poolId = sdk.BigEndianToUint64(key[1:9])
	positionId = sdk.BigEndianToUint64(key[9:17])
	return
}

func ParseTickInfoKey(key []byte) (poolId uint64, tick int32) {
	poolId = sdk.BigEndianToUint64(key[1:9])
	tick = BytesToTick(key[9:])
	return
}

func TickToBytes(tick int32) []byte {
	bz := make([]byte, 5)
	if tick >= 0 {
		bz[0] = 1
	}
	binary.BigEndian.PutUint32(bz[1:], uint32(tick))
	return bz
}

func BytesToTick(bz []byte) int32 {
	return int32(binary.BigEndian.Uint32(bz[1:]))
}
