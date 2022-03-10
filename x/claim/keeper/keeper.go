package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmosquad-labs/squad/x/claim/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        sdk.StoreKey
	bankKeeper      types.BankKeeper
	distrKeeper     types.DistrKeeper
	govKeeper       types.GovKeeper
	liquidityKeeper types.LiquidityKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	bk types.BankKeeper,
	dk types.DistrKeeper,
	gk types.GovKeeper,
	lk types.LiquidityKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		bankKeeper:      bk,
		distrKeeper:     dk,
		govKeeper:       gk,
		liquidityKeeper: lk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
