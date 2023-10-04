package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (k Keeper) GetLastPublicPositionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastPublicPositionIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastPublicPositionId(ctx sdk.Context, publicPositionId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastPublicPositionIdKey, sdk.Uint64ToBigEndian(publicPositionId))
}

func (k Keeper) GetNextPublicPositionIdWithUpdate(ctx sdk.Context) uint64 {
	publicPositionId := k.GetLastPublicPositionId(ctx)
	publicPositionId++
	k.SetLastPublicPositionId(ctx, publicPositionId)
	return publicPositionId
}

// GetPublicPosition returns public position object by the given id.
func (k Keeper) GetPublicPosition(ctx sdk.Context, publicPositionId uint64) (publicPosition types.PublicPosition, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPublicPositionKey(publicPositionId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &publicPosition)
	return publicPosition, true
}

func (k Keeper) LookupPublicPosition(ctx sdk.Context, publicPositionId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPublicPositionKey(publicPositionId))
}

// GetAllPublicPositions returns all public position objects stored in the store.
func (k Keeper) GetAllPublicPositions(ctx sdk.Context) (publicPositions []types.PublicPosition) {
	publicPositions = []types.PublicPosition{}
	k.IterateAllPublicPositions(ctx, func(publicPosition types.PublicPosition) (stop bool) {
		publicPositions = append(publicPositions, publicPosition)
		return false
	})
	return publicPositions
}

// SetPublicPosition stores public position object with the given id.
func (k Keeper) SetPublicPosition(ctx sdk.Context, publicPosition types.PublicPosition) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&publicPosition)
	store.Set(types.GetPublicPositionKey(publicPosition.Id), bz)
}

// IterateAllPublicPositions iterates through all public position objects
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function for each time.
func (k Keeper) IterateAllPublicPositions(ctx sdk.Context, cb func(publicPosition types.PublicPosition) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.PublicPositionKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var publicPosition types.PublicPosition
		k.cdc.MustUnmarshal(iter.Value(), &publicPosition)
		if cb(publicPosition) {
			break
		}
	}
}

func (k Keeper) IteratePublicPositionsByPool(ctx sdk.Context, poolId uint64, cb func(publicPosition types.PublicPosition) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetPublicPositionsByPoolIteratorPrefix(poolId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, publicPositionId := types.ParsePublicPositionsByPoolIndexKey(iter.Key())
		publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
		if !found { // sanity check
			panic("public position not found")
		}
		if cb(publicPosition) {
			break
		}
	}
}

// DeletePublicPosition deletes the public position object from the store.
func (k Keeper) DeletePublicPosition(ctx sdk.Context, publicPosition types.PublicPosition) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPublicPositionKey(publicPosition.Id))
}

func (k Keeper) SetPublicPositionsByPoolIndex(ctx sdk.Context, publicPosition types.PublicPosition) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPublicPositionsByPoolIndexKey(publicPosition.PoolId, publicPosition.Id), []byte{})
}

func (k Keeper) SetPublicPositionByParamsIndex(ctx sdk.Context, publicPosition types.PublicPosition) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPublicPositionByParamsIndexKey(
		k.GetModuleAddress(), publicPosition.PoolId, publicPosition.LowerTick, publicPosition.UpperTick), []byte{})
}

func (k Keeper) LookupPublicPositionByParams(ctx sdk.Context, poolId uint64, lowerTick, upperTick int32) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPublicPositionByParamsIndexKey(
		k.GetModuleAddress(), poolId, lowerTick, upperTick))
}

// GetNextRewardsAuctionEndTime returns the last rewards auction end time.
func (k Keeper) GetNextRewardsAuctionEndTime(ctx sdk.Context) (t time.Time, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastRewardsAuctionEndTimeKey)
	if bz == nil {
		return
	}
	t, err := sdk.ParseTimeBytes(bz)
	if err != nil {
		panic(err)
	}
	return t, true
}

// SetNextRewardsAuctionEndTime stores the last rewards auction end time.
func (k Keeper) SetNextRewardsAuctionEndTime(ctx sdk.Context, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastRewardsAuctionEndTimeKey, sdk.FormatTimeBytes(t))
}

