package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding liquidstaking type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.LiquidValidatorsKey):
			var cA, cB types.LiquidValidator
			cdc.MustUnmarshal(kvA.Value, &cA)
			cdc.MustUnmarshal(kvA.Value, &cB)
			return fmt.Sprintf("%v\n%v", cA, cB)

		default:
			panic(fmt.Sprintf("invalid liquidstaking key prefix %X", kvA.Key[:1]))
		}
	}
}
