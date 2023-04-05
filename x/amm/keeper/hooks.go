package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.ExchangeHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterRestingSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotLimitOrder, qty sdk.Int) {
	noGasCtx := ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())
	pool, found := h.k.GetPoolByReserveAddress(noGasCtx, sdk.MustAccAddressFromBech32(order.Orderer))
	if found {
		h.k.IncrementExecutedQuantity(ctx, pool.Id, qty)
	}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, market exchangetypes.SpotMarket, _ sdk.AccAddress, isBuy bool, lastPrice sdk.Dec, qty, quoteAmt sdk.Int) {
	h.k.IteratePoolsByMarket(ctx, market.Id, func(pool types.Pool) (stop bool) {
		executedQty := h.k.GetExecutedQuantity(ctx, pool.Id)
		if executedQty.IsPositive() {
			// TODO: update pool's current sqrt price and replace orders
		}
		return false
	})
}
