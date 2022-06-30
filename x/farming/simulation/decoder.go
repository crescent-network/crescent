package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding farming type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.PlanKeyPrefix):
			var pA, pB types.BasePlan
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.StakingKeyPrefix):
			var sA, sB types.Staking
			cdc.MustUnmarshal(kvA.Value, &sA)
			cdc.MustUnmarshal(kvB.Value, &sB)
			return fmt.Sprintf("%v\n%v", sA, sB)

		case bytes.Equal(kvA.Key[:1], types.QueuedStakingKeyPrefix):
			var sA, sB types.QueuedStaking
			cdc.MustUnmarshal(kvA.Value, &sA)
			cdc.MustUnmarshal(kvB.Value, &sB)
			return fmt.Sprintf("%v\n%v", sA, sB)

		case bytes.Equal(kvA.Key[:1], types.HistoricalRewardsKeyPrefix):
			var rA, rB types.HistoricalRewards
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)

		case bytes.Equal(kvA.Key[:1], types.OutstandingRewardsKeyPrefix):
			var rA, rB types.OutstandingRewards
			cdc.MustUnmarshal(kvA.Value, &rA)
			cdc.MustUnmarshal(kvB.Value, &rB)
			return fmt.Sprintf("%v\n%v", rA, rB)

		default:
			panic(fmt.Sprintf("invalid farming key prefix %X", kvA.Key[:1]))
		}
	}
}
