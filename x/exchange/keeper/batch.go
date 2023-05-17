package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.Order, err error) {
	market, found := k.GetMarket(ctx, marketId)
	if !found { // sanity check
		panic("market not found")
	}
	return k.CreateOrder(ctx, market, ordererAddr, isBuy, price, qty, qty, false)
}

func (k Keeper) RunBatch(ctx sdk.Context, market types.Market, orders []types.Order) {
	var bestBuyPrice, bestSellPrice sdk.Dec
	for _, order := range orders {
		if order.IsBuy {
			if bestBuyPrice.IsNil() || order.Price.GT(bestBuyPrice) {
				bestBuyPrice = order.Price
			}
		} else {
			if bestSellPrice.IsNil() || order.Price.LT(bestSellPrice) {
				bestSellPrice = order.Price
			}
		}
	}

	var buyObs, sellObs *types.TempOrderBookSide
	if !bestSellPrice.IsNil() {
		buyObs = k.ConstructTempOrderBookSide(ctx, market, true, &bestSellPrice, nil, nil)
	} else {
		buyObs = types.NewTempOrderBookSide(true)
	}
	if !bestBuyPrice.IsNil() {
		sellObs = k.ConstructTempOrderBookSide(ctx, market, false, &bestBuyPrice, nil, nil)
	} else {
		sellObs = types.NewTempOrderBookSide(false)
	}
	for _, order := range orders {
		tempOrder := types.NewTempOrder(order, market, nil)
		if order.IsBuy {
			buyObs.AddOrder(tempOrder)
		} else {
			sellObs.AddOrder(tempOrder)
		}
	}

	marketState := k.MustGetMarketState(ctx, market.Id)
	if marketState.LastPrice == nil {
		panic("not implemented")
	}
	lastPrice := *marketState.LastPrice

	// Phase 1
	buyLevelIdx, sellLevelIdx := 0, 0
	for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
		buyLevel := buyObs.Levels[buyLevelIdx]
		sellLevel := sellObs.Levels[sellLevelIdx]
		if buyLevel.Price.LT(lastPrice) || sellLevel.Price.GT(lastPrice) {
			break
		}
		// Both sides are maker
		_, sellFull, buyFull := market.MatchOrderBookLevels(sellLevel, false, buyLevel, false, lastPrice)
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
	}
	if buyLevelIdx >= len(buyObs.Levels) || sellLevelIdx >= len(sellObs.Levels) {
		return
	}

	// Phase2
	isPriceIncreasing := sellObs.Levels[sellLevelIdx].Price.GT(lastPrice)
	for buyLevelIdx < len(buyObs.Levels) && sellLevelIdx < len(sellObs.Levels) {
		buyLevel := buyObs.Levels[buyLevelIdx]
		sellLevel := sellObs.Levels[sellLevelIdx]
		if buyLevel.Price.LT(sellLevel.Price) {
			break
		}
		var price sdk.Dec
		if isPriceIncreasing {
			price = sellLevel.Price
		} else {
			price = buyLevel.Price
		}
		_, sellFull, buyFull := market.MatchOrderBookLevels(sellLevel, isPriceIncreasing, buyLevel, !isPriceIncreasing, price)
		lastPrice = price
		if buyFull {
			buyLevelIdx++
		}
		if sellFull {
			sellLevelIdx++
		}
	}

	var tempOrders []*types.TempOrder
	for _, level := range buyObs.Levels {
		tempOrders = append(tempOrders, level.Orders...)
	}
	for _, level := range sellObs.Levels {
		tempOrders = append(tempOrders, level.Orders...)
	}
	if err := k.FinalizeMatching(ctx, market, tempOrders); err != nil {
		panic(err)
	}
	if !marketState.LastPrice.Equal(lastPrice) {
		marketState.LastPrice = &lastPrice
		k.SetMarketState(ctx, market.Id, marketState)
	}
}
