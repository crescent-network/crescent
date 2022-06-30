package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// GetStaking returns a staking for given staking denom and farmer.
func (k Keeper) GetStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) (staking types.Staking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStakingKey(stakingCoinDenom, farmerAcc))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &staking)
	found = true
	return
}

// SetStaking sets a staking for given staking coin denom and farmer.
func (k Keeper) SetStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&staking)
	store.Set(types.GetStakingKey(stakingCoinDenom, farmerAcc), bz)
	store.Set(types.GetStakingIndexKey(farmerAcc, stakingCoinDenom), []byte{})
}

// DeleteStaking deletes a staking for given staking coin denom and farmer.
func (k Keeper) DeleteStaking(ctx sdk.Context, stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetStakingKey(stakingCoinDenom, farmerAcc))
	store.Delete(types.GetStakingIndexKey(farmerAcc, stakingCoinDenom))
}

// IterateStakings iterates through all stakings stored in the store
// and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateStakings(ctx sdk.Context, cb func(stakingCoinDenom string, farmerAcc sdk.AccAddress, staking types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.StakingKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var staking types.Staking
		k.cdc.MustUnmarshal(iter.Value(), &staking)
		stakingCoinDenom, farmerAcc := types.ParseStakingKey(iter.Key())
		if cb(stakingCoinDenom, farmerAcc, staking) {
			break
		}
	}
}

// IterateStakingsByFarmer iterates through all stakings by a farmer
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateStakingsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(stakingCoinDenom string, staking types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetStakingsByFarmerPrefix(farmerAcc))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		farmerAcc, stakingCoinDenom := types.ParseStakingIndexKey(iter.Key())
		staking, _ := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
		if cb(stakingCoinDenom, staking) {
			break
		}
	}
}

// GetAllStakedCoinsByFarmer returns all coins that are staked by a farmer.
func (k Keeper) GetAllStakedCoinsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins {
	stakedCoins := sdk.NewCoins()
	k.IterateStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, staking types.Staking) (stop bool) {
		stakedCoins = stakedCoins.Add(sdk.NewCoin(stakingCoinDenom, staking.Amount))
		return false
	})
	return stakedCoins
}

// GetQueuedStaking returns a queued staking for given staking coin denom
// and farmer.
func (k Keeper) GetQueuedStaking(ctx sdk.Context, endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress) (queuedStaking types.QueuedStaking, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetQueuedStakingKey(endTime, stakingCoinDenom, farmerAcc))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &queuedStaking)
	found = true
	return
}

// SetQueuedStaking sets a queued staking for given staking coin denom
// and farmer.
func (k Keeper) SetQueuedStaking(ctx sdk.Context, endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&queuedStaking)
	store.Set(types.GetQueuedStakingKey(endTime, stakingCoinDenom, farmerAcc), bz)
	store.Set(types.GetQueuedStakingIndexKey(farmerAcc, stakingCoinDenom, endTime), []byte{})
}

// DeleteQueuedStaking deletes a queued staking for given staking coin denom
// and farmer.
func (k Keeper) DeleteQueuedStaking(ctx sdk.Context, endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetQueuedStakingKey(endTime, stakingCoinDenom, farmerAcc))
	store.Delete(types.GetQueuedStakingIndexKey(farmerAcc, stakingCoinDenom, endTime))
}

// IterateQueuedStakings iterates through all queued stakings stored in
// the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateQueuedStakings(ctx sdk.Context, cb func(endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.QueuedStakingKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var queuedStaking types.QueuedStaking
		k.cdc.MustUnmarshal(iter.Value(), &queuedStaking)
		endTime, stakingCoinDenom, farmerAcc := types.ParseQueuedStakingKey(iter.Key())
		if cb(endTime, stakingCoinDenom, farmerAcc, queuedStaking) {
			break
		}
	}
}

