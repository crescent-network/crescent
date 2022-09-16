package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
)

// Farm handles types.MsgFarm to liquid farm.
func (k Keeper) Farm(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, farmingCoin sdk.Coin) error {
	pool, found := k.liquidityKeeper.GetPool(ctx, poolId)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", poolId)
	}

	liquidFarm, found := k.GetLiquidFarm(ctx, pool.Id)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "liquid farm by pool %d not found", pool.Id)
	}

	if farmingCoin.Amount.LT(liquidFarm.MinFarmAmount) {
		return sdkerrors.Wrapf(types.ErrSmallerThanMinimumAmount, "%s is smaller than %s", farmingCoin.Amount, liquidFarm.MinFarmAmount)
	}

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	if err := k.bankKeeper.SendCoins(ctx, farmer, reserveAddr, sdk.NewCoins(farmingCoin)); err != nil {
		return err
	}

	poolCoinDenom := liquiditytypes.PoolCoinDenom(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinTotalSupplyAmt := k.bankKeeper.GetSupply(ctx, types.LiquidFarmCoinDenom(pool.Id)).Amount
	lpCoinTotalQueuedAmt := k.farmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(ctx, reserveAddr, poolCoinDenom)
	lpCoinTotalStaked, found := k.farmingKeeper.GetStaking(ctx, poolCoinDenom, reserveAddr)
	if !found {
		lpCoinTotalStaked.Amount = sdk.ZeroInt()
	}

	mintedAmt := types.CalculateFarmMintingAmount(
		lfCoinTotalSupplyAmt,
		lpCoinTotalQueuedAmt,
		lpCoinTotalStaked.Amount,
		farmingCoin.Amount,
	)
	mintedCoin := sdk.NewCoin(lfCoinDenom, mintedAmt)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(mintedCoin)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, farmer, sdk.NewCoins(mintedCoin)); err != nil {
		return err
	}

	if err := k.farmingKeeper.Stake(ctx, reserveAddr, sdk.NewCoins(farmingCoin)); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFarm,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyFarmingCoin, farmingCoin.String()),
			sdk.NewAttribute(types.AttributeKeyMintedCoin, mintedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyLiquidFarmReserveAddress, reserveAddr.String()),
		),
	})

	return nil
}

// UnfarmInfo holds information about unfarm.
type UnfarmInfo struct {
	Farmer       sdk.AccAddress
	UnfarmedCoin sdk.Coin
}

// Unfarm handles types.MsgUnfarm to unfarm LFCoin.
// It doesn't validate if the liquid farm exists because farmers still need to be able to
// unfarm their LFCoin although the liquid farm object is removed in params.
func (k Keeper) Unfarm(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, burningCoin sdk.Coin) (UnfarmInfo, error) {
	pool, found := k.liquidityKeeper.GetPool(ctx, poolId)
	if !found {
		return UnfarmInfo{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", poolId)
	}

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(pool.Id)
	lfCoinTotalSupplyAmt := k.bankKeeper.GetSupply(ctx, types.LiquidFarmCoinDenom(pool.Id)).Amount
	lpCoinTotalQueuedAmt := k.farmingKeeper.GetAllQueuedStakingAmountByFarmerAndDenom(ctx, reserveAddr, poolCoinDenom)
	lpCoinTotalStaked, found := k.farmingKeeper.GetStaking(ctx, poolCoinDenom, reserveAddr)
	if !found {
		lpCoinTotalStaked.Amount = sdk.ZeroInt()
	}
	compoundingRewards, found := k.GetCompoundingRewards(ctx, pool.Id)
	if !found {
		compoundingRewards.Amount = sdk.ZeroInt()
	}

	_, found = k.GetLiquidFarm(ctx, poolId)
	if !found {
		if !lpCoinTotalStaked.Amount.IsZero() || !lpCoinTotalQueuedAmt.IsZero() {
			panic(fmt.Errorf("unexpected amount; staked amount: %s; queued amount: %s", lpCoinTotalStaked.Amount, lpCoinTotalQueuedAmt))
		}
		// Handle a case when liquid farm is removed in params
		// Since the reserve account must have unstaked all staked coins from the farming module,
		// the module must use the reserve account balance (staked + queued) and make queued amount zero
		lpCoinTotalStaked.Amount = k.bankKeeper.SpendableCoins(ctx, reserveAddr).AmountOf(poolCoinDenom)
	}

	unfarmedAmt := types.CalculateUnfarmedAmount(
		lfCoinTotalSupplyAmt,
		lpCoinTotalQueuedAmt,
		lpCoinTotalStaked.Amount,
		burningCoin.Amount,
		compoundingRewards.Amount,
	)
	unfarmedCoin := sdk.NewCoin(poolCoinDenom, unfarmedAmt)

	if found {
		// Unstake unfarm coin in the farming module and release it to the farmer
		if err := k.farmingKeeper.Unstake(ctx, reserveAddr, sdk.NewCoins(unfarmedCoin)); err != nil {
			return UnfarmInfo{}, err
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, reserveAddr, farmer, sdk.NewCoins(unfarmedCoin)); err != nil {
		return UnfarmInfo{}, err
	}

	// Burn the unfarming LFCoin by sending it to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, farmer, types.ModuleName, sdk.NewCoins(burningCoin)); err != nil {
		return UnfarmInfo{}, err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(burningCoin)); err != nil {
		return UnfarmInfo{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnfarm,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyBurningCoin, burningCoin.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmedCoin.String()),
		),
	})

	return UnfarmInfo{Farmer: farmer, UnfarmedCoin: unfarmedCoin}, nil
}

