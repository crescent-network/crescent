package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

// GetOrderExtraGas returns the current order extra gas parameter.
func (k Keeper) GetOrderExtraGas(ctx sdk.Context) (gas sdk.Gas) {
	k.paramSpace.Get(ctx, types.KeyOrderExtraGas, &gas)
	return
}
