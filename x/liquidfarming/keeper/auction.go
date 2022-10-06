package keeper

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// PlaceBid handles types.MsgPlaceBid and stores bid object.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionId uint64, poolId uint64, bidder sdk.AccAddress, biddingCoin sdk.Coin) (types.Bid, error) {
	liquidFarm, found := k.GetLiquidFarm(ctx, poolId)
	if !found {
		return types.Bid{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "liquid farm by pool %d not found", poolId)
	}

	auction, found := k.GetRewardsAuction(ctx, auctionId, poolId)
	if !found {
		return types.Bid{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "auction by pool %d not found", poolId)
	}

	if auction.BiddingCoinDenom != biddingCoin.Denom {
		return types.Bid{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected denom %s, but got %s", auction.BiddingCoinDenom, biddingCoin.Denom)
	}

	if biddingCoin.Amount.LT(liquidFarm.MinBidAmount) {
		return types.Bid{}, sdkerrors.Wrapf(types.ErrSmallerThanMinimumAmount, "%s is smaller than %s", biddingCoin.Amount, liquidFarm.MinBidAmount)
	}

	winningBid, found := k.GetWinningBid(ctx, auctionId, poolId)
	if found {
		if biddingCoin.Amount.LTE(winningBid.Amount.Amount) {
			return types.Bid{}, sdkerrors.Wrapf(types.ErrNotBiggerThanWinningBidAmount, "%s is not bigger than %s", biddingCoin.Amount, winningBid.Amount.Amount)
		}
	}

	// Refund the previous bid if exists
	previousBid, found := k.GetBid(ctx, auctionId, bidder)
	if found {
		if err := k.bankKeeper.SendCoins(ctx, auction.GetPayingReserveAddress(), previousBid.GetBidder(), sdk.NewCoins(previousBid.Amount)); err != nil {
			return types.Bid{}, err
		}
		k.DeleteBid(ctx, previousBid)
	}

	if err := k.bankKeeper.SendCoins(ctx, bidder, auction.GetPayingReserveAddress(), sdk.NewCoins(biddingCoin)); err != nil {
		return types.Bid{}, err
	}

	bid := types.NewBid(
		poolId,
		bidder.String(),
		biddingCoin,
	)
	k.SetBid(ctx, bid)
	k.SetWinningBid(ctx, auction.Id, bid)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePlaceBid,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyAuctionId, strconv.FormatUint(auction.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyBidder, bidder.String()),
			sdk.NewAttribute(types.AttributeKeyBiddingCoin, biddingCoin.String()),
		),
	})

	return bid, nil
}

// RefundBid handles types.MsgRefundBid and refunds bid amount to the bidder and
// delete the bid object.
func (k Keeper) RefundBid(ctx sdk.Context, auctionId uint64, poolId uint64, bidder sdk.AccAddress) error {
	auction, found := k.GetRewardsAuction(ctx, auctionId, poolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "auction by pool %d not found", poolId)
	}

	winningBid, found := k.GetWinningBid(ctx, auctionId, poolId)
	if found && winningBid.Bidder == bidder.String() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "winning bid can't be refunded")
	}

	bid, found := k.GetBid(ctx, poolId, bidder)
	if !found {
		return sdkerrors.Wrap(sdkerrors.ErrNotFound, "bid not found")
	}

	if err := k.bankKeeper.SendCoins(ctx, auction.GetPayingReserveAddress(), bidder, sdk.NewCoins(bid.Amount)); err != nil {
		return err
	}

	k.DeleteBid(ctx, bid)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRefundBid,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyBidder, bidder.String()),
			sdk.NewAttribute(types.AttributeKeyRefundCoin, bid.Amount.String()),
		),
	})

	return nil
}

// getNextAuctionIdWithUpdate increments rewards auction id by one and store it.
func (k Keeper) getNextAuctionIdWithUpdate(ctx sdk.Context, poolId uint64) uint64 {
	id := k.GetLastRewardsAuctionId(ctx, poolId) + 1
	k.SetLastRewardsAuctionId(ctx, id, poolId)
	return id
}

