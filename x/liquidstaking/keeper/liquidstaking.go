package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (k Keeper) BondedBondDenom(ctx sdk.Context) (res string) {
	k.paramSpace.Get(ctx, types.KeyBondedBondDenom, &res)
	return
}

func (k Keeper) NetAmount(ctx sdk.Context) sdk.Dec {
	// delegation power, bondDenom balance, remaining reward, unbonding amount of types.LiquidStakingProxyAcc
	balance := k.bankKeeper.GetBalance(ctx, types.LiquidStakingProxyAcc, k.stakingKeeper.BondDenom(ctx)).Amount
	totalRewards, _, totalLiquidTokens := k.CheckTotalRewards(ctx, types.LiquidStakingProxyAcc)
	// TODO: check unbonding value included on NetAmount
	fmt.Println("[balance, totalLiquidTokens, totalRewards]", balance, totalLiquidTokens, totalRewards)
	return balance.ToDec().Add(totalLiquidTokens).Add(totalRewards)
}

func (k Keeper) LiquidDelegate(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.ActiveLiquidValidators, stakingAmt sdk.Int, whitelistedValMap types.WhitelistedValMap) (newShares sdk.Dec, err error) {
	totalNewShares := sdk.ZeroDec()
	// crumb may occur due to a decimal point error in dividing the staking amount into the weight of liquid validators, It added on first active liquid validator
	weightedShares, crumb := types.DivideByWeight(activeVals, stakingAmt, whitelistedValMap)
	weightedShares[0] = weightedShares[0].Add(crumb)
	for i, val := range activeVals {
		validator, _ := k.stakingKeeper.GetValidator(ctx, val.GetOperator())
		newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, weightedShares[i], stakingtypes.Unbonded, validator, true)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		totalNewShares = totalNewShares.Add(newShares)
	}
	return totalNewShares, nil
}

// LiquidStaking ...
func (k Keeper) LiquidStaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, stakingCoin sdk.Coin) (newShares sdk.Dec, bTokenMintAmount sdk.Int, err error) {

	// check bond denomination
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if stakingCoin.Denom != bondDenom {
		return sdk.ZeroDec(), bTokenMintAmount, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", stakingCoin.Denom, bondDenom,
		)
	}

	params := k.GetParams(ctx)
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	if activeVals.Len() == 0 {
		return sdk.ZeroDec(), bTokenMintAmount, fmt.Errorf("there's no active liquid validators")
	}

	netAmount := k.NetAmount(ctx)

	// send staking coin to liquid staking proxy account to proxy delegation
	err = k.bankKeeper.SendCoins(ctx, liquidStaker, proxyAcc, sdk.NewCoins(stakingCoin))
	if err != nil {
		return sdk.ZeroDec(), bTokenMintAmount, err
	}

	// mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
	bondedBondDenom := k.BondedBondDenom(ctx)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, bondedBondDenom)
	bTokenMintAmount = stakingCoin.Amount
	if bTokenTotalSupply.IsPositive() {
		bTokenMintAmount = types.NativeTokenToBToken(stakingCoin.Amount, bTokenTotalSupply.Amount, netAmount)
	}

	// mint on module acc and send
	mintCoin := sdk.NewCoins(sdk.NewCoin(bondedBondDenom, bTokenMintAmount))
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

