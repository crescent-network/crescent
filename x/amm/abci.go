package amm

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	if err := k.TerminateEndedFarmingPlans(ctx); err != nil {
		panic(err)
	}
	if err := k.AllocateFarmingRewards(ctx); err != nil {
		panic(err)
	}
	//k.RunValidations(ctx)
}
