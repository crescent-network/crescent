package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/mint/keeper"
	"github.com/crescent-network/crescent/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, blockTime time.Duration) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored params
	params := k.GetParams(ctx)

	schedules := k.GetInflationSchedules()
	blockInflation := sdk.ZeroInt()
	for _, period := range schedules {
		if !period.EndTime.Before(ctx.BlockTime()) && !period.StartTime.After(ctx.BlockTime()) {
			// TODO: need to Get/Set LastBlockTime
			//blockTime := ctx.BlockTime().Sub(k.GetLastBlockTime(ctx))
			if blockTime > params.BlockTimeThreshold {
				blockTime = params.BlockTimeThreshold
			}
			blockInflation = period.Amount.MulRaw(blockTime.Nanoseconds()).QuoRaw(period.EndTime.Sub(period.StartTime).Nanoseconds())
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

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
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
			),
		)
	}
}
