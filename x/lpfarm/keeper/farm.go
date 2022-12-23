package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v4/x/lpfarm/types"
)

// Farm locks the coin.
// The farmer's rewards accrued in the given coin's denom are sent to the farmer.
// Farm creates a new farm object for the given coin's denom, if there wasn't.
func (k Keeper) Farm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error) {
	farmingReserveAddr := types.DeriveFarmingReserveAddress(coin.Denom)
	if err := k.bankKeeper.SendCoins(
		ctx, farmerAddr, farmingReserveAddr, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}

	_, found := k.GetFarm(ctx, coin.Denom)
	if !found {
		k.initializeFarm(ctx, coin.Denom)
	}

	position, found := k.GetPosition(ctx, farmerAddr, coin.Denom)
	if !found {
		k.incrementFarmPeriod(ctx, coin.Denom)
		position = types.Position{
			Farmer:        farmerAddr.String(),
			Denom:         coin.Denom,
			FarmingAmount: sdk.ZeroInt(),
		}
	} else {
		withdrawnRewards, err = k.withdrawRewards(ctx, position)
		if err != nil {
			return nil, err
		}
	}

	farm, _ := k.GetFarm(ctx, coin.Denom)
	farm.TotalFarmingAmount = farm.TotalFarmingAmount.Add(coin.Amount)
	k.SetFarm(ctx, coin.Denom, farm)

	position.FarmingAmount = position.FarmingAmount.Add(coin.Amount)
	k.updatePosition(ctx, position)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFarm{
		Farmer:           farmerAddr.String(),
		Coin:             coin,
		WithdrawnRewards: withdrawnRewards,
	}); err != nil {
		return nil, err
	}

	return withdrawnRewards, nil
}

// Unfarm unlocks the coin.
// The farmer's rewards accrued in the given coin's denom are sent to the farmer.
// If the remaining farming coin amount becomes zero, the farming position is
// deleted.
func (k Keeper) Unfarm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error) {
	position, found := k.GetPosition(ctx, farmerAddr, coin.Denom)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	if position.FarmingAmount.LT(coin.Amount) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "not enough farming amount")
	}

	withdrawnRewards, err = k.withdrawRewards(ctx, position)
	if err != nil {
		return nil, err
	}

	position.FarmingAmount = position.FarmingAmount.Sub(coin.Amount)
	if position.FarmingAmount.IsZero() {
		k.DeletePosition(ctx, farmerAddr, coin.Denom)
	} else {
		k.updatePosition(ctx, position)
	}

	farm, found := k.GetFarm(ctx, coin.Denom)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "farm not found")
	}
	farm.TotalFarmingAmount = farm.TotalFarmingAmount.Sub(coin.Amount)
	k.SetFarm(ctx, coin.Denom, farm)

	farmingReserveAddr := types.DeriveFarmingReserveAddress(coin.Denom)
	if err := k.bankKeeper.SendCoins(ctx, farmingReserveAddr, farmerAddr, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventUnfarm{
		Farmer:           farmerAddr.String(),
		Coin:             coin,
		WithdrawnRewards: withdrawnRewards,
	}); err != nil {
		return nil, err
	}

	return withdrawnRewards, nil
}

// Harvest sends the farmer's rewards accrued in the denom to the farmer.
func (k Keeper) Harvest(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (withdrawnRewards sdk.Coins, err error) {
	position, found := k.GetPosition(ctx, farmerAddr, denom)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}

	withdrawnRewards, err = k.withdrawRewards(ctx, position)
	if err != nil {
		return nil, err
	}

	k.updatePosition(ctx, position)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventHarvest{
		Farmer:           farmerAddr.String(),
		Denom:            denom,
		WithdrawnRewards: withdrawnRewards,
	}); err != nil {
		return nil, err
	}

	return withdrawnRewards, nil
}

// Rewards returns the farmer's rewards accrued in the denom so far.
// Rewards is a convenient query method existing for external modules.
func (k Keeper) Rewards(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) sdk.DecCoins {
	position, found := k.GetPosition(ctx, farmerAddr, denom)
	if !found {
		return nil
	}
	cacheCtx, _ := ctx.CacheContext()
	endPeriod := k.incrementFarmPeriod(cacheCtx, denom)
	return k.calculateRewards(cacheCtx, position, endPeriod)
}

// TotalRewards returns the farmer's rewards accrued in all denoms so far.
// TotalRewards is a convenient query method existing for external modules.
func (k Keeper) TotalRewards(ctx sdk.Context, farmerAddr sdk.AccAddress) (rewards sdk.DecCoins) {
	k.IteratePositionsByFarmer(ctx, farmerAddr, func(position types.Position) (stop bool) {
		cacheCtx, _ := ctx.CacheContext()
		endPeriod := k.incrementFarmPeriod(cacheCtx, position.Denom)
		rewards = rewards.Add(k.calculateRewards(cacheCtx, position, endPeriod)...)
		return false
	})
	return rewards
}

func (k Keeper) calculateRewards(ctx sdk.Context, position types.Position, endPeriod uint64) sdk.DecCoins {
	if position.StartingBlockHeight == ctx.BlockHeight() {
		return nil
	}
	startPeriod := position.PreviousPeriod
	return k.rewardsBetweenPeriods(
		ctx, position.Denom, startPeriod, endPeriod, position.FarmingAmount)
}

