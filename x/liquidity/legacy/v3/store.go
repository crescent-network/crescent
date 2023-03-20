package v3

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MigrateOrders(store sdk.KVStore, cdc codec.BinaryCodec) error {
	//iter := sdk.KVStorePrefixIterator(store, types.OrderKeyPrefix)
	//defer iter.Close()
	//
	//for ; iter.Valid(); iter.Next() {
	//	var oldOrder v2liquidity.Order
	//	if err := cdc.Unmarshal(iter.Value(), &oldOrder); err != nil {
	//		return err
	//	}
	//
	//	newOrder := types.Order{
	//		// There's no way to determine whether the order was made through
	//		// MsgLimitOrder or MsgMarketOrder, set the order type as OrderTypeLimit
	//		// as a fallback.
	//		Type:      types.OrderTypeLimit,
	//		Id:        oldOrder.Id,
	//		PairId:    oldOrder.PairId,
	//		MsgHeight: oldOrder.MsgHeight,
	//		Orderer:   oldOrder.Orderer,
	//		// Only the type has changed, not the value, so simply type-cast here
	//		Direction:          types.OrderDirection(oldOrder.Direction),
	//		OfferCoin:          oldOrder.OfferCoin,
	//		RemainingOfferCoin: oldOrder.RemainingOfferCoin,
	//		ReceivedCoin:       oldOrder.ReceivedCoin,
	//		Price:              oldOrder.Price,
	//		Amount:             oldOrder.Amount,
	//		OpenAmount:         oldOrder.OpenAmount,
	//		BatchId:            oldOrder.BatchId,
	//		ExpireAt:           oldOrder.ExpireAt,
	//		// Only the type has changed, not the value, so simply type-cast here
	//		Status: types.OrderStatus(oldOrder.Status),
	//	}
	//
	//	bz, err := cdc.Marshal(&newOrder)
	//	if err != nil {
	//		return err
	//	}
	//	store.Set(iter.Key(), bz)
	//}
	//
	return nil
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	if err := MigrateOrders(store, cdc); err != nil {
		return err
	}
	return nil
}
