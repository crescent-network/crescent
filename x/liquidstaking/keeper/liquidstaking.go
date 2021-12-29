package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/farming/x/liquidstaking/types"
)

func (k Keeper) NetAmountTBD(ctx sdk.Context) sdk.Dec {
	// TODO: delegation amount + unbonding amount + reward amount for the liquid staking module account
	// balance of types.LiquidStakingProxyAcc, remaining reward, unbonding amount
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	balance := k.bankKeeper.GetBalance(ctx, types.LiquidStakingProxyAcc, bondDenom)
	ubds := k.stakingKeeper.GetAllUnbondingDelegations(ctx, types.LiquidStakingProxyAcc)
	liquidPower := sdk.ZeroDec()
	unbondingPower := sdk.ZeroInt()
	totalRewards := sdk.ZeroDec()
	//var delRewards []distrtypes.DelegationDelegatorReward

	cachedCtx, _ := ctx.CacheContext()
	k.stakingKeeper.IterateDelegations(
		// TODO: cache ctx
		cachedCtx, types.LiquidStakingProxyAcc,
		func(_ int64, del stakingtypes.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(cachedCtx, valAddr)
			endingPeriod := k.distrKeeper.IncrementValidatorPeriod(cachedCtx, val)
			delReward := k.distrKeeper.CalculateDelegationRewards(cachedCtx, val, del, endingPeriod)

			//delRewards = append(delRewards, distrtypes.NewDelegationDelegatorReward(valAddr, delReward))
			liquidPower = liquidPower.Add(del.GetShares())
			totalRewards = totalRewards.Add(delReward.AmountOf(bondDenom))
			return false
		},
	)

	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			// TODO: Balance or InitialBalance(without slashing)
			unbondingPower = unbondingPower.Add(entry.Balance)
		}
	}

	// TODO: unbonding power is on delegation? no
	// TODO: decimal handling
	fmt.Println("[balance, liquidPower, totalRewards, unbondingPower]", balance, liquidPower, totalRewards, unbondingPower)
	return balance.Amount.ToDec().Add(liquidPower).Add(totalRewards).Add(unbondingPower.ToDec())
}

func (k Keeper) NetAmount(ctx sdk.Context) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, types.LiquidBondDenom).Amount
}

// LiquidStaking ...
// TODO: distribute activeValidators or make upper level function
func (k Keeper) LiquidStaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, stakingCoin sdk.Coin) (newShares sdk.Dec, err error) {

	netAmount := k.NetAmountTBD(ctx)
	//netAmount := k.NetAmount(ctx).ToDec()
	// send staking coin to liquid staking proxy account to proxy delegation
	err = k.bankKeeper.SendCoins(ctx, liquidStaker, proxyAcc, sdk.NewCoins(stakingCoin))
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// TODO: mint btoken, MintAmount = TotalSupply * StakeAmount/NetAmount
	// TODO: types.BtokenDenom to be params.LiquidBondDenom
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, types.LiquidBondDenom)
	mintAmt := stakingCoin.Amount
	stakingAmt := stakingCoin.Amount.ToDec()
	if bTokenTotalSupply.IsPositive() {
		// TODO: review decimal issue
		mintAmt = bTokenTotalSupply.Amount.ToDec().Mul(stakingAmt).QuoTruncate(netAmount).TruncateInt()
	}
	fmt.Println("[NetAmount, NetAmountTBD, mint]", k.NetAmount(ctx), netAmount, mintAmt, stakingAmt, bTokenTotalSupply)
	// TODO: mint on module and send
	mintCoin := sdk.NewCoins(sdk.NewCoin(types.LiquidBondDenom, mintAmt))
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, mintCoin)
	if err != nil {
		// TODO: make custom err
		return sdk.ZeroDec(), err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStaker, mintCoin)
	if err != nil {
		// TODO: make custom err
		return sdk.ZeroDec(), err
	}

	// check bond denomination
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if stakingCoin.Denom != bondDenom {
		return sdk.ZeroDec(), sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", stakingCoin.Denom, bondDenom,
		)
	}

	// TODO: rebalancing and sum validation, total should be same with decimal error correction
	activeVals := k.GetActiveLiquidValidators(ctx)
	share := stakingAmt.QuoTruncate(sdk.NewDec(int64(len(activeVals)))).TruncateInt()
	totalNewShares := sdk.ZeroDec()
	for _, val := range activeVals {
		// TODO: a validator to whitelisted validator list with weight
		// NOTE: source funds are always unbonded
		validator, found := k.stakingKeeper.GetValidator(ctx, val.GetOperator())
		if !found {
			panic("validator not founded")
		}
		newShares, err := k.stakingKeeper.Delegate(ctx, proxyAcc, share, stakingtypes.Unbonded, validator, true)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		totalNewShares = totalNewShares.Add(newShares)
	}

	return totalNewShares, nil
}

