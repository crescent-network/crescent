package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

// TODO: Deprecated
//func (k Keeper) TryRedelegations(ctx sdk.Context, redelegations []types.Redelegation) (completionTime time.Time, err error) {
//	cachedCtx, writeCache := ctx.CacheContext()
//	for _, re := range redelegations {
//		// TODO: ValidateUnbondAmount check
//		shares, err := k.stakingKeeper.ValidateUnbondAmount(
//			cachedCtx, re.Delegator, re.SrcValidator, re.Amount,
//		)
//		if err != nil {
//			return time.Time{}, err
//		}
//		completionTime, err = k.stakingKeeper.BeginRedelegation(cachedCtx, re.Delegator, re.SrcValidator, re.DstValidator, shares)
//		if err != nil {
//			return time.Time{}, err
//		}
//	}
//	writeCache()
//	// TODO: bug on liquidValsMap pointer set, need to optimize UpdateLiquidTokens
//	//k.UpdateLiquidTokens(ctx, proxyAcc)
//	return completionTime, nil
//}

func (k Keeper) TryRedelegation(ctx sdk.Context, re types.Redelegation) (completionTime time.Time, err error) {
	cachedCtx, writeCache := ctx.CacheContext()
	// TODO: ValidateUnbondAmount check
	shares, err := k.stakingKeeper.ValidateUnbondAmount(
		cachedCtx, re.Delegator, re.SrcValidator, re.Amount,
	)
	if err != nil {
		return time.Time{}, err
	}
	completionTime, err = k.stakingKeeper.BeginRedelegation(cachedCtx, re.Delegator, re.SrcValidator, re.DstValidator, shares)
	if err != nil {
		return time.Time{}, err
	}
	writeCache()
	// TODO: bug on liquidValsMap pointer set, need to optimize UpdateLiquidTokens
	//k.UpdateLiquidTokens(ctx, proxyAcc)
	return completionTime, nil
}

//// DivideByCurrentWeight divide the input value by the ratio of the weight of the liquid validator's liquid token and return it with crumb
//// TODO: consider deprecate DivideByCurrentWeight and use DivideByCurrentWeightDec
//func (k Keeper) DivideByCurrentWeight(ctx sdk.Context, vs types.LiquidValidators, input sdk.Int) (outputs []sdk.Int, crumb sdk.Int) {
//	totalLDelShares := vs.TotalDelShares(ctx, k.stakingKeeper)
//	totalShares := sdk.ZeroInt()
//	sharePerWeight := input.ToDec().QuoTruncate(totalLDelShares.ToDec())
//	for _, val := range vs {
//		weightedShare := sharePerWeight.MulInt(val.GetDelShares(ctx, k.stakingKeeper)).TruncateInt()
//		totalShares = totalShares.Add(weightedShare)
//		outputs = append(outputs, weightedShare)
//	}
//	return outputs, input.Sub(totalShares)
//}

// DivideByCurrentWeightDec divide the input value by the ratio of the weight of the liquid validator's liquid token and return it with crumb as decimal
func (k Keeper) DivideByCurrentWeightDec(ctx sdk.Context, vs types.LiquidValidators, input sdk.Dec) (outputs []sdk.Dec, crumb sdk.Dec) {
	totalLDelShares := vs.TotalDelShares(ctx, k.stakingKeeper)
	totalShares := sdk.ZeroDec()
	sharePerWeight := input.QuoTruncate(totalLDelShares.ToDec())
	for _, val := range vs {
		weightedShare := sharePerWeight.MulInt(val.GetDelShares(ctx, k.stakingKeeper))
		totalShares = totalShares.Add(weightedShare)
		outputs = append(outputs, weightedShare)
	}
	return outputs, input.Sub(totalShares)
}

//AddStakingTargetMap is make add staking target map for one-way rebalancing, it can be called recursively.
// TODO: not using on this version for simplify
func (k Keeper) AddStakingTargetMap(ctx sdk.Context, activeVals types.LiquidValidators, addStakingAmt sdk.Int) map[string]sdk.Int {
	targetMap := make(map[string]sdk.Int)
	if addStakingAmt.IsNil() || !addStakingAmt.IsPositive() || activeVals.Len() == 0 {
		return targetMap
	}
	totalDelShares := activeVals.TotalDelShares(ctx, k.stakingKeeper)
	totalWeight := activeVals.TotalWeight()
	ToBeTotalDelShares := totalDelShares.Add(addStakingAmt)
	existOverWeightedVal := false

	sharePerWeight := ToBeTotalDelShares.Quo(totalWeight)
	crumb := ToBeTotalDelShares.Sub(sharePerWeight.Mul(totalWeight))

	i := 0
	for _, val := range activeVals {
		weightedShare := val.Weight.Mul(sharePerWeight)
		if val.GetDelShares(ctx, k.stakingKeeper).GT(weightedShare) {
			existOverWeightedVal = true
		} else {
			activeVals[i] = val
			i++
			targetMap[val.OperatorAddress] = weightedShare.Sub(val.GetDelShares(ctx, k.stakingKeeper))
		}
	}
	// remove overWeightedVals for recursive call
	activeVals = activeVals[:i]

	if !existOverWeightedVal {
		if v, ok := targetMap[activeVals[0].OperatorAddress]; ok {
			targetMap[activeVals[0].OperatorAddress] = v.Add(crumb)
		} else {
			targetMap[activeVals[0].OperatorAddress] = crumb
		}
		return targetMap
	} else {
		fmt.Println("[AddStakingTargetMap] recursive call for", activeVals, addStakingAmt, totalDelShares, ToBeTotalDelShares, totalWeight, sharePerWeight, crumb)
		return k.AddStakingTargetMap(ctx, activeVals, addStakingAmt)
	}
}

