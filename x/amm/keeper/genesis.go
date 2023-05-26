package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if genState.LastPoolId > 0 {
		k.SetLastPoolId(ctx, genState.LastPoolId)
	}
	if genState.LastPositionId > 0 {
		k.SetLastPoolId(ctx, genState.LastPositionId)
	}
	for _, poolRecord := range genState.PoolRecords {
		k.SetPool(ctx, poolRecord.Pool)
		k.SetPoolByReserveAddressIndex(ctx, poolRecord.Pool)
		k.SetPoolByMarketIndex(ctx, poolRecord.Pool)
		k.SetPoolState(ctx, poolRecord.Pool.Id, poolRecord.State)
	}
	for _, position := range genState.Positions {
		k.SetPosition(ctx, position)
		k.SetPositionByParamsIndex(ctx, position)
		k.SetPositionsByPoolIndex(ctx, position)
	}
	for _, tickInfoRecord := range genState.TickInfoRecords {
		k.SetTickInfo(ctx, tickInfoRecord.PoolId, tickInfoRecord.Tick, tickInfoRecord.TickInfo)
	}
	if genState.LastFarmingPlanId > 0 {
		k.SetLastFarmingPlanId(ctx, genState.LastFarmingPlanId)
	}
	if genState.NumPrivateFarmingPlans > 0 {
		k.SetNumPrivateFarmingPlans(ctx, genState.NumPrivateFarmingPlans)
	}
	for _, plan := range genState.FarmingPlans {
		k.SetFarmingPlan(ctx, plan)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	poolRecords := []types.PoolRecord{}
	k.IterateAllPools(ctx, func(pool types.Pool) (stop bool) {
		poolRecords = append(poolRecords, types.PoolRecord{
			Pool:  pool,
			State: k.MustGetPoolState(ctx, pool.Id),
		})
		return false
	})
	positions := []types.Position{}
	k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
		positions = append(positions, position)
		return false
	})
	tickInfoRecords := []types.TickInfoRecord{}
	k.IterateAllTickInfos(ctx, func(poolId uint64, tick int32, tickInfo types.TickInfo) (stop bool) {
		tickInfoRecords = append(tickInfoRecords, types.TickInfoRecord{
			PoolId:   poolId,
			Tick:     tick,
			TickInfo: tickInfo,
		})
		return false
	})
	farmingPlans := []types.FarmingPlan{}
	k.IterateAllFarmingPlans(ctx, func(plan types.FarmingPlan) (stop bool) {
		farmingPlans = append(farmingPlans, plan)
		return false
	})
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.GetLastPoolId(ctx),
		k.GetLastPositionId(ctx),
		poolRecords,
		positions,
		tickInfoRecords,
		k.GetLastFarmingPlanId(ctx),
		k.GetNumPrivateFarmingPlans(ctx),
		farmingPlans)
}
