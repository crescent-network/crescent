package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (k Keeper) Farm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error) {
	farmingReserveAddr := types.DeriveFarmingReserveAddress(coin.Denom)
	if err := k.bankKeeper.SendCoins(
		ctx, farmerAddr, farmingReserveAddr, sdk.NewCoins(coin)); err != nil {
		return nil, err
	}

	farm, found := k.GetFarm(ctx, coin.Denom)
	if !found {
		farm = k.initializeFarm(ctx, coin.Denom)
	}
	farm.TotalFarmingAmount = farm.TotalFarmingAmount.Add(coin.Amount)
	k.SetFarm(ctx, coin.Denom, farm)

	position, found := k.GetPosition(ctx, farmerAddr, coin.Denom)
	if !found {
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
	position.FarmingAmount = position.FarmingAmount.Add(coin.Amount)
	k.updatePosition(ctx, position)

	// TODO: emit an event

	return withdrawnRewards, nil
}

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

	// TODO: emit an event

	return withdrawnRewards, nil
}

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

	// TODO: emit an event

	return withdrawnRewards, nil
}

func (k Keeper) Rewards(ctx sdk.Context, position types.Position, endPeriod uint64) sdk.DecCoins {
	if position.StartingBlockHeight == ctx.BlockHeight() {
		return nil
	}
	startPeriod := position.PreviousPeriod
	return k.rewardsBetweenPeriods(
		ctx, position.Denom, startPeriod, endPeriod, position.FarmingAmount)
}

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

func (k Keeper) IncrementFarmPeriod(ctx sdk.Context, denom string) (prevPeriod uint64) {
	farm, found := k.GetFarm(ctx, denom)
	if !found { // Sanity check
		panic("farm not found")
	}
	unitRewards := sdk.DecCoins{}
	if farm.TotalFarmingAmount.IsZero() {
		// TODO: do something special?
	} else {
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

func (k Keeper) incrementReferenceCount(ctx sdk.Context, denom string, period uint64) {
	hist, found := k.GetHistoricalRewards(ctx, denom, period)
	if !found { // Sanity check
		panic("historical rewards not found")
	}
	if hist.ReferenceCount > 2 { // Sanity check
		panic("reference count must never exceed 2")
	}
	hist.ReferenceCount++
	k.SetHistoricalRewards(ctx, denom, period, hist)
}

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

func (k Keeper) withdrawRewards(ctx sdk.Context, position types.Position) (sdk.Coins, error) {
	endPeriod := k.IncrementFarmPeriod(ctx, position.Denom)
	rewards := k.Rewards(ctx, position, endPeriod)

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
		// `found` has already been checked in k.IncrementFarmPeriod.
		farm, _ := k.GetFarm(ctx, position.Denom)
		farm.OutstandingRewards = farm.OutstandingRewards.
			Sub(sdk.NewDecCoinsFromCoins(truncatedRewards...))
		k.SetFarm(ctx, position.Denom, farm)
	}

	k.decrementReferenceCount(ctx, position.Denom, position.PreviousPeriod)
	return truncatedRewards, nil
}
