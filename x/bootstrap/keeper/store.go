package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// GetBootstrapPool returns bootstrap pool object for a given id
func (k Keeper) GetBootstrapPool(ctx sdk.Context, id uint64) (mm types.BootstrapPool, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBootstrapPoolKey(id))
	if bz == nil {
		return
	}
	k.cdc.MustUnmarshal(bz, &mm)
	found = true
	return
}

// SetBootstrapPool sets a bootstrap pool.
func (k Keeper) SetBootstrapPool(ctx sdk.Context, pool types.BootstrapPool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.GetBootstrapPoolKey(pool.Id), bz)
}

// GetLastBootstrapPoolId returns the last bootstrap pool id.
func (k Keeper) GetLastBootstrapPoolId(ctx sdk.Context) (id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastBootstrapPoolIdKey)
	if bz == nil {
		id = 0 // initialize the pair id
	} else {
		var val gogotypes.UInt64Value
		k.cdc.MustUnmarshal(bz, &val)
		id = val.GetValue()
	}
	return
}

// SetLastBootstrapPoolId stores the last pool id.
func (k Keeper) SetLastBootstrapPoolId(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})
	store.Set(types.LastBootstrapPoolIdKey, bz)
}

// getNextBootstrapPoolIdWithUpdate increments bootstrap pool id by one and set it.
func (k Keeper) getNextBootstrapPoolIdWithUpdate(ctx sdk.Context) uint64 {
	id := k.GetLastBootstrapPoolId(ctx) + 1
	k.SetLastBootstrapPoolId(ctx, id)
	return id
}

// GetOrder returns the particular order.
func (k Keeper) GetOrder(ctx sdk.Context, poolId, id uint64) (order types.Order, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOrderKey(poolId, id))
	if bz == nil {
		return
	}
	order = types.MustUnmarshalOrder(k.cdc, bz)
	return order, true
}

// SetOrder stores an order for the batch execution.
func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshaOrder(k.cdc, order)
	store.Set(types.GetOrderKey(order.BootstrapPoolId, order.Id), bz)
}

func (k Keeper) SetOrderIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOrderIndexKey(order.GetOrderer(), order.BootstrapPoolId, order.Id), []byte{})
}

// GetLastOrderId returns the last order id of the pool.
func (k Keeper) GetLastOrderId(ctx sdk.Context, poolId uint64) (id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLastOrderIdIndexKey(poolId))
	if bz == nil {
		id = 0 // initialize the order id
	} else {
		var val gogotypes.UInt64Value
		k.cdc.MustUnmarshal(bz, &val)
		id = val.GetValue()
	}
	return
}

func (k Keeper) SetLastOrderId(ctx sdk.Context, poolId, orderId uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: orderId})
	store.Set(types.GetLastOrderIdIndexKey(poolId), bz)
}

// getNextOrderIdWithUpdate increments the pool's last order id and returns it.
func (k Keeper) getNextOrderIdWithUpdate(ctx sdk.Context, poolId uint64) uint64 {
	id := k.GetLastOrderId(ctx, poolId) + 1
	k.SetLastOrderId(ctx, poolId, id)
	return id
}

