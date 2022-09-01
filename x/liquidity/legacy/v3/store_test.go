package v3_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	v2liquidity "github.com/crescent-network/crescent/v2/x/liquidity/legacy/v2"
	v3liquidity "github.com/crescent-network/crescent/v2/x/liquidity/legacy/v3"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func TestMigrateOrders(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	storeKey := sdk.NewKVStoreKey("liquidity")
	ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(storeKey)

	oldOrder := v2liquidity.Order{
		Id:                 1,
		PairId:             1,
		MsgHeight:          10000,
		Orderer:            utils.TestAddress(0).String(),
		Direction:          v2liquidity.OrderDirectionSell,
		OfferCoin:          utils.ParseCoin("1000000denom1"),
		RemainingOfferCoin: utils.ParseCoin("500000denom1"),
		ReceivedCoin:       utils.ParseCoin("250000denom2"),
		Price:              utils.ParseDec("2"),
		Amount:             sdk.NewInt(500000),
		OpenAmount:         sdk.NewInt(250000),
		BatchId:            1000,
		ExpireAt:           utils.ParseTime("2022-01-01T12:00:00Z"),
		Status:             v2liquidity.OrderStatusPartiallyMatched,
	}
	oldOrderValue := cdc.MustMarshal(&oldOrder)
	key := types.GetOrderKey(oldOrder.PairId, oldOrder.Id)
	store.Set(key, oldOrderValue)

	require.NoError(t, v3liquidity.MigrateStore(ctx, storeKey, cdc))

	newOrder := types.Order{
		Type:               types.OrderTypeLimit,
		Id:                 1,
		PairId:             1,
		MsgHeight:          10000,
		Orderer:            utils.TestAddress(0).String(),
		Direction:          types.OrderDirectionSell,
		OfferCoin:          utils.ParseCoin("1000000denom1"),
		RemainingOfferCoin: utils.ParseCoin("500000denom1"),
		ReceivedCoin:       utils.ParseCoin("250000denom2"),
		Price:              utils.ParseDec("2"),
		Amount:             sdk.NewInt(500000),
		OpenAmount:         sdk.NewInt(250000),
		BatchId:            1000,
		ExpireAt:           utils.ParseTime("2022-01-01T12:00:00Z"),
		Status:             types.OrderStatusPartiallyMatched,
	}
	newOrderValue := cdc.MustMarshal(&newOrder)
	require.Equal(t, newOrderValue, store.Get(key))
}