// LiquidUnstaking ...
func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, unstakingBtoken sdk.Coin,
) (time.Time, sdk.Dec, []stakingtypes.UnbondingDelegation, error) {

	// check bond denomination
	params := k.GetParams(ctx)
	bondedBondDenom := k.BondedBondDenom(ctx)
	if unstakingBtoken.Denom != bondedBondDenom {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", unstakingBtoken.Denom, bondedBondDenom,
		)
	}

	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	activeVals := k.GetActiveLiquidValidators(ctx, whitelistedValMap)
	if activeVals.Len() == 0 {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("there's no active liquid validators")
	}

	// UnstakeAmount = NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, bondedBondDenom)
	unstakingAll := false
	//if !bTokenTotalSupply.IsPositive() {
	if unstakingBtoken.Amount.GT(bTokenTotalSupply.Amount) {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("unstakingAll supply is not positive")
	} else if unstakingBtoken.Amount.Equal(bTokenTotalSupply.Amount) {
		// TODO: verify with netAmount for rewards, balance
		unstakingAll = true
	}
	netAmount := k.NetAmount(ctx)
	unbondingAmount := types.BTokenToNativeToken(unstakingBtoken.Amount, bTokenTotalSupply.Amount, netAmount, params.UnstakeFeeRate)
	totalReturnAmount := sdk.ZeroInt()

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidStaker, types.ModuleName, sdk.NewCoins(unstakingBtoken))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(bondedBondDenom, unstakingBtoken.Amount)))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}

	// crumb may occur due to a decimal error in dividing the unstaking bToken into the weight of liquid validators, which is also accumulated in the netAmount value
	unbondingAmounts, crumb := k.DivideByCurrentWeight(ctx, activeVals, unbondingAmount)
	// TODO: ValidateUnbondAmount for sufficient delShares
	unbondingAmounts[0] = unbondingAmounts[0].Add(crumb)
	var ubdTime time.Time
	var ubds []stakingtypes.UnbondingDelegation
	for i, val := range activeVals {
		var ubd stakingtypes.UnbondingDelegation
		var returnAmount sdk.Int
		del, found := k.stakingKeeper.GetDelegation(ctx, proxyAcc, val.GetOperator())
		weightedShare := sdk.ZeroDec()
		// TODO: add test case
		if unstakingAll {
			weightedShare = del.Shares
		} else {
			weightedShare, err = k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, val.GetOperator(), unbondingAmounts[i].TruncateInt())
			if err != nil {
				return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
			}
		}
		fmt.Println("[liquid UBD]", weightedShare.String(), del.Shares.String(), found, unbondingAmounts[i], activeVals.Len(), unbondingAmount.String())
		ubdTime, returnAmount, ubd, err = k.LiquidUnbond(ctx, proxyAcc, liquidStaker, val.GetOperator(), weightedShare)
		if err != nil {
			return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
		}
		ubds = append(ubds, ubd)
		totalReturnAmount = totalReturnAmount.Add(returnAmount)
		k.SetLiquidValidator(ctx, val)
	}
	return ubdTime, totalReturnAmount.ToDec(), ubds, nil
}

// LiquidUnbond ...
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

func (k Keeper) CheckTotalRewards(ctx sdk.Context, proxyAcc sdk.AccAddress) (sdk.Dec, sdk.Dec, sdk.Dec) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
	totalLiquidTokens := sdk.ZeroDec()
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
				liquidTokens := val.TokensFromSharesTruncated(delShares)
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

func (k Keeper) GetLiquidValidatorStates(ctx sdk.Context) (liquidValidatorStates []types.LiquidValidatorState) {
	lvs := k.GetAllLiquidValidators(ctx)
	whitelistedValMap := k.GetParams(ctx).WhitelistedValMap()
	for _, lv := range lvs {
		active := k.ActiveCondition(ctx, lv, whitelistedValMap.IsListed(lv.OperatorAddress))
		lvState := types.LiquidValidatorState{
			OperatorAddress: lv.OperatorAddress,
			Weight:          lv.GetWeight(whitelistedValMap, active),
			Status:          lv.GetStatus(active),
			// TODO: could be deprecated or reduce duplicated sk
			DelShares:    lv.GetDelShares(ctx, k.stakingKeeper),
			LiquidTokens: lv.GetLiquidTokens(ctx, k.stakingKeeper),
		}
		liquidValidatorStates = append(liquidValidatorStates, lvState)
	}
	return
}

func (k Keeper) ActiveCondition(ctx sdk.Context, v types.LiquidValidator, whitelisted bool) bool {
	val, found := k.stakingKeeper.GetValidator(ctx, v.GetOperator())
	if !found {
		return false
	}
	tombstoned := false
	if val.Jailed {
		consPk, err := val.ConsPubKey()
		if err != nil {
			tombstoned = false
		} else {
			tombstonedByValOper := k.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(val.GetOperator()))
			tombstonedByValCons := k.slashingKeeper.IsTombstoned(ctx, sdk.ConsAddress(consPk.Address()))
			tombstoned = tombstonedByValCons
			// TODO: WIP check ConsAddress, not operator, consensus Pubkey address
			//if !sdk.ConsAddress(consPk.Address()).Equals(sdk.ConsAddress(val.GetOperator())) {
			if tombstonedByValOper != tombstonedByValCons {
				panic("need to debug for tombstone checking with valCons or valOper")
			}
		}
	}
	return types.ActiveCondition(val, whitelisted, tombstoned)
}

// TODO: test
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

//// Deprecated: GetLiquidUnbonding
//func (k Keeper) GetLiquidUnbonding(ctx sdk.Context, proxyAcc sdk.AccAddress) []stakingtypes.UnbondingDelegation {
//	return k.stakingKeeper.GetAllUnbondingDelegations(ctx, proxyAcc)
//}

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