// CreateRewardsAuction creates new rewards auction and store it.
func (k Keeper) CreateRewardsAuction(ctx sdk.Context, poolId uint64, duration time.Duration) {
	nextAuctionId := k.getNextAuctionIdWithUpdate(ctx, poolId)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(poolId)
	startTime := ctx.BlockTime()
	endTime := startTime.Add(duration)
	k.SetRewardsAuction(ctx, types.NewRewardsAuction(
		nextAuctionId,
		poolId,
		poolCoinDenom,
		startTime,
		endTime,
	))
}

// RefundAllBids refunds all bids at once as the rewards auction is finished and delete all bids.
func (k Keeper) RefundAllBids(ctx sdk.Context, auction types.RewardsAuction, includeWinningBid bool) error {
	winningBid, found := k.GetWinningBid(ctx, auction.Id, auction.PoolId)
	if !found {
		winningBid = types.Bid{}
	}

	inputs := []banktypes.Input{}
	outputs := []banktypes.Output{}
	for _, bid := range k.GetBidsByPoolId(ctx, auction.PoolId) {
		if includeWinningBid || bid.Bidder != winningBid.Bidder {
			inputs = append(inputs, banktypes.NewInput(auction.GetPayingReserveAddress(), sdk.NewCoins(bid.Amount)))
			outputs = append(outputs, banktypes.NewOutput(bid.GetBidder(), sdk.NewCoins(bid.Amount)))
		}
		k.DeleteBid(ctx, bid)
	}
	if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
		return err
	}

	return nil
}

// FinishRewardsAuction finishes the ongoing rewards auction.
func (k Keeper) FinishRewardsAuction(ctx sdk.Context, auction types.RewardsAuction, feeRate sdk.Dec) error {
	liquidFarmReserveAddr := types.LiquidFarmReserveAddress(auction.PoolId)
	auctionPayingReserveAddr := auction.GetPayingReserveAddress()
	poolCoinDenom := liquiditytypes.PoolCoinDenom(auction.PoolId)
	position, found := k.farmKeeper.GetPosition(ctx, liquidFarmReserveAddr, poolCoinDenom)
	if !found {
		position.FarmingAmount = sdk.ZeroInt() // TODO: develop simulation and debug
	}
	rewards := k.farmKeeper.Rewards(ctx, liquidFarmReserveAddr, poolCoinDenom)
	truncatedRewards, _ := rewards.TruncateDecimal()

	var (
		compoundingRewards types.CompoundingRewards
		status             types.AuctionStatus
	)

	// Finishing rewards auction can have two different scenarios depending on winning bid existence
	// When there is winning bid, harvest farming rewards first and send them to the winner and
	// stake the winning bid amount in the farming module for farmers so that it acts as auto compounding functionality.
	winningBid, found := k.GetWinningBid(ctx, auction.Id, auction.PoolId)
	if found {
		if _, err := k.farmKeeper.Harvest(ctx, liquidFarmReserveAddr, poolCoinDenom); err != nil {
			return err
		}

		deducted, err := types.DeductFees(auction.PoolId, truncatedRewards, feeRate)
		if err != nil {
			return err
		}

		if err := k.bankKeeper.SendCoins(ctx, liquidFarmReserveAddr, winningBid.GetBidder(), deducted); err != nil {
			return err
		}

		if err := k.RefundAllBids(ctx, auction, false); err != nil {
			return err
		}

		if err := k.bankKeeper.SendCoins(ctx, auctionPayingReserveAddr, liquidFarmReserveAddr, sdk.NewCoins(winningBid.Amount)); err != nil {
			return err
		}

		if _, err := k.farmKeeper.Farm(ctx, liquidFarmReserveAddr, winningBid.Amount); err != nil {
			return err
		}

		k.DeleteWinningBid(ctx, auction.Id, auction.PoolId)
		compoundingRewards.Amount = winningBid.Amount.Amount
		status = types.AuctionStatusFinished
	} else {
		compoundingRewards.Amount = sdk.ZeroInt()
		status = types.AuctionStatusSkipped
	}

	auction.SetWinner(winningBid.Bidder)
	auction.SetWinningAmount(winningBid.Amount)
	auction.SetRewards(truncatedRewards)
	auction.SetStatus(status)
	k.SetCompoundingRewards(ctx, auction.PoolId, compoundingRewards)
	k.SetRewardsAuction(ctx, auction)

	return nil
}
