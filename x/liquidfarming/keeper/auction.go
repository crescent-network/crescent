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

// CreateRewardsAuction creates new rewards auction and store it.
func (k Keeper) CreateRewardsAuction(ctx sdk.Context, poolId uint64, endTime time.Time) {
	k.SetRewardsAuction(ctx, types.NewRewardsAuction(
		k.getNextAuctionIdWithUpdate(ctx, poolId),
		poolId,
		ctx.BlockTime(),
		endTime,
	))
}

// FinishRewardsAuction finishes ongoing rewards auction by looking up the existence of winning bid.
// Compound accumulated farming rewards for farmers and refund all bids that are placed for the auction if winning bid exists.
// If not, set the compounding rewards to zero and update the auction status AuctionStatusSkipped.
func (k Keeper) FinishRewardsAuction(ctx sdk.Context, auction types.RewardsAuction, feeRate sdk.Dec) error {
	liquidFarmReserveAddr := types.LiquidFarmReserveAddress(auction.PoolId)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(auction.PoolId)
	farmingRewards := k.lpfarmKeeper.Rewards(ctx, liquidFarmReserveAddr, poolCoinDenom)
	truncatedRewards, _ := farmingRewards.TruncateDecimal() // TODO: farm module may use sdk.DecCoins for sdk.Coins in the future

	winningBid, found := k.GetWinningBid(ctx, auction.Id, auction.PoolId)
	if !found {
		k.skipRewardsAuction(ctx, truncatedRewards, auction)
	} else {
		_, found = k.lpfarmKeeper.GetPosition(ctx, liquidFarmReserveAddr, poolCoinDenom)
		if found {
			if err := k.payoutRewards(ctx, auction.PoolId, feeRate, liquidFarmReserveAddr, poolCoinDenom, winningBid); err != nil {
				return err
			}
		}

		// Compound rewards even if there is no one farmed.
		auctionPayingReserveAddr := auction.GetPayingReserveAddress()
		if err := k.compoundRewards(ctx, auctionPayingReserveAddr, liquidFarmReserveAddr, winningBid.Amount); err != nil {
			return err
		}

		if err := k.refundAllBids(ctx, auction, false); err != nil {
			return err
		}

		auction.SetWinner(winningBid.Bidder)
		auction.SetWinningAmount(winningBid.Amount)
		auction.SetRewards(truncatedRewards)
		auction.SetStatus(types.AuctionStatusFinished)
		k.SetRewardsAuction(ctx, auction)
		k.SetCompoundingRewards(ctx, auction.PoolId, types.CompoundingRewards{
			Amount: winningBid.Amount.Amount,
		})
	}

	return nil
}

// getNextAuctionIdWithUpdate increments rewards auction id by one and store it.
func (k Keeper) getNextAuctionIdWithUpdate(ctx sdk.Context, poolId uint64) uint64 {
	auctionId := k.GetLastRewardsAuctionId(ctx, poolId) + 1
	k.SetLastRewardsAuctionId(ctx, auctionId, poolId)
	return auctionId
}

func (k Keeper) skipRewardsAuction(ctx sdk.Context, rewards sdk.Coins, auction types.RewardsAuction) {
	auction.SetRewards(rewards)
	auction.SetStatus(types.AuctionStatusSkipped)
	k.SetRewardsAuction(ctx, auction)
	k.SetCompoundingRewards(ctx, auction.PoolId, types.CompoundingRewards{
		Amount: sdk.ZeroInt(),
	})
}

// refundAllBids refunds all bids at once as the rewards auction is finished and delete all bids.
func (k Keeper) refundAllBids(ctx sdk.Context, auction types.RewardsAuction, includeWinningBid bool) error {
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

	if includeWinningBid {
		k.DeleteWinningBid(ctx, auction.Id, auction.PoolId)
	}

	return nil
}

// payoutRewards pays accumulated farming rewards to the winner for the auction.
// It first harvests farming rewards from the farm module and calculates sum of rewards from
// both withdrawn rewards reserve and liquid farm reserve accounts.
// Then it deducts fees from the rewards by the fee rate and sends it to the winner.
func (k Keeper) payoutRewards(
	ctx sdk.Context,
	poolId uint64,
	feeRate sdk.Dec,
	liquidFarmReserveAddr sdk.AccAddress,
	poolCoinDenom string,
	winningBid types.Bid,
) error {
	withdrawnRewards, err := k.lpfarmKeeper.Harvest(ctx, liquidFarmReserveAddr, poolCoinDenom)
	if err != nil {
		return err
	}

	// As the farm module is designed with F1 fee distribution mechanism,
	// farming rewards are accumulated every block and rewards are automatically withdrawn
	// when an account executes Farm/Unfarm if the account already has position.
	// The module reserves any auto withdrawn rewards in the withdrawn rewards reserve account.
	// So, farming rewards must add the balance of withdrawn rewards reserve account.
	withdrawnRewardsReserveAddr := types.WithdrawnRewardsReserveAddress(poolId)
	spendable := k.bankKeeper.SpendableCoins(ctx, withdrawnRewardsReserveAddr)
	totalRewards := spendable.Add(withdrawnRewards...)

	if !totalRewards.IsZero() {
		deducted, fees := types.DeductFees(totalRewards, feeRate)

		if err := k.bankKeeper.SendCoins(ctx, withdrawnRewardsReserveAddr, liquidFarmReserveAddr, spendable); err != nil {
			return err
		}

		if err := k.bankKeeper.SendCoins(ctx, liquidFarmReserveAddr, winningBid.GetBidder(), deducted); err != nil {
			return err
		}

		if !fees.IsZero() {
			feeCollectorAddr, err := sdk.AccAddressFromBech32(k.GetFeeCollector(ctx))
			if err != nil {
				return err
			}

			if err := k.bankKeeper.SendCoins(ctx, liquidFarmReserveAddr, feeCollectorAddr, fees); err != nil {
				return err
			}
		}
	}

	return nil
}

func (k Keeper) compoundRewards(ctx sdk.Context, auctionPayingReserveAddr, liquidFarmReserveAddr sdk.AccAddress, amount sdk.Coin) error {
	if err := k.bankKeeper.SendCoins(ctx, auctionPayingReserveAddr, liquidFarmReserveAddr, sdk.NewCoins(amount)); err != nil {
		return err
	}
	if _, err := k.lpfarmKeeper.Farm(ctx, liquidFarmReserveAddr, amount); err != nil {
		return err
	}
	return nil
}