// IterateAllOrders iterates through all orders in the store and all
// cb for each order.
func (k Keeper) IterateAllOrders(ctx sdk.Context, cb func(order types.Order) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		order := types.MustUnmarshalOrder(k.cdc, iter.Value())
		stop, err := cb(order)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// IterateOrdersByPool iterates through all the orders within the pool
// and call cb for each order.
func (k Keeper) IterateOrdersByPool(ctx sdk.Context, poolId uint64, cb func(order types.Order) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrdersByPoolKeyPrefix(poolId))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		order := types.MustUnmarshalOrder(k.cdc, iter.Value())
		stop, err := cb(order)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// IterateOrdersByOrderer iterates through orders in the store by an orderer
// and call cb on each order.
func (k Keeper) IterateOrdersByOrderer(ctx sdk.Context, orderer sdk.AccAddress, cb func(order types.Order) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrderIndexKeyByOrdererPrefix(orderer))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		_, poolId, orderId := types.ParseOrderIndexKey(iter.Key())
		order, _ := k.GetOrder(ctx, poolId, orderId)
		stop, err := cb(order)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetAllOrders returns all orders in the store.
func (k Keeper) GetAllOrders(ctx sdk.Context) (orders []types.Order) {
	orders = []types.Order{}
	_ = k.IterateAllOrders(ctx, func(order types.Order) (stop bool, err error) {
		orders = append(orders, order)
		return false, nil
	})
	return
}

// GetOrdersByPool returns orders within the pool.
func (k Keeper) GetOrdersByPool(ctx sdk.Context, poolId uint64) (orders []types.Order) {
	_ = k.IterateOrdersByPool(ctx, poolId, func(order types.Order) (stop bool, err error) {
		orders = append(orders, order)
		return false, nil
	})
	return
}

// GetOrdersByOrderer returns orders by the orderer.
func (k Keeper) GetOrdersByOrderer(ctx sdk.Context, orderer sdk.AccAddress) (orders []types.Order) {
	_ = k.IterateOrdersByOrderer(ctx, orderer, func(order types.Order) (stop bool, err error) {
		orders = append(orders, order)
		return false, nil
	})
	return
}

// DeleteOrder deletes an order.
func (k Keeper) DeleteOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOrderKey(order.BootstrapPoolId, order.Id))
	k.DeleteOrderIndex(ctx, order)
}

func (k Keeper) DeleteOrderIndex(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOrderIndexKey(order.GetOrderer(), order.BootstrapPoolId, order.Id))
}

