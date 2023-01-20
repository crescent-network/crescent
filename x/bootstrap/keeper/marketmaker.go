package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// GetBootstrap returns market maker object for a given
// address and pair id.
func (k Keeper) GetBootstrap(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) (mm types.Bootstrap, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBootstrapKey(mmAddr, pairId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &mm)
	found = true
	return
}

// SetBootstrap sets a market maker.
func (k Keeper) SetBootstrap(ctx sdk.Context, mm types.Bootstrap) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&mm)
	mmAddr := mm.GetAccAddress()
	store.Set(types.GetBootstrapKey(mmAddr, mm.PairId), bz)
	store.Set(types.GetBootstrapIndexByPairIdKey(mm.PairId, mmAddr), []byte{})
}

// DeleteBootstrap deletes market maker for a given address and pair id.
func (k Keeper) DeleteBootstrap(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetBootstrapKey(mmAddr, pairId))
	store.Delete(types.GetBootstrapIndexByPairIdKey(pairId, mmAddr))
}

// GetDeposit returns market maker deposit object for a given
// address and pair id.
func (k Keeper) GetDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) (mm types.Deposit, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDepositKey(mmAddr, pairId))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &mm)
	found = true
	return
}

// SetDeposit sets a deposit.
func (k Keeper) SetDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64, amount sdk.Coins) {
	var deposit types.Deposit
	deposit.Amount = amount
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&deposit)
	store.Set(types.GetDepositKey(mmAddr, pairId), bz)
}

// DeleteDeposit deletes deposit object for a given address and pair id.
func (k Keeper) DeleteDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDepositKey(mmAddr, pairId))
}

// IterateBootstraps iterates through all market makers
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateBootstraps(ctx sdk.Context, cb func(mm types.Bootstrap) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.BootstrapKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.Bootstrap
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// IterateBootstrapsByAddr iterates through all market makers by an address
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateBootstrapsByAddr(ctx sdk.Context, mmAddr sdk.AccAddress, cb func(mm types.Bootstrap) (stop bool)) {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GetBootstrapByAddrPrefix(mmAddr))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.Bootstrap
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// IterateBootstrapsByPairId iterates through all market makers by an pair id
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateBootstrapsByPairId(ctx sdk.Context, pairId uint64, cb func(mm types.Bootstrap) (stop bool)) {
	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GetBootstrapByPairIdPrefix(pairId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		pairId, mmAddr := types.ParseBootstrapIndexByPairIdKey(iter.Key())
		mm, _ := k.GetBootstrap(ctx, mmAddr, pairId)
		if cb(mm) {
			break
		}
	}
}

// GetAllBootstraps returns all market makers
func (k Keeper) GetAllBootstraps(ctx sdk.Context) []types.Bootstrap {
	mms := []types.Bootstrap{}
	k.IterateBootstraps(ctx, func(mm types.Bootstrap) (stop bool) {
		mms = append(mms, mm)
		return false
	})
	return mms
}

// IterateDeposits iterates through all apply deposits
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateDeposits(ctx sdk.Context, cb func(id types.Deposit) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DepositKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.Deposit
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// IterateDepositRecords iterates through all deposits
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateDepositRecords(ctx sdk.Context, cb func(idr types.DepositRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DepositKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var id types.Deposit
		k.cdc.MustUnmarshal(iter.Value(), &id)
		mmAddr, pairId := types.ParseDepositKey(iter.Key())
		record := types.DepositRecord{
			Address: mmAddr.String(),
			PairId:  pairId,
			Amount:  id.Amount,
		}
		if cb(record) {
			break
		}
	}
}

// GetAllDeposits returns all deposits
func (k Keeper) GetAllDeposits(ctx sdk.Context) []types.Deposit {
	ids := []types.Deposit{}
	k.IterateDeposits(ctx, func(id types.Deposit) (stop bool) {
		ids = append(ids, id)
		return false
	})
	return ids
}

// GetAllDepositRecords returns all deposit records
func (k Keeper) GetAllDepositRecords(ctx sdk.Context) []types.DepositRecord {
	idrs := []types.DepositRecord{}
	k.IterateDepositRecords(ctx, func(idr types.DepositRecord) (stop bool) {
		idrs = append(idrs, idr)
		return false
	})
	return idrs
}

// GetIncentive returns claimable incentive object for a given address.
func (k Keeper) GetIncentive(ctx sdk.Context, mmAddr sdk.AccAddress) (incentive types.Incentive, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetIncentiveKey(mmAddr))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &incentive)
	found = true
	return
}

// SetIncentive sets claimable incentive.
func (k Keeper) SetIncentive(ctx sdk.Context, incentive types.Incentive) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&incentive)
	store.Set(types.GetIncentiveKey(incentive.GetAccAddress()), bz)
}

