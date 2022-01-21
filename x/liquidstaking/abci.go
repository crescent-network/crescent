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
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	params := k.GetParams(ctx)
	liquidValidators := k.GetAllLiquidValidators(ctx)
	// TODO: pointer map looks uncertainty, need to fix
	liquidValsMap := liquidValidators.Map()
	valsMap := k.GetValidatorsMap(ctx)
	whitelistedValMap := make(map[string]types.WhitelistedValidator)
	for _, wv := range params.WhitelistedValidators {
		whitelistedValMap[wv.ValidatorAddress] = wv
	}

	// delisting to delisted
	liquidValidators.DelistingToDelisted(valsMap)

	// active -> delisting
	liquidValidators.ActiveToDelisting(valsMap, whitelistedValMap, params.CommissionRate)

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if lv, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			// whitelist -> active
			// added on whitelist -> active set
			// TODO: consider add params.MaxActiveLiquidValidator
			// TODO: k.SetLiquidValidator(ctx, *lv) set on TryRedelegations if succeed or pre-active without rebalancing
			lv = &types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
				Status:          types.ValidatorStatusActive,
				LiquidTokens:    sdk.ZeroInt(),
				Weight:          wv.Weight,
			}
			k.SetLiquidValidator(ctx, *lv)
			liquidValsMap[lv.OperatorAddress] = lv
			liquidValidators = append(liquidValidators, *lv)
		} else {
			// TODO: weight change update

			// delisted -> active
			if lv.Status == types.ValidatorStatusDelisted {
				// TODO: k.SetLiquidValidator(ctx, *lv) set on TryRedelegations if succeed
				// TODO: check active conditions, not jailed, tombstoned, unbonded
				lv.UpdateStatus(types.ValidatorStatusActive)
			}
		}
		whitelistedValMap[wv.ValidatorAddress] = wv
	}

	// rebalancing based updated liquid validators status with threshold, try by cachedCtx
	redelegations := types.Rebalancing(types.LiquidStakingProxyAcc, liquidValidators, types.RebalancingTrigger)
	_, err := k.TryRedelegations(ctx, types.LiquidStakingProxyAcc, redelegations)
	if err != nil {
		fmt.Println("[TryRedelegations] failed due to redelegation restriction", redelegations)
	}

	// withdraw rewards and re-staing when over threshold
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