// IterateMatureQueuedStakings iterates through all the queued stakings
// that are mature at currTime.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateMatureQueuedStakings(ctx sdk.Context, currTime time.Time, cb func(endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator(types.QueuedStakingKeyPrefix, types.GetQueuedStakingEndBytes(currTime))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var queuedStaking types.QueuedStaking
		k.cdc.MustUnmarshal(iter.Value(), &queuedStaking)
		endTime, stakingCoinDenom, farmerAcc := types.ParseQueuedStakingKey(iter.Key())
		if cb(endTime, stakingCoinDenom, farmerAcc, queuedStaking) {
			break
		}
	}
}

// IterateQueuedStakingsByFarmer iterates through all queued stakings
// by farmer stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateQueuedStakingsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress, cb func(stakingCoinDenom string, endTime time.Time, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetQueuedStakingsByFarmerPrefix(farmerAcc))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, stakingCoinDenom, endTime := types.ParseQueuedStakingIndexKey(iter.Key())
		queuedStaking, _ := k.GetQueuedStaking(ctx, endTime, stakingCoinDenom, farmerAcc)
		if cb(stakingCoinDenom, endTime, queuedStaking) {
			break
		}
	}
}

// IterateQueuedStakingsByFarmerAndDenom iterates through all the queued stakings
// by farmer address and staking coin denom.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateQueuedStakingsByFarmerAndDenom(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string, cb func(endTime time.Time, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetQueuedStakingsByFarmerAndDenomPrefix(farmerAcc, stakingCoinDenom))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, _, endTime := types.ParseQueuedStakingIndexKey(iter.Key())
		queuedStaking, _ := k.GetQueuedStaking(ctx, endTime, stakingCoinDenom, farmerAcc)
		if cb(endTime, queuedStaking) {
			break
		}
	}
}

// IterateQueuedStakingsByFarmerAndDenomReverse iterates through all the queued
// stakings by farmer address and staking coin denom, in reverse order.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateQueuedStakingsByFarmerAndDenomReverse(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string, cb func(endTime time.Time, queuedStaking types.QueuedStaking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStoreReversePrefixIterator(store, types.GetQueuedStakingsByFarmerAndDenomPrefix(farmerAcc, stakingCoinDenom))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, _, endTime := types.ParseQueuedStakingIndexKey(iter.Key())
		queuedStaking, _ := k.GetQueuedStaking(ctx, endTime, stakingCoinDenom, farmerAcc)
		if cb(endTime, queuedStaking) {
			break
		}
	}
}

// GetAllQueuedStakingAmountByFarmerAndDenom returns the amount of all queued
// stakings by the farmer for given staking coin denom.
func (k Keeper) GetAllQueuedStakingAmountByFarmerAndDenom(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoinDenom string) sdk.Int {
	amt := sdk.ZeroInt()
	k.IterateQueuedStakingsByFarmerAndDenom(ctx, farmerAcc, stakingCoinDenom, func(endTime time.Time, queuedStaking types.QueuedStaking) (stop bool) {
		if endTime.After(ctx.BlockTime()) { // sanity check
			amt = amt.Add(queuedStaking.Amount)
		}
		return false
	})
	return amt
}

// GetAllQueuedCoinsByFarmer returns all coins that are queued for staking
// by a farmer.
func (k Keeper) GetAllQueuedCoinsByFarmer(ctx sdk.Context, farmerAcc sdk.AccAddress) sdk.Coins {
	stakedCoins := sdk.NewCoins()
	k.IterateQueuedStakingsByFarmer(ctx, farmerAcc, func(stakingCoinDenom string, endTime time.Time, queuedStaking types.QueuedStaking) (stop bool) {
		if endTime.After(ctx.BlockTime()) { // sanity check
			stakedCoins = stakedCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
		}
		return false
	})
	return stakedCoins
}

// GetTotalStakings returns total stakings for given staking coin denom.
func (k Keeper) GetTotalStakings(ctx sdk.Context, stakingCoinDenom string) (totalStakings types.TotalStakings, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalStakingsKey(stakingCoinDenom))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &totalStakings)
	found = true
	return
}

