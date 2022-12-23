package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/lpfarm/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	if genState.LastBlockTime != nil {
		k.SetLastBlockTime(ctx, *genState.LastBlockTime)
	}
	if genState.LastPlanId > 0 {
		k.SetLastPlanId(ctx, genState.LastPlanId)
	}
	k.SetNumPrivatePlans(ctx, genState.NumPrivatePlans)
	for _, plan := range genState.Plans {
		k.SetPlan(ctx, plan)
	}
	for _, farm := range genState.Farms {
		k.SetFarm(ctx, farm.Denom, farm.Farm)
	}
	for _, position := range genState.Positions {
		k.SetPosition(ctx, position)
	}
	for _, hist := range genState.HistoricalRewards {
		k.SetHistoricalRewards(ctx, hist.Denom, hist.Period, hist.HistoricalRewards)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var lastBlockTimePtr *time.Time
	lastBlockTime, found := k.GetLastBlockTime(ctx)
	if found {
		lastBlockTimePtr = &lastBlockTime
	}

	lastPlanId, _ := k.GetLastPlanId(ctx)

	plans := []types.Plan{}
	k.IterateAllPlans(ctx, func(plan types.Plan) (stop bool) {
		plans = append(plans, plan)
		return false
	})

	farms := []types.FarmRecord{}
	k.IterateAllFarms(ctx, func(denom string, farm types.Farm) (stop bool) {
		farms = append(farms, types.FarmRecord{
			Denom: denom,
			Farm:  farm,
		})
		return false
	})

	positions := []types.Position{}
	k.IterateAllPositions(ctx, func(position types.Position) (stop bool) {
		positions = append(positions, position)
		return false
	})

	hists := []types.HistoricalRewardsRecord{}
	k.IterateAllHistoricalRewards(
		ctx, func(denom string, period uint64, hist types.HistoricalRewards) (stop bool) {
			hists = append(hists, types.HistoricalRewardsRecord{
				Denom:             denom,
				Period:            period,
				HistoricalRewards: hist,
			})
			return false
		})

	return types.NewGenesisState(
		k.GetParams(ctx), lastBlockTimePtr, lastPlanId, k.GetNumPrivatePlans(ctx),
		plans, farms, positions, hists)
}
