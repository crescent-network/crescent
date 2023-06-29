package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, execQty sdk.Int, paid, received sdk.Coin, err error) {
	orderId, order, execQty, paid, received, err = k.placeLimitOrder(
		ctx, types.OrderTypeLimit, marketId, ordererAddr, isBuy, price, qty, lifespan, false)
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

func (k Keeper) PlaceBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (order types.Order, err error) {
	_, order, _, _, _, err = k.placeLimitOrder(
		ctx, types.OrderTypeLimit, marketId, ordererAddr, isBuy, price, qty, lifespan, true)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventPlaceBatchLimitOrder{
		MarketId: marketId,
		OrderId:  order.Id,
		Orderer:  ordererAddr.String(),
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
		Deadline: ctx.BlockTime().Add(lifespan),
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceMMLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, execQty sdk.Int, paid, received sdk.Coin, err error) {
	orderId, order, execQty, paid, received, err = k.placeLimitOrder(
		ctx, types.OrderTypeMM, marketId, ordererAddr, isBuy, price, qty, lifespan, false)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventPlaceMMLimitOrder{
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

func (k Keeper) PlaceMMBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (order types.Order, err error) {
	_, order, _, _, _, err = k.placeLimitOrder(
		ctx, types.OrderTypeMM, marketId, ordererAddr, isBuy, price, qty, lifespan, true)
	if err != nil {
		return
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventPlaceMMBatchLimitOrder{
		MarketId: marketId,
		OrderId:  order.Id,
		Orderer:  ordererAddr.String(),
		IsBuy:    isBuy,
		Price:    price,
		Quantity: qty,
		Lifespan: lifespan,
		Deadline: ctx.BlockTime().Add(lifespan),
	}); err != nil {
		return
	}
	return
}

func (k Keeper) placeLimitOrder(
	ctx sdk.Context, typ types.OrderType, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration, isBatch bool) (orderId uint64, order types.Order, execQty sdk.Int, paid, received sdk.Coin, err error) {
	if !qty.IsPositive() { // sanity check
		panic("quantity must be positive")
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	if typ == types.OrderTypeMM {
		numMMOrders, found := k.GetNumMMOrders(ctx, ordererAddr, marketId)
		if found {
			if maxNum := k.GetMaxNumMMOrders(ctx); numMMOrders >= maxNum {
				err = sdkerrors.Wrapf(types.ErrMaxNumMMOrdersExceeded, "%d", maxNum)
				return
			}
		}
		k.SetNumMMOrders(ctx, ordererAddr, marketId, numMMOrders+1)
	}

	marketState := k.MustGetMarketState(ctx, market.Id)
	if marketState.LastPrice != nil {
		maxPriceRatio := k.GetMaxOrderPriceRatio(ctx)
		minPrice := marketState.LastPrice.Mul(utils.OneDec.Sub(maxPriceRatio))
		maxPrice := marketState.LastPrice.Mul(utils.OneDec.Add(maxPriceRatio))
		if isBuy && price.GT(maxPrice) {
			err = sdkerrors.Wrapf(types.ErrOrderPriceOutOfRange, "price is higher than the limit %s", maxPrice)
			return
		} else if !isBuy && price.LT(minPrice) {
			err = sdkerrors.Wrapf(types.ErrOrderPriceOutOfRange, "price is lower than the limit %s", minPrice)
			return
		}
	}

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	if isBatch {
		execQty = utils.ZeroInt
	} else {
		execQty, paid, received, _, err = k.executeOrder(
			ctx, market, ordererAddr, isBuy, &price, &qty, nil, false, false)
		if err != nil {
			return
		}
	}

	openQty := qty.Sub(execQty)
	if isBatch || openQty.IsPositive() {
		deadline := ctx.BlockTime().Add(lifespan)
		order, err = k.newOrder(
			ctx, orderId, typ, market, ordererAddr, isBuy, price,
			qty, openQty, deadline, false)
		if err != nil {
			return
		}
		k.SetOrder(ctx, order)
		k.SetOrderBookOrder(ctx, order)
		k.SetOrdersByOrdererIndex(ctx, order)
	}
	return
}

func (k Keeper) PlaceMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Int) (orderId uint64, execQty sdk.Int, paid, received sdk.Coin, err error) {
	if !qty.IsPositive() { // sanity check
		panic("quantity must be positive")
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	execQty, paid, received, _, err = k.executeOrder(
		ctx, market, ordererAddr, isBuy, nil, &qty, nil, false, false)
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

func (k Keeper) newOrder(
	ctx sdk.Context, orderId uint64, typ types.OrderType, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty, openQty sdk.Int, deadline time.Time, isTemp bool) (types.Order, error) {
	deposit := types.DepositAmount(isBuy, price, openQty)
	msgHeight := int64(0)
	if !isTemp {
		msgHeight = ctx.BlockHeight()
	}
	order := types.NewOrder(
		orderId, typ, ordererAddr, market.Id, isBuy, price, qty,
		msgHeight, openQty, deposit, deadline)
	if err := k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit), isTemp); err != nil {
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
	if order.MsgHeight == ctx.BlockHeight() {
		err = sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest, "cannot cancel order placed in the same block")
		return
	}
	market := k.MustGetMarket(ctx, order.MarketId)
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
	ordererAddr := order.MustGetOrdererAddress()
	refundedDeposit = market.DepositCoin(order.IsBuy, order.RemainingDeposit)
	if err = k.ReleaseCoin(
		ctx, market, ordererAddr, refundedDeposit, queueSend); err != nil {
		return
	}
	if order.Type == types.OrderTypeMM {
		numMMOrders, found := k.GetNumMMOrders(ctx, ordererAddr, market.Id)
		if !found { // sanity check
			panic("num mm orders not found")
		}
		if numMMOrders == 1 {
			k.DeleteNumMMOrders(ctx, ordererAddr, market.Id)
		} else {
			k.SetNumMMOrders(ctx, ordererAddr, market.Id, numMMOrders-1)
		}
	}
	k.DeleteOrder(ctx, order)
	k.DeleteOrderBookOrder(ctx, order)
	k.DeleteOrdersByOrdererIndex(ctx, order)
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
