package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

// InitGenesis initializes the marketmaker module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := types.ValidateGenesis(genState); err != nil {
		panic(err)
	}

	ctx, writeCache := ctx.CacheContext()

	// init to prevent nil slice, []types.IncentivePairs(nil)
	if genState.Params.IncentivePairs == nil || len(genState.Params.IncentivePairs) == 0 {
		genState.Params.IncentivePairs = []types.IncentivePair{}
	}

	// validations
	if err := k.ValidateDepositReservedAmount(ctx); err != nil {
		panic(err)
	}

	if err := k.ValidateIncentiveReservedAmount(ctx, genState.Incentives); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)

	for _, record := range genState.MarketMakers {
		if err := record.Validate(); err != nil {
			panic(err)
		}
		k.SetMarketMaker(ctx, record)
	}

	for _, record := range genState.Incentives {
		if err := record.Validate(); err != nil {
			panic(err)
		}
		k.SetIncentive(ctx, record)
	}

	for _, record := range genState.DepositRecords {
		if err := record.Validate(); err != nil {
			panic(err)
		}
		k.SetDeposit(ctx, record.GetAccAddress(), record.PairId, record.Amount)
	}

	writeCache()
}

// ExportGenesis returns the marketmaker module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	// init to prevent nil slice, []types.IncentivePairs(nil)
	if params.IncentivePairs == nil || len(params.IncentivePairs) == 0 {
		params.IncentivePairs = []types.IncentivePair{}
	}

	mms := k.GetAllMarketMakers(ctx)
	incentives := k.GetAllIncentives(ctx)
	depositRecords := k.GetAllDepositRecords(ctx)

	return types.NewGenesisState(
		params,
		mms,
		incentives,
		depositRecords,
	)
}
