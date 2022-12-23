package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/crescent-network/crescent/v4/x/liquidity/legacy/v2"
	v3 "github.com/crescent-network/crescent/v4/x/liquidity/legacy/v3"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}
