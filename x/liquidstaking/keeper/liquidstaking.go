package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) LiquidBondDenom(ctx sdk.Context) (res string) {
	k.paramSpace.Get(ctx, types.KeyLiquidBondDenom, &res)
	return
}

// NetAmount calculates the sum of bondedDenom balance, delegation power(slash applied LiquidTokens), the remaining reward of types.LiquidStakingProxyAcc
// During Liquid Unstacking, btoken immediately burns and the unbonding queue belongs to the requester, so the the unbonding values are excluded on netAmount.
func (k Keeper) NetAmount(ctx sdk.Context) sdk.Dec {
	balance := k.bankKeeper.GetBalance(ctx, types.LiquidStakingProxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
	totalRewards, _, totalLiquidTokens := k.CheckTotalRewards(ctx, types.LiquidStakingProxyAcc)
	return balance.ToDec().Add(totalLiquidTokens.ToDec()).Add(totalRewards)
}

// LiquidStaking mints bToken worth of staking coin value according to NetAmount and performs LiquidDelegate.
func (k Keeper) LiquidStaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, stakingCoin sdk.Coin) (newShares sdk.Dec, bTokenMintAmount sdk.Int, err error) {

	// check bond denomination
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if stakingCoin.Denom != bondDenom {
		return sdk.ZeroDec(), bTokenMintAmount, sdkerrors.Wrapf(
			types.ErrInvalidBondDenom, "invalid coin denomination: got %s, expected %s", stakingCoin.Denom, bondDenom,
		)
	}

	params := k.GetParams(ctx)
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	if activeVals.Len() == 0 || !activeVals.TotalWeight(whitelistedValMap).IsPositive() {
		return sdk.ZeroDec(), bTokenMintAmount, types.ErrActiveLiquidValidatorsNotExists
	}

	// NetAmount must be calculated before send
	netAmount := k.NetAmount(ctx)

	// send staking coin to liquid staking proxy account to proxy delegation
	err = k.bankKeeper.SendCoins(ctx, liquidStaker, proxyAcc, sdk.NewCoins(stakingCoin))
	if err != nil {
		return sdk.ZeroDec(), bTokenMintAmount, err
	}

	// mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
	liquidBondDenom := k.LiquidBondDenom(ctx)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom)
	bTokenMintAmount = stakingCoin.Amount
	if bTokenTotalSupply.IsPositive() {
		bTokenMintAmount = types.NativeTokenToBToken(stakingCoin.Amount, bTokenTotalSupply.Amount, netAmount)
	}

	// mint on module acc and send
	mintCoin := sdk.NewCoins(sdk.NewCoin(liquidBondDenom, bTokenMintAmount))
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoin)
	if err != nil {
		return sdk.ZeroDec(), bTokenMintAmount, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStaker, mintCoin)
	if err != nil {
		return sdk.ZeroDec(), bTokenMintAmount, err
	}
	newShares, err = k.LiquidDelegate(ctx, proxyAcc, activeVals, stakingCoin.Amount, whitelistedValMap)
	return newShares, bTokenMintAmount, err
}

// LiquidDelegate delegates staking amount to active validators by proxy account.
func (k Keeper) LiquidDelegate(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.ActiveLiquidValidators, stakingAmt sdk.Int, whitelistedValMap types.WhitelistedValMap) (newShares sdk.Dec, err error) {
	totalNewShares := sdk.ZeroDec()
	// crumb may occur due to a decimal point error in dividing the staking amount into the weight of liquid validators, It added on first active liquid validator
	weightedAmt, crumb := types.DivideByWeight(activeVals, stakingAmt, whitelistedValMap)
	if len(weightedAmt) == 0 {
		return sdk.ZeroDec(), types.ErrInvalidActiveLiquidValidators
	}
	weightedAmt[0] = weightedAmt[0].Add(crumb)
	for i, val := range activeVals {
		if !weightedAmt[i].IsPositive() {
			continue
		}
		validator, _ := k.stakingKeeper.GetValidator(ctx, val.GetOperator())
		newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, weightedAmt[i], stakingtypes.Unbonded, validator, true)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		totalNewShares = totalNewShares.Add(newShares)
	}
	return totalNewShares, nil
}

