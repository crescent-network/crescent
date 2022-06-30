package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

// Keeper of the liquidstaking store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	stakingKeeper   types.StakingKeeper
	distrKeeper     types.DistrKeeper
	liquidityKeeper types.LiquidityKeeper
	farmingKeeper   types.FarmingKeeper
	slashingKeeper  types.SlashingKeeper
}

// NewKeeper returns a liquidstaking keeper. It handles:
// - creating new ModuleAccounts for each pool ReserveAccount
// - sending to and from ModuleAccounts
// - minting, burning PoolCoins
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, stakingKeeper types.StakingKeeper,
	distrKeeper types.DistrKeeper, liquidityKeeper types.LiquidityKeeper,
	farmingKeeper types.FarmingKeeper, slashingKeeper types.SlashingKeeper,
) Keeper {
	// ensure liquidstaking module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:        key,
		cdc:             cdc,
		paramSpace:      paramSpace,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		stakingKeeper:   stakingKeeper,
		distrKeeper:     distrKeeper,
		liquidityKeeper: liquidityKeeper,
		farmingKeeper:   farmingKeeper,
		slashingKeeper:  slashingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetParams gets the parameters for the liquidstaking module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the liquidstaking module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetCodec return codec.Codec object used by the keeper
func (k Keeper) GetCodec() codec.BinaryCodec { return k.cdc }
