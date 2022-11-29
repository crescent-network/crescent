package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
)

// LiquidFarm handles types.MsgLiquidFarm to farm.
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
	position, found := k.lpfarmKeeper.GetPosition(ctx, reserveAddr, farmingCoin.Denom)
	if !found {
		position.FarmingAmount = sdk.ZeroInt()
	}
	mintingAmt := types.CalculateLiquidFarmAmount(
		lfCoinTotalSupplyAmt,
		position.FarmingAmount,
		farmingCoin.Amount,
	)
	mintingCoin := sdk.NewCoin(lfCoinDenom, mintingAmt)

	// Mint new LFCoin amount and send it to the farmer
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(mintingCoin)); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, farmer, sdk.NewCoins(mintingCoin)); err != nil {
		return err
	}

	// Reserve account farms the farming coin amount for liquid farmer
	withdrawnRewards, err := k.lpfarmKeeper.Farm(ctx, reserveAddr, farmingCoin)
	if err != nil {
		return err
	}

	// As the farm module is designed with F1 fee distribution mechanism,
	// Farming rewards are automatically withdrawn if the module account already has position.
	// In order to keep in track of the rewards, the module reserves them in the WithdrawnRewardsReserveAddress.
	if !withdrawnRewards.IsZero() {
		withdrawnRewardsReserveAddr := types.WithdrawnRewardsReserveAddress(poolId)
		if err := k.bankKeeper.SendCoins(ctx, reserveAddr, withdrawnRewardsReserveAddr, withdrawnRewards); err != nil {
			return err
		}
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

// LiquidUnfarm handles types.MsgLiquidUnfarm to unfarm LFCoin.
// It doesn't validate if the liquid farm exists because farmers still need to be able to
// unfarm their LFCoin in case the liquid farm object is removed in params by governance proposal.
func (k Keeper) LiquidUnfarm(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, unfarmingCoin sdk.Coin) (unfarmedCoin sdk.Coin, err error) {
	pool, found := k.liquidityKeeper.GetPool(ctx, poolId)
	if !found {
		return sdk.Coin{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pool %d not found", poolId)
	}

	reserveAddr := types.LiquidFarmReserveAddress(pool.Id)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(pool.Id)
	lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
	lfCoinTotalSupplyAmt := k.bankKeeper.GetSupply(ctx, lfCoinDenom).Amount
	lpCoinTotalFarmingAmt := sdk.ZeroInt()
	position, found := k.lpfarmKeeper.GetPosition(ctx, reserveAddr, poolCoinDenom)
	if found {
		lpCoinTotalFarmingAmt = position.FarmingAmount
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
	unfarmedCoin = sdk.NewCoin(poolCoinDenom, unfarmingAmt)

	withdrawnRewards := sdk.Coins{}
	if found {
		// Unfarm the farmed coin in the farm module and release it to the farmer
		withdrawnRewards, err = k.lpfarmKeeper.Unfarm(ctx, reserveAddr, unfarmedCoin)
		if err != nil {
			return sdk.Coin{}, err
		}
	}

	// As the farm module is designed with F1 fee distribution mechanism,
	// Farming rewards are automatically withdrawn if the module account already has position.
	// In order to keep in track of the rewards, the module reserves them in the WithdrawnRewardsReserveAddress.
	if !withdrawnRewards.IsZero() {
		withdrawnRewardsReserveAddr := types.WithdrawnRewardsReserveAddress(poolId)
		if err := k.bankKeeper.SendCoins(ctx, reserveAddr, withdrawnRewardsReserveAddr, withdrawnRewards); err != nil {
			return sdk.Coin{}, err
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, reserveAddr, farmer, sdk.NewCoins(unfarmedCoin)); err != nil {
		return sdk.Coin{}, err
	}

	// Burn the LFCoin amount
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, farmer, types.ModuleName, sdk.NewCoins(unfarmingCoin)); err != nil {
		return sdk.Coin{}, err
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(unfarmingCoin)); err != nil {
		return sdk.Coin{}, err
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

	return unfarmedCoin, nil
}

// LiquidUnfarmAndWithdraw handles types.MsgUnfarmAndWithdraw to unfarm LFCoin and withdraw pool coin from the pool.
func (k Keeper) LiquidUnfarmAndWithdraw(ctx sdk.Context, poolId uint64, farmer sdk.AccAddress, unfarmingCoin sdk.Coin) error {
	unfarmedCoin, err := k.LiquidUnfarm(ctx, poolId, farmer, unfarmingCoin)
	if err != nil {
		return err
	}

	_, err = k.liquidityKeeper.Withdraw(ctx, &liquiditytypes.MsgWithdraw{
		PoolId:     poolId,
		Withdrawer: farmer.String(),
		PoolCoin:   unfarmedCoin,
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
			sdk.NewAttribute(types.AttributeKeyUnfarmedCoin, unfarmedCoin.String()),
		),
	})

	return nil
}

// HandleRemovedLiquidFarm unfarms all farmed coin from the farm module to stop having
// farming rewards accumulated and sends the harvested rewards to the fee collector.
// It refunds all placed bids and updates an appropriate states.
func (k Keeper) HandleRemovedLiquidFarm(ctx sdk.Context, liquidFarm types.LiquidFarm) {
	feeCollectorAddr, _ := sdk.AccAddressFromBech32(k.GetFeeCollector(ctx))
	reserveAddr := types.LiquidFarmReserveAddress(liquidFarm.PoolId)
	rewardsReserveAddr := types.WithdrawnRewardsReserveAddress(liquidFarm.PoolId)
	poolCoinDenom := liquiditytypes.PoolCoinDenom(liquidFarm.PoolId)

	position, found := k.lpfarmKeeper.GetPosition(ctx, reserveAddr, poolCoinDenom)
	if found {
		// Unfarm all farmed coin to stop having rewards accumulated in the farm module and
		// send the farming rewards to the fee collector.
		withdrawnRewards, err := k.lpfarmKeeper.Unfarm(ctx, reserveAddr, sdk.NewCoin(poolCoinDenom, position.FarmingAmount))
		if err != nil {
			panic(err)
		}

		if !withdrawnRewards.IsZero() {
			if err := k.bankKeeper.SendCoins(ctx, reserveAddr, feeCollectorAddr, withdrawnRewards); err != nil {
				panic(err)
			}
		}
	}

	// Send all auto withdrawn rewards by the farm module to the fee collector
	rewardsReserveBalance := k.bankKeeper.SpendableCoins(ctx, rewardsReserveAddr)
	if !rewardsReserveBalance.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, rewardsReserveAddr, feeCollectorAddr, rewardsReserveBalance); err != nil {
			panic(err)
		}
	}

	// Finish the ongoing rewards auction by refunding all bids and
	// set status to AuctionStatusFinished
	auction, found := k.GetLastRewardsAuction(ctx, liquidFarm.PoolId)
	if found {
		if err := k.refundAllBids(ctx, auction, true); err != nil {
			panic(err)
		}

		auction.SetStatus(types.AuctionStatusFinished)
		auction.SetFeeRate(liquidFarm.FeeRate)
		k.SetRewardsAuction(ctx, auction)
	}

	k.SetCompoundingRewards(ctx, liquidFarm.PoolId, types.CompoundingRewards{Amount: sdk.ZeroInt()})
	k.DeleteLiquidFarm(ctx, liquidFarm)
}
