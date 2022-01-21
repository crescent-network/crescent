package liquidstaking

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

// BeginBlocker collects liquidStakings for the current block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	params := k.GetParams(ctx)
	liquidValidators := k.GetAllLiquidValidators(ctx)
	liquidValsMap := liquidValidators.Map()
	valsMap := k.GetValidatorsMap(ctx)
	whitelistedValMap := make(map[string]types.WhitelistedValidator)
	// TODO: check delisting -> delisted for mature redelegation queue

	////var whiteListToActiveQueue []types.WhitelistedValidator
	//var delistingToActiveQueue []types.LiquidValidator
	//var activeToDelistingQueue []types.LiquidValidator
	//var delistedToActiveQueue []types.LiquidValidator

	// active -> delisting
	//activeLiquidValidators := k.GetActiveLiquidValidators(ctx)

	liquidValidators.ActiveToDelisting(valsMap, whitelistedValMap, params.CommissionRate)

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if lv, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			// TODO: added on whitelist -> active set active when rebalancing succeeded
			// TODO: or it could be added without rebalancing
			// TODO: consider add params.MaxActiveLiquidValidator
			// whitelist -> active
			lv = &types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
				Status:          types.ValidatorStatusActive,
				LiquidTokens:    sdk.ZeroInt(),
				Weight:          wv.Weight,
			}
			k.SetLiquidValidator(ctx, *lv)
			liquidValsMap[lv.OperatorAddress] = lv
		} else {
			// TODO: delisted -> active
			if lv.Status == types.ValidatorStatusDelisted {
				// TODO: if not jailed, tombstoned, unbonded, rebalancing succeed
				//lv.UpdateStatus(types.ValidatorStatusActive)
			}
		}
		whitelistedValMap[wv.ValidatorAddress] = wv
	}

	// TODO: rebalancing based updated liquid validators status
	// TODO: Set status

	// TODO: rebalancing first or re-staking first
	activeVals := k.GetActiveLiquidValidators(ctx)
	totalLiquidTokens := activeVals.TotalLiquidTokens()
	if totalLiquidTokens.IsPositive() {
		// Withdraw rewards of LiquidStakingProxyAcc and re-staking
		totalRewards, _ := k.CheckRewardsAndLiquidPower(ctx, types.LiquidStakingProxyAcc)
		// checking over types.RewardTrigger and execute GetRewards
		// TODO: test triggering
		balance := k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
		rewardsThreshold := types.RewardTrigger.MulInt(totalLiquidTokens).TruncateInt()
		if balance.Add(totalRewards.TruncateInt()).GTE(rewardsThreshold) {
			// re-staking with balance, due to auto-withdraw on add staking by f1
			_ = k.WithdrawLiquidRewards(ctx, types.LiquidStakingProxyAcc)
			balance = k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
			_, err := k.LiquidDelegate(ctx, types.LiquidStakingProxyAcc, k.GetActiveLiquidValidators(ctx), balance)
			if err != nil {
				panic(err)
			}
		}
	}
}
