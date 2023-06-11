package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.LastPoolIdKey):
			idA := sdk.BigEndianToUint64(kvA.Value)
			idB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", idA, idB)

		case bytes.Equal(kvA.Key, types.LastPositionIdKey):
			idA := sdk.BigEndianToUint64(kvA.Value)
			idB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", idA, idB)

		case bytes.Equal(kvA.Key[:1], types.PoolKeyPrefix):
			var pA, pB types.Pool
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.PoolStateKeyPrefix):
			var psA, psB types.PoolState
			cdc.MustUnmarshal(kvA.Value, &psA)
			cdc.MustUnmarshal(kvB.Value, &psB)
			return fmt.Sprintf("%v\n%v", psA, psB)

		case bytes.Equal(kvA.Key[:1], types.PositionKeyPrefix):
			var pA, pB types.Position
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		case bytes.Equal(kvA.Key[:1], types.TickInfoKeyPrefix):
			var tA, tB types.TickInfo
			cdc.MustUnmarshal(kvA.Value, &tA)
			cdc.MustUnmarshal(kvB.Value, &tB)
			return fmt.Sprintf("%v\n%v", tA, tB)

		case bytes.Equal(kvA.Key[:1], types.LastFarmingPlanIdKey):
			idA := sdk.BigEndianToUint64(kvA.Value)
			idB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", idA, idB)

		case bytes.Equal(kvA.Key[:1], types.FarmingPlanKeyPrefix):
			var pA, pB types.FarmingPlan
			cdc.MustUnmarshal(kvA.Value, &pA)
			cdc.MustUnmarshal(kvB.Value, &pB)
			return fmt.Sprintf("%v\n%v", pA, pB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		}
	}
}
