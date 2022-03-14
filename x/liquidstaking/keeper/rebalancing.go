package keeper

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) GetProxyAccBalance(ctx sdk.Context, proxyAcc sdk.AccAddress) (balance sdk.Coin) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	return sdk.NewCoin(bondDenom, k.bankKeeper.SpendableCoins(ctx, proxyAcc).AmountOf(bondDenom))
}

// TryRedelegation attempts redelegation, which is applied only when successful through cached context because there is a constraint that fails if already receiving redelegation.
func (k Keeper) TryRedelegation(ctx sdk.Context, re types.Redelegation) (completionTime time.Time, err error) {
	cachedCtx, writeCache := ctx.CacheContext()
	srcVal := re.SrcValidator.GetOperator()
	dstVal := re.DstValidator.GetOperator()
	// calculate delShares from tokens with validation
	shares, err := k.stakingKeeper.ValidateUnbondAmount(
		cachedCtx, re.Delegator, srcVal, re.Amount,
	)
	if err != nil {
		return time.Time{}, err
	}

	// when last, full redelegation of shares from delegation
	if re.Last {
		shares = re.SrcValidator.GetDelShares(ctx, k.stakingKeeper)
	}
	completionTime, err = k.stakingKeeper.BeginRedelegation(cachedCtx, re.Delegator, srcVal, dstVal, shares)
	if err != nil {
		return time.Time{}, err
	}
	writeCache()
	return completionTime, nil
}

// Rebalance argument liquidVals containing ValidatorStatusActive which is containing just added on whitelist(liquidToken 0) and ValidatorStatusInactive to delist
func (k Keeper) Rebalance(ctx sdk.Context, proxyAcc sdk.AccAddress, liquidVals types.LiquidValidators, whitelistedValMap types.WhitelistedValMap, rebalancingTrigger sdk.Dec) (redelegations []types.Redelegation) {
	logger := k.Logger(ctx)
	totalLiquidTokens, _ := liquidVals.TotalLiquidTokens(ctx, k.stakingKeeper, false)
	if !totalLiquidTokens.IsPositive() {
		return []types.Redelegation{}
	}

	weightMap, totalWeight := k.GetWeightMap(ctx, liquidVals, whitelistedValMap)

	// no active liquid validators
	if !totalWeight.IsPositive() {
		return []types.Redelegation{}
	}

	// calculate rebalancing target map
	targetMap := map[string]sdk.Int{}
	totalTargetMap := sdk.ZeroInt()
	for _, val := range liquidVals {
		targetMap[val.OperatorAddress] = totalLiquidTokens.Mul(weightMap[val.OperatorAddress]).Quo(totalWeight)
		totalTargetMap = totalTargetMap.Add(targetMap[val.OperatorAddress])
	}
	crumb := totalLiquidTokens.Sub(totalTargetMap)
	if !totalTargetMap.IsPositive() {
		return []types.Redelegation{}
	}
	// crumb to first non zero liquid validator
	for _, val := range liquidVals {
		if targetMap[val.OperatorAddress].IsPositive() {
			targetMap[val.OperatorAddress] = targetMap[val.OperatorAddress].Add(crumb)
			break
		}
	}

	failCount := 0
	rebalancingThresholdAmt := rebalancingTrigger.Mul(totalLiquidTokens.ToDec()).TruncateInt()
	for i := 0; i < liquidVals.Len(); i++ {
		// sync totalLiquidTokens, liquidTokenMap applied rebalancing
		var liquidTokenMap map[string]sdk.Int
		totalLiquidTokens, liquidTokenMap = liquidVals.TotalLiquidTokens(ctx, k.stakingKeeper, false)

		// get min, max of liquid token gap
		minVal, maxVal, amountNeeded, last := liquidVals.MinMaxGap(targetMap, liquidTokenMap)
		if amountNeeded.IsZero() || (i == 0 && !amountNeeded.GT(rebalancingThresholdAmt)) {
			break
		}

		// try redelegation from max validator to min validator
		redelegation := types.Redelegation{
			Delegator:    proxyAcc,
			SrcValidator: maxVal,
			DstValidator: minVal,
			Amount:       amountNeeded,
			Last:         last,
		}
		redelegations = append(redelegations, redelegation)
		_, err := k.TryRedelegation(ctx, redelegation)
		if err != nil {
			logger.Error("rebalancing failed due to redelegation restriction", "redelegations", redelegations, "error", err)
			failCount++
		}
	}
	if len(redelegations) != 0 {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeBeginRebalancing,
				sdk.NewAttribute(types.AttributeKeyDelegator, types.LiquidStakingProxyAcc.String()),
				sdk.NewAttribute(types.AttributeKeyRedelegationCount, strconv.Itoa(len(redelegations))),
				sdk.NewAttribute(types.AttributeKeyRedelegationFailCount, strconv.Itoa(failCount)),
			),
		})
		logger.Info(types.EventTypeBeginRebalancing,
			types.AttributeKeyDelegator, types.LiquidStakingProxyAcc.String(),
			types.AttributeKeyRedelegationCount, strconv.Itoa(len(redelegations)),
			types.AttributeKeyRedelegationFailCount, strconv.Itoa(failCount))
	}
	return redelegations
}