// LiquidUnstaking ...
// TODO: distribute activeValidators or make upper level function
func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, amount sdk.Coin,
) (time.Time, []stakingtypes.UnbondingDelegation, error) {

	// TODO: UnstakeAmount = NetAmount * BTokenAmount/TotalSupply * (1-UnstakeFeeRate), review decimal truncation
	params := k.GetParams(ctx)
	// TODO: handle zero supply
	bTokenTotalSupply := k.bankKeeper.GetSupply(ctx, types.LiquidBondDenom)
	if !bTokenTotalSupply.IsPositive() {
		return time.Time{}, []stakingtypes.UnbondingDelegation{}, fmt.Errorf("LiquidBondDenom supply is not positive")
	}
	amountDec := amount.Amount.ToDec()
	netAmount := k.NetAmountTBD(ctx)
	//netAmount := k.NetAmount(ctx).ToDec()
	unstakeAmount := netAmount.Mul(amountDec.QuoTruncate(bTokenTotalSupply.Amount.ToDec())).Mul(sdk.OneDec().Sub(params.UnstakeFeeRate)).TruncateDec()
	fmt.Println(unstakeAmount)
	fmt.Println("[NetAmount, NetAmountTBD]", k.NetAmount(ctx), k.NetAmountTBD(ctx))

	// TODO: burn or reserve queue for burning btoken, is unstake make reduce power immediately?
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidStaker, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		// TODO: make custom err
		return time.Time{}, []stakingtypes.UnbondingDelegation{}, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.LiquidBondDenom, amount.Amount)))
	if err != nil {
		// TODO: make custom err
		return time.Time{}, []stakingtypes.UnbondingDelegation{}, err
	}

	activeVals := k.GetActiveLiquidValidators(ctx)
	// TODO: Get 1/n power len(activeVals) with rebalancing, checking sum of shares under total unstakeAmount
	share := unstakeAmount.QuoTruncate(sdk.NewDec(int64(len(activeVals))))
	var ubdTime time.Time
	var ubds []stakingtypes.UnbondingDelegation
	for _, val := range activeVals {
		var ubd stakingtypes.UnbondingDelegation
		ubdTime, ubd, err = k.LiquidUnbond(ctx, proxyAcc, liquidStaker, val.GetOperator(), share)
		if err != nil {
			// TODO: should be revertable, only in msg_server
			panic(err)
		}
		ubds = append(ubds, ubd)
	}
	return ubdTime, ubds, nil
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
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins); err != nil {
			panic(err)
		}
	}

	completionTime := ctx.BlockHeader().Time.Add(k.stakingKeeper.UnbondingTime(ctx))
	ubd := k.stakingKeeper.SetUnbondingDelegationEntry(ctx, liquidStaker, valAddr, ctx.BlockHeight(), completionTime, returnAmount)
	k.stakingKeeper.InsertUBDQueue(ctx, ubd, completionTime)

	return completionTime, ubd, nil
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
func (k Keeper) GetActiveLiquidValidators(ctx sdk.Context) (vals []types.LiquidValidator) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LiquidValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.MustUnmarshalLiquidValidator(k.cdc, iterator.Value())
		if val.Status == types.ValidatorStatusWhiteListed {
			vals = append(vals, val)
		}
	}

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
