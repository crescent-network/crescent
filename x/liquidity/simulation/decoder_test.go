package simulation_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/simulation"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func TestDecodeLiquidityStore(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	pair := types.NewPair(1, "denom1", "denom2")
	pool := types.NewBasicPool(1, 1, utils.TestAddress(0))
	depositReq := types.DepositRequest{
		Id:             1,
		PoolId:         1,
		MsgHeight:      1,
		Depositor:      sdk.AccAddress(crypto.AddressHash([]byte("depositor"))).String(),
		MintedPoolCoin: sdk.NewInt64Coin("pool1", 0),
		Status:         types.RequestStatusNotExecuted,
	}
	withdrawReq := types.WithdrawRequest{
		Id:         1,
		PoolId:     1,
		MsgHeight:  1,
		Withdrawer: sdk.AccAddress(crypto.AddressHash([]byte("withdrawer"))).String(),
		PoolCoin:   sdk.NewInt64Coin("pool1", 1000000),
		Status:     types.RequestStatusNotExecuted,
	}
	order := types.Order{
		Id:                 1,
		PairId:             1,
		MsgHeight:          1,
		Orderer:            sdk.AccAddress(crypto.AddressHash([]byte("orderer"))).String(),
		Direction:          types.OrderDirectionSell,
		OfferCoin:          sdk.NewInt64Coin("denom1", 1000000),
		RemainingOfferCoin: sdk.NewInt64Coin("denom1", 500000),
		ReceivedCoin:       sdk.NewInt64Coin("denom2", 500000),
		Price:              utils.ParseDec("1.0"),
		Amount:             sdk.NewInt(1000000),
		OpenAmount:         sdk.NewInt(500000),
		BatchId:            1,
		ExpireAt:           utils.ParseTime("2022-02-01T00:00:00Z"),
		Status:             types.OrderStatusPartiallyMatched,
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.PairKeyPrefix, Value: cdc.MustMarshal(&pair)},
			{Key: types.PoolKeyPrefix, Value: cdc.MustMarshal(&pool)},
			{Key: types.DepositRequestKeyPrefix, Value: cdc.MustMarshal(&depositReq)},
			{Key: types.WithdrawRequestKeyPrefix, Value: cdc.MustMarshal(&withdrawReq)},
			{Key: types.OrderKeyPrefix, Value: cdc.MustMarshal(&order)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Pair", fmt.Sprintf("%v\n%v", pair, pair)},
		{"Pool", fmt.Sprintf("%v\n%v", pool, pool)},
		{"DepositRequest", fmt.Sprintf("%v\n%v", depositReq, depositReq)},
		{"WithdrawRequest", fmt.Sprintf("%v\n%v", withdrawReq, withdrawReq)},
		{"OrderRequest", fmt.Sprintf("%v\n%v", order, order)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
