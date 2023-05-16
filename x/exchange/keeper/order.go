package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.Order, execQty, execQuote sdk.Int, err error) {
	return k.PlaceOrder(ctx, marketId, ordererAddr, isBuy, &price, qty)
}

func (k Keeper) PlaceMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Int) (execQty, execQuote sdk.Int, err error) {
	_, execQty, execQuote, err = k.PlaceOrder(ctx, marketId, ordererAddr, isBuy, nil, qty)
	return
}

func (k Keeper) PlaceOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (order types.Order, execQty, execQuote sdk.Int, err error) {
	if !qty.IsPositive() {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "quantity must be positive")
		return
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	execQty, execQuote = k.executeOrder(
		ctx, market, ordererAddr, isBuy, priceLimit, &qty, nil, false)

	openQty := qty.Sub(execQty)
	if priceLimit != nil {
		if openQty.IsPositive() {
			order, err = k.CreateOrder(ctx, market, ordererAddr, isBuy, *priceLimit, qty, openQty, false)
			if err != nil {
				return
			}
			k.SetOrder(ctx, order)
			k.SetOrderBookOrder(ctx, order)
		}
	}
	return
}

func (k Keeper) CreateOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty, openQty sdk.Int, isTemp bool) (types.Order, error) {
	orderId := k.GetNextOrderIdWithUpdate(ctx)
	deposit := types.DepositAmount(isBuy, price, openQty)
	msgHeight := int64(0)
	if !isTemp {
		msgHeight = ctx.BlockHeight()
	}
	order := types.NewOrder(
		orderId, ordererAddr, market.Id, isBuy, price, qty, msgHeight, openQty, deposit)
	if err := k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit), false); err != nil {
		return order, err
	}
	return order, nil
}

func (k Keeper) CancelOrder(ctx sdk.Context, senderAddr sdk.AccAddress, orderId uint64) (order types.Order, err error) {
	order, found := k.GetOrder(ctx, orderId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
	}
	market, found := k.GetMarket(ctx, order.MarketId)
	if !found { // sanity check
		panic("market not found")
	}
	if senderAddr.String() != order.Orderer {
		return order, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "order is not created by the sender")
	}
	refundedCoin := market.DepositCoin(order.IsBuy, order.RemainingDeposit)
	if err := k.ReleaseCoin(ctx, market, senderAddr, refundedCoin, false); err != nil {
		return order, err
	}
	k.DeleteOrder(ctx, order)
	k.DeleteOrderBookOrder(ctx, order)
	return order, nil
}
