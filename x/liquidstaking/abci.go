package liquidstaking

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/liquidstaking/keeper"
	"github.com/tendermint/farming/x/liquidstaking/types"
)

// BeginBlocker collects liquidStakings for the current block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// TODO: Unimplemented beginblock logic
	//if err != nil {
	//	panic(err)
	//}
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	params := k.GetParams(ctx)
	liquidValsMap := k.GetAllLiquidValidatorsMap(ctx)

	whitelistedValMap := make(map[string]types.WhitelistedValidator)
	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if _, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			k.SetLiquidValidator(ctx, types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
				Status:          types.ValidatorStatusActive,
				LiquidTokens:    sdk.ZeroInt(),
				Weight:          wv.Weight,
			})
		}
		whitelistedValMap[wv.ValidatorAddress] = wv
	}
	// TODO: rebalancing and delisting logic
	for _, lv := range k.GetAllLiquidValidators(ctx) {
		if wv, ok := whitelistedValMap[lv.OperatorAddress]; !ok && lv.Status == types.ValidatorStatusActive {
			lv.Status = types.ValidatorStatusDelisting
			k.SetLiquidValidator(ctx, lv)
			fmt.Println("[delisting liquid validator]", lv, wv)
		}
	}
}
