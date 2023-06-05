package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// PlaceBid handles types.MsgPlaceBid and stores bid object.
func (k Keeper) PlaceBid(
	ctx sdk.Context, bidderAddr sdk.AccAddress, liquidFarmId, auctionId uint64, share sdk.Coin) (bid types.Bid, err error) {
	liquidFarm, found := k.GetLiquidFarm(ctx, liquidFarmId)
	if !found {
		return bid, sdkerrors.Wrap(sdkerrors.ErrNotFound, "liquid farm not found")
	}
	if shareDenom := types.ShareDenom(liquidFarmId); share.Denom != shareDenom {
		return bid, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "share denom != %s", shareDenom)
	}
	if share.Amount.LT(liquidFarm.MinBidAmount) {
		return bid, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "share amount must not be smaller than %s", liquidFarm.MinBidAmount)
	}

	auction, found := k.GetRewardsAuction(ctx, liquidFarmId, auctionId)
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
	prevBid, found := k.GetBid(ctx, liquidFarm.Id, auction.Id, bidderAddr)
	if found {
		if err := k.refundBid(ctx, liquidFarm, prevBid); err != nil {
			return bid, err
		}
	}

	// Reserve bidding amount
	if err := k.bankKeeper.SendCoins(
		ctx, bidderAddr, sdk.MustAccAddressFromBech32(liquidFarm.BidReserveAddress),
		sdk.NewCoins(share)); err != nil {
		return bid, err
	}

	bid = types.NewBid(liquidFarm.Id, auction.Id, bidderAddr, share)
	k.SetBid(ctx, bid)
	auction.SetWinningBid(&bid)
	k.SetRewardsAuction(ctx, auction)

	// TODO: emit typed event
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypePlaceBid,
	//		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
	//		sdk.NewAttribute(types.AttributeKeyAuctionId, strconv.FormatUint(auction.Id, 10)),
	//		sdk.NewAttribute(types.AttributeKeyBidder, bidder.String()),
	//		sdk.NewAttribute(types.AttributeKeyBiddingCoin, biddingCoin.String()),
	//	),
	//})

	return bid, nil
}

func (k Keeper) refundBid(ctx sdk.Context, liquidFarm types.LiquidFarm, bid types.Bid) error {
	if err := k.bankKeeper.SendCoins(
		ctx, sdk.MustAccAddressFromBech32(liquidFarm.BidReserveAddress),
		sdk.MustAccAddressFromBech32(bid.Bidder), sdk.NewCoins(bid.Share)); err != nil {
		return err
	}
	k.DeleteBid(ctx, bid)

	// TODO: emit typed event
	//ctx.EventManager().EmitEvents(sdk.Events{
	//	sdk.NewEvent(
	//		types.EventTypeRefundBid,
	//		sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
	//		sdk.NewAttribute(types.AttributeKeyBidder, bidder.String()),
	//		sdk.NewAttribute(types.AttributeKeyRefundCoin, bid.Amount.String()),
	//	),
	//})
	return nil
}

func (k Keeper) GetLastRewardsAuction(ctx sdk.Context, liquidFarmId uint64) (auction types.RewardsAuction, found bool) {
	liquidFarm, found := k.GetLiquidFarm(ctx, liquidFarmId)
	if !found {
		return auction, false
	}
	if liquidFarm.LastRewardsAuctionId == 0 {
		return auction, false
	}
	return k.GetRewardsAuction(ctx, liquidFarm.Id, liquidFarm.LastRewardsAuctionId)
}

func (k Keeper) GetPreviousRewardsAuction(ctx sdk.Context, liquidFarm types.LiquidFarm) (auction types.RewardsAuction, found bool) {
	if liquidFarm.LastRewardsAuctionId <= 1 {
		return
	}
	return k.GetRewardsAuction(ctx, liquidFarm.Id, liquidFarm.LastRewardsAuctionId-1)
}