// SetTotalStakings sets total stakings for given staking coin denom.
func (k Keeper) SetTotalStakings(ctx sdk.Context, stakingCoinDenom string, totalStakings types.TotalStakings) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalStakings)
	store.Set(types.GetTotalStakingsKey(stakingCoinDenom), bz)
}

// DeleteTotalStakings deletes total stakings for given staking coin denom.
func (k Keeper) DeleteTotalStakings(ctx sdk.Context, stakingCoinDenom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTotalStakingsKey(stakingCoinDenom))
}

// IncreaseTotalStakings increases total stakings for given staking coin denom
// by given amount.
func (k Keeper) IncreaseTotalStakings(ctx sdk.Context, stakingCoinDenom string, amount sdk.Int) {
	totalStakings, found := k.GetTotalStakings(ctx, stakingCoinDenom)
	if !found {
		totalStakings.Amount = sdk.ZeroInt()
	}
	totalStakings.Amount = totalStakings.Amount.Add(amount)
	k.SetTotalStakings(ctx, stakingCoinDenom, totalStakings)
	if totalStakings.Amount.Equal(amount) {
		k.afterStakingCoinAdded(ctx, stakingCoinDenom)
	}
}

// DecreaseTotalStakings decreases total stakings for given staking coin denom
// by given amount.
func (k Keeper) DecreaseTotalStakings(ctx sdk.Context, stakingCoinDenom string, amount sdk.Int) {
	totalStakings, found := k.GetTotalStakings(ctx, stakingCoinDenom)
	if !found {
		panic("total stakings not found")
	}
	if totalStakings.Amount.LT(amount) {
		panic("cannot set negative total stakings")
	}
	if amount.Equal(totalStakings.Amount) {
		k.DeleteTotalStakings(ctx, stakingCoinDenom)
		if err := k.afterStakingCoinRemoved(ctx, stakingCoinDenom); err != nil {
			panic(err)
		}
	} else {
		totalStakings.Amount = totalStakings.Amount.Sub(amount)
		k.SetTotalStakings(ctx, stakingCoinDenom, totalStakings)
	}
}

// IterateTotalStakings iterates through all total stakings
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateTotalStakings(ctx sdk.Context, cb func(stakingCoinDenom string, totalStakings types.TotalStakings) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.TotalStakingKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var totalStakings types.TotalStakings
		k.cdc.MustUnmarshal(iter.Value(), &totalStakings)
		stakingCoinDenom := types.ParseTotalStakingsKey(iter.Key())
		if cb(stakingCoinDenom, totalStakings) {
			break
		}
	}
}

// ReserveStakingCoins sends staking coins to the staking reserve account.
func (k Keeper) ReserveStakingCoins(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoins sdk.Coins) error {
	if stakingCoins.Len() == 1 {
		reserveAcc := types.StakingReserveAcc(stakingCoins[0].Denom)
		if err := k.bankKeeper.SendCoins(ctx, farmerAcc, reserveAcc, stakingCoins); err != nil {
			return err
		}
		if !k.bankKeeper.BlockedAddr(ctx, reserveAcc) {
			k.bankKeeper.AddBlockedAddr(ctx, reserveAcc)
		}
	} else {
		var inputs []banktypes.Input
		var outputs []banktypes.Output
		for _, coin := range stakingCoins {
			reserveAcc := types.StakingReserveAcc(coin.Denom)
			inputs = append(inputs, banktypes.NewInput(farmerAcc, sdk.Coins{coin}))
			outputs = append(outputs, banktypes.NewOutput(reserveAcc, sdk.Coins{coin}))
			if !k.bankKeeper.BlockedAddr(ctx, reserveAcc) {
				k.bankKeeper.AddBlockedAddr(ctx, reserveAcc)
			}
		}
		if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return err
		}
	}
	return nil
}

