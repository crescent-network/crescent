package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.LastMarketIdKey):
			idA := sdk.BigEndianToUint64(kvA.Value)
			idB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", idA, idB)

		case bytes.Equal(kvA.Key, types.LastOrderIdKey):
			idA := sdk.BigEndianToUint64(kvA.Value)
			idB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", idA, idB)

		case bytes.Equal(kvA.Key[:1], types.MarketKeyPrefix):
			var mA, mB types.Market
			cdc.MustUnmarshal(kvA.Value, &mA)
			cdc.MustUnmarshal(kvB.Value, &mB)
			return fmt.Sprintf("%v\n%v", mA, mB)

		case bytes.Equal(kvA.Key[:1], types.MarketStateKeyPrefix):
			var msA, msB types.MarketState
			cdc.MustUnmarshal(kvA.Value, &msA)
			cdc.MustUnmarshal(kvB.Value, &msB)
			return fmt.Sprintf("%v\n%v", msA, msB)

		case bytes.Equal(kvA.Key[:1], types.OrderKeyPrefix):
			var oA, oB types.Order
			cdc.MustUnmarshal(kvA.Value, &oA)
			cdc.MustUnmarshal(kvB.Value, &oB)
			return fmt.Sprintf("%v\n%v", oA, oB)

		default:
			panic(fmt.Sprintf("invalid exchange key prefix %X", kvA.Key[:1]))
		}
	}
}