// activeVals containing ValidatorStatusActive which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
func (k Keeper) Rebalancing(ctx sdk.Context, proxyAcc sdk.AccAddress, liquidVals types.LiquidValidators, rebalancingTrigger sdk.Dec) (redelegations []types.Redelegation) {
	totalLDelShares := liquidVals.TotalDelShares(ctx, k.stakingKeeper)
	totalWeight := liquidVals.TotalWeight()
	threshold := rebalancingTrigger.MulInt(totalLDelShares)

	var targetWeight sdk.Int
	targetMap := map[string]sdk.Int{}
	for _, val := range liquidVals {
		if val.Status == types.ValidatorStatusActive {
			targetWeight = val.Weight
		} else if val.Status == types.ValidatorStatusDelisting ||
			val.Status == types.ValidatorStatusDelisted {
			targetWeight = sdk.ZeroInt()
		} else {
			//val.Status == types.ValidatorStatusUnspecified
			targetWeight = sdk.ZeroInt()
		}
		targetMap[val.OperatorAddress] = totalLDelShares.Mul(targetWeight).Quo(totalWeight)
	}

	for i := 0; i < len(liquidVals); i++ {
		maxVal, minVal, amountNeeded := liquidVals.MinMaxGap(ctx, k.stakingKeeper, targetMap)
		if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold.TruncateInt())) {
			break
		}
		// TODO: refactor using map
		addedVal := 0
		subtractedVal := 0
		for idx := range liquidVals {
			if liquidVals[idx].OperatorAddress == maxVal.OperatorAddress {
				addedVal = idx
			}
			if liquidVals[idx].OperatorAddress == minVal.OperatorAddress {
				subtractedVal = idx
			}
		}
		redelegation := types.Redelegation{
			Delegator:    proxyAcc,
			SrcValidator: liquidVals[subtractedVal].GetOperator(),
			DstValidator: liquidVals[addedVal].GetOperator(),
			Amount:       amountNeeded,
		}
		// TODO: deprecate return redelegations
		redelegations = append(redelegations, redelegation)
		_, err := k.TryRedelegation(ctx, redelegation)
		if err != nil {
			fmt.Println("[TryRedelegations] failed due to redelegation restriction", redelegations)
		}
	}
	for _, r := range redelegations {
		fmt.Println("[rebalancing]", r.Amount.String(), r.SrcValidator.String(), "->", r.DstValidator.String())
	}
	return redelegations
}

func (k Keeper) EndBlocker(ctx sdk.Context) {
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
	liquidValidators.ActiveToDelisting(valsMap, whitelistedValMap)

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if lv, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			// whitelist -> active
			// added on whitelist -> active set
			// TODO: k.SetLiquidValidator(ctx, *lv) set on TryRedelegations if succeed or pre-active without rebalancing
			lv = &types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
				Status:          types.ValidatorStatusActive,
				//LiquidTokens:    sdk.ZeroInt(),
				Weight: wv.TargetWeight,
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
	k.Rebalancing(ctx, types.LiquidStakingProxyAcc, liquidValidators, types.RebalancingTrigger)
	//_, err := k.TryRedelegations(ctx, redelegations)
	//if err != nil {
	//	fmt.Println("[TryRedelegations] failed due to redelegation restriction", redelegations)
	//}

	// withdraw rewards and re-staing when over threshold
	activeVals := k.GetActiveLiquidValidators(ctx)
	totalLDelShares := activeVals.TotalDelShares(ctx, k.stakingKeeper)
	if totalLDelShares.IsPositive() {
		// Withdraw rewards of LiquidStakingProxyAcc and re-staking
		totalRewards, _ := k.CheckRewardsAndDelShares(ctx, types.LiquidStakingProxyAcc)
		// checking over types.RewardTrigger and execute GetRewards
		// TODO: test triggering
		balance := k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
		rewardsThreshold := types.RewardTrigger.MulInt(totalLDelShares).TruncateInt()
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
