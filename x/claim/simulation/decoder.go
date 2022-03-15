package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/cosmosquad-labs/squad/x/claim/types"
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

		case bytes.Equal(kvA.Key[:1], types.StartTimeKeyPrefix):
			stA, err := sdk.ParseTimeBytes(kvA.Value)
			if err != nil {
				panic(err)
			}
			stB, err := sdk.ParseTimeBytes(kvB.Value)
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", stA, stB)

		case bytes.Equal(kvA.Key[:1], types.EndTimeKeyPrefix):
			etA, err := sdk.ParseTimeBytes(kvA.Value)
			if err != nil {
				panic(err)
			}
			etB, err := sdk.ParseTimeBytes(kvB.Value)
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", etA, etB)

		default:
			panic(fmt.Sprintf("invalid claim key prefix %X", kvA.Key[:1]))
		}
	}
}
