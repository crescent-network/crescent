package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/simulation"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := chain.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	market := types.NewMarket(
		10, "ucre", "uusd",
		types.DefaultFees,
		types.DefaultOrderQuantityLimits, types.DefaultOrderQuoteLimits)
	marketState := types.NewMarketState(utils.ParseDecP("12.345"))
	order := types.NewOrder(
		1, types.OrderTypeLimit, utils.TestAddress(1), 10, false, utils.ParseDec("12.345"), sdk.NewInt(100_000000),
		200, sdk.NewInt(90_000000), sdk.NewInt(90_000000), utils.ParseTime("2023-06-01T00:00:00Z"))

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.LastMarketIdKey, Value: sdk.Uint64ToBigEndian(10)},
			{Key: types.LastOrderIdKey, Value: sdk.Uint64ToBigEndian(100)},
			{Key: types.GetMarketKey(10), Value: cdc.MustMarshal(&market)},
			{Key: types.GetMarketStateKey(10), Value: cdc.MustMarshal(&marketState)},
			{Key: types.GetOrderKey(100), Value: cdc.MustMarshal(&order)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"LastMarketId", fmt.Sprintf("%d\n%d", 10, 10)},
		{"LastOrderId", fmt.Sprintf("%d\n%d", 100, 100)},
		{"Market", fmt.Sprintf("%v\n%v", market, market)},
		{"MarketState", fmt.Sprintf("%v\n%v", marketState, marketState)},
		{"Order", fmt.Sprintf("%v\n%v", order, order)},
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