//// DeleteBootstrap deletes market maker for a given address and pair id.
//func (k Keeper) DeleteBootstrap(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) {
//	store := ctx.KVStore(k.storeKey)
//	store.Delete(types.GetBootstrapPoolKey(mmAddr, pairId))
//	store.Delete(types.GetBootstrapIndexByPairIdKey(pairId, mmAddr))
//}
//
//// GetDeposit returns market maker deposit object for a given
//// address and pair id.
//func (k Keeper) GetDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) (mm types.Deposit, found bool) {
//	store := ctx.KVStore(k.storeKey)
//	bz := store.Get(types.GetOrderKey(mmAddr, pairId))
//	if bz == nil {
//		return
//	}
//	k.cdc.MustUnmarshal(bz, &mm)
//	found = true
//	return
//}
//
//// SetDeposit sets a deposit.
//func (k Keeper) SetDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64, amount sdk.Coins) {
//	var deposit types.Deposit
//	deposit.Amount = amount
//	store := ctx.KVStore(k.storeKey)
//	bz := k.cdc.MustMarshal(&deposit)
//	store.Set(types.GetOrderKey(mmAddr, pairId), bz)
//}
//
//// DeleteDeposit deletes deposit object for a given address and pair id.
//func (k Keeper) DeleteDeposit(ctx sdk.Context, mmAddr sdk.AccAddress, pairId uint64) {
//	store := ctx.KVStore(k.storeKey)
//	store.Delete(types.GetOrderKey(mmAddr, pairId))
//}
//
//// IterateBootstraps iterates through all market makers
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateBootstraps(ctx sdk.Context, cb func(mm types.Bootstrap) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iter := sdk.KVStorePrefixIterator(store, types.BootstrapKeyPrefix)
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		var record types.Bootstrap
//		k.cdc.MustUnmarshal(iter.Value(), &record)
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// IterateBootstrapsByAddr iterates through all market makers by an address
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateBootstrapsByAddr(ctx sdk.Context, mmAddr sdk.AccAddress, cb func(mm types.Bootstrap) (stop bool)) {
//	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GetBootstrapByAddrPrefix(mmAddr))
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		var record types.Bootstrap
//		k.cdc.MustUnmarshal(iter.Value(), &record)
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// IterateBootstrapsByPairId iterates through all market makers by an pair id
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateBootstrapsByPairId(ctx sdk.Context, pairId uint64, cb func(mm types.Bootstrap) (stop bool)) {
//	iter := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GetBootstrapByPairIdPrefix(pairId))
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		pairId, mmAddr := types.ParseBootstrapIndexByPairIdKey(iter.Key())
//		mm, _ := k.GetBootstrapPool(ctx, mmAddr, pairId)
//		if cb(mm) {
//			break
//		}
//	}
//}
//
//// GetAllBootstraps returns all market makers
//func (k Keeper) GetAllBootstraps(ctx sdk.Context) []types.Bootstrap {
//	mms := []types.Bootstrap{}
//	k.IterateBootstraps(ctx, func(mm types.Bootstrap) (stop bool) {
//		mms = append(mms, mm)
//		return false
//	})
//	return mms
//}
//
//// IterateDeposits iterates through all apply deposits
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateDeposits(ctx sdk.Context, cb func(id types.Deposit) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iter := sdk.KVStorePrefixIterator(store, types.DepositKeyPrefix)
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		var record types.Deposit
//		k.cdc.MustUnmarshal(iter.Value(), &record)
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// IterateDepositRecords iterates through all deposits
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateDepositRecords(ctx sdk.Context, cb func(idr types.DepositRecord) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iter := sdk.KVStorePrefixIterator(store, types.DepositKeyPrefix)
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		var id types.Deposit
//		k.cdc.MustUnmarshal(iter.Value(), &id)
//		mmAddr, pairId := types.ParseDepositKey(iter.Key())
//		record := types.DepositRecord{
//			Address: mmAddr.String(),
//			PairId:  pairId,
//			Amount:  id.Amount,
//		}
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// GetAllDeposits returns all deposits
//func (k Keeper) GetAllDeposits(ctx sdk.Context) []types.Deposit {
//	ids := []types.Deposit{}
//	k.IterateDeposits(ctx, func(id types.Deposit) (stop bool) {
//		ids = append(ids, id)
//		return false
//	})
//	return ids
//}
//
//// GetAllDepositRecords returns all deposit records
//func (k Keeper) GetAllDepositRecords(ctx sdk.Context) []types.DepositRecord {
//	idrs := []types.DepositRecord{}
//	k.IterateDepositRecords(ctx, func(idr types.DepositRecord) (stop bool) {
//		idrs = append(idrs, idr)
//		return false
//	})
//	return idrs
//}
//
//// GetIncentive returns claimable incentive object for a given address.
//func (k Keeper) GetIncentive(ctx sdk.Context, mmAddr sdk.AccAddress) (incentive types.Incentive, found bool) {
//	store := ctx.KVStore(k.storeKey)
//	bz := store.Get(types.GetIncentiveKey(mmAddr))
//	if bz == nil {
//		return
//	}
//	k.cdc.MustUnmarshal(bz, &incentive)
//	found = true
//	return
//}
//
//// SetIncentive sets claimable incentive.
//func (k Keeper) SetIncentive(ctx sdk.Context, incentive types.Incentive) {
//	store := ctx.KVStore(k.storeKey)
//	bz := k.cdc.MustMarshal(&incentive)
//	store.Set(types.GetIncentiveKey(incentive.GetAccAddress()), bz)
//}
//
//// DeleteIncentive deletes market maker claimable incentive for a given address.
//func (k Keeper) DeleteIncentive(ctx sdk.Context, mmAddr sdk.AccAddress) {
//	store := ctx.KVStore(k.storeKey)
//	store.Delete(types.GetIncentiveKey(mmAddr))
//}
//
//// IterateIncentives iterates through all incentives
//// stored in the store and invokes callback function for each item.
//// Stops the iteration when the callback function returns true.
//func (k Keeper) IterateIncentives(ctx sdk.Context, cb func(incentive types.Incentive) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iter := sdk.KVStorePrefixIterator(store, types.IncentiveKeyPrefix)
//	defer iter.Close()
//	for ; iter.Valid(); iter.Next() {
//		var record types.Incentive
//		k.cdc.MustUnmarshal(iter.Value(), &record)
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// GetAllIncentives returns all incentives
//func (k Keeper) GetAllIncentives(ctx sdk.Context) []types.Incentive {
//	incentives := []types.Incentive{}
//	k.IterateIncentives(ctx, func(incentive types.Incentive) (stop bool) {
//		incentives = append(incentives, incentive)
//		return false
//	})
//	return incentives
//}