// LiquidUnstaking burns unstakingBtoken and performs LiquidUnbond to active liquid validators with del shares worth of shares according to NetAmount with each validators current weight.
func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, unstakingBtoken sdk.Coin,
) (time.Time, sdk.Int, []stakingtypes.UnbondingDelegation, error) {

	// check bond denomination
	params := k.GetParams(ctx)
	liquidBondDenom := k.LiquidBondDenom(ctx)
	if unstakingBtoken.Denom != liquidBondDenom {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, sdkerrors.Wrapf(
			types.ErrInvalidLiquidBondDenom, "invalid coin denomination: got %s, expected %s", unstakingBtoken.Denom, liquidBondDenom,
		)
	}

	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	if activeVals.Len() == 0 {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, types.ErrActiveLiquidValidatorsNotExists
	}

	// UnstakeAmount = NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom)
	unstakingAll := false
	if unstakingBtoken.Amount.GT(bTokenTotalSupply.Amount) {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, types.ErrInvalidBTokenSupply
	} else if unstakingBtoken.Amount.Equal(bTokenTotalSupply.Amount) {
		// TODO: need to policy for last liquid unstaking for remaining rewards, balance
		unstakingAll = true
	}
	netAmount := k.NetAmount(ctx)

	unbondingAmount := types.BTokenToNativeToken(unstakingBtoken.Amount, bTokenTotalSupply.Amount, netAmount)
	unbondingAmount = types.DeductFeeRate(unbondingAmount, params.UnstakeFeeRate)

	totalReturnAmount := sdk.ZeroInt()

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidStaker, types.ModuleName, sdk.NewCoins(unstakingBtoken))
	if err != nil {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(liquidBondDenom, unstakingBtoken.Amount)))
	if err != nil {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, err
	}

	totalLiquidTokens, liquidTokenMap := activeVals.TotalLiquidTokens(ctx, k.stakingKeeper)
	// crumb may occur due to a decimal error in dividing the unstaking bToken into the weight of liquid validators, add it to first sufficient active validator
	unbondingAmounts, crumb := types.DivideByCurrentWeight(activeVals, unbondingAmount, totalLiquidTokens, liquidTokenMap)
	if len(unbondingAmounts) == 0 {
		return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, types.ErrInvalidActiveLiquidValidators
	}
	var ubdTime time.Time
	var ubds []stakingtypes.UnbondingDelegation
	for i, val := range activeVals {
		var ubd stakingtypes.UnbondingDelegation
		var returnAmount sdk.Int
		var weightedShare sdk.Dec
		// unstaking all when last unstaking request(unstakingBtoken == bTokenTotalSupply)
		del, found := k.stakingKeeper.GetDelegation(ctx, proxyAcc, val.GetOperator())
		if unstakingAll && found && del.Shares.IsPositive() {
			weightedShare = del.Shares
		} else if crumb.IsPositive() {
			// crumb to first sufficient active liquid validator
			weightedShareCrumbAdded, err := k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, val.GetOperator(), unbondingAmounts[i].Add(crumb).TruncateInt())
			if err == nil {
				crumb = sdk.ZeroDec()
				weightedShare = weightedShareCrumbAdded
			}
		}
		if weightedShare.IsNil() {
			// calculate delShares from tokens with validation
			weightedShare, err = k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, val.GetOperator(), unbondingAmounts[i].TruncateInt())
			if err != nil {
				return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, err
			}
		}
		if !weightedShare.IsPositive() {
			continue
		}
		// unbond with weightedShare
		ubdTime, returnAmount, ubd, err = k.LiquidUnbond(ctx, proxyAcc, liquidStaker, val.GetOperator(), weightedShare)
		if err != nil {
			return time.Time{}, sdk.ZeroInt(), []stakingtypes.UnbondingDelegation{}, err
		}
		ubds = append(ubds, ubd)
		totalReturnAmount = totalReturnAmount.Add(returnAmount)
	}
	return ubdTime, totalReturnAmount, ubds, nil
}

