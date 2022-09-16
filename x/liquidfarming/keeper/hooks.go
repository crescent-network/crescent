package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ farmingtypes.FarmingHooks = Hooks{}

// Hooks creates new hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterAllocateRewards hook is triggered in the farming module when an epoch is advanced and
// AllocateRewards is successfully executed till the end logic.
// It creates the first rewards auction if liquid farm doesn't have any auction before.
// If there is an ongoing rewards auction, finish the auction and create the next one.
func (h Hooks) AfterAllocateRewards(ctx sdk.Context) {
	for _, liquidFarm := range h.k.GetAllLiquidFarms(ctx) {
		auctionId := h.k.GetLastRewardsAuctionId(ctx, liquidFarm.PoolId)
		auction, found := h.k.GetRewardsAuction(ctx, liquidFarm.PoolId, auctionId)
		if found {
			if err := h.k.FinishRewardsAuction(ctx, auction, liquidFarm.FeeRate); err != nil {
				panic(err)
			}
		}
		h.k.CreateRewardsAuction(ctx, liquidFarm.PoolId)
	}
}
