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

	resp := []bindings.PairResponse{}
	for _, pair := range pairs {
		p := &bindings.PairResponse{
			Id:             pair.Id,
			BaseCoinDenom:  pair.BaseCoinDenom,
			QuoteCoinDenom: pair.QuoteCoinDenom,
			EscrowAddress:  pair.EscrowAddress,
		}
		resp = append(resp, *p)
	}
	return &bindings.PairsResponse{Pairs: resp}
}

func (qp QueryPlugin) Pair(ctx sdk.Context, pairId uint64) (*bindings.PairResponse, error) {
	pair, found := qp.liquidityKeeper.GetPair(ctx, pairId)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", pairId)
	}

	return &bindings.PairResponse{
		Id:             pair.Id,
		BaseCoinDenom:  pair.BaseCoinDenom,
		QuoteCoinDenom: pair.QuoteCoinDenom,
		EscrowAddress:  pair.EscrowAddress,
	}, nil
}
