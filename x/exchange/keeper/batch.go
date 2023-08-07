package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) RunBatchMatching(ctx sdk.Context, market types.Market) (err error) {
	// Find the best buy(bid) and sell(ask) prices to limit the price to load
	// on the other side.
	bestBuyPrice, found := k.getBestPrice(ctx, market, true)
	if !found { // Nothing to match, exit early
		return nil
	}
	bestSellPrice, found := k.getBestPrice(ctx, market, false)
	if !found { // Nothing to match, exit early
		return nil
	}

	// Construct order book sides with the price limits we obtained previously.
	escrow := types.NewEscrow(market.MustGetEscrowAddress())
	buyObs := k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
		IsBuy:      true,
		PriceLimit: &bestSellPrice,
	}, escrow)
	sellObs := k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
		IsBuy:      false,
		PriceLimit: &bestBuyPrice,
	}, escrow)

	var lastPrice sdk.Dec
	marketState := k.MustGetMarketState(ctx, market.Id)
	defer func() {
		// If there was an error, exit early.
		if err != nil {
			return
		}
		// If there was no matching, exit early, too.
		if lastPrice.IsNil() {
			return
		}

		// Apply the match results.
		memOrders := append(append(([]*types.MemOrder)(nil), buyObs.Orders()...), sellObs.Orders()...)
		if err = k.finalizeMatching(ctx, market, memOrders, escrow); err != nil {
			return
		}
		marketState.LastPrice = &lastPrice
		marketState.LastMatchingHeight = ctx.BlockHeight()
		k.SetMarketState(ctx, market.Id, marketState)
	}()

	mCtx := types.NewMatchingContext(market, false)
	if marketState.LastPrice == nil {
		lastPrice = mCtx.RunSinglePriceAuction(buyObs, sellObs)
	} else {
		lastPrice = mCtx.BatchMatchOrderBookSides(buyObs, sellObs, *marketState.LastPrice)
	}
	return nil
}
