package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/farming/x/farming/types"
)

// Keeper of the farming store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	distrKeeper   types.DistributionKeeper

	blockedAddrs map[string]bool
}

// NewKeeper returns a farming keeper. It handles:
// - creating new ModuleAccounts for each pool ReserveAccount
// - sending to and from ModuleAccounts
// - minting, burning PoolCoins
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, distrKeeper types.DistributionKeeper,
	blockedAddrs map[string]bool,
) Keeper {
	// TODO: TBD module account for farming
	//// ensure farming module account is set
	//if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	//}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		blockedAddrs:  blockedAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetParams gets the parameters for the farming module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the farming module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetNextPlanIDWithUpdate returns and increments the global Plan ID counter.
// If the global plan number is not set, it initializes it with value 1.
func (k Keeper) GetNextPlanIDWithUpdate(ctx sdk.Context) uint64 {
	var id uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GlobalFarmingPlanIDKey)
	if bz == nil {
		// initialize the PlanId
		id = 1
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

		id = val.GetValue()
	}
	bz = k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id + 1})
	store.Set(types.GlobalFarmingPlanIDKey, bz)
	return id
}

func (k Keeper) decodePlan(bz []byte) types.PlanI {
	acc, err := k.UnmarshalPlan(bz)
	if err != nil {
		panic(err)
	}

	return acc
}

// MarshalPlan protobuf serializes an Plan interface
func (k Keeper) MarshalPlan(plan types.PlanI) ([]byte, error) { // nolint:interfacer
	return k.cdc.MarshalInterface(plan)
}

// UnmarshalPlan returns an Plan interface from raw encoded plan
// bytes of a Proto-based Plan type
func (k Keeper) UnmarshalPlan(bz []byte) (types.PlanI, error) {
	var acc types.PlanI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

// GetCodec return codec.Codec object used by the keeper
func (k Keeper) GetCodec() codec.BinaryCodec { return k.cdc }
