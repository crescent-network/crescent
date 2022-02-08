package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

func (k Keeper) TryRedelegation(ctx sdk.Context, re types.Redelegation) (completionTime time.Time, err error) {
	cachedCtx, writeCache := ctx.CacheContext()
	// TODO: check small left delShares for zero targetWeight
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
	fmt.Println("[rebalancing]", re.Amount.String(), re.SrcValidator.String(), "->", re.DstValidator.String())
	return completionTime, nil
}

// DivideByCurrentWeight divide the input value by the ratio of the weight of the liquid validator's liquid token and return it with crumb
// which is may occur while dividing according to the weight of liquid validators by decimal error.
func (k Keeper) DivideByCurrentWeight(ctx sdk.Context, avs types.ActiveLiquidValidators, input sdk.Dec) (outputs []sdk.Dec, crumb sdk.Dec) {
	totalLiquidTokens := avs.TotalLiquidTokens(ctx, k.stakingKeeper)
	totalOutput := sdk.ZeroDec()
	unitInput := input.QuoTruncate(totalLiquidTokens)
	for _, val := range avs {
		output := unitInput.MulTruncate(val.GetLiquidTokens(ctx, k.stakingKeeper))
		totalOutput = totalOutput.Add(output)
		outputs = append(outputs, output)
	}
	return outputs, input.Sub(totalOutput)
}

// Rebalancing argument liquidVals containing ValidatorStatusActive which is containing just added on whitelist(liquidToken 0) and ValidatorStatusInActive to delist
func (k Keeper) Rebalancing(ctx sdk.Context, proxyAcc sdk.AccAddress, liquidVals types.LiquidValidators, whitelistedValMap types.WhitelistedValMap, rebalancingTrigger sdk.Dec) (redelegations []types.Redelegation) {
	totalLiquidTokens := liquidVals.TotalLiquidTokens(ctx, k.stakingKeeper)
	if !totalLiquidTokens.IsPositive() {
		return []types.Redelegation{}
	}
	weightMap, totalWeight := k.GetWeightMap(ctx, liquidVals, whitelistedValMap)
	if !totalWeight.IsPositive() {
		// TODO: no active liquid validators, make policy for this case
		return []types.Redelegation{}
	}
	threshold := rebalancingTrigger.Mul(totalLiquidTokens).TruncateInt()
	totalLiquidTokensInt := totalLiquidTokens.TruncateInt()

	targetMap := map[string]sdk.Int{}
	for _, val := range liquidVals {
		// TODO: check decimal loss occurred
		targetMap[val.OperatorAddress] = totalLiquidTokensInt.Mul(weightMap[val.OperatorAddress]).Quo(totalWeight)
	}
	squadtypes.PP(targetMap)
	squadtypes.PP(weightMap)

	lenLiquidVals := liquidVals.Len()
	for i := 0; i < lenLiquidVals; i++ {
		// TODO: check small left delShares for zero targetWeight
		minVal, maxVal, amountNeeded := liquidVals.MinMaxGap(ctx, k.stakingKeeper, targetMap, threshold)
		// TODO: consider threshold policy apply every redelegatoin or maxGap
		if amountNeeded.IsZero() || amountNeeded.LT(threshold) {
			//if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold)) {
			break
		}
		redelegation := types.Redelegation{
			Delegator:    proxyAcc,
			SrcValidator: maxVal.GetOperator(),
			DstValidator: minVal.GetOperator(),
			Amount:       amountNeeded,
		}
		redelegations = append(redelegations, redelegation)
		_, err := k.TryRedelegation(ctx, redelegation)
		if err != nil {
			fmt.Println("[TryRedelegations] failed due to redelegation restriction", redelegations)
		}
	}
	return redelegations
}

