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
	isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
	orderId, order, res, err = k.placeLimitOrder(
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
		ExecutedQuantity: res.ExecutedQuantity,
		Paid:             res.Paid,
		Received:         res.Received,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (order types.Order, err error) {
	_, order, _, err = k.placeLimitOrder(
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
	isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
	orderId, order, res, err = k.placeLimitOrder(
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
		ExecutedQuantity: res.ExecutedQuantity,
		Paid:             res.Paid,
		Received:         res.Received,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceMMBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (order types.Order, err error) {
	_, order, _, err = k.placeLimitOrder(
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
	isBuy bool, price, qty sdk.Dec, lifespan time.Duration, isBatch bool) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
	if !qty.IsPositive() { // sanity check
		panic("quantity must be positive")
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	var (
		maxNumMMOrders, numMMOrders uint32
	)
	if typ == types.OrderTypeMM {
		numMMOrders, _ = k.GetNumMMOrders(ctx, ordererAddr, marketId)
		maxNumMMOrders = k.GetMaxNumMMOrders(ctx)
		if numMMOrders+1 > maxNumMMOrders {
			err = sdkerrors.Wrapf(types.ErrMaxNumMMOrdersExceeded, "%d > %d", numMMOrders+1, maxNumMMOrders)
			return
		}
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
	if !isBatch {
		res, err = k.executeOrder(
			ctx, market, ordererAddr, isBuy, &price, &qty, nil, false, false)
		if err != nil {
			return
		}
	}

	openQty := qty
	if !isBatch {
		openQty = openQty.Sub(res.ExecutedQuantity)
	}
	if isBatch || openQty.IsPositive() {
		deadline := ctx.BlockTime().Add(lifespan)
		deposit := types.DepositAmount(isBuy, price, openQty).Ceil().TruncateInt()
		order = types.NewOrder(
			orderId, typ, ordererAddr, market.Id, isBuy, price, qty,
			ctx.BlockHeight(), openQty, deposit.ToDec(), deadline)
		if err = k.EscrowCoin(ctx, market, ordererAddr, sdk.NewCoin(market.DepositDenom(isBuy), deposit), false); err != nil {
			return
		}
		k.SetOrder(ctx, order)
		k.SetOrderBookOrderIndex(ctx, order)
		k.SetOrdersByOrdererIndex(ctx, order)

		if typ == types.OrderTypeMM {
			k.SetNumMMOrders(ctx, ordererAddr, marketId, numMMOrders+1)
		}
	}
	return
}

func (k Keeper) PlaceMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Dec) (orderId uint64, res types.ExecuteOrderResult, err error) {
	if !qty.IsPositive() { // sanity check
		panic("quantity must be positive")
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	res, err = k.executeOrder(
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
		ExecutedQuantity: res.ExecutedQuantity,
		Paid:             res.Paid,
		Received:         res.Received,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) CancelOrder(ctx sdk.Context, ordererAddr sdk.AccAddress, orderId uint64) (order types.Order, err error) {
	var found bool
	order, found = k.GetOrder(ctx, orderId)
	if !found {
		return order, sdkerrors.Wrap(sdkerrors.ErrNotFound, "order not found")
	}
	if order.MsgHeight == ctx.BlockHeight() {
		return order, sdkerrors.Wrap(
			sdkerrors.ErrInvalidRequest, "cannot cancel order placed in the same block")
	}
	market := k.MustGetMarket(ctx, order.MarketId)
	if ordererAddr.String() != order.Orderer {
		return order, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "order is not created by the sender")
	}
	if err = k.cancelOrder(ctx, market, order, false); err != nil {
		return order, err
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventCancelOrder{
		Orderer: ordererAddr.String(),
		OrderId: orderId,
	}); err != nil {
		return order, err
	}
	return order, nil
}

func (k Keeper) CancelAllOrders(ctx sdk.Context, ordererAddr sdk.AccAddress, marketId uint64) (orders []types.Order, err error) {
	market, found := k.GetMarket(ctx, marketId)
	if !found {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}
	var cancelledOrderIds []uint64
	k.IterateOrdersByOrdererAndMarket(ctx, ordererAddr, market.Id, func(order types.Order) (stop bool) {
		if order.MsgHeight == ctx.BlockHeight() {
			return false
		}
		if err = k.cancelOrder(ctx, market, order, true); err != nil {
			return true
		}
		orders = append(orders, order)
		cancelledOrderIds = append(cancelledOrderIds, order.Id)
		return false
	})
	if err != nil {
		return nil, err
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventCancelAllOrders{
		Orderer:           ordererAddr.String(),
		MarketId:          marketId,
		CancelledOrderIds: cancelledOrderIds,
	}); err != nil {
		return nil, err
	}
	return orders, nil
}

func (k Keeper) cancelOrder(ctx sdk.Context, market types.Market, order types.Order, queueSend bool) error {
	ordererAddr := order.MustGetOrdererAddress()
	deposit, _ := market.DepositCoin(order.IsBuy, order.RemainingDeposit).TruncateDecimal()
	if err := k.ReleaseCoin(
		ctx, market, ordererAddr, deposit, queueSend); err != nil {
		return err
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
	k.DeleteOrderBookOrderIndex(ctx, order)
	k.DeleteOrdersByOrdererIndex(ctx, order)
	return nil
}

func (k Keeper) CancelExpiredOrders(ctx sdk.Context) (err error) {
	blockTime := ctx.BlockTime()
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		// TODO: optimize by using timestamp queue
		k.IterateOrdersByMarket(ctx, market.Id, func(order types.Order) (stop bool) {
			if !blockTime.Before(order.Deadline) {
				if err = k.cancelOrder(ctx, market, order, true); err != nil {
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
	if err != nil {
		return err
	}
	return k.ExecuteSendCoins(ctx)
}
