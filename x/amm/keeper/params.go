package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
)

// GetParams returns the parameters for the module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

// SetParams sets the parameters for the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) GetPoolCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPoolCreationFee, &fee)
	return
}

func (k Keeper) SetPoolCreationFee(ctx sdk.Context, fee sdk.Coins) {
	k.paramSpace.Set(ctx, types.KeyPoolCreationFee, fee)
}

func (k Keeper) GetDefaultTickSpacing(ctx sdk.Context) (tickSpacing uint32) {
	k.paramSpace.Get(ctx, types.KeyDefaultTickSpacing, &tickSpacing)
	return
}

func (k Keeper) SetDefaultTickSpacing(ctx sdk.Context, tickSpacing uint32) {
	k.paramSpace.Set(ctx, types.KeyDefaultTickSpacing, tickSpacing)
}

func (k Keeper) GetPrivateFarmingPlanCreationFee(ctx sdk.Context) (fee sdk.Coins) {
	k.paramSpace.Get(ctx, types.KeyPrivateFarmingPlanCreationFee, &fee)
	return
}

func (k Keeper) SetPrivateFarmingPlanCreationFee(ctx sdk.Context, fee sdk.Coins) {
	k.paramSpace.Set(ctx, types.KeyPrivateFarmingPlanCreationFee, fee)
}

func (k Keeper) GetMaxNumPrivateFarmingPlans(ctx sdk.Context) (max uint32) {
	k.paramSpace.Get(ctx, types.KeyMaxNumPrivateFarmingPlans, &max)
	return
}

func (k Keeper) SetMaxNumPrivateFarmingPlans(ctx sdk.Context, max uint32) {
	k.paramSpace.Set(ctx, types.KeyPrivateFarmingPlanCreationFee, max)
}

func (k Keeper) GetMaxFarmingBlockTime(ctx sdk.Context) (blockTime time.Duration) {
	k.paramSpace.Get(ctx, types.KeyMaxFarmingBlockTime, &blockTime)
	return
}

func (k Keeper) SetMaxFarmingBlockTime(ctx sdk.Context, blockTime time.Duration) {
	k.paramSpace.Set(ctx, types.KeyMaxFarmingBlockTime, blockTime)
}