// WithdrawRewardsAndReStaking withdraw rewards and re-staking when over threshold
func (k Keeper) WithdrawRewardsAndReStaking(ctx sdk.Context, whitelistedValMap types.WhitelistedValMap) {
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	totalLiquidTokens := activeVals.TotalLiquidTokens(ctx, k.stakingKeeper)
	if totalLiquidTokens.IsPositive() {
		// Withdraw rewards of LiquidStakingProxyAcc and re-staking
		totalRewards, _, _ := k.CheckTotalRewards(ctx, types.LiquidStakingProxyAcc)
		// checking over types.RewardTrigger and execute GetRewards
		balance := k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
		rewardsThreshold := types.RewardTrigger.Mul(totalLiquidTokens).TruncateInt()
		if balance.Add(totalRewards.TruncateInt()).GTE(rewardsThreshold) {
			// re-staking with balance, due to auto-withdraw on add staking by f1
			k.WithdrawLiquidRewards(ctx, types.LiquidStakingProxyAcc)
			balance = k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
			_, err := k.LiquidDelegate(ctx, types.LiquidStakingProxyAcc, activeVals, balance, whitelistedValMap)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (k Keeper) UpdateLiquidValidatorSet(ctx sdk.Context) {
	params := k.GetParams(ctx)
	liquidValidators := k.GetAllLiquidValidators(ctx)
	liquidValsMap := liquidValidators.Map()
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)

	// TODO: add tombstone handling, remove if power 0 when tomestoned

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if _, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			lv := types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
			}
			if k.ActiveCondition(ctx, lv, true) {
				k.SetLiquidValidator(ctx, lv)
				liquidValidators = append(liquidValidators, lv)
			}
		}
	}

	// rebalancing based updated liquid validators status with threshold, try by cachedCtx
	k.Rebalancing(ctx, types.LiquidStakingProxyAcc, liquidValidators, whitelistedValMap, types.RebalancingTrigger)

	// remove inactive with zero liquidToken liquidvalidator
	for _, lv := range liquidValidators {
		// TODO: GetLiquidTokens or GetDelShares
		if !k.ActiveCondition(ctx, lv, whitelistedValMap.IsListed(lv.OperatorAddress)) && !lv.GetDelShares(ctx, k.stakingKeeper).IsPositive() {
			k.RemoveLiquidValidator(ctx, lv)
		}
	}

	// withdraw rewards and re-staking when over threshold
	k.WithdrawRewardsAndReStaking(ctx, whitelistedValMap)
}

//// Deprecated: AddStakingTargetMap is make add staking target map for one-way rebalancing, it can be called recursively, not using on this version for simplify.
//func (k Keeper) AddStakingTargetMap(ctx sdk.Context, activeVals types.ActiveLiquidValidators, addStakingAmt sdk.Int) map[string]sdk.Int {
//	targetMap := make(map[string]sdk.Int)
//	if addStakingAmt.IsNil() || !addStakingAmt.IsPositive() || activeVals.Len() == 0 {
//		return targetMap
//	}
//	params := k.GetParams(ctx)
//	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
//	totalLiquidTokens := activeVals.TotalLiquidTokens(ctx, k.stakingKeeper)
//	totalWeight := activeVals.TotalWeight(whitelistedValMap)
//	ToBeTotalDelShares := totalLiquidTokens.TruncateInt().Add(addStakingAmt)
//	existOverWeightedVal := false
//
//	sharePerWeight := ToBeTotalDelShares.Quo(totalWeight)
//	crumb := ToBeTotalDelShares.Sub(sharePerWeight.Mul(totalWeight))
//
//	i := 0
//	for _, val := range activeVals {
//		weightedShare := val.GetWeight(whitelistedValMap, true).Mul(sharePerWeight)
//		if val.GetLiquidTokens(ctx, k.stakingKeeper).TruncateInt().GT(weightedShare) {
//			existOverWeightedVal = true
//		} else {
//			activeVals[i] = val
//			i++
//			targetMap[val.OperatorAddress] = weightedShare.Sub(val.GetLiquidTokens(ctx, k.stakingKeeper).TruncateInt())
//		}
//	}
//	// remove overWeightedVals for recursive call
//	activeVals = activeVals[:i]
//
//	if !existOverWeightedVal {
//		if v, ok := targetMap[activeVals[0].OperatorAddress]; ok {
//			targetMap[activeVals[0].OperatorAddress] = v.Add(crumb)
//		} else {
//			targetMap[activeVals[0].OperatorAddress] = crumb
//		}
//		return targetMap
//	} else {
//		return k.AddStakingTargetMap(ctx, activeVals, addStakingAmt)
//	}
//}
