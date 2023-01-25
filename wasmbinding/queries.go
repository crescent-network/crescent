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

func (qp QueryPlugin) Pairs(ctx sdk.Context) *bindings.PairsResponse {
	pairs := qp.liquidityKeeper.GetAllPairs(ctx)

	pairsResponse := []bindings.PairResponse{}
	for _, pair := range pairs {
		p := &bindings.PairResponse{
			Id:             pair.Id,
			BaseCoinDenom:  pair.BaseCoinDenom,
			QuoteCoinDenom: pair.QuoteCoinDenom,
			EscrowAddress:  pair.EscrowAddress,
		}
		pairsResponse = append(pairsResponse, *p)
	}
	return &bindings.PairsResponse{Pairs: pairsResponse}
}

func (qp QueryPlugin) Pair(ctx sdk.Context, pairId uint64) (*bindings.PairResponse, error) {
	pair, found := qp.liquidityKeeper.GetPair(ctx, pairId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pair not found")
	}
	return &bindings.PairResponse{
		Id:             pair.Id,
		BaseCoinDenom:  pair.BaseCoinDenom,
		QuoteCoinDenom: pair.QuoteCoinDenom,
		EscrowAddress:  pair.EscrowAddress,
	}, nil
}

func (qp QueryPlugin) Pools(ctx sdk.Context) *bindings.PoolsResponse {
	pools := qp.liquidityKeeper.GetAllPools(ctx)
	return &bindings.PoolsResponse{Pools: pools}
}

func (qp QueryPlugin) Pool(ctx sdk.Context, poolId uint64) (*bindings.PoolResponse, error) {
	pool, found := qp.liquidityKeeper.GetPool(ctx, poolId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "pool not found")
	}
	return &bindings.PoolResponse{Pool: pool}, nil
}
