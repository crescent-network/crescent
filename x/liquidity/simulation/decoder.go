package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.PairKeyPrefix):
			var pairA, pairB types.Pair
			cdc.MustUnmarshal(kvA.Value, &pairA)
			cdc.MustUnmarshal(kvB.Value, &pairB)
			return fmt.Sprintf("%v\n%v", pairA, pairB)

		case bytes.Equal(kvA.Key[:1], types.PoolKeyPrefix):
			var poolA, poolB types.Pool
			cdc.MustUnmarshal(kvA.Value, &poolA)
			cdc.MustUnmarshal(kvB.Value, &poolB)
			return fmt.Sprintf("%v\n%v", poolA, poolB)

		case bytes.Equal(kvA.Key[:1], types.DepositRequestKeyPrefix):
			var reqA, reqB types.DepositRequest
			cdc.MustUnmarshal(kvA.Value, &reqA)
			cdc.MustUnmarshal(kvB.Value, &reqB)
			return fmt.Sprintf("%v\n%v", reqA, reqB)

		case bytes.Equal(kvA.Key[:1], types.WithdrawRequestKeyPrefix):
			var reqA, reqB types.WithdrawRequest
			cdc.MustUnmarshal(kvA.Value, &reqA)
			cdc.MustUnmarshal(kvB.Value, &reqB)
			return fmt.Sprintf("%v\n%v", reqA, reqB)

		case bytes.Equal(kvA.Key[:1], types.OrderKeyPrefix):
			var reqA, reqB types.Order
			cdc.MustUnmarshal(kvA.Value, &reqA)
			cdc.MustUnmarshal(kvB.Value, &reqB)
			return fmt.Sprintf("%v\n%v", reqA, reqB)

		default:
			panic(fmt.Sprintf("invalid liquidity key prefix %X", kvA.Key[:1]))
		}
	}
}