// GetRewardsAuction returns the reward auction object by the given auction id pool id.
func (k Keeper) GetRewardsAuction(ctx sdk.Context, publicPositionId, auctionId uint64) (auction types.RewardsAuction, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRewardsAuctionKey(publicPositionId, auctionId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &auction)
	return auction, true
}

func (k Keeper) LookupRewardsAuction(ctx sdk.Context, publicPositionId, auctionId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetRewardsAuctionKey(publicPositionId, auctionId))
}

// SetRewardsAuction stores rewards auction.
func (k Keeper) SetRewardsAuction(ctx sdk.Context, auction types.RewardsAuction) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&auction)
	store.Set(types.GetRewardsAuctionKey(auction.PublicPositionId, auction.Id), bz)
}

// GetAllRewardsAuctions returns all rewards auctions in the store.
func (k Keeper) GetAllRewardsAuctions(ctx sdk.Context) (auctions []types.RewardsAuction) {
	auctions = []types.RewardsAuction{}
	k.IterateAllRewardsAuctions(ctx, func(auction types.RewardsAuction) (stop bool) {
		auctions = append(auctions, auction)
		return false
	})
	return auctions
}

// IterateAllRewardsAuctions iterates over all the stored auctions and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllRewardsAuctions(ctx sdk.Context, cb func(auction types.RewardsAuction) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardsAuctionKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var auction types.RewardsAuction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)
		if cb(auction) {
			break
		}
	}
}

func (k Keeper) IterateRewardsAuctionsByPublicPosition(ctx sdk.Context, publicPositionId uint64, cb func(auction types.RewardsAuction) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardsAuctionsByPublicPositionIteratorPrefix(publicPositionId))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var auction types.RewardsAuction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)
		if cb(auction) {
			break
		}
	}
}

func (k Keeper) DeleteRewardsAuction(ctx sdk.Context, auction types.RewardsAuction) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetRewardsAuctionKey(auction.PublicPositionId, auction.Id))
}

// GetBid returns the bid object by the given pool id and bidder address.
func (k Keeper) GetBid(ctx sdk.Context, publicPositionId, auctionId uint64, bidderAddr sdk.AccAddress) (bid types.Bid, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBidKey(publicPositionId, auctionId, bidderAddr))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &bid)
	return bid, true
}

// SetBid stores a bid object with the given pool id.
func (k Keeper) SetBid(ctx sdk.Context, bid types.Bid) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&bid)
	store.Set(
		types.GetBidKey(
			bid.PublicPositionId, bid.RewardsAuctionId, sdk.MustAccAddressFromBech32(bid.Bidder)), bz)
}

// GetAllBids returns all bids in the store.
func (k Keeper) GetAllBids(ctx sdk.Context) (bids []types.Bid) {
	bids = []types.Bid{}
	k.IterateAllBids(ctx, func(bid types.Bid) (stop bool) {
		bids = append(bids, bid)
		return false
	})
	return
}

// IterateAllBids iterates over all the stored bids and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllBids(ctx sdk.Context, cb func(bid types.Bid) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.BidKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var bid types.Bid
		k.cdc.MustUnmarshal(iterator.Value(), &bid)
		if cb(bid) {
			break
		}
	}
}

// In the kv store, there are only bids for auctions that are currently in progress.
// If req.AuctionId is not PublicPosition.LastRewardsAuctionId, the iterate is meaningless.
func (k Keeper) IterateBidsByRewardsAuction(ctx sdk.Context, publicPositionId, auctionId uint64, cb func(bid types.Bid) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetBidsByRewardsAuctionIteratorPrefix(publicPositionId, auctionId))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var bid types.Bid
		k.cdc.MustUnmarshal(iterator.Value(), &bid)
		if cb(bid) {
			break
		}
	}
}

// DeleteBid deletes the bid object.
func (k Keeper) DeleteBid(ctx sdk.Context, bid types.Bid) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(
		types.GetBidKey(
			bid.PublicPositionId, bid.RewardsAuctionId, sdk.MustAccAddressFromBech32(bid.Bidder)))
}
