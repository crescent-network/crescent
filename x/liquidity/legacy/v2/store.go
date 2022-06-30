package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1liquidity "github.com/crescent-network/crescent/v2/x/liquidity/legacy/v1"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func MigratePool(store sdk.KVStore, cdc codec.BinaryCodec) error {
	iter := sdk.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldPool v1liquidity.Pool
		if err := cdc.Unmarshal(iter.Value(), &oldPool); err != nil {
			return err
		}

		newPool := types.Pool{
			Type:                  types.PoolTypeBasic,
			Id:                    oldPool.Id,
			PairId:                oldPool.PairId,
			Creator:               "",
			ReserveAddress:        oldPool.ReserveAddress,
			PoolCoinDenom:         oldPool.PoolCoinDenom,
			MinPrice:              nil,
			MaxPrice:              nil,
			LastDepositRequestId:  oldPool.LastDepositRequestId,
			LastWithdrawRequestId: oldPool.LastWithdrawRequestId,
			Disabled:              oldPool.Disabled,
		}
		bz, err := cdc.Marshal(&newPool)
		if err != nil {
			return err
		}
		store.Set(iter.Key(), bz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	if err := MigratePool(store, cdc); err != nil {
		return err
	}
	return nil
}
