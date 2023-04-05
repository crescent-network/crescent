package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

var _ exchangetypes.ExchangeHooks = Hooks{}

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterSpotOrderExecuted(ctx sdk.Context, order exchangetypes.SpotLimitOrder, qty sdk.Int) {
	fmt.Println("pool order executed", order, qty)
}
