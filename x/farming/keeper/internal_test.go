package keeper

import (
	"github.com/crescent-network/crescent/v2/x/farming/types"
)

// UnsafeSetHooks updates the farming keeper's hooks, overriding any potential
// pre-existing hooks.
// WARNING: this function should only be used in tests.
func UnsafeSetHooks(k *Keeper, h types.FarmingHooks) {
	k.hooks = h
}
