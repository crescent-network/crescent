package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceSpotLimitOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, priceLimit sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder, rested bool, err error) {
	return k.PlaceSpotOrder(ctx, ordererAddr, marketId, isBuy, &priceLimit, qty)
}

func (k Keeper) PlaceSpotMarketOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, qty sdk.Int) error {
	_, _, err := k.PlaceSpotOrder(ctx, ordererAddr, marketId, isBuy, nil, qty)
	return err
}

func (k Keeper) PlaceSpotOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder, rested bool, err error) {
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		return types.SpotLimitOrder{}, false, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}

	lastPrice, executedQty, executedQuoteAmt, outputs := k.executeSpotOrder(ctx, market, ordererAddr, isBuy, priceLimit, qty)

	if !lastPrice.IsNil() {
		market.LastPrice = &lastPrice
		k.SetSpotMarket(ctx, market)
	}

	if executedQty.IsPositive() {
		executedQuoteCoin := sdk.Coins{sdk.NewCoin(market.QuoteDenom, executedQuoteAmt)}
		executedBaseCoin := sdk.Coins{sdk.NewCoin(market.BaseDenom, executedQty)}
		var inputs []banktypes.Input
		if isBuy {
			inputs = append(inputs,
				banktypes.NewInput(ordererAddr, executedQuoteCoin),
				banktypes.NewInput(k.accountKeeper.GetModuleAddress(types.ModuleName), executedBaseCoin))
		} else {
			inputs = append(inputs,
				banktypes.NewInput(ordererAddr, executedBaseCoin),
				banktypes.NewInput(k.accountKeeper.GetModuleAddress(types.ModuleName), executedQuoteCoin))
		}

		if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return types.SpotLimitOrder{}, false, err
		}
	}

	if priceLimit != nil { // limit order
		if remainingQty := qty.Sub(executedQty); remainingQty.IsPositive() {
			order = k.restSpotLimitOrder(ctx, ordererAddr, marketId, isBuy, *priceLimit, remainingQty)
			offerCoin := sdk.Coins{market.OfferCoin(isBuy, *priceLimit, remainingQty)}
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ordererAddr, types.ModuleName, offerCoin); err != nil {
				return types.SpotLimitOrder{}, false, err
			}
			rested = true
		}
	}
	return
}

func (k Keeper) executeSpotOrder(
	ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (lastPrice sdk.Dec, executedQty, executedQuoteAmt sdk.Int, outputs []banktypes.Output) {
	remainingQty := qty
	executedQuoteAmt = types.ZeroInt
	k.IterateSpotOrderBookOrders(ctx, market.Id, !isBuy, priceLimit, func(order types.SpotLimitOrder) (stop bool) {
		executableQty := types.MinInt(order.OpenQuantity, remainingQty)
		quoteAmt := types.QuoteAmount(isBuy, order.Price, executableQty)
		executedQuoteAmt = executedQuoteAmt.Add(quoteAmt)

		order.OpenQuantity = order.OpenQuantity.Sub(executableQty)
		var paidCoin, receivedCoin sdk.Coins
		if isBuy {
			paidCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
		} else {
			paidCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
		}
		outputs = append(outputs,
			banktypes.Output{Address: order.Orderer, Coins: paidCoin},
			banktypes.NewOutput(ordererAddr, receivedCoin))

		lastPrice = order.Price
		if order.OpenQuantity.IsZero() {
			k.DeleteSpotOrderBookOrder(ctx, order)
			k.DeleteSpotLimitOrder(ctx, order)
			// TODO: emit event
		} else {
			k.SetSpotLimitOrder(ctx, order)
		}
		k.AfterRestingSpotOrderExecuted(ctx, order, executableQty)

		remainingQty = remainingQty.Sub(executableQty)
		return remainingQty.IsZero()
	})
	executedQty = qty.Sub(remainingQty)
	k.AfterSpotOrderExecuted(ctx, market, ordererAddr, isBuy, lastPrice, executedQty, executedQuoteAmt)
	return
}

func (k Keeper) restSpotLimitOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder) {
	orderId := k.GetNextOrderIdWithUpdate(ctx)
	order = types.NewSpotLimitOrder(orderId, ordererAddr, marketId, isBuy, price, qty)
	k.SetSpotLimitOrder(ctx, order)
	k.SetSpotOrderBookOrder(ctx, order)
	return
}
