package wasmbinding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v4/wasmbinding/bindings"
	liquiditykeeper "github.com/crescent-network/crescent/v4/x/liquidity/keeper"
)

type QueryPlugin struct {
	liquidityKeeper *liquiditykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(lk *liquiditykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		liquidityKeeper: lk,
	}
}

// TODO: the function must return bindings.PairResponse? why not []liquiditytypes.Pair?
func (qp QueryPlugin) Pairs(ctx sdk.Context) *bindings.PairsResponse {
	pairs := qp.liquidityKeeper.GetAllPairs(ctx)
	return &bindings.PairsResponse{Pairs: pairs}
}

func (qp QueryPlugin) Pair(ctx sdk.Context, pairId uint64) (*bindings.PairResponse, error) {
	pair, found := qp.liquidityKeeper.GetPair(ctx, pairId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pair not found")
	}
	return &bindings.PairResponse{Pair: pair}, nil
}

func (qp QueryPlugin) Pools(ctx sdk.Context) {
	// TODO: not implemented yet
}

func (qp QueryPlugin) Pool(ctx sdk.Context) {
	// TODO: not implemented yet
}
