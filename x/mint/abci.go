package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"

	"github.com/crescent-network/crescent/v2/x/mint/keeper"
	"github.com/crescent-network/crescent/v2/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored params
	params := k.GetParams(ctx)

	lastBlockTime := k.GetLastBlockTime(ctx)
	// if not set LastBlockTime(e.g. fist block), skip minting inflation
	if lastBlockTime == nil {
		k.SetLastBlockTime(ctx, ctx.BlockTime())
		return
	}

	inflationSchedules := k.GetInflationSchedules(ctx)
	blockInflation := sdk.ZeroInt()
	var blockDurationForInflation time.Duration
	for _, schedule := range inflationSchedules {
		if utils.DateRangeIncludes(schedule.StartTime, schedule.EndTime, ctx.BlockTime()) {
			blockDurationForInflation = ctx.BlockTime().Sub(*lastBlockTime)
			if blockDurationForInflation > params.BlockTimeThreshold {
				blockDurationForInflation = params.BlockTimeThreshold
			}
			// blockInflation = InflationAmountThisPeriod * min(CurrentBlockTime-LastBlockTime,BlockTimeThreshold)/(InflationPeriodEndDate-InflationPeriodStartDate)
			blockInflation = schedule.Amount.MulRaw(blockDurationForInflation.Nanoseconds()).QuoRaw(schedule.EndTime.Sub(schedule.StartTime).Nanoseconds())
			break
		}
	}

	if blockInflation.IsPositive() {
		mintedCoin := sdk.NewCoin(params.MintDenom, blockInflation)
		mintedCoins := sdk.NewCoins(mintedCoin)
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the mint pool
		err = k.SendInflationToMintPool(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
				sdk.NewAttribute(types.AttributeKeyBlockDuration, blockDurationForInflation.String()),
			),
		)
	}
	k.SetLastBlockTime(ctx, ctx.BlockTime())
}
