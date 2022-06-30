package farming

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/farming/keeper"
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := k.Logger(ctx)

	k.ProcessQueuedCoins(ctx, ctx.BlockTime())

	if err := k.TerminateEndedPlans(ctx); err != nil {
		logger.Error("failed to terminate plan", "err", err.Error())
	}

	// CurrentEpochDays is initialized with the value of NextEpochDays in genesis, and
	// it is used here to prevent from affecting the epoch days for farming rewards allocation.
	// Suppose NextEpochDays is 7 days, and it is proposed to change the value to 1 day through governance proposal.
	// Although the proposal is passed, farming rewards allocation should continue to proceed with 7 days,
	// and then it gets updated.
	currentEpochDays := k.GetCurrentEpochDays(ctx)

	lastEpochTime, found := k.GetLastEpochTime(ctx)
	if !found {
		k.SetLastEpochTime(ctx, ctx.BlockTime())
	} else {
		y, m, d := lastEpochTime.AddDate(0, 0, int(currentEpochDays)).Date()
		y2, m2, d2 := ctx.BlockTime().Date()
		if !time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC).Before(time.Date(y, m, d, 0, 0, 0, 0, time.UTC)) {
			if err := k.AllocateRewards(ctx); err != nil {
				panic(err)
			}
			k.SetLastEpochTime(ctx, ctx.BlockTime())

			if params := k.GetParams(ctx); params.NextEpochDays != currentEpochDays {
				k.SetCurrentEpochDays(ctx, params.NextEpochDays)
			}
		}
	}

}
