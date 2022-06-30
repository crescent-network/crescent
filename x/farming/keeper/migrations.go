package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/crescent-network/crescent/v2/x/farming/legacy/v2"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	currentEpochDays := m.keeper.GetCurrentEpochDays(ctx)

	return v2.MigrateStore(ctx, m.keeper.storeKey, currentEpochDays)
}