// initializeFarm creates a new farm object in the store, along with historical
// rewards for period 0.
func (k Keeper) initializeFarm(ctx sdk.Context, denom string) types.Farm {
	farm := types.Farm{
		TotalFarmingAmount: sdk.ZeroInt(),
		CurrentRewards:     sdk.DecCoins{},
		OutstandingRewards: sdk.DecCoins{},
		Period:             1,
	}
	k.SetFarm(ctx, denom, farm)
	k.SetHistoricalRewards(ctx, denom, 0, types.HistoricalRewards{
		CumulativeUnitRewards: sdk.DecCoins{},
		ReferenceCount:        1,
	})
	return farm
}

// updatePosition updates the position's starting info.
func (k Keeper) updatePosition(ctx sdk.Context, position types.Position) {
	farm, found := k.GetFarm(ctx, position.Denom)
	if !found { // Sanity check
		panic("farm not found")
	}
	prevPeriod := farm.Period - 1
	k.incrementReferenceCount(ctx, position.Denom, prevPeriod)
	position.PreviousPeriod = prevPeriod
	position.StartingBlockHeight = ctx.BlockHeight()
	k.SetPosition(ctx, position)
}

// incrementFarmPeriod increments the farm object's period by settling
// the historical rewards for the farm's current period.
func (k Keeper) incrementFarmPeriod(ctx sdk.Context, denom string) (prevPeriod uint64) {
	farm, found := k.GetFarm(ctx, denom)
	if !found { // Sanity check
		panic("farm not found")
	}
	unitRewards := sdk.DecCoins{}
	if farm.TotalFarmingAmount.IsPositive() {
		unitRewards = farm.CurrentRewards.QuoDecTruncate(sdk.NewDecFromInt(farm.TotalFarmingAmount))
	}
	hist, found := k.GetHistoricalRewards(ctx, denom, farm.Period-1)
	if !found { // Sanity check
		panic("historical rewards not found")
	}
	k.decrementReferenceCount(ctx, denom, farm.Period-1)
	k.SetHistoricalRewards(ctx, denom, farm.Period, types.HistoricalRewards{
		CumulativeUnitRewards: hist.CumulativeUnitRewards.Add(unitRewards...),
		ReferenceCount:        1,
	})
	farm.CurrentRewards = sdk.DecCoins{}
	prevPeriod = farm.Period
	farm.Period++
	k.SetFarm(ctx, denom, farm)
	return prevPeriod
}

// incrementReferenceCount increments the reference count of the historical
// rewards.
func (k Keeper) incrementReferenceCount(ctx sdk.Context, denom string, period uint64) {
	hist, found := k.GetHistoricalRewards(ctx, denom, period)
	if !found { // Sanity check
		panic("historical rewards not found")
	}
	if hist.ReferenceCount >= 2 { // Sanity check
		panic("reference count must never exceed 2")
	}
	hist.ReferenceCount++
	k.SetHistoricalRewards(ctx, denom, period, hist)
}

// decrementReferenceCount decrements the reference count of the historical
// rewards.
// If the reference count goes down to 0, the object is deleted.
func (k Keeper) decrementReferenceCount(ctx sdk.Context, denom string, period uint64) {
	hist, found := k.GetHistoricalRewards(ctx, denom, period)
	if !found { // Sanity check
		panic("historical rewards not found")
	}
	if hist.ReferenceCount == 0 {
		panic("reference count must not be negative")
	}
	hist.ReferenceCount--
	if hist.ReferenceCount == 0 {
		k.DeleteHistoricalRewards(ctx, denom, period)
	} else {
		k.SetHistoricalRewards(ctx, denom, period, hist)
	}
}

func (k Keeper) rewardsBetweenPeriods(ctx sdk.Context, denom string, startPeriod, endPeriod uint64, amt sdk.Int) sdk.DecCoins {
	start, found := k.GetHistoricalRewards(ctx, denom, startPeriod)
	if !found {
		panic("historical rewards not found")
	}
	end, found := k.GetHistoricalRewards(ctx, denom, endPeriod)
	if !found {
		panic("historical rewards not found")
	}
	diff := end.CumulativeUnitRewards.Sub(start.CumulativeUnitRewards)
	return diff.MulDecTruncate(sdk.NewDecFromInt(amt))
}

// withdrawRewards withdraws accrued rewards for the position and increments
// the farm's period.
func (k Keeper) withdrawRewards(ctx sdk.Context, position types.Position) (sdk.Coins, error) {
	endPeriod := k.incrementFarmPeriod(ctx, position.Denom)
	rewards := k.calculateRewards(ctx, position, endPeriod)

	truncatedRewards, _ := rewards.TruncateDecimal()
	if !truncatedRewards.IsZero() {
		farmerAddr, err := sdk.AccAddressFromBech32(position.Farmer)
		if err != nil {
			return nil, err
		}
		if err := k.bankKeeper.SendCoins(
			ctx, types.RewardsPoolAddress, farmerAddr, truncatedRewards); err != nil {
			return nil, err
		}
		// `found` has already been checked in k.incrementFarmPeriod.
		farm, _ := k.GetFarm(ctx, position.Denom)
		farm.OutstandingRewards = farm.OutstandingRewards.
			Sub(sdk.NewDecCoinsFromCoins(truncatedRewards...))
		k.SetFarm(ctx, position.Denom, farm)
	}

	k.decrementReferenceCount(ctx, position.Denom, position.PreviousPeriod)
	return truncatedRewards, nil
}
