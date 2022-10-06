package liquidfarming

import (
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
)

// BeginBlocker compares all LiquidFarms stored in KVstore with all LiquidFarms registered in params.
// Execute an appropriate operation when either new LiquidFarm is added or existing LiquidFarm is removed by
// going through governance proposal.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	params := k.GetParams(ctx)
	liquidFarmsInStore := k.GetAllLiquidFarms(ctx)
	liquidFarmsInParams := params.LiquidFarms

	liquidFarmByPoolId := map[uint64]types.LiquidFarm{} // PoolId => LiquidFarm
	for _, liquidFarm := range liquidFarmsInStore {
		liquidFarmByPoolId[liquidFarm.PoolId] = liquidFarm
	}

	// Compare all liquid farms stored in KVStore with the ones registered in params
	// If new liquid farm is added through governance proposal, store it in KVStore.
	// Otherwise, delete from the liquidFarmByPoolId
	for _, liquidFarm := range liquidFarmsInParams {
		if _, found := liquidFarmByPoolId[liquidFarm.PoolId]; !found {
			k.SetLiquidFarm(ctx, liquidFarm)
			continue
		}
		delete(liquidFarmByPoolId, liquidFarm.PoolId)
	}

	// Sort map keys for deterministic execution
	var poolIds []uint64
	for poolId := range liquidFarmByPoolId {
		poolIds = append(poolIds, poolId)
	}
	sort.Slice(poolIds, func(i, j int) bool {
		return poolIds[i] < poolIds[j]
	})

	// Handle a case when LiquidFarm is removed in params by governance proposal
	for _, poolId := range poolIds {
		k.HandleRemovedLiquidFarm(ctx, liquidFarmByPoolId[poolId])
	}

	endTime := k.GetLastRewardsAuctionEndTime(ctx)
	if endTime.IsZero() {
		k.SetLastRewardsAuctionEndTime(ctx, ctx.BlockTime().Add(params.RewardsAuctionDuration))
		return
	}

	// Iterate all LiquidFarms in KVStore to create rewards auction if it is not found.
	// If there is an ongoing rewards auction, finish it.
	if !ctx.BlockTime().Before(endTime) { // AuctionEndTime <= Current BlockTime
		for _, liquidFarm := range liquidFarmsInStore {
			auctionId := k.GetLastRewardsAuctionId(ctx, liquidFarm.PoolId)
			auction, found := k.GetRewardsAuction(ctx, auctionId, liquidFarm.PoolId)
			if found {
				// Note that order matters in this logic.
				// The module needs to finish the auction and create new one
				if err := k.FinishRewardsAuction(ctx, auction, liquidFarm.FeeRate); err != nil {
					panic(err)
				}
			}
			k.CreateRewardsAuction(ctx, liquidFarm.PoolId, params.RewardsAuctionDuration)
		}
		k.SetLastRewardsAuctionEndTime(ctx, ctx.BlockTime().Add(params.RewardsAuctionDuration))
	}
}
