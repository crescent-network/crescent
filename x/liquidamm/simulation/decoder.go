package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding liquidamm type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.PublicPositionKeyPrefix):
			var lA, lB types.PublicPosition
			cdc.MustUnmarshal(kvA.Value, &lA)
			cdc.MustUnmarshal(kvB.Value, &lB)
			return fmt.Sprintf("%v\n%v", lA, lB)

		case bytes.Equal(kvA.Key[:1], types.RewardsAuctionKeyPrefix):
			var rA, rB types.RewardsAuction
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)

		case bytes.Equal(kvA.Key[:1], types.BidKeyPrefix):
			var bA, bB types.Bid
			cdc.MustUnmarshal(kvA.Value, &bA)
			cdc.MustUnmarshal(kvB.Value, &bB)
			return fmt.Sprintf("%v\n%v", bA, bB)

		default:
			panic(fmt.Sprintf("invalid liquidamm key prefix %X", kvA.Key[:1]))
		}
	}
}
