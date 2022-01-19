package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (k Keeper) LiquidBondDenom(ctx sdk.Context) (res string) {
	k.paramSpace.Get(ctx, types.KeyLiquidBondDenom, &res)
	return
}

func (k Keeper) NetAmount(ctx sdk.Context) sdk.Dec {
	// delegation power, bondDenom balance, remaining reward, unbonding amount of types.LiquidStakingProxyAcc
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	balance := k.bankKeeper.GetBalance(ctx, types.LiquidStakingProxyAcc, bondDenom)
	ubds := k.stakingKeeper.GetAllUnbondingDelegations(ctx, types.LiquidStakingProxyAcc)
	liquidPower := sdk.ZeroDec()
	unbondingPower := sdk.ZeroInt()
	totalRewards := sdk.ZeroDec()

	// Cache ctx for calculate rewards
	cachedCtx, _ := ctx.CacheContext()
	k.stakingKeeper.IterateDelegations(
		cachedCtx, types.LiquidStakingProxyAcc,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(cachedCtx, valAddr)
			endingPeriod := k.distrKeeper.IncrementValidatorPeriod(cachedCtx, val)
			delReward := k.distrKeeper.CalculateDelegationRewards(cachedCtx, val, del, endingPeriod)
			liquidPower = liquidPower.Add(del.GetShares())
			totalRewards = totalRewards.Add(delReward.AmountOf(bondDenom))
			return false
		},
	)

	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			// use Balance(slashing applied) not InitialBalance(without slashing)
			unbondingPower = unbondingPower.Add(entry.Balance)
		}
	}

	fmt.Println("[balance, liquidPower, totalRewards, unbondingPower]", balance, liquidPower, totalRewards, unbondingPower)
	return balance.Amount.ToDec().Add(liquidPower).Add(totalRewards).Add(unbondingPower.ToDec())
}

func (k Keeper) LiquidDelegate(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.LiquidValidators, stakingAmt sdk.Int) (newShares sdk.Dec, err error) {
	totalNewShares := sdk.ZeroDec()
	targetMap := types.AddStakingTargetMap(activeVals, stakingAmt)
	for valStr, amt := range targetMap {
		val, err := sdk.ValAddressFromBech32(valStr)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, val)
		if !found {
			panic("validator not founded")
		}
		// TODO: consider checking active val
		newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, amt, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		// TODO: consider kv optimize
		liquidVal, found := k.GetLiquidValidator(ctx, val)
		if !found {
			panic("liquid validator not founded")
		}
		liquidVal.LiquidTokens = liquidVal.LiquidTokens.Add(amt)
		k.SetLiquidValidator(ctx, liquidVal)
		totalNewShares = totalNewShares.Add(newShares)
	}
	return totalNewShares, nil
}

// LiquidStaking ...
func (k Keeper) LiquidStaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, stakingCoin sdk.Coin) (newShares sdk.Dec, btokenMintAmount sdk.Int, err error) {

	// check bond denomination
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if stakingCoin.Denom != bondDenom {
		return sdk.ZeroDec(), btokenMintAmount, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", stakingCoin.Denom, bondDenom,
		)
	}

	activeVals := k.GetActiveLiquidValidators(ctx)
	if activeVals.Len() == 0 || !activeVals.TotalWeight().IsPositive() {
		return sdk.ZeroDec(), btokenMintAmount, fmt.Errorf("there's no active liquid validators")
	}

	netAmount := k.NetAmount(ctx)

	// send staking coin to liquid staking proxy account to proxy delegation
	err = k.bankKeeper.SendCoins(ctx, liquidStaker, proxyAcc, sdk.NewCoins(stakingCoin))
	if err != nil {
		return sdk.ZeroDec(), btokenMintAmount, err
	}

	// mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
	liquidBondDenom := k.LiquidBondDenom(ctx)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom)
	btokenMintAmount = stakingCoin.Amount
	if bTokenTotalSupply.IsPositive() {
		btokenMintAmount = bTokenTotalSupply.Amount.ToDec().Mul(stakingCoin.Amount.ToDec()).QuoTruncate(netAmount).TruncateInt()
	}

	// mint on module acc and send
	mintCoin := sdk.NewCoins(sdk.NewCoin(liquidBondDenom, btokenMintAmount))
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoin)
	if err != nil {
		return sdk.ZeroDec(), btokenMintAmount, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStaker, mintCoin)
	if err != nil {
		return sdk.ZeroDec(), btokenMintAmount, err
	}
	newShares, err = k.LiquidDelegate(ctx, proxyAcc, activeVals, stakingCoin.Amount)
	return newShares, btokenMintAmount, err
}

