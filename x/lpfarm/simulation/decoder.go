package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.PlanKeyPrefix):
			var pA, pB types.Plan
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.FarmKeyPrefix):
			var fA, fB types.Farm
			cdc.MustUnmarshal(kvA.Value, &fA)
			cdc.MustUnmarshal(kvB.Value, &fB)
			return fmt.Sprintf("%v\n%v", fA, fB)

		case bytes.Equal(kvA.Key[:1], types.PositionKeyPrefix):
			var pA, pB types.Position
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.HistoricalRewardsKeyPrefix):
			var hA, hB types.HistoricalRewards
			cdc.MustUnmarshal(kvA.Value, &hA)
			cdc.MustUnmarshal(kvB.Value, &hB)
			return fmt.Sprintf("%v\n%v", hA, hB)

		default:
			panic(fmt.Sprintf("invalid lpfarm key prefix %X", kvA.Key[:1]))
		}
	}
}
