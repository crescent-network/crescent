package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Int) {
	return k.bankKeeper.GetBalance(ctx, proxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
}

func (k Keeper) TryRedelegation(ctx sdk.Context, re types.Redelegation) (completionTime time.Time, err error) {
	cachedCtx, writeCache := ctx.CacheContext()
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
	return completionTime, nil
}

// DivideByCurrentWeight divide the input value by the ratio of the weight of the liquid validator's liquid token and return it with crumb
// which is may occur while dividing according to the weight of liquid validators by decimal error.
func (k Keeper) DivideByCurrentWeight(ctx sdk.Context, vs types.LiquidValidators, input sdk.Dec) (outputs []sdk.Dec, crumb sdk.Dec) {
	_, totalLiquidTokens := vs.TotalDelSharesAndLiquidTokens(ctx, k.stakingKeeper)
	totalShares := sdk.ZeroDec()
	sharePerWeight := input.QuoTruncate(totalLiquidTokens)
	for _, val := range vs {
		weightedShare := sharePerWeight.Mul(val.GetDelShares(ctx, k.stakingKeeper))
		totalShares = totalShares.Add(weightedShare)
		outputs = append(outputs, weightedShare)
	}
	return outputs, input.Sub(totalShares)
}

// activeVals containing ValidatorStatusActive which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
func (k Keeper) Rebalancing(ctx sdk.Context, proxyAcc sdk.AccAddress, liquidVals types.LiquidValidators, rebalancingTrigger sdk.Dec, whitelistedValMap types.WhitelistedValMap) (redelegations []types.Redelegation) {
	_, totalLiquidTokens := liquidVals.TotalDelSharesAndLiquidTokens(ctx, k.stakingKeeper)
	totalWeight := liquidVals.TotalWeight(whitelistedValMap)
	threshold := rebalancingTrigger.Mul(totalLiquidTokens)

	var targetWeight sdk.Int
	targetMap := map[string]sdk.Int{}
	for _, val := range liquidVals {
		if k.ActiveCondition(ctx, val, whitelistedValMap) {
			targetWeight = val.GetWeight(whitelistedValMap)
		} else {
			targetWeight = sdk.ZeroInt()
		}
		targetMap[val.OperatorAddress] = totalLiquidTokens.MulInt(targetWeight).QuoInt(totalWeight).TruncateInt()
	}

	for i := 0; i < len(liquidVals); i++ {
		maxVal, minVal, amountNeeded := liquidVals.MinMaxGap(ctx, k.stakingKeeper, targetMap)
		if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold.TruncateInt())) {
			break
		}
		redelegation := types.Redelegation{
			Delegator:    proxyAcc,
			SrcValidator: minVal.GetOperator(),
			DstValidator: maxVal.GetOperator(),
			Amount:       amountNeeded,
		}
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

// WithdrawRewardsAndReStaking withdraw rewards and re-staking when over threshold
func (k Keeper) WithdrawRewardsAndReStaking(ctx sdk.Context, valMap map[string]stakingtypes.Validator, whitelistedValMap types.WhitelistedValMap) {
	activeVals := k.GetActiveLiquidValidators(ctx, valMap, whitelistedValMap)
	_, totalLiquidTokens := activeVals.TotalDelSharesAndLiquidTokens(ctx, k.stakingKeeper)
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

func (k Keeper) EndBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)
	liquidValidators := k.GetAllLiquidValidators(ctx)
	liquidValsMap := liquidValidators.Map()
	valMap := k.GetValidatorsMap(ctx)
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if _, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			lv := &types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
			}
			if types.ActiveCondition(valMap[lv.OperatorAddress], true) {
				k.SetLiquidValidator(ctx, *lv)
				liquidValidators = append(liquidValidators, *lv)
			}
		}
	}

	// rebalancing based updated liquid validators status with threshold, try by cachedCtx
	k.Rebalancing(ctx, types.LiquidStakingProxyAcc, liquidValidators, types.RebalancingTrigger, whitelistedValMap)

	// withdraw rewards and re-staking when over threshold
	k.WithdrawRewardsAndReStaking(ctx, valMap, whitelistedValMap)
}

// Deprecated: AddStakingTargetMap is make add staking target map for one-way rebalancing, it can be called recursively, not using on this version for simplify.
func (k Keeper) AddStakingTargetMap(ctx sdk.Context, activeVals types.LiquidValidators, addStakingAmt sdk.Int) map[string]sdk.Int {
	targetMap := make(map[string]sdk.Int)
	if addStakingAmt.IsNil() || !addStakingAmt.IsPositive() || activeVals.Len() == 0 {
		return targetMap
	}
	params := k.GetParams(ctx)
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	_, totalLiquidTokens := activeVals.TotalDelSharesAndLiquidTokens(ctx, k.stakingKeeper)
	totalWeight := activeVals.TotalWeight(whitelistedValMap)
	ToBeTotalDelShares := totalLiquidTokens.TruncateInt().Add(addStakingAmt)
	existOverWeightedVal := false

	sharePerWeight := ToBeTotalDelShares.Quo(totalWeight)
	crumb := ToBeTotalDelShares.Sub(sharePerWeight.Mul(totalWeight))

	i := 0
	for _, val := range activeVals {
		weightedShare := val.GetWeight(whitelistedValMap).Mul(sharePerWeight)
		if val.GetDelShares(ctx, k.stakingKeeper).TruncateInt().GT(weightedShare) {
			existOverWeightedVal = true
		} else {
			activeVals[i] = val
			i++
			targetMap[val.OperatorAddress] = weightedShare.Sub(val.GetDelShares(ctx, k.stakingKeeper).TruncateInt())
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
		return k.AddStakingTargetMap(ctx, activeVals, addStakingAmt)
	}
}
