package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

// InitGenesis initializes the liquidstaking module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := types.ValidateGenesis(genState); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)
	moduleAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	//for _, record := range genState.BiquidStakingRecords {
	//	k.SetTotalCollectedCoins(ctx, record.Name, record.TotalCollectedCoins)
	//}
}

// ExportGenesis returns the liquidstaking module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	var liquidValidators []types.LiquidValidator

	//k.IterateAllTotalCollectedCoins(ctx, func(record types.BiquidStakingRecord) (stop bool) {
	//	liquidStakingRecords = append(liquidStakingRecords, record)
	//	return false
	//})

	return types.NewGenesisState(params, liquidValidators)
}
