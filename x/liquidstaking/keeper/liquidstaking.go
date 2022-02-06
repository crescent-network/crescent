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

func (k Keeper) CheckRewardsAndDelShares(ctx sdk.Context, proxyAcc sdk.AccAddress) (sdk.Dec, sdk.Dec) {
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	totalDelShares := sdk.ZeroDec()
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
			totalDelShares = totalDelShares.Add(del.GetShares())
			totalRewards = totalRewards.Add(delReward.AmountOf(bondDenom))
			return false
		},
	)

	return totalRewards, totalDelShares
}

func (k Keeper) NetAmount(ctx sdk.Context) sdk.Dec {
	// delegation power, bondDenom balance, remaining reward, unbonding amount of types.LiquidStakingProxyAcc
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	balance := k.bankKeeper.GetBalance(ctx, types.LiquidStakingProxyAcc, bondDenom)
	ubds := k.stakingKeeper.GetAllUnbondingDelegations(ctx, types.LiquidStakingProxyAcc)
	unbondingPower := sdk.ZeroInt()
	// TODO: Add invariant checking for equal CheckRewardsAndDelShares with sum of GetDelShares
	totalRewards, totalDelShares := k.CheckRewardsAndDelShares(ctx, types.LiquidStakingProxyAcc)

	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			// use Balance(slashing applied) not InitialBalance(without slashing)
			unbondingPower = unbondingPower.Add(entry.Balance)
		}
	}

	fmt.Println("[balance, totalDelShares, totalRewards, unbondingPower]", balance, totalDelShares, totalRewards, unbondingPower)
	return balance.Amount.ToDec().Add(totalDelShares).Add(totalRewards).Add(unbondingPower.ToDec())
}

