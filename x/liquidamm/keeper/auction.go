package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// PlaceBid handles types.MsgPlaceBid and stores bid object.
func (k Keeper) PlaceBid(
	ctx sdk.Context, bidderAddr sdk.AccAddress, publicPositionId, auctionId uint64, share sdk.Coin) (bid types.Bid, err error) {
	publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
	if !found {
		return bid, sdkerrors.Wrap(sdkerrors.ErrNotFound, "public position not found")
	}
	if shareDenom := types.ShareDenom(publicPositionId); share.Denom != shareDenom {
		return bid, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "share denom != %s", shareDenom)
	}
	if share.Amount.LT(publicPosition.MinBidAmount) {
		return bid, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "share amount must not be smaller than %s", publicPosition.MinBidAmount)
	}

	auction, found := k.GetRewardsAuction(ctx, publicPositionId, auctionId)
	if !found {
		return bid, sdkerrors.Wrap(sdkerrors.ErrNotFound, "rewards auction not found")
	}
	if auction.Status != types.AuctionStatusStarted {
		return bid, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "rewards auction is not started")
	}

	if auction.WinningBid != nil {
		if share.Amount.LTE(auction.WinningBid.Share.Amount) {
			return bid, sdkerrors.Wrapf(
				types.ErrInsufficientBidAmount,
				"share amount must be greater than winning bid's share %s", auction.WinningBid.Share.Amount)
		}
	}

	// Refund the previous bid if the bidder has placed bid before
	prevBid, found := k.GetBid(ctx, publicPosition.Id, auction.Id, bidderAddr)
	if found {
		if err := k.refundBid(ctx, publicPosition, prevBid); err != nil {
			return bid, err
		}
	}

	// Reserve bidding amount
	if err := k.bankKeeper.SendCoins(
		ctx, bidderAddr, sdk.MustAccAddressFromBech32(publicPosition.BidReserveAddress),
		sdk.NewCoins(share)); err != nil {
		return bid, err
	}

	bid = types.NewBid(publicPosition.Id, auction.Id, bidderAddr, share)
	k.SetBid(ctx, bid)
	auction.SetWinningBid(&bid)
	k.SetRewardsAuction(ctx, auction)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventPlaceBid{
		Bidder:           bidderAddr.String(),
		PublicPositionId: publicPositionId,
		RewardsAuctionId: auctionId,
		Share:            share,
	}); err != nil {
		return bid, err
	}
	return bid, nil
}

