package liquidfarming

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

// BeginBlocker compares all LiquidFarms stored in KVstore with all LiquidFarms registered in params.
// Execute an appropriate operation when either new LiquidFarm is added or existing LiquidFarm is removed
// by going through governance proposal.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	y, m, d := ctx.BlockTime().Date()

	endTime, found := k.GetNextRewardsAuctionEndTime(ctx)
	if !found {
		initialEndTime := time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC) // the next day 00:00 UTC
		k.SetNextRewardsAuctionEndTime(ctx, initialEndTime)
	} else {
		currentTime := ctx.BlockTime()
		if !currentTime.Before(endTime) {
			duration := k.GetRewardsAuctionDuration(ctx)
			nextEndTime := endTime.Add(duration)

			// Handle a case when a chain is halted for a long time
			if !currentTime.Before(nextEndTime) {
				nextEndTime = time.Date(y, m, d+1, 0, 0, 0, 0, time.UTC) // the next day 00:00 UTC
			}

			k.IterateAllLiquidFarms(ctx, func(liquidFarm types.LiquidFarm) (stop bool) {
				if liquidFarm.LastRewardsAuctionId != 0 {
					auction, found := k.GetRewardsAuction(ctx, liquidFarm.Id, liquidFarm.LastRewardsAuctionId)
					if !found { // sanity check
						panic("rewards auction not found")
					}
					if auction.WinningBid == nil {
						if err := k.SkipRewardsAuction(ctx, liquidFarm, auction); err != nil {
							panic(err)
						}
					} else {
						if err := k.FinishRewardsAuction(ctx, liquidFarm, auction); err != nil {
							panic(err)
						}
					}
				}
				k.StartNewRewardsAuction(ctx, liquidFarm, nextEndTime)
				return false
			})
			k.SetNextRewardsAuctionEndTime(ctx, nextEndTime)
		}
	}
}
