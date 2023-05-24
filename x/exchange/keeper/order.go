package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, execQty sdk.Int, paid, received sdk.Coin, err error) {
	orderId, order, execQty, paid, received, err = k.PlaceOrder(
		ctx, marketId, ordererAddr, isBuy, &price, qty, lifespan)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventPlaceLimitOrder{
		MarketId:         marketId,
		OrderId:          orderId,
		Orderer:          ordererAddr.String(),
		IsBuy:            isBuy,
		Price:            price,
		Quantity:         qty,
		Lifespan:         lifespan,
		Deadline:         ctx.BlockTime().Add(lifespan),
		ExecutedQuantity: execQty,
		Paid:             paid,
		Received:         received,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Int) (orderId uint64, execQty sdk.Int, paid, received sdk.Coin, err error) {
	orderId, _, execQty, paid, received, err = k.PlaceOrder(
		ctx, marketId, ordererAddr, isBuy, nil, qty, 0)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventPlaceMarketOrder{
		MarketId:         marketId,
		OrderId:          orderId,
		Orderer:          ordererAddr.String(),
		IsBuy:            isBuy,
		Quantity:         qty,
		ExecutedQuantity: execQty,
		Paid:             paid,
		Received:         received,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, execQty sdk.Int, paid, received sdk.Coin, err error) {
	if !qty.IsPositive() {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "quantity must be positive")
		return
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	execQty, paid, received = k.executeOrder(
		ctx, market, ordererAddr, isBuy, priceLimit, &qty, nil, false)

	openQty := qty.Sub(execQty)
	if priceLimit != nil {
		if openQty.IsPositive() {
			deadline := ctx.BlockTime().Add(lifespan)
			order, err = k.newOrder(
				ctx, orderId, market, ordererAddr, isBuy, *priceLimit,
				qty, openQty, deadline, false)
			if err != nil {
				return
			}
			k.SetOrder(ctx, order)
			k.SetOrderBookOrder(ctx, order)
		}
	}
	return
}

func (k Keeper) newOrder(
	ctx sdk.Context, orderId uint64, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty, openQty sdk.Int, deadline time.Time, isTemp bool) (types.Order, error) {
	deposit := types.DepositAmount(isBuy, price, openQty)
	msgHeight := int64(0)
	if !isTemp {
		msgHeight = ctx.BlockHeight()
	}
	order := types.NewOrder(
		orderId, ordererAddr, market.Id, isBuy, price, qty,
		msgHeight, openQty, deposit, deadline)
	if err := k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit), false); err != nil {
		return order, err
	}
	return order, nil
}

func (k Keeper) CancelOrder(ctx sdk.Context, ordererAddr sdk.AccAddress, orderId uint64) (order types.Order, refundedDeposit sdk.Coin, err error) {
	order, found := k.GetOrder(ctx, orderId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
		return
	}
	market, found := k.GetMarket(ctx, order.MarketId)
	if !found { // sanity check
		panic("market not found")
	}
	if ordererAddr.String() != order.Orderer {
		err = sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "order is not created by the sender")
		return
	}
	refundedDeposit, err = k.cancelOrder(ctx, market, order, false)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventCancelOrder{
		Orderer:         ordererAddr.String(),
		OrderId:         orderId,
		RefundedDeposit: refundedDeposit,
	}); err != nil {
		return
	}
	return order, refundedDeposit, nil
}

func (k Keeper) cancelOrder(ctx sdk.Context, market types.Market, order types.Order, queueSend bool) (refundedDeposit sdk.Coin, err error) {
	refundedDeposit = market.DepositCoin(order.IsBuy, order.RemainingDeposit)
	if err = k.ReleaseCoin(
		ctx, market, sdk.MustAccAddressFromBech32(order.Orderer), refundedDeposit, queueSend); err != nil {
		return
	}
	k.DeleteOrder(ctx, order)
	k.DeleteOrderBookOrder(ctx, order)
	return
}

func (k Keeper) CancelExpiredOrders(ctx sdk.Context) (err error) {
	blockTime := ctx.BlockTime()
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		// TODO: optimize by using timestamp queue
		k.IterateOrdersByMarket(ctx, market.Id, func(order types.Order) (stop bool) {
			if !blockTime.Before(order.Deadline) {
				if _, err = k.cancelOrder(ctx, market, order, true); err != nil {
					return true
				}
				if err = ctx.EventManager().EmitTypedEvent(&types.EventOrderExpired{
					OrderId: order.Id,
				}); err != nil {
					return true
				}
			}
			return false
		})
		return err != nil
	})
	return k.ExecuteSendCoins(ctx)
}
