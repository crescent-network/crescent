package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

// GetCodec return codec.Codec object used by the keeper
func (k Keeper) GetCodec() codec.BinaryCodec { return k.cdc }

// GetStakingCreationFeePool returns module account for collecting Staking Creation Fee
func (k Keeper) GetStakingCreationFeePool(ctx sdk.Context) authtypes.ModuleAccountI { // nolint:interfacer
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetStakingReservePoolAcc returns module account for Staking Reserve Pool account
func (k Keeper) GetStakingReservePoolAcc(ctx sdk.Context) sdk.AccAddress { // nolint:interfacer
	return types.StakingReserveAcc
}

// GetRewardsReservePoolAcc returns temporary module account for Reward coins Reserve Pool account
func (k Keeper) GetRewardsReservePoolAcc(ctx sdk.Context) sdk.AccAddress { // nolint:interfacer
	return types.RewardsReserveAcc
}

// GetFarmingFeeCollectorAcc returns module account for the farming fee collector account.
func (k Keeper) GetFarmingFeeCollectorAcc(ctx sdk.Context) sdk.AccAddress {
	params := k.GetParams(ctx)
	return sdk.AccAddress(params.FarmingFeeCollector)
}

// GetLastDistributedTime returns the last distributed time for a plan.
func (k Keeper) GetLastDistributedTime(ctx sdk.Context, planID uint64) (time.Time, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLastDistributedTimeKey(planID))
	if bz == nil {
		return time.Time{}, false
	}
	var ts gogotypes.Timestamp
	k.cdc.MustUnmarshal(bz, &ts)
	t, err := gogotypes.TimestampFromProto(&ts)
	if err != nil {
		panic(err)
	}
	return t, true
}

// SetLastDistributedTime sets the last distributed time for a plan.
func (k Keeper) SetLastDistributedTime(ctx sdk.Context, planID uint64, t time.Time) {
	store := ctx.KVStore(k.storeKey)
	ts, err := gogotypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	bz := k.cdc.MustMarshal(ts)
	store.Set(types.GetLastDistributedTimeKey(planID), bz)
}

// GetTotalDistributedRewardCoins returns the total distributed reward coins for a plan so far.
func (k Keeper) GetTotalDistributedRewardCoins(ctx sdk.Context, planID uint64) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTotalDistributedRewardCoinsKey(planID))
	if bz == nil {
		return nil
	}
	var rewardCoins types.RewardCoins
	k.cdc.MustUnmarshal(bz, &rewardCoins)
	return rewardCoins.RewardCoins
}

// SetTotalDistributedRewardCoins sets the total distributed reward coins for a plan so far.
func (k Keeper) SetTotalDistributedRewardCoins(ctx sdk.Context, planID uint64, amt sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.RewardCoins{RewardCoins: amt})
	store.Set(types.GetTotalDistributedRewardCoinsKey(planID), bz)
}
