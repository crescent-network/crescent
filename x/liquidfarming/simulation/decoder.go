package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding liquidfarming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.LiquidFarmKeyPrefix):
			var lA, lB types.LiquidFarm
			cdc.MustUnmarshal(kvA.Value, &lA)
			cdc.MustUnmarshal(kvB.Value, &lB)
			return fmt.Sprintf("%v\n%v", lA, lB)

		case bytes.Equal(kvA.Key[:1], types.CompoundingRewardsKeyPrefix):
			var cA, cB types.CompoundingRewards
			cdc.MustUnmarshal(kvA.Value, &cA)
			cdc.MustUnmarshal(kvB.Value, &cB)
			return fmt.Sprintf("%v\n%v", cA, cB)

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
			panic(fmt.Sprintf("invalid liquid farm key prefix %X", kvA.Key[:1]))
		}
	}
}