func (k Keeper) refundBid(ctx sdk.Context, publicPosition types.PublicPosition, bid types.Bid) error {
	if err := k.bankKeeper.SendCoins(
		ctx, sdk.MustAccAddressFromBech32(publicPosition.BidReserveAddress),
		sdk.MustAccAddressFromBech32(bid.Bidder), sdk.NewCoins(bid.Share)); err != nil {
		return err
	}
	k.DeleteBid(ctx, bid)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventBidRefunded{
		Bidder:           bid.Bidder,
		PublicPositionId: bid.PublicPositionId,
		RewardsAuctionId: bid.RewardsAuctionId,
		Share:            bid.Share,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetLastRewardsAuction(ctx sdk.Context, publicPositionId uint64) (auction types.RewardsAuction, found bool) {
	publicPosition, found := k.GetPublicPosition(ctx, publicPositionId)
	if !found {
		return auction, false
	}
	if publicPosition.LastRewardsAuctionId == 0 {
		return auction, false
	}
	return k.GetRewardsAuction(ctx, publicPosition.Id, publicPosition.LastRewardsAuctionId)
}

// I think we can combine it with GetLastRewardsAuction to create a generalized function.
func (k Keeper) GetPreviousRewardsAuction(ctx sdk.Context, publicPosition types.PublicPosition) (auction types.RewardsAuction, found bool) {
	if publicPosition.LastRewardsAuctionId <= 1 {
		return
	}
	return k.GetRewardsAuction(ctx, publicPosition.Id, publicPosition.LastRewardsAuctionId-1)
}

// AdvanceRewardsAuctions advances all rewards auctions' epoch by one and
// sets the next auction end time.
func (k Keeper) AdvanceRewardsAuctions(ctx sdk.Context, nextEndTime time.Time) (err error) {
	maxNumRecentAuctions := k.GetMaxNumRecentRewardsAuctions(ctx)
	k.IterateAllPublicPositions(ctx, func(publicPosition types.PublicPosition) (stop bool) {
		if publicPosition.LastRewardsAuctionId != 0 {
			auction, found := k.GetRewardsAuction(ctx, publicPosition.Id, publicPosition.LastRewardsAuctionId)
			if !found { // sanity check
				panic("rewards auction not found")
			}
			if auction.WinningBid == nil {
				if err = k.SkipRewardsAuction(ctx, publicPosition, auction); err != nil {
					return true
				}
			} else {
				if err = k.FinishRewardsAuction(ctx, publicPosition, auction); err != nil {
					return true
				}
			}
		}
		// Prune old rewards auctions.
		lastAuctionId := k.StartNewRewardsAuction(ctx, publicPosition, nextEndTime)
		k.IterateRewardsAuctionsByPublicPosition(ctx, publicPosition.Id, func(auction types.RewardsAuction) (stop bool) {
			if auction.Id+uint64(maxNumRecentAuctions) >= lastAuctionId {
				return true
			}
			k.DeleteRewardsAuction(ctx, auction) // Modifying kv store inside an iterator is undesirable. And the key prefix used by the iterator and DeleteRewardsAuction is the same.
			return false
		})
		return false
	})
	k.SetNextRewardsAuctionEndTime(ctx, nextEndTime)
	return
}

// StartNewRewardsAuction creates a new rewards auction and increment
// the public position's last rewards auction id.
func (k Keeper) StartNewRewardsAuction(ctx sdk.Context, publicPosition types.PublicPosition, endTime time.Time) (auctionId uint64) {
	publicPosition.LastRewardsAuctionId++
	auctionId = publicPosition.LastRewardsAuctionId
	k.SetPublicPosition(ctx, publicPosition)
	startTime := ctx.BlockTime()
	auction := types.NewRewardsAuction(
		publicPosition.Id, auctionId, startTime, endTime, types.AuctionStatusStarted)
	k.SetRewardsAuction(ctx, auction)
	return
}

// FinishRewardsAuction finishes ongoing rewards auction by looking up the existence of winning bid.
// Compound accumulated farming rewards for farmers and refund all bids that are placed for the auction if winning bid exists.
// If not, set the compounding rewards to zero and update the auction status AuctionStatusSkipped.
func (k Keeper) FinishRewardsAuction(ctx sdk.Context, publicPosition types.PublicPosition, auction types.RewardsAuction) error {
	if auction.WinningBid == nil { // sanity check
		panic("auction has no winning bid")
	}
	winningBid := *auction.WinningBid

	position := k.MustGetAMMPosition(ctx, publicPosition)
	fee, farmingRewards, err := k.ammKeeper.CollectibleCoins(ctx, position.Id)
	if err != nil {
		return err
	}
	rewards := fee.Add(farmingRewards...)
	var protocolFee sdk.Coins
	if rewards.IsAllPositive() {
		moduleAccAddr := k.GetModuleAddress()
		// First, collect all rewards.
		if err := k.ammKeeper.Collect(ctx, moduleAccAddr, moduleAccAddr, position.Id, rewards); err != nil {
			return err
		}
		var deductedRewards sdk.Coins
		deductedRewards, protocolFee = types.DeductFees(rewards, publicPosition.FeeRate)
		if deductedRewards.IsAllPositive() {
			// Then send deducted rewards to the winning bidder.
			winningBidderAddr := sdk.MustAccAddressFromBech32(winningBid.Bidder)
			if err := k.bankKeeper.SendCoins(ctx, moduleAccAddr, winningBidderAddr, deductedRewards); err != nil {
				return err
			}
		}
		// Fees have been accrued in the module account.
		// Now burn the winning bid's share.
		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, sdk.MustAccAddressFromBech32(publicPosition.BidReserveAddress),
			types.ModuleName, sdk.NewCoins(winningBid.Share)); err != nil {
			return err
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(winningBid.Share)); err != nil {
			return err
		}
		k.DeleteBid(ctx, winningBid)
	}

	k.IterateBidsByRewardsAuction(ctx, publicPosition.Id, auction.Id, func(bid types.Bid) (stop bool) {
		if err = k.refundBid(ctx, publicPosition, bid); err != nil {
			return true
		}
		return false
	})
	if err != nil {
		return err
	}

	auction.SetRewards(rewards)
	auction.SetFees(protocolFee)
	auction.SetStatus(types.AuctionStatusFinished)
	k.SetRewardsAuction(ctx, auction)

	// TODO: emit event

	return nil
}

// SkipRewardsAuction skips rewards auction since there is no bid.
func (k Keeper) SkipRewardsAuction(ctx sdk.Context, publicPosition types.PublicPosition, auction types.RewardsAuction) error {
	if auction.WinningBid != nil { // sanity check
		panic("auction has winning bid")
	}

	position, found := k.GetAMMPosition(ctx, publicPosition)
	var rewards sdk.Coins
	if found {
		fee, farmingRewards, err := k.ammKeeper.CollectibleCoins(ctx, position.Id)
		if err != nil {
			return err
		}
		rewards = fee.Add(farmingRewards...)
	}

	auction.SetRewards(rewards)
	auction.SetStatus(types.AuctionStatusSkipped)
	k.SetRewardsAuction(ctx, auction)

	// TODO: emit event

	return nil
}
