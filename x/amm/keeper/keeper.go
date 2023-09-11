package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// Keeper of the module's store.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	exchangeKeeper types.ExchangeKeeper
	markerKeeper   types.MarkerKeeper
}

// NewKeeper creates a new Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSpace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	exchangeKeeper types.ExchangeKeeper,
	markerKeeper types.MarkerKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     paramSpace,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		exchangeKeeper: exchangeKeeper,
		markerKeeper:   markerKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
