package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

// InitGenesis initializes the budget module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := types.ValidateGenesis(genState); err != nil {
		panic(err)
	}
	// init to prevent nil slice, []types.Budget(nil)
	if genState.Params.Budgets == nil || len(genState.Params.Budgets) == 0 {
		genState.Params.Budgets = []types.Budget{}
	}

	k.SetParams(ctx, genState.Params)
	moduleAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	for _, record := range genState.BudgetRecords {
		k.SetTotalCollectedCoins(ctx, record.Name, record.TotalCollectedCoins)
	}
}

// ExportGenesis returns the budget module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	// init to prevent nil slice, []types.Budget(nil)
	if params.Budgets == nil || len(params.Budgets) == 0 {
		params.Budgets = []types.Budget{}
	}

	budgetRecords := make([]types.BudgetRecord, 0)
	k.IterateAllTotalCollectedCoins(ctx, func(record types.BudgetRecord) (stop bool) {
		budgetRecords = append(budgetRecords, record)
		return false
	})

	return types.NewGenesisState(params, budgetRecords)
}
