package farming

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	for _, plan := range k.GetAllPlans(ctx) {
		if !plan.GetTerminated() && ctx.BlockTime().After(plan.GetEndTime()) {
			if err := k.TerminatePlan(ctx, plan); err != nil {
				panic(err)
			}
		}
	}

	params := k.GetParams(ctx)

	lastEpochTime, found := k.GetLastEpochTime(ctx)
	if !found {
		k.SetLastEpochTime(ctx, ctx.BlockTime())
	} else {
		y, m, d := lastEpochTime.AddDate(0, 0, int(params.EpochDays)).Date()
		y2, m2, d2 := ctx.BlockTime().Date()
		if !time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC).Before(time.Date(y, m, d, 0, 0, 0, 0, time.UTC)) {
			if err := k.AdvanceEpoch(ctx); err != nil {
				panic(err)
			}
		}
	}
}
