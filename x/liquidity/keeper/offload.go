package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidity/types"
)

func (k Keeper) SetOffloadedOrder(ctx sdk.Context, order types.Order) {
	store := dbadapter.Store{DB: k.offChainDB}
	bz := k.cdc.MustMarshal(&order)
	store.Set(types.GetOffloadedOrderKey(ctx.BlockHeight(), order.PairId, order.Id), bz)
}
