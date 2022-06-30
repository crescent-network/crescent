package keeper

import (
	"fmt"
	"strconv"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

var (
	enableAdvanceEpoch = "false" // Set this to "true" using build flags to enable AdvanceEpoch msg handling.
	enableRatioPlan    = "false" // Set this to "true" using build flags to enable creation of RatioPlans.

	// EnableAdvanceEpoch indicates whether msgServer accepts MsgAdvanceEpoch or not.
	// Never set this to true in production mode. Doing that will expose serious attack vector.
	EnableAdvanceEpoch = false
	// EnableRatioPlan indicates whether msgServer and proposal handler accept
	// creation of RatioPlans.
	// Default is false, which means that RatioPlans can't be created through a
	// MsgCreateRatioPlan msg and a PublicPlanProposal.
	EnableRatioPlan = false
)

func init() {
	var err error
	EnableAdvanceEpoch, err = strconv.ParseBool(enableAdvanceEpoch)
	if err != nil {
		panic(err)
	}
	EnableRatioPlan, err = strconv.ParseBool(enableRatioPlan)
	if err != nil {
		panic(err)
	}
}

// Keeper of the farming store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
}

// NewKeeper returns a farming keeper. It handles:
// - creating new ModuleAccounts for each pool ReserveAccount
// - sending to and from ModuleAccounts
// - minting, burning PoolCoins
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper,
) Keeper {
	// ensure farming module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

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
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetParams returns the parameters for the farming module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the farming module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetCodec returns codec.Codec object used by the keeper>
func (k Keeper) GetCodec() codec.BinaryCodec { return k.cdc }
