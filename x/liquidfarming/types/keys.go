package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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
	LastRewardsAuctionEndTimeKey  = []byte{0xe1} // key to retrieve the auction end time
	LastRewardsAuctionIdKeyPrefix = []byte{0xe2}
	LiquidFarmKeyPrefix           = []byte{0xe3}
	CompoundingRewardsKeyPrefix   = []byte{0xe4}
	RewardsAuctionKeyPrefix       = []byte{0xe5}
	BidKeyPrefix                  = []byte{0xe6}
	WinningBidKeyPrefix           = []byte{0xe7}
)

// GetLastRewardsAuctionIdKey returns the store key to retrieve the last rewards auction
// by the given pool id.
func GetLastRewardsAuctionIdKey(poolId uint64) []byte {
	return append(LastRewardsAuctionIdKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetLiquidFarmKey returns the store key to retrieve the liquid farm object
// by the given pool id.
func GetLiquidFarmKey(poolId uint64) []byte {
	return append(LiquidFarmKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetCompoundingRewardsKey returns the store key to retrieve the compounding rewards object
// by the given pool id.
func GetCompoundingRewardsKey(poolId uint64) []byte {
	return append(CompoundingRewardsKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetRewardsAuctionKey returns the store key to retrieve the rewards auction object
// by the given pool id and auction id.
func GetRewardsAuctionKey(auctionId, poolId uint64) []byte {
	return append(append(RewardsAuctionKeyPrefix, sdk.Uint64ToBigEndian(auctionId)...), sdk.Uint64ToBigEndian(poolId)...)
}

// GetBidKey returns the store key to retrieve the bid object
// by the given pool id and bidder address.
func GetBidKey(poolId uint64, bidder sdk.AccAddress) []byte {
	return append(append(BidKeyPrefix, sdk.Uint64ToBigEndian(poolId)...), address.MustLengthPrefix(bidder)...)
}

// GetBidByPoolIdPrefix returns the prefix to iterate all bids
// by the given pool id.
func GetBidByPoolIdPrefix(poolId uint64) []byte {
	return append(BidKeyPrefix, sdk.Uint64ToBigEndian(poolId)...)
}

// GetWinningBidKey returns the store key to retrieve the winning bid
// by the given pool id and the auction id.
func GetWinningBidKey(auctionId uint64, poolId uint64) []byte {
	return append(append(WinningBidKeyPrefix, sdk.Uint64ToBigEndian(auctionId)...), sdk.Uint64ToBigEndian(poolId)...)
}

// LengthPrefixTimeBytes returns length-prefixed bytes representation
// of time.Time.
func LengthPrefixTimeBytes(t time.Time) []byte {
	bz := sdk.FormatTimeBytes(t)
	return append([]byte{byte(len(bz))}, bz...)
}