// LiquidUnbond unbond delegation shares to active validators by proxy account.
func (k Keeper) LiquidUnbond(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, valAddr sdk.ValAddress, shares sdk.Dec,
) (time.Time, sdk.Int, stakingtypes.UnbondingDelegation, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, sdk.ZeroInt(), stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, proxyAcc, valAddr, shares)
	if err != nil {
		return time.Time{}, sdk.ZeroInt(), stakingtypes.UnbondingDelegation{}, err
	}

	// transfer the validator tokens to the not bonded pool
	if validator.IsBonded() {
		coins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), returnAmount))
		if err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins); err != nil {
			panic(err)
		}
	}

	completionTime := ctx.BlockHeader().Time.Add(k.stakingKeeper.UnbondingTime(ctx))
	ubd := k.stakingKeeper.SetUnbondingDelegationEntry(ctx, liquidStaker, valAddr, ctx.BlockHeight(), completionTime, returnAmount)
	k.stakingKeeper.InsertUBDQueue(ctx, ubd, completionTime)

	return completionTime, returnAmount, ubd, nil
}

func (k Keeper) CheckTotalRewards(ctx sdk.Context, proxyAcc sdk.AccAddress) (sdk.Dec, sdk.Dec, sdk.Int) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroInt()
	totalRewards := sdk.ZeroDec()

	// Cache ctx for calculate rewards
	cachedCtx, _ := ctx.CacheContext()
	k.stakingKeeper.IterateDelegations(
		cachedCtx, proxyAcc,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(cachedCtx, valAddr)
			endingPeriod := k.distrKeeper.IncrementValidatorPeriod(cachedCtx, val)
			delReward := k.distrKeeper.CalculateDelegationRewards(cachedCtx, val, del, endingPeriod)
			delShares := del.GetShares()
			if delShares.IsPositive() {
				totalDelShares = totalDelShares.Add(delShares)
				liquidTokens := val.TokensFromSharesTruncated(delShares).TruncateInt()
				totalLiquidTokens = totalLiquidTokens.Add(liquidTokens)
				totalRewards = totalRewards.Add(delReward.AmountOf(bondDenom))
			}
			return false
		},
	)

	return totalRewards, totalDelShares, totalLiquidTokens
}

func (k Keeper) WithdrawLiquidRewards(ctx sdk.Context, proxyAcc sdk.AccAddress) sdk.Int {
	totalRewards := sdk.ZeroInt()
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	k.stakingKeeper.IterateDelegations(
		ctx, proxyAcc,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			reward, err := k.distrKeeper.WithdrawDelegationRewards(ctx, proxyAcc, valAddr)
			if err != nil {
				panic(err)
			}
			totalRewards = totalRewards.Add(reward.AmountOf(bondDenom))
			return false
		},
	)
	return totalRewards
}

// GetLiquidValidator get a single liquid validator
func (k Keeper) GetLiquidValidator(ctx sdk.Context, addr sdk.ValAddress) (val types.LiquidValidator, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetLiquidValidatorKey(addr))
	if value == nil {
		return val, false
	}

	val = types.MustUnmarshalLiquidValidator(k.cdc, value)
	return val, true
}

// SetLiquidValidator set the main record holding liquid validator details
func (k Keeper) SetLiquidValidator(ctx sdk.Context, val types.LiquidValidator) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalLiquidValidator(k.cdc, &val)
	store.Set(types.GetLiquidValidatorKey(val.GetOperator()), bz)
}

// RemoveLiquidValidator remove a liquid validator on kv store
func (k Keeper) RemoveLiquidValidator(ctx sdk.Context, val types.LiquidValidator) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLiquidValidatorKey(val.GetOperator()))
}

// GetAllLiquidValidators get the set of all liquid validators with no limits, used during genesis dump
func (k Keeper) GetAllLiquidValidators(ctx sdk.Context) (vals types.LiquidValidators) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		vals = append(vals, val)
	}

	return vals
}

// GetActiveLiquidValidators get the set of active liquid validators.
func (k Keeper) GetActiveLiquidValidators(ctx sdk.Context, whitelistedValMap types.WhitelistedValMap) (vals types.ActiveLiquidValidators) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		if k.ActiveCondition(ctx, val, whitelistedValMap.IsListed(val.OperatorAddress)) {
			vals = append(vals, val)
		}
	}
	return vals
}

