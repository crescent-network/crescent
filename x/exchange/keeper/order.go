package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceSpotLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.SpotOrder, execQty, execQuote sdk.Int, err error) {
	return k.PlaceSpotOrder(ctx, marketId, ordererAddr, isBuy, &price, qty)
}

func (k Keeper) PlaceSpotMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Int) (execQty, execQuote sdk.Int, err error) {
	_, execQty, execQuote, err = k.PlaceSpotOrder(ctx, marketId, ordererAddr, isBuy, nil, qty)
	return
}

func (k Keeper) PlaceSpotOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (order types.SpotOrder, execQty, execQuote sdk.Int, err error) {
	if !qty.IsPositive() {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "quantity must be positive")
		return
	}

	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	execQty, execQuote = k.executeSpotOrder(
		ctx, market, ordererAddr, isBuy, priceLimit, &qty, nil)

	openQty := qty.Sub(execQty)
	if priceLimit != nil {
		if openQty.IsPositive() {
			orderId := k.GetNextOrderIdWithUpdate(ctx)
			deposit := types.DepositAmount(isBuy, *priceLimit, openQty)
			order = types.NewSpotOrder(
				orderId, ordererAddr, market.Id, isBuy, *priceLimit, qty, openQty, deposit)
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
	isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int) (totalExecQty, totalExecQuote sdk.Int) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	var lastPrice sdk.Dec
	totalExecQty = utils.ZeroInt
	totalExecQuote = utils.ZeroInt
	k.constructTransientSpotOrderBook(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit)
	k.IterateTransientSpotOrderBookSide(ctx, market.Id, !isBuy, func(order types.TransientSpotOrder) (stop bool) {
		if priceLimit != nil &&
			((isBuy && order.Order.Price.GT(*priceLimit)) ||
				(!isBuy && order.Order.Price.LT(*priceLimit))) {
			return true
		}
		if qtyLimit != nil && !qtyLimit.Sub(totalExecQty).IsPositive() {
			return true
		}
		if quoteLimit != nil && !quoteLimit.Sub(totalExecQuote).IsPositive() {
			return true
		}

		var execQty sdk.Int
		if qtyLimit != nil {
			execQty = utils.MinInt(order.ExecutableQuantity(), qtyLimit.Sub(totalExecQty))
		} else { // quoteLimit != nil
			execQty = utils.MinInt(
				order.ExecutableQuantity(),
				quoteLimit.Sub(totalExecQuote).ToDec().QuoTruncate(order.Order.Price).TruncateInt())
		}
		execQuote := types.QuoteAmount(isBuy, order.Order.Price, execQty)
		totalExecQty = totalExecQty.Add(execQty)
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
			ctx, ordererAddr, sdk.MustAccAddressFromBech32(order.Order.Orderer), sdk.NewCoins(paidCoin)); err != nil {
			panic(err)
		}
		if err := k.ReleaseCoin(ctx, market, ordererAddr, receivedCoin); err != nil {
			panic(err)
		}
		order.Order.OpenQuantity = order.Order.OpenQuantity.Sub(execQty)
		order.Order.RemainingDeposit = order.Order.RemainingDeposit.Sub(receivedCoin.Amount)
		order.Updated = true
		lastPrice = order.Order.Price

		k.SetTransientSpotOrderBookOrder(ctx, order)
		k.AfterSpotOrderExecuted(ctx, order.Order, execQty)
		return false
	})
	k.settleTransientSpotOrderBook(ctx, market)
	if !lastPrice.IsNil() {
		state := k.MustGetSpotMarketState(ctx, market.Id)
		state.LastPrice = &lastPrice
		k.SetSpotMarketState(ctx, market.Id, state)
	}
	return
}

func (k Keeper) CancelSpotOrder(ctx sdk.Context, senderAddr sdk.AccAddress, orderId uint64) (order types.SpotOrder, err error) {
	order, found := k.GetSpotOrder(ctx, orderId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
	}
	market, found := k.GetSpotMarket(ctx, order.MarketId)
	if !found { // sanity check
		panic("market not found")
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