// LiquidUnstaking ...
func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, amount sdk.Coin,
) (time.Time, sdk.Dec, []stakingtypes.UnbondingDelegation, error) {

	// check bond denomination
	params := k.GetParams(ctx)
	liquidBondDenom := k.LiquidBondDenom(ctx)
	if amount.Denom != liquidBondDenom {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", amount.Denom, liquidBondDenom,
		)
	}

	activeVals := k.GetActiveLiquidValidators(ctx)
	if activeVals.Len() == 0 || !activeVals.TotalWeight().IsPositive() {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("there's no active liquid validators")
	}

	// UnstakeAmount = NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate), review decimal truncation
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, liquidBondDenom)
	if !bTokenTotalSupply.IsPositive() {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("DefaultLiquidBondDenom supply is not positive")
	}
	amountDec := amount.Amount.ToDec()
	netAmount := k.NetAmount(ctx)
	unbondingAmount := netAmount.Mul(amountDec.QuoTruncate(bTokenTotalSupply.Amount.ToDec())).Mul(sdk.OneDec().Sub(params.UnstakeFeeRate)).TruncateDec()

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidStaker, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(liquidBondDenom, amount.Amount)))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}

	share := unbondingAmount.QuoInt(activeVals.TotalWeight()).TruncateInt()
	leftAmount := unbondingAmount.TruncateInt()
	var weightedShare sdk.Int

	var ubdTime time.Time
	var ubds []stakingtypes.UnbondingDelegation
	lenActiveVals := activeVals.Len()
	for i := lenActiveVals - 1; i >= 0; i-- {
		if i == 0 {
			// To minimize the decimal error, use the remaining amount for the last validator.
			weightedShare = leftAmount
		} else {
			weightedShare = share.Mul(activeVals[i].Weight)
		}
		var ubd stakingtypes.UnbondingDelegation
		del, found := k.stakingKeeper.GetDelegation(ctx, proxyAcc, activeVals[i].GetOperator())
		fmt.Println("[liquid UBD]", weightedShare.String(), del.Shares.String(), found)
		ubdTime, ubd, err = k.LiquidUnbond(ctx, proxyAcc, liquidStaker, activeVals[i].GetOperator(), weightedShare.ToDec())
		if err != nil {
			return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
		}
		ubds = append(ubds, ubd)
		activeVals[i].LiquidTokens = activeVals[i].LiquidTokens.Sub(weightedShare)
		k.SetLiquidValidator(ctx, activeVals[i])
		leftAmount = leftAmount.Sub(weightedShare)
	}
	return ubdTime, unbondingAmount, ubds, nil
}

// LiquidUnbond ...
func (k Keeper) LiquidUnbond(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
) (time.Time, stakingtypes.UnbondingDelegation, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, proxyAcc, valAddr, sharesAmount)
	if err != nil {
		return time.Time{}, stakingtypes.UnbondingDelegation{}, err
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

	return completionTime, ubd, nil
}

func (k Keeper) WithdrawLiquidRewards(ctx sdk.Context, proxyAcc sdk.AccAddress) (totalRewards sdk.Int) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	k.stakingKeeper.IterateDelegations(
		ctx, proxyAcc,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			reward, err := k.distrKeeper.WithdrawDelegationRewards(ctx, proxyAcc, valAddr)
			if err != nil {
				// TODO: tmp panic for debugging
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

// GetAllLiquidValidators get the set of all liquid validators with no limits, used during genesis dump
func (k Keeper) GetAllLiquidValidators(ctx sdk.Context) (vals []types.LiquidValidator) {
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
// TODO: refactor []types.LiquidValidator for types.LiquidValidators for totalWeights and len and minMaxGap
func (k Keeper) GetActiveLiquidValidators(ctx sdk.Context) (vals types.LiquidValidators) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	//totalWeight = sdk.ZeroInt()
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		if val.Status == types.ValidatorStatusActive {
			vals = append(vals, val)
			//totalWeight = totalWeight.Add(val.Weight)
		}
	}

	////lenVals = len(vals)
	//if lenVals == 0 || !totalWeight.IsPositive() {
	//	// TODO: make a error type for this
	//	err = fmt.Errorf("there's no active liquid validators")
	//}

	return vals
}

// GetAllLiquidValidatorsMap get the set of all liquid validators as map with no limits
func (k Keeper) GetAllLiquidValidatorsMap(ctx sdk.Context) map[string]types.LiquidValidator {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	defer iterator.Close()

	valsMap := make(map[string]types.LiquidValidator)
	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		valsMap[val.OperatorAddress] = val
	}

	return valsMap
}