// ReleaseStakingCoins sends staking coins back to the farmer.
func (k Keeper) ReleaseStakingCoins(ctx sdk.Context, farmerAcc sdk.AccAddress, stakingCoins sdk.Coins) error {
	if stakingCoins.Len() == 1 {
		if err := k.bankKeeper.SendCoins(ctx, types.StakingReserveAcc(stakingCoins[0].Denom), farmerAcc, stakingCoins); err != nil {
			return err
		}
	} else {
		var inputs []banktypes.Input
		var outputs []banktypes.Output
		for _, coin := range stakingCoins {
			inputs = append(inputs, banktypes.NewInput(types.StakingReserveAcc(coin.Denom), sdk.Coins{coin}))
			outputs = append(outputs, banktypes.NewOutput(farmerAcc, sdk.Coins{coin}))
		}
		if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return err
		}
	}
	return nil
}

// afterStakingCoinAdded is called after a new staking coin denom appeared
// during ProcessQueuedCoins.
func (k Keeper) afterStakingCoinAdded(ctx sdk.Context, stakingCoinDenom string) {
	currentEpoch := k.GetCurrentEpoch(ctx, stakingCoinDenom)
	if currentEpoch == 0 {
		currentEpoch = 1
	}
	k.SetCurrentEpoch(ctx, stakingCoinDenom, currentEpoch)
	k.SetHistoricalRewards(ctx, stakingCoinDenom, currentEpoch-1, types.HistoricalRewards{CumulativeUnitRewards: sdk.DecCoins{}})
	k.SetOutstandingRewards(ctx, stakingCoinDenom, types.OutstandingRewards{Rewards: sdk.DecCoins{}})
}

// afterStakingCoinRemoved is called after a staking coin denom got removed
// during Unstake.
func (k Keeper) afterStakingCoinRemoved(ctx sdk.Context, stakingCoinDenom string) error {
	// Send remaining outstanding rewards to the farming fee collector.
	// A staking coin is removed only after there is no farmers
	// have rewards.
	// Note that there should never be any remaining integral rewards
	// in general situations, so this exists for confidence.
	outstanding, _ := k.GetOutstandingRewards(ctx, stakingCoinDenom)
	coins, _ := outstanding.Rewards.TruncateDecimal() // Ignore remainder, since it cannot be sent.
	if !coins.IsZero() {
		params := k.GetParams(ctx)
		feeCollectorAcc, _ := sdk.AccAddressFromBech32(params.FarmingFeeCollector) // Already validated
		if err := k.bankKeeper.SendCoins(ctx, types.RewardsReserveAcc, feeCollectorAcc, coins); err != nil {
			return err
		}
	}

	k.DeleteOutstandingRewards(ctx, stakingCoinDenom)
	k.DeleteAllHistoricalRewards(ctx, stakingCoinDenom)
	return nil
}

// Stake stores staking coins to queued coins, and it will be processed in the next epoch.
func (k Keeper) Stake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error {
	if err := k.ReserveStakingCoins(ctx, farmerAcc, amount); err != nil {
		return err
	}

	currentEpochDays := k.GetCurrentEpochDays(ctx)
	endTime := ctx.BlockTime().Add(time.Duration(currentEpochDays) * types.Day)

	numStakingCoinDenoms := 0
	for _, coin := range amount {
		queuedStaking, found := k.GetQueuedStaking(ctx, endTime, coin.Denom, farmerAcc)
		if !found {
			queuedStaking.Amount = sdk.ZeroInt()
		}
		queuedStaking.Amount = queuedStaking.Amount.Add(coin.Amount)
		k.SetQueuedStaking(ctx, endTime, coin.Denom, farmerAcc, queuedStaking)

		_, found = k.GetStaking(ctx, coin.Denom, farmerAcc)
		if found {
			numStakingCoinDenoms++
		}
	}

	if numStakingCoinDenoms > 0 {
		params := k.GetParams(ctx)
		ctx.GasMeter().ConsumeGas(sdk.Gas(numStakingCoinDenoms)*params.DelayedStakingGasFee, "DelayedStakingGasFee")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyStakingCoins, amount.String()),
		),
	})

	return nil
}

