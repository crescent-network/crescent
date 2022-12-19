package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	liquiditykeeper "github.com/crescent-network/crescent/v3/x/liquidity/keeper"
)

type QueryPlugin struct {
	liquidityKeeper *liquiditykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(lpk *liquiditykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		liquidityKeeper: lpk,
	}
}

func (qp QueryPlugin) QueryPairs(ctx sdk.Context) {
	// TODO: not implemented yet
}

func (qp QueryPlugin) QueryPair(ctx sdk.Context) {
	// TODO: not implemented yet
}

func (qp QueryPlugin) QueryPools(ctx sdk.Context) {
	// TODO: not implemented yet
}

func (qp QueryPlugin) QueryPool(ctx sdk.Context) {
	// TODO: not implemented yet
}

// TODO: add Orders and Order