//// CollectBiquidStakings collects all the valid liquidStakings registered in params.BiquidStakings and
//// distributes the total collected coins to destination address.
//func (k Keeper) CollectBiquidStakings(ctx sdk.Context) error {
//	params := k.GetParams(ctx)
//	var liquidStakings []types.BiquidStaking
//	if params.EpochBlocks > 0 && ctx.BlockHeight()%int64(params.EpochBlocks) == 0 {
//		liquidStakings = types.CollectibleBiquidStakings(params.BiquidStakings, ctx.BlockTime())
//	}
//	if len(liquidStakings) == 0 {
//		return nil
//	}
//
//	// Get a map GetBiquidStakingsBySourceMap that has a list of liquidStakings and their total rate, which
//	// contain the same SourceAddress
//	liquidStakingsBySourceMap := types.GetBiquidStakingsBySourceMap(liquidStakings)
//	for source, liquidStakingsBySource := range liquidStakingsBySourceMap {
//		sourceAcc, err := sdk.AccAddressFromBech32(source)
//		if err != nil {
//			return err
//		}
//		sourceBalances := sdk.NewDecCoinsFromCoins(k.bankKeeper.GetAllBalances(ctx, sourceAcc)...)
//		if sourceBalances.IsZero() {
//			continue
//		}
//
//		var inputs []banktypes.Input
//		var outputs []banktypes.Output
//		liquidStakingsBySource.CollectionCoins = make([]sdk.Coins, len(liquidStakingsBySource.BiquidStakings))
//		for i, liquidStaking := range liquidStakingsBySource.BiquidStakings {
//			destinationAcc, err := sdk.AccAddressFromBech32(liquidStaking.DestinationAddress)
//			if err != nil {
//				return err
//			}
//
//			collectionCoins, _ := sourceBalances.MulDecTruncate(liquidStaking.Rate).TruncateDecimal()
//			if collectionCoins.Empty() || !collectionCoins.IsValid() {
//				continue
//			}
//
//			inputs = append(inputs, banktypes.NewInput(sourceAcc, collectionCoins))
//			outputs = append(outputs, banktypes.NewOutput(destinationAcc, collectionCoins))
//			liquidStakingsBySource.CollectionCoins[i] = collectionCoins
//		}
//
//		if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
//			return err
//		}
//
//		for i, liquidStaking := range liquidStakingsBySource.BiquidStakings {
//			k.AddTotalCollectedCoins(ctx, liquidStaking.Name, liquidStakingsBySource.CollectionCoins[i])
//			ctx.EventManager().EmitEvents(sdk.Events{
//				sdk.NewEvent(
//					types.EventTypeBiquidStakingCollected,
//					sdk.NewAttribute(types.AttributeValueName, liquidStaking.Name),
//					sdk.NewAttribute(types.AttributeValueDestinationAddress, liquidStaking.DestinationAddress),
//					sdk.NewAttribute(types.AttributeValueSourceAddress, liquidStaking.SourceAddress),
//					sdk.NewAttribute(types.AttributeValueRate, liquidStaking.Rate.String()),
//					sdk.NewAttribute(types.AttributeValueAmount, liquidStakingsBySource.CollectionCoins[i].String()),
//				),
//			})
//		}
//	}
//	return nil
//}
//
//// GetTotalCollectedCoins returns total collected coins for a liquidstaking.
//func (k Keeper) GetTotalCollectedCoins(ctx sdk.Context, liquidStakingName string) sdk.Coins {
//	store := ctx.KVStore(k.storeKey)
//	bz := store.Get(types.GetTotalCollectedCoinsKey(liquidStakingName))
//	if bz == nil {
//		return nil
//	}
//	var collectedCoins types.TotalCollectedCoins
//	k.cdc.MustUnmarshal(bz, &collectedCoins)
//	return collectedCoins.TotalCollectedCoins
//}
//
//// IterateAllTotalCollectedCoins iterates over all the stored TotalCollectedCoins and performs a callback function.
//// Stops iteration when callback returns true.
//func (k Keeper) IterateAllTotalCollectedCoins(ctx sdk.Context, cb func(record types.BiquidStakingRecord) (stop bool)) {
//	store := ctx.KVStore(k.storeKey)
//	iterator := sdk.KVStorePrefixIterator(store, types.TotalCollectedCoinsKeyPrefix)
//
//	defer iterator.Close()
//	for ; iterator.Valid(); iterator.Next() {
//		var record types.BiquidStakingRecord
//		var collectedCoins types.TotalCollectedCoins
//		k.cdc.MustUnmarshal(iterator.Value(), &collectedCoins)
//		record.Name = types.ParseTotalCollectedCoinsKey(iterator.Key())
//		record.TotalCollectedCoins = collectedCoins.TotalCollectedCoins
//		if cb(record) {
//			break
//		}
//	}
//}
//
//// SetTotalCollectedCoins sets total collected coins for a liquidstaking.
//func (k Keeper) SetTotalCollectedCoins(ctx sdk.Context, liquidStakingName string, amount sdk.Coins) {
//	store := ctx.KVStore(k.storeKey)
//	collectedCoins := types.TotalCollectedCoins{TotalCollectedCoins: amount}
//	bz := k.cdc.MustMarshal(&collectedCoins)
//	store.Set(types.GetTotalCollectedCoinsKey(liquidStakingName), bz)
//}
//
//// AddTotalCollectedCoins increases total collected coins for a liquidstaking.
//func (k Keeper) AddTotalCollectedCoins(ctx sdk.Context, liquidStakingName string, amount sdk.Coins) {
//	collectedCoins := k.GetTotalCollectedCoins(ctx, liquidStakingName)
//	collectedCoins = collectedCoins.Add(amount...)
//	k.SetTotalCollectedCoins(ctx, liquidStakingName, collectedCoins)
//}
