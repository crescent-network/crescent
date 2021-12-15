package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/farming/x/liquidstaking/types"
)

func (k Keeper) LiquidStaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, stakingCoin sdk.Coin,
	validator stakingtypes.Validator) (newShares sdk.Dec, err error) {

	// send staking coin to liquid staking proxy account to proxy delegation
	err = k.bankKeeper.SendCoins(ctx, liquidStaker, proxyAcc, sdk.NewCoins(stakingCoin))
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// check bond denomination
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if stakingCoin.Denom != bondDenom {
		return sdk.ZeroDec(), sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", stakingCoin.Denom, bondDenom,
		)
	}

	// NOTE: source funds are always unbonded
	newShares, err = k.stakingKeeper.Delegate(ctx, proxyAcc, stakingCoin.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return newShares, nil
}

func (k Keeper) LiquidUnstaking(
	ctx sdk.Context, proxyAcc, liquidStaker sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
) (time.Time, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return time.Time{}, stakingtypes.ErrNoDelegatorForAddress
	}

	if k.stakingKeeper.HasMaxUnbondingDelegationEntries(ctx, proxyAcc, valAddr) {
		return time.Time{}, stakingtypes.ErrMaxUnbondingDelegationEntries
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, proxyAcc, valAddr, sharesAmount)
	if err != nil {
		return time.Time{}, err
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

	return completionTime, nil
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

// SetValidator set the main record holding liquid validator details
func (k Keeper) SetValidator(ctx sdk.Context, val types.LiquidValidator) {
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