// WithdrawRewardsAndReStaking withdraw rewards and re-staking when over threshold
func (k Keeper) WithdrawRewardsAndReStaking(ctx sdk.Context, whitelistedValMap types.WhitelistedValMap) {
	totalRemainingRewards, _, totalLiquidTokens := k.CheckDelegationStates(ctx, types.LiquidStakingProxyAcc)

	// checking over types.RewardTrigger and execute GetRewards
	proxyAccBalance := k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)
	rewardsThreshold := types.RewardTrigger.Mul(totalLiquidTokens.ToDec())

	// skip If it doesn't exceed the rewards threshold
	if !proxyAccBalance.Amount.ToDec().Add(totalRemainingRewards).GT(rewardsThreshold) {
		return
	}

	// Withdraw rewards of LiquidStakingProxyAcc and re-staking
	k.WithdrawLiquidRewards(ctx, types.LiquidStakingProxyAcc)

	// re-staking with proxyAccBalance, due to auto-withdraw on add staking by f1
	proxyAccBalance = k.GetProxyAccBalance(ctx, types.LiquidStakingProxyAcc)

	// skip when no active liquid validator
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	if len(activeVals) == 0 {
		return
	}

	// re-staking
	_, err := k.LiquidDelegate(ctx, types.LiquidStakingProxyAcc, activeVals, proxyAccBalance.Amount, whitelistedValMap)
	if err != nil {
		panic(err)
	}
	logger := k.Logger(ctx)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeReStake,
			sdk.NewAttribute(types.AttributeKeyDelegator, types.LiquidStakingProxyAcc.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, proxyAccBalance.String()),
		),
	})
	logger.Info(types.EventTypeReStake,
		types.AttributeKeyDelegator, types.LiquidStakingProxyAcc.String(),
		sdk.AttributeKeyAmount, proxyAccBalance.String())
}

func (k Keeper) UpdateLiquidValidatorSet(ctx sdk.Context) []types.Redelegation {
	logger := k.Logger(ctx)
	params := k.GetParams(ctx)
	liquidValidators := k.GetAllLiquidValidators(ctx)
	liquidValsMap := liquidValidators.Map()
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)

	// Set Liquid validators for added whitelist validators
	for _, wv := range params.WhitelistedValidators {
		if _, ok := liquidValsMap[wv.ValidatorAddress]; !ok {
			lv := types.LiquidValidator{
				OperatorAddress: wv.ValidatorAddress,
			}
			if k.IsActiveLiquidValidator(ctx, lv, whitelistedValMap) {
				k.SetLiquidValidator(ctx, lv)
				liquidValidators = append(liquidValidators, lv)
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						types.EventTypeAddLiquidValidator,
						sdk.NewAttribute(types.AttributeKeyLiquidValidator, lv.OperatorAddress),
					),
				})
				logger.Info(types.EventTypeAddLiquidValidator, types.AttributeKeyLiquidValidator, lv.OperatorAddress)
			}
		}
	}

	// rebalancing based updated liquid validators status with threshold, try by cachedCtx
	// tombstone status also handled on Rebalance
	reds := k.Rebalance(ctx, types.LiquidStakingProxyAcc, liquidValidators, whitelistedValMap, types.RebalancingTrigger)

	for _, lv := range liquidValidators {
		if !k.IsActiveLiquidValidator(ctx, lv, whitelistedValMap) {
			// unbond all delShares to proxyAcc if delShares exist on inactive validators
			delShares := lv.GetDelShares(ctx, k.stakingKeeper)
			if delShares.IsPositive() {
				completionTime, returnAmount, _, err := k.LiquidUnbond(ctx, types.LiquidStakingProxyAcc, types.LiquidStakingProxyAcc, lv.GetOperator(), delShares)
				if err != nil {
					panic(err)
				}
				unbondingAmount := sdk.Coin{Denom: k.stakingKeeper.BondDenom(ctx), Amount: returnAmount}.String()
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						types.EventTypeUnbondInactiveLiquidTokens,
						sdk.NewAttribute(types.AttributeKeyLiquidValidator, lv.OperatorAddress),
						sdk.NewAttribute(types.AttributeKeyUnbondingAmount, unbondingAmount),
						sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
					),
				})
				logger.Info(types.EventTypeUnbondInactiveLiquidTokens,
					types.AttributeKeyLiquidValidator, lv.OperatorAddress,
					types.AttributeKeyUnbondingAmount, unbondingAmount,
					types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339))
			}
			_, found := k.stakingKeeper.GetDelegation(ctx, types.LiquidStakingProxyAcc, lv.GetOperator())
			if !found {
				k.RemoveLiquidValidator(ctx, lv)
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						types.EventTypeRemoveLiquidValidator,
						sdk.NewAttribute(types.AttributeKeyLiquidValidator, lv.OperatorAddress),
					),
				})
				logger.Info(types.EventTypeRemoveLiquidValidator, types.AttributeKeyLiquidValidator, lv.OperatorAddress)
			}
		}
	}

	// withdraw rewards and re-staking when over threshold
	k.WithdrawRewardsAndReStaking(ctx, whitelistedValMap)
	return reds
}
