package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (k Keeper) GetLastLiquidFarmId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastLiquidFarmIdKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLastLiquidFarmId(ctx sdk.Context, liquidFarmId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastLiquidFarmIdKey, sdk.Uint64ToBigEndian(liquidFarmId))
}

func (k Keeper) GetNextLiquidFarmIdWithUpdate(ctx sdk.Context) uint64 {
	liquidFarmId := k.GetLastLiquidFarmId(ctx)
	liquidFarmId++
	k.SetLastLiquidFarmId(ctx, liquidFarmId)
	return liquidFarmId
}

// GetLiquidFarm returns liquid farm object by the given id.
func (k Keeper) GetLiquidFarm(ctx sdk.Context, liquidFarmId uint64) (liquidFarm types.LiquidFarm, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLiquidFarmKey(liquidFarmId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &liquidFarm)
	return liquidFarm, true
}

func (k Keeper) LookupLiquidFarm(ctx sdk.Context, liquidFarmId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetLiquidFarmKey(liquidFarmId))
}

// GetAllLiquidFarms returns all liquid farm objects stored in the store.
func (k Keeper) GetAllLiquidFarms(ctx sdk.Context) (liquidFarms []types.LiquidFarm) {
	liquidFarms = []types.LiquidFarm{}
	k.IterateAllLiquidFarms(ctx, func(liquidFarm types.LiquidFarm) (stop bool) {
		liquidFarms = append(liquidFarms, liquidFarm)
		return false
	})
	return liquidFarms
}

// SetLiquidFarm stores liquid farm object with the given id.
func (k Keeper) SetLiquidFarm(ctx sdk.Context, liquidFarm types.LiquidFarm) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&liquidFarm)
	store.Set(types.GetLiquidFarmKey(liquidFarm.Id), bz)
}

// IterateAllLiquidFarms iterates through all liquid farm objects
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function for each time.
func (k Keeper) IterateAllLiquidFarms(ctx sdk.Context, cb func(liquidFarm types.LiquidFarm) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.LiquidFarmKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var liquidFarm types.LiquidFarm
		k.cdc.MustUnmarshal(iter.Value(), &liquidFarm)
		if cb(liquidFarm) {
			break
		}
	}
}

// DeleteLiquidFarm deletes the liquid farm object from the store.
func (k Keeper) DeleteLiquidFarm(ctx sdk.Context, liquidFarm types.LiquidFarm) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLiquidFarmKey(liquidFarm.Id))
}

// GetNextRewardsAuctionEndTime returns the last rewards auction end time.
func (k Keeper) GetNextRewardsAuctionEndTime(ctx sdk.Context) (t time.Time, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextRewardsAuctionEndTimeKey)
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
	store.Set(types.NextRewardsAuctionEndTimeKey, sdk.FormatTimeBytes(t))
}

// GetRewardsAuction returns the reward auction object by the given auction id pool id.
func (k Keeper) GetRewardsAuction(ctx sdk.Context, liquidFarmId, auctionId uint64) (auction types.RewardsAuction, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRewardsAuctionKey(liquidFarmId, auctionId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &auction)
	return auction, true
}

func (k Keeper) LookupRewardsAuction(ctx sdk.Context, liquidFarmId, auctionId uint64) (found bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetRewardsAuctionKey(liquidFarmId, auctionId))
}

// SetRewardsAuction stores rewards auction.
func (k Keeper) SetRewardsAuction(ctx sdk.Context, auction types.RewardsAuction) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&auction)
	store.Set(types.GetRewardsAuctionKey(auction.LiquidFarmId, auction.Id), bz)
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

func (k Keeper) IterateRewardsAuctionsByLiquidFarm(ctx sdk.Context, liquidFarmId uint64, cb func(auction types.RewardsAuction) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardsAuctionsByLiquidFarmIteratorPrefix(liquidFarmId))
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
	store.Delete(types.GetRewardsAuctionKey(auction.LiquidFarmId, auction.Id))
}

// GetBid returns the bid object by the given pool id and bidder address.
func (k Keeper) GetBid(ctx sdk.Context, liquidFarmId, auctionId uint64, bidderAddr sdk.AccAddress) (bid types.Bid, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBidKey(liquidFarmId, auctionId, bidderAddr))
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
			bid.LiquidFarmId, bid.RewardsAuctionId, sdk.MustAccAddressFromBech32(bid.Bidder)), bz)
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

func (k Keeper) IterateBidsByRewardsAuction(ctx sdk.Context, liquidFarmId, auctionId uint64, cb func(bid types.Bid) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetBidsByRewardsAuctionIteratorPrefix(liquidFarmId, auctionId))
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
			bid.LiquidFarmId, bid.RewardsAuctionId, sdk.MustAccAddressFromBech32(bid.Bidder)))
}
