package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// LiquidFarm handles types.MsgLiquidFarm to liquid farm.
func (k Keeper) LiquidFarm(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, farmingCoin sdk.Coin) error {
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

	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinTotalSupplyAmt := k.bankKeeper.GetSupply(ctx, lfCoinDenom).Amount
	farm, found := k.farmKeeper.GetFarm(ctx, farmingCoin.Denom)
	if !found {
		farm.TotalFarmingAmount = sdk.ZeroInt()
	}

	mintingAmt := types.CalculateLiquidFarmAmount(
		lfCoinTotalSupplyAmt,
		farm.TotalFarmingAmount,
		farmingCoin.Amount,
	)
	mintingCoin := sdk.NewCoin(lfCoinDenom, mintingAmt)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(mintingCoin)); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, farmer, sdk.NewCoins(mintingCoin)); err != nil {
		return err
	}

	if _, err := k.farmKeeper.Farm(ctx, reserveAddr, farmingCoin); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLiquidFarm,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyFarmingCoin, farmingCoin.String()),
			sdk.NewAttribute(types.AttributeKeyMintedCoin, mintingCoin.String()),
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

// LiquidUnfarm handles types.MsgLiquidUnfarm to unfarm LFCoin.
// It doesn't validate if the liquid farm exists because farmers still need to be able to
// unfarm their LFCoin in case the liquid farm object is removed in params by governance proposal.
func (k Keeper) LiquidUnfarm(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, unfarmingCoin sdk.Coin) (UnfarmInfo, error) {
	pool, found := k.liquidityKeeper.GetPool(ctx, poolId)
	if !found {
		return UnfarmInfo{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", poolId)
	}

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinTotalSupplyAmt := k.bankKeeper.GetSupply(ctx, lfCoinDenom).Amount
	lpCoinTotalFarmingAmt := sdk.ZeroInt()
	farm, found := k.farmKeeper.GetFarm(ctx, poolCoinDenom)
	if found {
		lpCoinTotalFarmingAmt = farm.TotalFarmingAmount
	}
	compoundingRewardsAmt := sdk.ZeroInt()
	compoundingRewards, found := k.GetCompoundingRewards(ctx, pool.Id)
	if found {
		compoundingRewardsAmt = compoundingRewards.Amount
	}

	_, found = k.GetLiquidFarm(ctx, poolId)
	if !found {
		// Handle a case when the liquid farm is removed in params
		// Since the reserve account must have unfarm all farmed coin from the farm module,
		// the module must use the reserve account balance
		lpCoinTotalFarmingAmt = k.bankKeeper.SpendableCoins(ctx, reserveAddr).AmountOf(poolCoinDenom)
	}

	unfarmingAmt := types.CalculateLiquidUnfarmAmount(
		lfCoinTotalSupplyAmt,
		lpCoinTotalFarmingAmt,
		unfarmingCoin.Amount,
		compoundingRewardsAmt,
	)
	unfarmedCoin := sdk.NewCoin(poolCoinDenom, unfarmingAmt)

	if found {
		// Unfarm the farmed coin in the farm module and release it to the farmer
		if _, err := k.farmKeeper.Unfarm(ctx, reserveAddr, unfarmedCoin); err != nil {
			return UnfarmInfo{}, err
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, reserveAddr, farmer, sdk.NewCoins(unfarmedCoin)); err != nil {
		return UnfarmInfo{}, err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, farmer, types.ModuleName, sdk.NewCoins(unfarmingCoin)); err != nil {
		return UnfarmInfo{}, err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(unfarmingCoin)); err != nil {
		return UnfarmInfo{}, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLiquidUnfarm,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmingCoin, unfarmingCoin.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmedCoin.String()),
		),
	})

	return UnfarmInfo{Farmer: farmer, UnfarmedCoin: unfarmedCoin}, nil
}

// LiquidUnfarmAndWithdraw handles types.MsgUnfarmAndWithdraw to unfarm LFCoin and withdraw pool coin from the pool.
func (k Keeper) LiquidUnfarmAndWithdraw(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, unfarmingCoin sdk.Coin) error {
	unfarmInfo, err := k.LiquidUnfarm(ctx, poolId, farmer, unfarmingCoin)
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
			types.EventTypeLiquidUnfarmAndWithdraw,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
			sdk.NewAttribute(types.AttributeKeyFarmer, farmer.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmingCoin, unfarmingCoin.String()),
			sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmInfo.UnfarmedCoin.String()),
		),
	})

	return nil
}

// HandleRemovedLiquidFarm unfarms all farmed coin from the farm module to stop having
// farming rewards accumulated and handle an ongoing rewards auction.
// Then finally delete the LiquidFarm object in the store.
func (k Keeper) HandleRemovedLiquidFarm(ctx sdk.Context, liquidFarm types.LiquidFarm) {
	reserveAddr := types.LiquidFarmReserveAddress(liquidFarm.PoolId)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(liquidFarm.PoolId)

	// Unstake all farmed coin to stop having rewards accumulated in the farm module
	position, found := k.farmKeeper.GetPosition(ctx, reserveAddr, poolCoinDenom)
	if found {
		if _, err := k.farmKeeper.Unfarm(ctx, reserveAddr, sdk.NewCoin(poolCoinDenom, position.FarmingAmount)); err != nil {
			panic(err)
		}
		// TODO: rewards remain in the reserve account. How do we deal with them?
	}

	// Handle the ongoing rewards auction by refunding all bids and
	// set status to AuctionStatusFinished
	auctionId := k.GetLastRewardsAuctionId(ctx, liquidFarm.PoolId)
	auction, found := k.GetRewardsAuction(ctx, auctionId, liquidFarm.PoolId)
	if found {
		if err := k.RefundAllBids(ctx, auction, true); err != nil {
			panic(err)
		}
		k.DeleteWinningBid(ctx, auctionId, liquidFarm.PoolId)
		auction.SetStatus(types.AuctionStatusFinished)
		k.SetRewardsAuction(ctx, auction)
	}

	k.SetCompoundingRewards(ctx, liquidFarm.PoolId, types.CompoundingRewards{Amount: sdk.ZeroInt()})
	k.DeleteLiquidFarm(ctx, liquidFarm)
}