func (k Keeper) LiquidDelegate(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.LiquidValidators, stakingAmt sdk.Int, whitelistedValMap types.WhitelistedValMap) (newShares sdk.Dec, err error) {
	totalNewShares := sdk.ZeroDec()
	// crumb may occur due to a decimal point error in dividing the staking amount into the weight of liquid validators, which is also accumulated in the netAmount value
	weightedShares, _ := types.DivideByWeight(activeVals, stakingAmt, whitelistedValMap)
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

// Deprecated: LiquidStakingWithBalancing for using simple weight distribution, not rebalancing, not using on this version for simplify.
func (k Keeper) LiquidStakingWithBalancing(ctx sdk.Context, proxyAcc sdk.AccAddress, activeVals types.LiquidValidators, stakingAmt sdk.Int) (newShares sdk.Dec, err error) {
	totalNewShares := sdk.ZeroDec()
	targetMap := k.AddStakingTargetMap(ctx, activeVals, stakingAmt)
	for valStr, amt := range targetMap {
		val, err := sdk.ValAddressFromBech32(valStr)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		validator, found := k.stakingKeeper.GetValidator(ctx, val)
		if !found {
			panic("validator not founded")
		}
		newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, amt, stakingtypes.Unbonded, validator, true)
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
	valMap := k.GetValidatorsMap(ctx)
	activeVals := k.GetActiveLiquidValidators(ctx, valMap, whitelistedValMap)
	if activeVals.Len() == 0 || !activeVals.TotalWeight(whitelistedValMap).IsPositive() {
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
	// TODO: need to re-calc mintAmount for actual delegation amount using sum of divided with truncation
	newShares, err = k.LiquidDelegate(ctx, proxyAcc, activeVals, stakingCoin.Amount, whitelistedValMap)
	return newShares, bTokenMintAmount, err
}

// LiquidUnstaking ...
func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, amount sdk.Coin,
) (time.Time, sdk.Dec, []stakingtypes.UnbondingDelegation, error) {

	// check bond denomination
	params := k.GetParams(ctx)
	bondedBondDenom := k.BondedBondDenom(ctx)
	if amount.Denom != bondedBondDenom {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", amount.Denom, bondedBondDenom,
		)
	}

	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	valMap := k.GetValidatorsMap(ctx)
	activeVals := k.GetActiveLiquidValidators(ctx, valMap, whitelistedValMap)
	if activeVals.Len() == 0 || !activeVals.TotalWeight(whitelistedValMap).IsPositive() {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("there's no active liquid validators")
	}

	// UnstakeAmount = NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate)
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, bondedBondDenom)
	if !bTokenTotalSupply.IsPositive() {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, fmt.Errorf("DefaultBondedBondDenom supply is not positive")
	}
	// TODO: last unstaking if bTokenTotalSupply == amount {}
	netAmount := k.NetAmount(ctx)
	unbondingAmount := types.BTokenToNativeToken(amount.Amount, bTokenTotalSupply.Amount, netAmount, params.UnstakeFeeRate)
	totalReturnAmount := sdk.ZeroInt()

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidStaker, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(bondedBondDenom, amount.Amount)))
	if err != nil {
		return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
	}

	// crumb may occur due to a decimal error in dividing the amount into the weight of liquid validators, which is also accumulated in the netAmount value
	unbondingShares, _ := k.DivideByCurrentWeight(ctx, activeVals, unbondingAmount)
	var ubdTime time.Time
	var ubds []stakingtypes.UnbondingDelegation
	for i, val := range activeVals {
		weightedShare := unbondingShares[i]
		var ubd stakingtypes.UnbondingDelegation
		var returnAmount sdk.Int
		del, found := k.stakingKeeper.GetDelegation(ctx, proxyAcc, val.GetOperator())
		weightedShare, err = k.stakingKeeper.ValidateUnbondAmount(ctx, proxyAcc, val.GetOperator(), weightedShare.TruncateInt())
		fmt.Println("[liquid UBD]", weightedShare.String(), del.Shares.String(), found, unbondingShares[i], activeVals.Len(), unbondingAmount.String())
		if err != nil {
			return time.Time{}, sdk.ZeroDec(), []stakingtypes.UnbondingDelegation{}, err
		}
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
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
) (time.Time, sdk.Int, stakingtypes.UnbondingDelegation, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, sdk.ZeroInt(), stakingtypes.UnbondingDelegation{}, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, proxyAcc, valAddr, sharesAmount)
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
func (k Keeper) GetActiveLiquidValidators(ctx sdk.Context, valMap map[string]stakingtypes.Validator, whitelistedValMap types.WhitelistedValMap) (vals types.LiquidValidators) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		if val.ActiveCondition(valMap[val.OperatorAddress], whitelistedValMap.IsListed(val.OperatorAddress)) {
			vals = append(vals, val)
		}
		//if _, ok := whitelistedValMap[val.OperatorAddress]; ok &&
		//	val.ActiveCondition(valMap[val.OperatorAddress], ok) {
		//	vals = append(vals, val)
		//}
	}
	return vals
}

// GetValidatorsMap get the set of all validators as map with no limits
// TODO: it could optimize to containing only to be used validators
func (k Keeper) GetValidatorsMap(ctx sdk.Context) map[string]stakingtypes.Validator {
	valMap := make(map[string]stakingtypes.Validator)
	vals := k.stakingKeeper.GetAllValidators(ctx)
	for _, val := range vals {
		valMap[val.OperatorAddress] = val
	}
	return valMap
}

func (k Keeper) GetLiquidValidatorStates(ctx sdk.Context) (liquidValidatorStates []types.LiquidValidatorState) {
	lvs := k.GetAllLiquidValidators(ctx)
	valMap := k.GetValidatorsMap(ctx)
	whitelistedValMap := k.GetParams(ctx).WhitelistedValMap()
	for _, lv := range lvs {
		lvState := types.LiquidValidatorState{
			OperatorAddress: lv.OperatorAddress,
			Weight:          lv.GetWeight(whitelistedValMap),
			Status:          lv.GetStatus(valMap[lv.OperatorAddress], whitelistedValMap.IsListed(lv.OperatorAddress)),
			DelShares:       lv.GetDelShares(ctx, k.stakingKeeper),
		}
		liquidValidatorStates = append(liquidValidatorStates, lvState)
	}
	return
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
