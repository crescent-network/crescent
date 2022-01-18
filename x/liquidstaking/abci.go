package liquidstaking

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

// BeginBlocker collects liquidStakings for the current block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// TODO: Unimplemented beginblock logic
	// TODO: withdraw rewards and re-staking for active validator
	//totalRewards := k.WithdrawLiquidRewards(ctx, types.LiquidStakingProxyAcc)
	// TODO: re-staking only the rewards or with balance
	// TODO: re-staking with rebalancing?
	//k.GetActiveLiquidValidators()
	//k.stakingKeeper.
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
