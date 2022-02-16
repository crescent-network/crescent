package claim

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/claim/keeper"
	"github.com/cosmosquad-labs/squad/x/claim/types"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Terminate airdrop if the airdrop end time has passed
	for _, airdrop := range k.GetAllAirdrops(ctx) {
		if ctx.BlockTime().After(airdrop.EndTime) {
			if err := k.TerminateAirdrop(ctx, airdrop); err != nil {
				panic(err)
			}
		}
	}
}
