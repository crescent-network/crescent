package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/claim/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding claim type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.AirdropKeyPrefix):
			var adA, adB types.Airdrop
			cdc.MustUnmarshal(kvA.Value, &adA)
			cdc.MustUnmarshal(kvB.Value, &adB)
			return fmt.Sprintf("%v\n%v", adA, adB)

		case bytes.Equal(kvA.Key[:1], types.ClaimRecordKeyPrefix):
			var crA, crB types.ClaimRecord
			cdc.MustUnmarshal(kvA.Value, &crA)
			cdc.MustUnmarshal(kvB.Value, &crB)
			return fmt.Sprintf("%v\n%v", crA, crB)

		default:
			panic(fmt.Sprintf("invalid claim key prefix %X", kvA.Key[:1]))
		}
	}
}