// Unstake unstakes an amount of staking coins from the staking reserve account.
// It causes accumulated rewards to be withdrawn to the farmer.
func (k Keeper) Unstake(ctx sdk.Context, farmerAcc sdk.AccAddress, amount sdk.Coins) error {
	totalUnharvestedRewards := sdk.Coins{}
	for _, coin := range amount {
		unstaked := sdk.ZeroInt()
		k.IterateQueuedStakingsByFarmerAndDenomReverse(ctx, farmerAcc, coin.Denom, func(endTime time.Time, queuedStaking types.QueuedStaking) (stop bool) {
			if endTime.After(ctx.BlockTime()) { // sanity check
				amtToUnstake := sdk.MinInt(coin.Amount.Sub(unstaked), queuedStaking.Amount)
				queuedStaking.Amount = queuedStaking.Amount.Sub(amtToUnstake)
				if queuedStaking.Amount.IsZero() {
					k.DeleteQueuedStaking(ctx, endTime, coin.Denom, farmerAcc)
				} else {
					k.SetQueuedStaking(ctx, endTime, coin.Denom, farmerAcc, queuedStaking)
				}
				unstaked = unstaked.Add(amtToUnstake)
				if unstaked.Equal(coin.Amount) { // Fully unstaked from queued stakings, so stop.
					return true
				}
			}
			return false
		})

		amtToUnstake := coin.Amount.Sub(unstaked)
		if amtToUnstake.IsPositive() {
			// If there is more to unstake, then unstake from staked coins.
			staking, found := k.GetStaking(ctx, coin.Denom, farmerAcc)
			if !found {
				staking.Amount = sdk.ZeroInt()
			}
			if staking.Amount.LT(amtToUnstake) {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInsufficientFunds, "not enough staked coins, %s%s is less than %s%s",
					unstaked.Add(staking.Amount), coin.Denom, unstaked.Add(amtToUnstake), coin.Denom)
			}

			if found {
				// Harvest rewards(send rewards to the farmer) when unstaking
				// whole staked coins.
				harvest := amtToUnstake.Equal(staking.Amount)

				if _, err := k.WithdrawRewards(ctx, farmerAcc, coin.Denom, harvest); err != nil {
					return err
				}

				if harvest {
					unharvested, found := k.GetUnharvestedRewards(ctx, farmerAcc, coin.Denom)
					if found {
						totalUnharvestedRewards = totalUnharvestedRewards.Add(unharvested.Rewards...)
						k.DeleteUnharvestedRewards(ctx, farmerAcc, coin.Denom)
					}
				}
			}

			staking.Amount = staking.Amount.Sub(amtToUnstake)
			if staking.Amount.IsPositive() {
				currentEpoch := k.GetCurrentEpoch(ctx, coin.Denom)
				staking.StartingEpoch = currentEpoch
				k.SetStaking(ctx, coin.Denom, farmerAcc, staking)
			} else {
				k.DeleteStaking(ctx, coin.Denom, farmerAcc)
			}

			k.DecreaseTotalStakings(ctx, coin.Denom, amtToUnstake)
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, types.UnharvestedRewardsReserveAcc, farmerAcc, totalUnharvestedRewards); err != nil {
		return err
	}

	if err := k.ReleaseStakingCoins(ctx, farmerAcc, amount); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyFarmer, farmerAcc.String()),
			sdk.NewAttribute(types.AttributeKeyUnstakingCoins, amount.String()),
		),
	})

	return nil
}

