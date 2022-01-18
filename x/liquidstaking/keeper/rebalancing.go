package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) UpdateLiquidValidators(ctx sdk.Context) {
	//liquidVals := k.GetAllLiquidValidators()
	// TODO: GET, SET, GETALL, Iterate LiquidValidators, indexing
	//for _, val := range liquidVals {
	//	if val.
	//}
	//k.stakingKeeper.GetLastTotalPower()
}

// activeVals containing ValidatorStatusActive which is containing just added on whitelist(power 0) and ValidatorStatusDelisting
func (k Keeper) Rebalancing(ctx sdk.Context, moduleAcc sdk.AccAddress, activeVals types.LiquidValidators, threshold sdk.Dec) (rebalancedLiquidVals types.LiquidValidators) {
	totalLiquidTokens := sdk.ZeroInt()
	totalWeight := sdk.ZeroInt()
	for _, val := range activeVals {
		totalLiquidTokens = totalLiquidTokens.Add(val.LiquidTokens)
		if val.Status == 1 {
			totalWeight = totalWeight.Add(val.Weight)
		}
	}

	var targetWeight sdk.Int
	targetMap := map[string]sdk.Int{}
	for _, val := range activeVals {
		if val.Status == 1 {
			targetWeight = val.Weight
		} else {
			targetWeight = sdk.ZeroInt()
		}
		targetMap[val.OperatorAddress] = totalLiquidTokens.Mul(targetWeight).Quo(totalWeight)
	}
	fmt.Println(targetMap)

	for i := 0; i < len(activeVals); i++ {
		maxVal, minVal, amountNeeded := activeVals.MinMaxGap(targetMap)
		if amountNeeded.IsZero() || (i == 0 && amountNeeded.LT(threshold.TruncateInt())) {
			break
		}
		for idx := range activeVals {
			if activeVals[idx].OperatorAddress == maxVal.OperatorAddress {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Add(amountNeeded)
			}
			if activeVals[idx].OperatorAddress == minVal.OperatorAddress {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Sub(amountNeeded)
			}
		}
	}
	fmt.Println(activeVals)

	return activeVals
}

//AddStakingTargetMap is make add staking target map for one-way rebalancing, it can be called recursively.
func AddStakingTargetMap(activeVals types.LiquidValidators, addStakingAmt sdk.Int) map[string]sdk.Int {
	totalLiquidTokens := activeVals.TotalLiquidTokens()
	totalWeight := activeVals.TotalWeight()
	ToBeTotalLiquidTokens := totalLiquidTokens.Add(addStakingAmt)
	targetMap := make(map[string]sdk.Int)
	existOverWeightedVal := false

	sharePerWeight := ToBeTotalLiquidTokens.Quo(totalWeight)
	crumb := ToBeTotalLiquidTokens.Sub(sharePerWeight.Mul(totalWeight))

	i := 0
	for _, val := range activeVals {
		weightedShare := val.Weight.Mul(sharePerWeight)
		if val.LiquidTokens.GT(weightedShare) {
			existOverWeightedVal = true
		} else {
			activeVals[i] = val
			i++
			targetMap[val.OperatorAddress] = weightedShare.Sub(val.LiquidTokens)
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
		fmt.Println("[AddStakingTargetMap] recursive call for", activeVals, addStakingAmt, totalLiquidTokens, ToBeTotalLiquidTokens, totalWeight, sharePerWeight, crumb)
		return AddStakingTargetMap(activeVals, addStakingAmt)
	}
	return targetMap
}

func (k Keeper) ProcessStaking(moduleAcc sdk.AccAddress, activeVals types.LiquidValidators, addStakingTokens sdk.Int, unstakingTokens sdk.Int) (rebalancedLiquidVals types.LiquidValidators) {
	// suppose that when unstaking process starts, the required amount of unstaking is transferred to (notBonded) moduleAcc
	// and when the unstaking queue matures, finally the tokens are transferred to staker's address
	// addStakingAmt : additional staking amount to be considered
	// unstakingTokens : unstaking amount to be considered

	type accountBalance struct {
		Address      string
		LiquidTokens sdk.Int
	}
	moduleAccBalance := accountBalance{
		Address:      moduleAcc.String(),
		LiquidTokens: sdk.ZeroInt(),
	}
	// temporary struct for tokens in moduleAcc (require to fix)

	if addStakingTokens.GT(unstakingTokens) {
		moduleAccBalance.LiquidTokens = moduleAccBalance.LiquidTokens.Add(unstakingTokens)
		addStakingTokens = addStakingTokens.Sub(unstakingTokens) // above 2 line must change to send action
		distributionAmount := addStakingTokens.Quo(sdk.NewInt(int64(len(activeVals))))
		for idx := range activeVals {
			if distributionAmount.Equal(sdk.ZeroInt()) {
				break
			}
			if activeVals[idx].Status == 1 {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Add(distributionAmount)
				addStakingTokens = addStakingTokens.Sub(distributionAmount) // must change to delegate action
			}
		}
	} else {
		moduleAccBalance.LiquidTokens = moduleAccBalance.LiquidTokens.Add(addStakingTokens)
		unstakingTokens = unstakingTokens.Sub(addStakingTokens) // must change to send action
		//<<<<<<< Updated upstream
		contributionAmount := unstakingTokens.Quo(sdk.NewInt(int64(len(activeVals))))
		for idx := range activeVals {
			if contributionAmount.Equal(sdk.ZeroInt()) {
				break
			}
			if activeVals[idx].Status == 1 {
				activeVals[idx].LiquidTokens = activeVals[idx].LiquidTokens.Sub(contributionAmount)
				moduleAccBalance.LiquidTokens = moduleAccBalance.LiquidTokens.Add(contributionAmount) // must change to undelegate action
			}
		}
	}
	fmt.Println(moduleAccBalance)
	fmt.Println(activeVals)

	return activeVals
}
