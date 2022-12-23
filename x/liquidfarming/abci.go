package liquidfarming

import (
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v4/x/liquidfarming/types"
)

// BeginBlocker compares all LiquidFarms stored in KVstore with all LiquidFarms registered in params.
// Execute an appropriate operation when either new LiquidFarm is added or existing LiquidFarm is removed
// by going through governance proposal.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	liquidFarmsInStore := k.GetLiquidFarmsInStore(ctx)
	liquidFarmsInParams := k.GetLiquidFarmsInParams(ctx)

	liquidFarmByPoolId := map[uint64]types.LiquidFarm{} // PoolId => LiquidFarm
	for _, liquidFarm := range liquidFarmsInStore {
		liquidFarmByPoolId[liquidFarm.PoolId] = liquidFarm
	}

	// Iterate through all liquid farms in params.
	// Store new one or delete the existing one from liquidFarmByPoolId
	// when it is added or removed in params by governance proposal.
	for _, liquidFarm := range liquidFarmsInParams {
		if _, found := liquidFarmByPoolId[liquidFarm.PoolId]; !found {
			k.SetLiquidFarm(ctx, liquidFarm)
		} else {
			delete(liquidFarmByPoolId, liquidFarm.PoolId)
		}
	}

	// Sort map keys for deterministic execution
	var poolIds []uint64
	for poolId := range liquidFarmByPoolId {
		poolIds = append(poolIds, poolId)
	}
	sort.Slice(poolIds, func(i, j int) bool {
		return poolIds[i] < poolIds[j]
	})

	for _, poolId := range poolIds {
		k.HandleRemovedLiquidFarm(ctx, liquidFarmByPoolId[poolId])
	}

	y, m, d := ctx.BlockTime().Date()

	endTime, found := k.GetLastRewardsAuctionEndTime(ctx)
	if !found {
		initialEndTime := time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC) // the next day 00:00 UTC
		k.SetLastRewardsAuctionEndTime(ctx, initialEndTime)
	} else {
		currentTime := ctx.BlockTime()
		if !currentTime.Before(endTime) {
			duration := k.GetRewardsAuctionDuration(ctx)
			nextEndTime := endTime.Add(duration)

			// Handle a case when a chain is halted for a long time
			if !currentTime.Before(nextEndTime) {
				nextEndTime = time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC) // the next day 00:00 UTC
			}

			for _, l := range liquidFarmsInStore {
				auction, found := k.GetLastRewardsAuction(ctx, l.PoolId)
				if found {
					if err := k.FinishRewardsAuction(ctx, auction, l.FeeRate); err != nil {
						panic(err)
					}
				}
				k.CreateRewardsAuction(ctx, l.PoolId, nextEndTime)
			}
			k.SetLastRewardsAuctionEndTime(ctx, nextEndTime)
		}
	}
}
