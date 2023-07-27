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
	LastPublicPositionIdKey      = []byte{0xe1}
	LastRewardsAuctionEndTimeKey = []byte{0xe3}
	PublicPositionKeyPrefix      = []byte{0xe4}
	RewardsAuctionKeyPrefix      = []byte{0xe5}
	BidKeyPrefix                 = []byte{0xe6}
)

// GetPublicPositionKey returns the store key to retrieve the public position object
// by the given id.
func GetPublicPositionKey(publicPositionId uint64) []byte {
	return utils.Key(PublicPositionKeyPrefix, sdk.Uint64ToBigEndian(publicPositionId))
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