// UnfarmAndWithdraw handles types.MsgUnfarmAndWithdraw to unfarm LFCoin and withdraw pool coin from the pool.
func (k Keeper) UnfarmAndWithdraw(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, burningCoin sdk.Coin) error {
	unfarmInfo, err := k.Unfarm(ctx, poolId, farmer, burningCoin)
	if err != nil {
		return err
	}

	_, err = k.liquidityKeeper.Withdraw(ctx, &liquiditytypes.MsgWithdraw{
		PoolId:     poolId,
		Withdrawer: farmer.String(),
		PoolCoin:   unfarmInfo.UnfarmedCoin,
	})
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnfarmAndWithdraw,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyBurningCoin, burningCoin.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmInfo.UnfarmedCoin.String()),
		),
	})

	return nil
}

// HandleRemovedLiquidFarm unstakes all staked pool coins from the farming module and
// remove the liquid farm object in the store
func (k Keeper) HandleRemovedLiquidFarm(ctx sdk.Context, liquidFarm types.LiquidFarm) {
	reserveAddr := types.LiquidFarmReserveAddress(liquidFarm.PoolId)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(liquidFarm.PoolId)
	stakedAmt := sdk.ZeroInt()
	staking, found := k.farmingKeeper.GetStaking(ctx, poolCoinDenom, reserveAddr)
	if found {
		stakedAmt = staking.Amount
	}

	stakedCoin := sdk.NewCoin(poolCoinDenom, stakedAmt)
	if !stakedCoin.IsZero() {
		// Unstake all staked coins so that there will be no rewards accumulating
		if err := k.farmingKeeper.Unstake(ctx, reserveAddr, sdk.NewCoins(stakedCoin)); err != nil {
			panic(err)
		}
	}

	// Handle a case when the last rewards auction id isn't set in the store
	auctionId := k.GetLastRewardsAuctionId(ctx, liquidFarm.PoolId)
	auction, found := k.GetRewardsAuction(ctx, liquidFarm.PoolId, auctionId)
	if found {
		if err := k.RefundAllBids(ctx, auction, types.Bid{}); err != nil {
			panic(err)
		}
		k.DeleteWinningBid(ctx, liquidFarm.PoolId, auctionId)
		auction.SetStatus(types.AuctionStatusFinished)
		k.SetRewardsAuction(ctx, auction)
	}

	k.SetCompoundingRewards(ctx, liquidFarm.PoolId, types.CompoundingRewards{Amount: sdk.ZeroInt()})
	k.DeleteLiquidFarm(ctx, liquidFarm)
}
