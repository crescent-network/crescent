package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	liquidstakingtypes "github.com/crescent-network/crescent/x/liquidstaking/types"
	"github.com/crescent-network/crescent/x/mint/keeper"
	"github.com/crescent-network/crescent/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored params
	params := k.GetParams(ctx)

	// temporary hardcoded
	inflations := []types.InflationPeriod{
		{
			StartTime: liquidstakingtypes.MustParseRFC3339("2022-01-01T00:00:00Z"),
			EndTime:   liquidstakingtypes.MustParseRFC3339("2022-12-31T23:59:59Z"),
			Amount:    sdk.NewInt(1000000000000),
		},
	}
	// TODO: Get/Set LastBlockTime
	tmpLastBlockTime := ctx.BlockTime()
	blockInflation := sdk.ZeroInt()
	for _, period := range inflations {
		if period.EndTime.After(ctx.BlockTime()) && period.StartTime.Before(ctx.BlockTime()) {
			blockTime := params.BlockTimeThreshold
			blockTimeDiff := ctx.BlockTime().Sub(tmpLastBlockTime)
			if params.BlockTimeThreshold > blockTimeDiff {
				blockTime = blockTimeDiff
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
				//sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	}

}
