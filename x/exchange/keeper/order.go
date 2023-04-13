package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceSpotLimitOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, priceLimit sdk.Dec, qty sdk.Int) (order types.SpotOrder, execQuote sdk.Int, err error) {
	return k.PlaceSpotOrder(ctx, ordererAddr, marketId, isBuy, &priceLimit, qty)
}

func (k Keeper) PlaceSpotMarketOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, qty sdk.Int) (order types.SpotOrder, execQuote sdk.Int, err error) {
	return k.PlaceSpotOrder(ctx, ordererAddr, marketId, isBuy, nil, qty)
}

func (k Keeper) PlaceSpotOrder(
	ctx sdk.Context, ordererAddr sdk.AccAddress, marketId string,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (order types.SpotOrder, execQuote sdk.Int, err error) {
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}
	if !qty.IsPositive() {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "quantity must be positive")
		return
	}

	var (
		firstPrice, lastPrice sdk.Dec
		execQty               sdk.Int
	)
	firstPrice, lastPrice, execQty, execQuote = k.executeSpotOrder(ctx, market, ordererAddr, isBuy, priceLimit, qty)

	if !lastPrice.IsNil() {
		state, found := k.GetSpotMarketState(ctx, market.Id)
		if !found { // sanity check
			panic("spot market state not found")
		}
		state.LastPrice = &lastPrice
		k.SetSpotMarketState(ctx, market.Id, state)
		k.AfterSpotOrderExecuted(ctx, market, ordererAddr, isBuy, firstPrice, lastPrice, execQty, execQuote)
	}

	openQty := qty.Sub(execQty)
	var deposit sdk.Int
	if priceLimit != nil {
		deposit = types.DepositAmount(isBuy, *priceLimit, openQty)
	} else {
		deposit = utils.ZeroInt
	}
	orderId := k.GetNextOrderIdWithUpdate(ctx)
	order = types.NewSpotOrder(
		orderId, ordererAddr, market.Id, isBuy, priceLimit, qty, openQty, deposit)

	if priceLimit != nil { // limit order
		if openQty.IsPositive() {
			if err = k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit)); err != nil {
				return
			}
			k.SetSpotOrder(ctx, order)
			k.SetSpotOrderBookOrder(ctx, order)
		}
	}
	return
}

func (k Keeper) executeSpotOrder(
	ctx sdk.Context, market types.SpotMarket, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (firstPrice, lastPrice sdk.Dec, totalExecQty, totalExecQuote sdk.Int) {
	totalExecQuote = utils.ZeroInt
	remainingQty := qty
	type Trade struct {
		order   types.SpotOrder
		execQty sdk.Int
	}
	var trades []Trade
	k.IterateSpotOrderBookOrders(ctx, market.Id, !isBuy, priceLimit, func(order types.SpotOrder) (stop bool) {
		if firstPrice.IsNil() {
			firstPrice = *order.Price
		}

		execQty := utils.MinInt(order.ExecutableQuantity(), remainingQty)
		execQuote := types.QuoteAmount(isBuy, *order.Price, execQty)
		totalExecQuote = totalExecQuote.Add(execQuote)

		var (
			paid                   sdk.Int
			paidCoin, receivedCoin sdk.Coin
		)
		if isBuy {
			paid = execQuote
			paidCoin = sdk.NewCoin(market.QuoteDenom, paid)
			receivedCoin = sdk.NewCoin(market.BaseDenom, execQty)
		} else {
			paid = execQty
			paidCoin = sdk.NewCoin(market.BaseDenom, paid)
			receivedCoin = sdk.NewCoin(market.QuoteDenom, execQuote)
		}
		if err := k.bankKeeper.SendCoins(
			ctx, ordererAddr, sdk.MustAccAddressFromBech32(order.Orderer), sdk.NewCoins(paidCoin)); err != nil {
			panic(err)
		}
		if err := k.ReleaseCoin(ctx, market, ordererAddr, receivedCoin); err != nil {
			panic(err)
		}
		order.OpenQuantity = order.OpenQuantity.Sub(execQty)
		order.RemainingDeposit = order.RemainingDeposit.Sub(receivedCoin.Amount)
		trades = append(trades, Trade{
			order:   order,
			execQty: execQty,
		})

		lastPrice = *order.Price
		if order.OpenQuantity.IsZero() {
			k.DeleteSpotOrderBookOrder(ctx, order)
			k.DeleteSpotOrder(ctx, order)
			// TODO: emit event
		} else {
			k.SetSpotOrder(ctx, order)
		}

		remainingQty = remainingQty.Sub(execQty)
		return remainingQty.IsZero()
	})
	for _, trade := range trades {
		k.AfterRestingSpotOrderExecuted(ctx, trade.order, trade.execQty) // TODO: rename the hook?
	}
	totalExecQty = qty.Sub(remainingQty)
	return
}

func (k Keeper) CancelSpotOrder(ctx sdk.Context, senderAddr sdk.AccAddress, marketId string, orderId uint64) (order types.SpotOrder, err error) {
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}
	order, found = k.GetSpotOrder(ctx, market.Id, orderId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
	}
	if senderAddr.String() != order.Orderer {
		return order, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "order is not created by the sender")
	}
	refundedCoin := market.DepositCoin(order.IsBuy, order.RemainingDeposit)
	if err := k.ReleaseCoin(ctx, market, senderAddr, refundedCoin); err != nil {
		return order, err
	}
	k.DeleteSpotOrder(ctx, order)
	k.DeleteSpotOrderBookOrder(ctx, order)
	return order, nil
}
