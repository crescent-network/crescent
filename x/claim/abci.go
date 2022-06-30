package claim

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/claim/keeper"
	"github.com/crescent-network/crescent/v2/x/claim/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Terminate airdrop if the airdrop end time has passed
	for _, airdrop := range k.GetAllAirdrops(ctx) {
		if !ctx.BlockTime().Before(airdrop.EndTime) { // BlockTime >= EndTime
			if err := k.TerminateAirdrop(ctx, airdrop); err != nil {
				panic(err)
			}
		}
	}
}
