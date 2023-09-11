package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

// CollectBudgets collects all the valid budgets registered in params.Budgets and
// distributes the total collected coins to destination address.
func (k Keeper) CollectBudgets(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	var budgets []types.Budget
	if params.EpochBlocks > 0 && ctx.BlockHeight()%int64(params.EpochBlocks) == 0 {
		budgets = types.CollectibleBudgets(params.Budgets, ctx.BlockTime())
	}
	if len(budgets) == 0 {
		return nil
	}

	// Get a map GetBudgetsBySourceMap that has a list of budgets and their total rate, which
	// contain the same SourceAddress
	budgetsBySourceMap, budgetSources := types.GetBudgetsBySourceMap(budgets)
	for _, source := range budgetSources {
		budgetsBySource := budgetsBySourceMap[source]
		sourceAcc, err := sdk.AccAddressFromBech32(source)
		if err != nil {
			return err
		}
		sourceBalances := sdk.NewDecCoinsFromCoins(k.bankKeeper.SpendableCoins(ctx, sourceAcc)...)
		if sourceBalances.IsZero() {
			continue
		}

		var inputs []banktypes.Input
		var outputs []banktypes.Output
		budgetsBySource.CollectionCoins = make([]sdk.Coins, len(budgetsBySource.Budgets))
		for i, budget := range budgetsBySource.Budgets {
			destinationAcc, err := sdk.AccAddressFromBech32(budget.DestinationAddress)
			if err != nil {
				return err
			}
			collectionCoins, _ := sourceBalances.MulDecTruncate(budget.Rate).TruncateDecimal()
			if collectionCoins.Empty() || collectionCoins.IsZero() || !collectionCoins.IsValid() {
				continue
			}

			inputs = append(inputs, banktypes.NewInput(sourceAcc, collectionCoins))
			outputs = append(outputs, banktypes.NewOutput(destinationAcc, collectionCoins))
			budgetsBySource.CollectionCoins[i] = collectionCoins
		}

		if err = k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return err
		}

		for i, budget := range budgetsBySource.Budgets {
			if budgetsBySource.CollectionCoins[i].Empty() || budgetsBySource.CollectionCoins[i].IsZero() || !budgetsBySource.CollectionCoins[i].IsValid() {
				continue
			}
			k.AddTotalCollectedCoins(ctx, budget.Name, budgetsBySource.CollectionCoins[i])
			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeBudgetCollected,
					sdk.NewAttribute(types.AttributeValueName, budget.Name),
					sdk.NewAttribute(types.AttributeValueAmount, budgetsBySource.CollectionCoins[i].String()),
				),
			})
		}
	}
	return nil
}

// GetTotalCollectedCoins returns total collected coins for a budget.
func (k Keeper) GetTotalCollectedCoins(ctx sdk.Context, budgetName string) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalCollectedCoinsKey(budgetName))
	if bz == nil {
		return nil
	}
	var collectedCoins types.TotalCollectedCoins
	k.cdc.MustUnmarshal(bz, &collectedCoins)
	return collectedCoins.TotalCollectedCoins
}

// IterateAllTotalCollectedCoins iterates over all the stored TotalCollectedCoins and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateAllTotalCollectedCoins(ctx sdk.Context, cb func(record types.BudgetRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.TotalCollectedCoinsKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.BudgetRecord
		var collectedCoins types.TotalCollectedCoins
		k.cdc.MustUnmarshal(iterator.Value(), &collectedCoins)
		record.Name = types.ParseTotalCollectedCoinsKey(iterator.Key())
		record.TotalCollectedCoins = collectedCoins.TotalCollectedCoins
		if cb(record) {
			break
		}
	}
}

// SetTotalCollectedCoins sets total collected coins for a budget.
func (k Keeper) SetTotalCollectedCoins(ctx sdk.Context, budgetName string, amount sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	collectedCoins := types.TotalCollectedCoins{TotalCollectedCoins: amount}
	bz := k.cdc.MustMarshal(&collectedCoins)
	store.Set(types.GetTotalCollectedCoinsKey(budgetName), bz)
}

// AddTotalCollectedCoins increases total collected coins for a budget.
func (k Keeper) AddTotalCollectedCoins(ctx sdk.Context, budgetName string, amount sdk.Coins) {
	collectedCoins := k.GetTotalCollectedCoins(ctx, budgetName)
	collectedCoins = collectedCoins.Add(amount...)
	k.SetTotalCollectedCoins(ctx, budgetName, collectedCoins)
}