func (k Keeper) GetAllLiquidValidatorStates(ctx sdk.Context) (liquidValidatorStates []types.LiquidValidatorState) {
	lvs := k.GetAllLiquidValidators(ctx)
	whitelistedValMap := k.GetParams(ctx).WhitelistedValMap()
	for _, lv := range lvs {
		active := k.ActiveCondition(ctx, lv, whitelistedValMap.IsListed(lv.OperatorAddress))
		lvState := types.LiquidValidatorState{
			OperatorAddress: lv.OperatorAddress,
			Weight:          lv.GetWeight(whitelistedValMap, active),
			Status:          lv.GetStatus(active),
			DelShares:       lv.GetDelShares(ctx, k.stakingKeeper),
			LiquidTokens:    lv.GetLiquidTokens(ctx, k.stakingKeeper),
		}
		liquidValidatorStates = append(liquidValidatorStates, lvState)
	}
	return
}

func (k Keeper) GetLiquidValidatorState(ctx sdk.Context, addr sdk.ValAddress) (liquidValidatorState types.LiquidValidatorState, found bool) {
	lv, found := k.GetLiquidValidator(ctx, addr)
	if !found {
		return types.LiquidValidatorState{
			OperatorAddress: addr.String(),
			Weight:          sdk.ZeroInt(),
			Status:          types.ValidatorStatusUnspecified,
			DelShares:       sdk.ZeroDec(),
			LiquidTokens:    sdk.ZeroInt(),
		}, false
	}
	whitelistedValMap := k.GetParams(ctx).WhitelistedValMap()
	active := k.ActiveCondition(ctx, lv, whitelistedValMap.IsListed(lv.OperatorAddress))
	return types.LiquidValidatorState{
		OperatorAddress: lv.OperatorAddress,
		Weight:          lv.GetWeight(whitelistedValMap, active),
		Status:          lv.GetStatus(active),
		DelShares:       lv.GetDelShares(ctx, k.stakingKeeper),
		LiquidTokens:    lv.GetLiquidTokens(ctx, k.stakingKeeper),
	}, true
}

func (k Keeper) ActiveCondition(ctx sdk.Context, v types.LiquidValidator, whitelisted bool) bool {
	val, found := k.stakingKeeper.GetValidator(ctx, v.GetOperator())
	if !found {
		return false
	}
	tombstoned := v.IsTombstoned(ctx, k.stakingKeeper, k.slashingKeeper)
	return types.ActiveCondition(val, whitelisted, tombstoned)
}

func (k Keeper) GetWeightMap(ctx sdk.Context, liquidVals types.LiquidValidators, whitelistedValMap types.WhitelistedValMap) (map[string]sdk.Int, sdk.Int) {
	weightMap := make(map[string]sdk.Int)
	totalWeight := sdk.ZeroInt()
	for _, val := range liquidVals {
		weight := val.GetWeight(whitelistedValMap, k.ActiveCondition(ctx, val, whitelistedValMap.IsListed(val.OperatorAddress)))
		totalWeight = totalWeight.Add(weight)
		weightMap[val.OperatorAddress] = weight
	}
	return weightMap, totalWeight
}

//// Deprecated: LiquidStakingWithBalancing for using simple weight distribution, not rebalancing, not using on this version for simplify.
//func (k Keeper) LiquidStakingWithBalancing(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.ActiveLiquidValidators, stakingAmt sdk.Int) (newShares sdk.Dec, err error) {
//	totalNewShares := sdk.ZeroDec()
//	targetMap := k.AddStakingTargetMap(ctx, activeVals, stakingAmt)
//	for valStr, amt := range targetMap {
//		val, err := sdk.ValAddressFromBech32(valStr)
//		if err != nil {
//			return sdk.ZeroDec(), err
//		}
//		validator, found := k.stakingKeeper.GetValidator(ctx, val)
//		if !found {
//			panic("validator not founded")
//		}
//		newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, amt, stakingtypes.Unbonded, validator, true)
//		if err != nil {
//			return sdk.ZeroDec(), err
//		}
//		totalNewShares = totalNewShares.Add(newShares)
//	}
//	return totalNewShares, nil
//}

//// Deprecated: GetValidatorsMap get the set of all validators as map with no limits
//func (k Keeper) GetValidatorsMap(ctx sdk.Context) map[string]stakingtypes.Validator {
//	valMap := make(map[string]stakingtypes.Validator)
//	vals := k.stakingKeeper.GetAllValidators(ctx)
//	for _, val := range vals {
//		valMap[val.OperatorAddress] = val
//	}
//	return valMap
//}