// ProcessQueuedCoins moves queued coins into staked coins.
// It causes accumulated rewards to be withdrawn as UnharvestedRewards,
// which can be claimed later by the farmer.
func (k Keeper) ProcessQueuedCoins(ctx sdk.Context, currTime time.Time) {
	type farmerDenomPair struct {
		farmerAcc        string
		stakingCoinDenom string
	}
	newStakingMap := map[farmerDenomPair]sdk.Int{} // (farmerAcc, stakingCoinDenom) => newStakingAmt
	newTotalStakingsMap := map[string]sdk.Int{}    // stakingCoinDenom => newTotalStakingsAmt

	k.IterateMatureQueuedStakings(ctx, currTime, func(endTime time.Time, stakingCoinDenom string, farmerAcc sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		newStakingKey := farmerDenomPair{farmerAcc.String(), stakingCoinDenom}
		newStakingAmt, ok := newStakingMap[newStakingKey]
		if !ok {
			newStakingAmt = sdk.ZeroInt()
		}
		newStakingMap[newStakingKey] = newStakingAmt.Add(queuedStaking.Amount)

		newTotalStakingAmt, ok := newTotalStakingsMap[stakingCoinDenom]
		if !ok {
			newTotalStakingAmt = sdk.ZeroInt()
		}
		newTotalStakingsMap[stakingCoinDenom] = newTotalStakingAmt.Add(queuedStaking.Amount)

		k.DeleteQueuedStaking(ctx, endTime, stakingCoinDenom, farmerAcc)

		return false
	})

	// Sort newTotalStakingsMap keys.
	var newTotalStakingsKeys []string
	for key := range newTotalStakingsMap {
		newTotalStakingsKeys = append(newTotalStakingsKeys, key)
	}
	sort.Strings(newTotalStakingsKeys)

	// Increase total stakings first.
	for _, key := range newTotalStakingsKeys {
		k.IncreaseTotalStakings(ctx, key, newTotalStakingsMap[key])
	}

	// Sort newStakingMap keys.
	var newStakingKeys []farmerDenomPair
	for key := range newStakingMap {
		newStakingKeys = append(newStakingKeys, key)
	}
	sort.Slice(newStakingKeys, func(i, j int) bool {
		if newStakingKeys[i].farmerAcc == newStakingKeys[j].farmerAcc {
			return newStakingKeys[i].stakingCoinDenom < newStakingKeys[j].stakingCoinDenom
		}
		return newStakingKeys[i].farmerAcc < newStakingKeys[j].farmerAcc
	})

	// Increase staking amount and withdraw rewards if there was already
	// a staking.
	for _, key := range newStakingKeys {
		stakingCoinDenom := key.stakingCoinDenom
		farmerAcc, _ := sdk.AccAddressFromBech32(key.farmerAcc)
		newStakingAmt := newStakingMap[key]

		staking, found := k.GetStaking(ctx, stakingCoinDenom, farmerAcc)
		if found {
			if _, err := k.WithdrawRewards(ctx, farmerAcc, stakingCoinDenom, false); err != nil {
				panic(err)
			}
		} else {
			staking.Amount = sdk.ZeroInt()
		}

		k.SetStaking(ctx, stakingCoinDenom, farmerAcc, types.Staking{
			Amount:        staking.Amount.Add(newStakingAmt),
			StartingEpoch: k.GetCurrentEpoch(ctx, stakingCoinDenom),
		})
	}
}

// ValidateStakingReservedAmount checks that the balance of
// StakingReserveAcc greater than the amount of staked, queued coins in all
// staking objects.
func (k Keeper) ValidateStakingReservedAmount(ctx sdk.Context) error {
	reservedCoins := sdk.NewCoins()
	k.IterateStakings(ctx, func(stakingCoinDenom string, _ sdk.AccAddress, staking types.Staking) (stop bool) {
		reservedCoins = reservedCoins.Add(sdk.NewCoin(stakingCoinDenom, staking.Amount))
		return false
	})
	k.IterateQueuedStakings(ctx, func(_ time.Time, stakingCoinDenom string, _ sdk.AccAddress, queuedStaking types.QueuedStaking) (stop bool) {
		reservedCoins = reservedCoins.Add(sdk.NewCoin(stakingCoinDenom, queuedStaking.Amount))
		return false
	})

	for _, coin := range reservedCoins {
		balanceStakingReserveAcc := k.bankKeeper.SpendableCoins(ctx, types.StakingReserveAcc(coin.Denom))
		if !balanceStakingReserveAcc.IsAllGTE(sdk.Coins{coin}) {
			return types.ErrInvalidStakingReservedAmount
		}
	}

	return nil
}
