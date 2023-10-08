package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Inside the getBestPrice function, it is calling the ConstructMemOrderBookSide function,
// which means that the ConstructMemOrderBookSide function is called 4 times in this RunBatchMatching logic.
// Since this logic is called for every market in the midblocker of every block,
// it can act as a significant overhead. How about improving the logic?
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
	if bestBuyPrice.LT(bestSellPrice) {
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

	marketState := k.MustGetMarketState(ctx, market.Id)
	mCtx := types.NewMatchingContext(market, false)
	var (
		lastPrice sdk.Dec
		matched   bool
	)
	if marketState.LastPrice == nil {
		lastPrice, matched = mCtx.RunSinglePriceAuction(buyObs, sellObs)
	} else {
		lastPrice, matched = mCtx.BatchMatchOrderBookSides(buyObs, sellObs, *marketState.LastPrice)
	}
	// If there was no matching, exit early.
	if !matched {
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
	return nil
}