// AdvanceRewardsAuctions advances all rewards auctions' epoch by one and
// sets the next auction end time.
func (k Keeper) AdvanceRewardsAuctions(ctx sdk.Context, nextEndTime time.Time) (err error) {
	k.IterateAllLiquidFarms(ctx, func(liquidFarm types.LiquidFarm) (stop bool) {
		if liquidFarm.LastRewardsAuctionId != 0 {
			auction, found := k.GetRewardsAuction(ctx, liquidFarm.Id, liquidFarm.LastRewardsAuctionId)
			if !found { // sanity check
				panic("rewards auction not found")
			}
			if auction.WinningBid == nil {
				if err = k.SkipRewardsAuction(ctx, liquidFarm, auction); err != nil {
					return true
				}
			} else {
				if err = k.FinishRewardsAuction(ctx, liquidFarm, auction); err != nil {
					return true
				}
			}
		}
		k.StartNewRewardsAuction(ctx, liquidFarm, nextEndTime)
		return false
	})
	k.SetNextRewardsAuctionEndTime(ctx, nextEndTime)
	return
}

// StartNewRewardsAuction creates a new rewards auction and increment
// the liquid farm's last rewards auction id.
func (k Keeper) StartNewRewardsAuction(ctx sdk.Context, liquidFarm types.LiquidFarm, endTime time.Time) {
	liquidFarm.LastRewardsAuctionId++
	auctionId := liquidFarm.LastRewardsAuctionId
	k.SetLiquidFarm(ctx, liquidFarm)
	startTime := ctx.BlockTime()
	auction := types.NewRewardsAuction(
		liquidFarm.Id, auctionId, startTime, endTime, types.AuctionStatusStarted)
	k.SetRewardsAuction(ctx, auction)
}

// FinishRewardsAuction finishes ongoing rewards auction by looking up the existence of winning bid.
// Compound accumulated farming rewards for farmers and refund all bids that are placed for the auction if winning bid exists.
// If not, set the compounding rewards to zero and update the auction status AuctionStatusSkipped.
func (k Keeper) FinishRewardsAuction(ctx sdk.Context, liquidFarm types.LiquidFarm, auction types.RewardsAuction) error {
	if auction.WinningBid == nil { // sanity check
		panic("auction has no winning bid")
	}
	winningBid := *auction.WinningBid

	position := k.MustGetLiquidFarmPosition(ctx, liquidFarm)
	rewards, err := k.ammKeeper.CollectibleCoins(ctx, position.Id)
	if err != nil {
		return err
	}
	var fees sdk.Coins
	if rewards.IsAllPositive() {
		moduleAccAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
		// First, collect all rewards.
		if err := k.ammKeeper.Collect(ctx, moduleAccAddr, moduleAccAddr, position.Id, rewards); err != nil {
			return err
		}
		var deductedRewards sdk.Coins
		deductedRewards, fees = types.DeductFees(rewards, liquidFarm.FeeRate)
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
			ctx, sdk.MustAccAddressFromBech32(liquidFarm.BidReserveAddress),
			types.ModuleName, sdk.NewCoins(winningBid.Share)); err != nil {
			return err
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(winningBid.Share)); err != nil {
			return err
		}
		k.DeleteBid(ctx, winningBid)
	}

	k.IterateBidsByRewardsAuction(ctx, liquidFarm.Id, auction.Id, func(bid types.Bid) (stop bool) {
		if err = k.refundBid(ctx, liquidFarm, bid); err != nil {
			return true
		}
		return false
	})
	if err != nil {
		return err
	}

	auction.SetRewards(rewards)
	auction.SetFees(fees)
	auction.SetStatus(types.AuctionStatusFinished)
	k.SetRewardsAuction(ctx, auction)

	// TODO: emit event

	return nil
}

// SkipRewardsAuction skips rewards auction since there is no bid.
func (k Keeper) SkipRewardsAuction(ctx sdk.Context, liquidFarm types.LiquidFarm, auction types.RewardsAuction) error {
	if auction.WinningBid != nil { // sanity check
		panic("auction has winning bid")
	}

	position, found := k.GetLiquidFarmPosition(ctx, liquidFarm)
	var rewards sdk.Coins
	if found {
		var err error
		rewards, err = k.ammKeeper.CollectibleCoins(ctx, position.Id)
		if err != nil {
			return err
		}
	}

	auction.SetRewards(rewards)
	auction.SetStatus(types.AuctionStatusSkipped)
	k.SetRewardsAuction(ctx, auction)

	// TODO: emit event

	return nil
}
