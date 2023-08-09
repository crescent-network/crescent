package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "liquidamm"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// keys for the store prefixes
var (
	LastPublicPositionIdKey             = []byte{0x81}
	LastRewardsAuctionEndTimeKey        = []byte{0x82}
	PublicPositionKeyPrefix             = []byte{0x83}
	PublicPositionsByPoolIndexKeyPrefix = []byte{0x84}
	RewardsAuctionKeyPrefix             = []byte{0x85}
	BidKeyPrefix                        = []byte{0x86}
)

// GetPublicPositionKey returns the store key to retrieve the public position object
// by the given id.
func GetPublicPositionKey(publicPositionId uint64) []byte {
	return utils.Key(PublicPositionKeyPrefix, sdk.Uint64ToBigEndian(publicPositionId))
}

func GetPublicPositionsByPoolIndexKey(poolId, publicPositionId uint64) []byte {
	return utils.Key(
		PublicPositionsByPoolIndexKeyPrefix,
		sdk.Uint64ToBigEndian(poolId),
		sdk.Uint64ToBigEndian(publicPositionId))
}

func GetPublicPositionsByPoolIteratorPrefix(poolId uint64) []byte {
	return utils.Key(PublicPositionsByPoolIndexKeyPrefix, sdk.Uint64ToBigEndian(poolId))
}

// GetRewardsAuctionKey returns the store key to retrieve the rewards auction object
// by the given public position id and rewards auction id.
func GetRewardsAuctionKey(publicPositionId, auctionId uint64) []byte {
	return utils.Key(
		RewardsAuctionKeyPrefix,
		sdk.Uint64ToBigEndian(publicPositionId),
		sdk.Uint64ToBigEndian(auctionId))
}

func GetRewardsAuctionsByPublicPositionIteratorPrefix(publicPositionId uint64) []byte {
	return utils.Key(RewardsAuctionKeyPrefix, sdk.Uint64ToBigEndian(publicPositionId))
}

// GetBidKey returns the store key to retrieve the bid object
// by the given public position id, rewards auction id and bidder address.
func GetBidKey(publicPositionId, auctionId uint64, bidderAddr sdk.AccAddress) []byte {
	return utils.Key(
		BidKeyPrefix,
		sdk.Uint64ToBigEndian(publicPositionId),
		sdk.Uint64ToBigEndian(auctionId),
		bidderAddr)
}

// GetBidsByRewardsAuctionIteratorPrefix returns the prefix to iterate all bids
// by the given rewards auction id.
func GetBidsByRewardsAuctionIteratorPrefix(publicPositionId, auctionId uint64) []byte {
	return utils.Key(
		BidKeyPrefix,
		sdk.Uint64ToBigEndian(publicPositionId),
		sdk.Uint64ToBigEndian(auctionId))
}

func ParsePublicPositionsByPoolIndexKey(key []byte) (poolId, publicPositionId uint64) {
	poolId = sdk.BigEndianToUint64(key[1:9])
	publicPositionId = sdk.BigEndianToUint64(key[9:17])
	return
}
