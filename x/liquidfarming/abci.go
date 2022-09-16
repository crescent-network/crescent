package liquidfarming

import (
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
)

// BeginBlocker compares all LiquidFarms stored in the store with all LiquidFarms registered in params.
// Execute an appropriate operation when either adding new LiquidFarm or removing one through governance proposal.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	liquidFarmSet := map[uint64]types.LiquidFarm{} // PoolId => LiquidFarm
	for _, liquidFarm := range k.GetAllLiquidFarms(ctx) {
		liquidFarmSet[liquidFarm.PoolId] = liquidFarm
	}

	// Compare all liquid farms stored in the store with all liquid farms registered in params
	// Store if new liquid farm is added and delete from the liquidFarmSet if it exists
	for _, liquidFarm := range k.GetParams(ctx).LiquidFarms {
		_, found := liquidFarmSet[liquidFarm.PoolId]
		if !found { // new LiquidFarm is added
			k.SetLiquidFarm(ctx, liquidFarm)
		} else {
			delete(liquidFarmSet, liquidFarm.PoolId)
		}
	}

	// Sort map keys for deterministic execution
	var poolIds []uint64
	for poolId := range liquidFarmSet {
		poolIds = append(poolIds, poolId)
	}
	sort.Slice(poolIds, func(i, j int) bool {
		return poolIds[i] < poolIds[j]
	})

	for _, poolId := range poolIds {
		k.HandleRemovedLiquidFarm(ctx, liquidFarmSet[poolId])
	}
}
