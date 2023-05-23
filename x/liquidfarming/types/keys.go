package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "liquidfarming"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// keys for the store prefixes
var (
	LastLiquidFarmIdKey          = []byte{0xe1}
	LastRewardsAuctionEndTimeKey = []byte{0xe3}
	LiquidFarmKeyPrefix          = []byte{0xe4}
	RewardsAuctionKeyPrefix      = []byte{0xe5}
	BidKeyPrefix                 = []byte{0xe6}
	PreviousWinningBidKeyPrefix  = []byte{0xe7}
)

// GetLiquidFarmKey returns the store key to retrieve the liquid farm object
// by the given id.
func GetLiquidFarmKey(liquidFarmId uint64) []byte {
	return utils.Key(LiquidFarmKeyPrefix, sdk.Uint64ToBigEndian(liquidFarmId))
}

// GetRewardsAuctionKey returns the store key to retrieve the rewards auction object
// by the given liquid farm id and rewards auction id.
func GetRewardsAuctionKey(liquidFarmId, auctionId uint64) []byte {
	return utils.Key(
		RewardsAuctionKeyPrefix,
		sdk.Uint64ToBigEndian(liquidFarmId),
		sdk.Uint64ToBigEndian(auctionId))
}

// GetBidKey returns the store key to retrieve the bid object
// by the given liquid farm id, rewards auction id and bidder address.
func GetBidKey(liquidFarmId, auctionId uint64, bidderAddr sdk.AccAddress) []byte {
	return utils.Key(
		BidKeyPrefix,
		sdk.Uint64ToBigEndian(liquidFarmId),
		sdk.Uint64ToBigEndian(auctionId),
		bidderAddr)
}

// GetBidsByRewardsAuctionIteratorPrefix returns the prefix to iterate all bids
// by the given rewards auction id.
func GetBidsByRewardsAuctionIteratorPrefix(liquidFarmId, auctionId uint64) []byte {
	return utils.Key(
		BidKeyPrefix,
		sdk.Uint64ToBigEndian(liquidFarmId),
		sdk.Uint64ToBigEndian(auctionId))
}

// GetPreviousWinningBidKey returns the store key to retrieve the previous winning bid
// by the given liquid farm id.
func GetPreviousWinningBidKey(liquidFarmId uint64) []byte {
	return utils.Key(PreviousWinningBidKeyPrefix, sdk.Uint64ToBigEndian(liquidFarmId))
}
