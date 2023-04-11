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
	if !qty.IsPositive() {
		return order, false, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "quantity must be positive")
	}
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		return order, false, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}

	firstPrice, lastPrice, executedQty, executedQuoteAmt, outputs := k.executeSpotOrder(ctx, market, ordererAddr, isBuy, priceLimit, qty)

	// TODO: merge if-statement conditions - seems like duplicated
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
			return order, false, err
		}
	}

	if !firstPrice.IsNil() {
		k.AfterSpotOrderExecuted(ctx, market, ordererAddr, isBuy, firstPrice, lastPrice, executedQty, executedQuoteAmt)
	}

	if priceLimit != nil { // limit order
		if remainingQty := qty.Sub(executedQty); remainingQty.IsPositive() {
			order = k.restSpotLimitOrder(ctx, ordererAddr, marketId, isBuy, *priceLimit, remainingQty)
			depositCoins := sdk.Coins{market.DepositCoin(isBuy, types.DepositAmount(isBuy, *priceLimit, remainingQty))}
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ordererAddr, types.ModuleName, depositCoins); err != nil {
				return order, false, err
			}
			rested = true
		}
	}
	return
}

func (k Keeper) executeSpotOrder(
	ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (firstPrice, lastPrice sdk.Dec, executedQty, executedQuoteAmt sdk.Int, outputs []banktypes.Output) {
	remainingQty := qty
	executedQuoteAmt = types.ZeroInt
	k.IterateSpotOrderBookOrders(ctx, market.Id, !isBuy, priceLimit, func(order types.SpotLimitOrder) (stop bool) {
		if firstPrice.IsNil() {
			firstPrice = order.Price
		}

		executableQty := types.MinInt(
			order.ExecutableQuantity(),
			remainingQty)
		quoteAmt := types.QuoteAmount(isBuy, order.Price, executableQty)
		executedQuoteAmt = executedQuoteAmt.Add(quoteAmt)

		k.AfterRestingSpotOrderExecuted(ctx, order, executableQty) // TODO: rename the hook?
		order.OpenQuantity = order.OpenQuantity.Sub(executableQty)
		var paidCoin, receivedCoin sdk.Coins
		if isBuy {
			paidCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
			order.DepositAmount = order.DepositAmount.Sub(executableQty)
		} else {
			paidCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
			order.DepositAmount = order.DepositAmount.Sub(quoteAmt)
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

		remainingQty = remainingQty.Sub(executableQty)
		return remainingQty.IsZero()
	})
	executedQty = qty.Sub(remainingQty)
	return
}

func (k Keeper) restSpotLimitOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder) {
	orderId := k.GetNextOrderIdWithUpdate(ctx)
	depositAmt := types.DepositAmount(isBuy, price, qty)
	order = types.NewSpotLimitOrder(orderId, ordererAddr, marketId, isBuy, price, qty, depositAmt)
	k.SetSpotLimitOrder(ctx, order)
	k.SetSpotOrderBookOrder(ctx, order)
	return
}

func (k Keeper) CancelSpotOrder(ctx sdk.Context, senderAddr sdk.AccAddress, marketId string, orderId uint64) (order types.SpotLimitOrder, err error) {
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}
	order, found = k.GetSpotLimitOrder(ctx, market.Id, orderId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
	}
	if senderAddr.String() != order.Orderer {
		return order, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "order is not created by the sender")
	}
	refundedCoins := sdk.NewCoins(market.DepositCoin(order.IsBuy, order.DepositAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, refundedCoins); err != nil {
		return order, err
	}
	k.DeleteSpotLimitOrder(ctx, order)
	k.DeleteSpotOrderBookOrder(ctx, order)
	return order, nil
}