// DeleteIncentive deletes market maker claimable incentive for a given address.
func (k Keeper) DeleteIncentive(ctx sdk.Context, mmAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetIncentiveKey(mmAddr))
}

// IterateIncentives iterates through all incentives
// stored in the store and invokes callback function for each item.
// Stops the iteration when the callback function returns true.
func (k Keeper) IterateIncentives(ctx sdk.Context, cb func(incentive types.Incentive) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.IncentiveKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var record types.Incentive
		k.cdc.MustUnmarshal(iter.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllIncentives returns all incentives
func (k Keeper) GetAllIncentives(ctx sdk.Context) []types.Incentive {
	incentives := []types.Incentive{}
	k.IterateIncentives(ctx, func(incentive types.Incentive) (stop bool) {
		incentives = append(incentives, incentive)
		return false
	})
	return incentives
}

func (k Keeper) ApplyBootstrap(ctx sdk.Context, mmAddr sdk.AccAddress, pairIds []uint64) error {
	params := k.GetParams(ctx)
	incentivePairsMap := params.IncentivePairsMap()

	totalDepositAmt := sdk.Coins{}
	for _, pairId := range pairIds {
		// Fail if the same market maker already exists
		_, found := k.GetBootstrap(ctx, mmAddr, pairId)
		if found {
			return types.ErrAlreadyExistBootstrap
		}
		totalDepositAmt = totalDepositAmt.Add(params.DepositAmount...)

		// fail for pairs that are not registered as incentive pairs on params
		if _, ok := incentivePairsMap[pairId]; !ok {
			return types.ErrUnregisteredPairId
		}
	}

	// total deposit amount = deposit amount * number of pairs
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, mmAddr, types.ModuleName, totalDepositAmt)
	if err != nil {
		return err
	}

	// create market maker, deposit object for each pair
	for _, pairId := range pairIds {
		k.SetDeposit(ctx, mmAddr, pairId, params.DepositAmount)
		k.SetBootstrap(ctx, types.Bootstrap{
			Address:  mmAddr.String(),
			PairId:   pairId,
			Eligible: false,
		})
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeApplyBootstrap,
			sdk.NewAttribute(types.AttributeKeyAddress, mmAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPairIds, strings.Trim(strings.Replace(fmt.Sprint(pairIds), " ", ",", -1), "[]")),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeApplyBootstrap),
			sdk.NewAttribute(sdk.AttributeKeySender, mmAddr.String()),
		),
	})

	return nil
}

func (k Keeper) ClaimIncentives(ctx sdk.Context, mmAddr sdk.AccAddress) error {
	incentive, found := k.GetIncentive(ctx, mmAddr)
	if !found {
		return types.ErrEmptyClaimableIncentive
	}

	if err := k.bankKeeper.SendCoins(ctx, types.ClaimableIncentiveReserveAcc, mmAddr, incentive.Claimable); err != nil {
		return err
	}

	k.DeleteIncentive(ctx, mmAddr)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaimIncentives,
			sdk.NewAttribute(types.AttributeKeyAddress, mmAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, incentive.Claimable.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.EventTypeClaimIncentives),
			sdk.NewAttribute(sdk.AttributeKeySender, mmAddr.String()),
		),
	})
	return nil
}

func (k Keeper) ValidateDepositReservedAmount(ctx sdk.Context) error {
	mmCount := 0
	depositCount := 0
	var totalAmt sdk.Coins
	k.IterateBootstraps(ctx, func(mm types.Bootstrap) (stop bool) {
		if !mm.Eligible {
			mmCount += 1
		}
		return false
	})
	k.IterateDeposits(ctx, func(id types.Deposit) (stop bool) {
		depositCount += 1
		totalAmt = totalAmt.Add(id.Amount...)
		return false
	})
	if mmCount != depositCount {
		return fmt.Errorf("market maker number differs from the actual value; have %d, want %d", mmCount, depositCount)
	}

	if !totalAmt.Empty() {
		reserveBalance := k.bankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		if !reserveBalance.IsAllGTE(totalAmt) {
			return fmt.Errorf("DepositReserveAcc differs from the actual value; have %s, want %s", reserveBalance, totalAmt)
		}
	}
	return nil
}

func (k Keeper) ValidateIncentiveReservedAmount(ctx sdk.Context, incentives []types.Incentive) error {
	var totalClaimable sdk.Coins
	for _, record := range incentives {
		totalClaimable = totalClaimable.Add(record.Claimable...)
	}
	if !totalClaimable.Empty() {
		reserveBalance := k.bankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)
		if !reserveBalance.IsAllGTE(totalClaimable) {
			return fmt.Errorf("ClaimableIncentiveReserveAcc differs from the actual value; have %s, want %s", reserveBalance, totalClaimable)
		}
	}
	return nil
}
