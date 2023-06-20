package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func HandleMarketParameterChangeProposal(ctx sdk.Context, k Keeper, p *types.MarketParameterChangeProposal) error {
	for _, change := range p.Changes {
		market, found := k.GetMarket(ctx, change.MarketId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "market %d not found", change.MarketId)
		}
		market.MakerFeeRate = change.MakerFeeRate
		market.TakerFeeRate = change.TakerFeeRate
		k.SetMarket(ctx, market)
		if err := ctx.EventManager().EmitTypedEvent(&types.EventMarketParameterChanged{
			MarketId:     change.MarketId,
			MakerFeeRate: change.MakerFeeRate,
			TakerFeeRate: change.TakerFeeRate,
		}); err != nil {
			return err
		}
	}
	return nil
}
