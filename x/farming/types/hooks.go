package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MultiFarmingHooks combines multiple farming hooks.
// All hook functions are run in array sequence
type MultiFarmingHooks []FarmingHooks

func NewMultiFarmingHooks(hooks ...FarmingHooks) MultiFarmingHooks {
	return hooks
}

func (h MultiFarmingHooks) AfterAllocateRewards(ctx sdk.Context) {
	for i := range h {
		h[i].AfterAllocateRewards(ctx)
	}
}
