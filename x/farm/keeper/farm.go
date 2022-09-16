package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (k Keeper) Farm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawnRewards sdk.Coins, err error) {
	if err := k.bankKeeper.SendCoins(
		ctx, farmerAddr, types.DeriveFarmingReserveAddress(coin.Denom), sdk.NewCoins(coin),
	); err != nil {
		return nil, err
	}

	farm, found := k.GetFarm(ctx, coin.Denom)
	if !found {
		farm = k.initializeFarm(ctx, coin.Denom)
	}
	farm.TotalFarmingAmount = farm.TotalFarmingAmount.Add(coin.Amount)
	k.SetFarm(ctx, coin.Denom, farm)

	prevPeriod := farm.Period - 1
	k.incrementReferenceCount(ctx, coin.Denom, prevPeriod)
	k.IncrementFarmPeriod(ctx, coin.Denom)

	position, found := k.GetPosition(ctx, farmerAddr, coin.Denom)
	if !found {
		position = types.Position{
			Farmer:        farmerAddr.String(),
			Denom:         coin.Denom,
			FarmingAmount: sdk.ZeroInt(),
		}
	} else {
		withdrawnRewards, err = k.withdrawRewards(ctx, farmerAddr, coin.Denom)
		if err != nil {
			return nil, err
		}
	}
	position.FarmingAmount = position.FarmingAmount.Add(coin.Amount)
	position.PreviousPeriod = prevPeriod
	position.StartingBlockHeight = ctx.BlockHeight()
	k.SetPosition(ctx, position)

	return withdrawnRewards, nil
}

func (k Keeper) Unfarm(ctx sdk.Context, farmerAddr sdk.AccAddress, coin sdk.Coin) (withdrawRewards sdk.Coins, err error) {
	return nil, nil
}

func (k Keeper) Harvest(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (withdrawnRewards sdk.Coins, err error) {
	return nil, nil
}

func (k Keeper) Rewards(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string, endPeriod uint64) sdk.DecCoins {
	position, found := k.GetPosition(ctx, farmerAddr, denom)
	if !found {
		return nil
	}
	if position.StartingBlockHeight == ctx.BlockHeight() {
		return nil
	}
	startPeriod := position.PreviousPeriod
	return k.rewardsBetweenPeriods(ctx, denom, startPeriod, endPeriod, position.FarmingAmount)
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

func (k Keeper) IncrementFarmPeriod(ctx sdk.Context, denom string) (prevPeriod uint64) {
	farm, found := k.GetFarm(ctx, denom)
	if !found { // Sanity check
		panic(fmt.Errorf("farm %s not found", denom))
	}
	unitRewards := sdk.DecCoins{}
	if farm.TotalFarmingAmount.IsZero() {
		// TODO: do something special?
	} else {
		unitRewards = farm.CurrentRewards.QuoDecTruncate(sdk.NewDecFromInt(farm.TotalFarmingAmount))
	}
	hist, found := k.GetHistoricalRewards(ctx, denom, farm.Period-1)
	if !found { // Sanity check
		panic(fmt.Errorf("historical rewards (%s, %d) not found", denom, farm.Period-1))
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
		panic(fmt.Errorf("historical rewards (%s, %d) not found", denom, period))
	}
	if hist.ReferenceCount > 2 {
		panic(fmt.Errorf("ref. count of historical rewards (%s, %d) must never exceed 2", denom, period))
	}
	hist.ReferenceCount++
	k.SetHistoricalRewards(ctx, denom, period, hist)
}

func (k Keeper) decrementReferenceCount(ctx sdk.Context, denom string, period uint64) {
	hist, found := k.GetHistoricalRewards(ctx, denom, period)
	if !found { // Sanity check
		panic(fmt.Errorf("historical rewards (%s, %d) not found", denom, period))
	}
	if hist.ReferenceCount == 0 {
		panic(fmt.Errorf("ref. count of historical rewards (%s, %d) must not be negative", denom, period))
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
		panic(fmt.Errorf("historical rewards (%s, %d) not found", denom, startPeriod))
	}
	end, found := k.GetHistoricalRewards(ctx, denom, endPeriod)
	if !found {
		panic(fmt.Errorf("historical rewards (%s, %d) not found", denom, endPeriod))
	}
	diff := end.CumulativeUnitRewards.Sub(start.CumulativeUnitRewards)
	return diff.MulDecTruncate(sdk.NewDecFromInt(amt))
}

func (k Keeper) withdrawRewards(ctx sdk.Context, farmerAddr sdk.AccAddress, denom string) (sdk.Coins, error) {
	position, found := k.GetPosition(ctx, farmerAddr, denom)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "position not found")
	}
	endPeriod := k.IncrementFarmPeriod(ctx, denom)
	rewards := k.Rewards(ctx, farmerAddr, denom, endPeriod)

	truncatedRewards, _ := rewards.TruncateDecimal()
	if !truncatedRewards.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, types.RewardsPoolAddress, farmerAddr, truncatedRewards); err != nil {
			return nil, err
		}
		farm, found := k.GetFarm(ctx, denom)
		if !found {
			return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "farm not found")
		}
		farm.OutstandingRewards = farm.OutstandingRewards.Sub(sdk.NewDecCoinsFromCoins(truncatedRewards...))
		k.SetFarm(ctx, denom, farm)
	}

	k.decrementReferenceCount(ctx, denom, position.PreviousPeriod)
	return truncatedRewards, nil
}
