package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) RunBatchMatching(ctx sdk.Context, market types.Market) (err error) {
	// Find the best buy(bid) and sell(ask) prices to limit the price to load
	// on the other side.
	bestBuyPrice, _ := k.GetBestPrice(ctx, market.Id, true)
	bestSellPrice, _ := k.GetBestPrice(ctx, market.Id, false)

	// Construct order book sides with the price limits we obtained previously.
	var buyObs, sellObs *types.MemOrderBookSide
	if !bestSellPrice.IsNil() {
		buyObs = k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
			IsBuy:      true,
			PriceLimit: &bestSellPrice,
		}, nil)
	} else {
		buyObs = types.NewMemOrderBookSide(true)
	}
	if !bestBuyPrice.IsNil() {
		sellObs = k.ConstructMemOrderBookSide(ctx, market, types.MemOrderBookSideOptions{
			IsBuy:      false,
			PriceLimit: &bestBuyPrice,
		}, nil)
	} else {
		sellObs = types.NewMemOrderBookSide(false)
	}

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
		if err = k.FinalizeMatching(ctx, market, memOrders, nil); err != nil {
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
