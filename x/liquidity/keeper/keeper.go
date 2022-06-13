package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/crescent-network/crescent/x/liquidity/types"
)

// Keeper of the liquidity store.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new liquidity Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSpace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the parameters for the liquidity module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

// SetParams sets the parameters for the liquidity module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetMaxPriceLimitRatio(ctx sdk.Context) (ratio sdk.Dec) {
	k.paramSpace.Get(ctx, types.KeyMaxPriceLimitRatio, &ratio)
	return
}

func (k Keeper) GetTickPrecision(ctx sdk.Context) (tickPrec uint32) {
	k.paramSpace.Get(ctx, types.KeyTickPrecision, &tickPrec)
	return
}

func (k Keeper) GetDustCollector(ctx sdk.Context) sdk.AccAddress {
	var dustCollectorAddr string
	k.paramSpace.Get(ctx, types.KeyDustCollectorAddress, &dustCollectorAddr)
	addr, err := sdk.AccAddressFromBech32(dustCollectorAddr)
	if err != nil {
		panic(err)
	}
	return addr
}
