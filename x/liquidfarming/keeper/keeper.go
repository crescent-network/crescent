package keeper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
)

var (
	// Set this "true" using build flags to enable AdvanceAuction msg handling.
	enableAdvanceAuction = "false"

	// EnableAdvanceAuction indicates whether msgServer accepts MsgAdvanceAuction or not.
	// Setting this true in production mode will expose unexpected vulnerability.
	EnableAdvanceAuction = false
)

func init() {
	var err error
	EnableAdvanceAuction, err = strconv.ParseBool(enableAdvanceAuction)
	if err != nil {
		panic(err)
	}
}

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        sdk.StoreKey
	paramSpace      paramtypes.Subspace
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	farmKeeper      types.FarmKeeper
	liquidityKeeper types.LiquidityKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	farmKeeper types.FarmKeeper,
	liquidityKeeper types.LiquidityKeeper,
) Keeper {
	// Ensure the module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// Set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		paramSpace:      paramSpace,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		farmKeeper:      farmKeeper,
		liquidityKeeper: liquidityKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the parameters for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetFeeCollector(ctx sdk.Context) (feeCollector string) {
	k.paramSpace.Get(ctx, types.KeyFeeCollector, &feeCollector)
	return
}

func (k Keeper) GetRewardsAuctionDuration(ctx sdk.Context) (duration time.Duration) {
	k.paramSpace.Get(ctx, types.KeyRewardsAuctionDuration, &duration)
	return
}

func (k Keeper) GetLiquidFarmsInParams(ctx sdk.Context) (liquidFarms []types.LiquidFarm) {
	k.paramSpace.Get(ctx, types.KeyLiquidFarms, &liquidFarms)
	return
}
