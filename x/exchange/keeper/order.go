package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
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
		FeePaid:          res.FeePaid,
		FeeReceived:      res.FeeReceived,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (order types.Order, err error) {
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
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
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
		FeePaid:          res.FeePaid,
		FeeReceived:      res.FeeReceived,
	}); err != nil {
		return
	}
	return
}

func (k Keeper) PlaceMMBatchLimitOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (order types.Order, err error) {
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
	isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration, isBatch bool) (orderId uint64, order types.Order, res types.ExecuteOrderResult, err error) {
	if !price.IsPositive() { // sanity check
		panic(fmt.Sprintf("price must be positive: %s", price))
	}
	if !qty.IsPositive() { // sanity check
		panic(fmt.Sprintf("quantity must be positive: %s", qty))
	}

	if maxOrderLifespan := k.GetMaxOrderLifespan(ctx); lifespan > maxOrderLifespan {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"lifespan is longer than the maximum: %s > %s", lifespan, maxOrderLifespan)
		return
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}
	if qty.LT(market.OrderQuantityLimits.Min) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity is less than the minimum order quantity allowed: %s < %s",
			qty, market.OrderQuantityLimits.Min)
		return
	} else if qty.GT(market.OrderQuantityLimits.Max) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity is greater than the maximum order quantity allowed: %s > %s",
			qty, market.OrderQuantityLimits.Max)
		return
	}
	if quote := price.MulInt(qty).TruncateInt(); quote.LT(market.OrderQuoteLimits.Min) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity * price is less than the minimum order quote allowed: %s < %s",
			quote, market.OrderQuoteLimits.Min)
		return
	} else if quote := price.MulInt(qty).Ceil().TruncateInt(); quote.GT(market.OrderQuoteLimits.Max) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity * price is greater than the maximum order quote allowed: %s > %s",
			quote, market.OrderQuoteLimits.Max)
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
		minPrice, maxPrice := types.OrderPriceLimit(*marketState.LastPrice, maxPriceRatio)
		if isBuy && price.GT(maxPrice) {
			err = sdkerrors.Wrapf(types.ErrOrderPriceOutOfRange, "price is higher than the limit %s", maxPrice)
			return
		} else if !isBuy && price.LT(minPrice) {
			err = sdkerrors.Wrapf(types.ErrOrderPriceOutOfRange, "price is lower than the limit %s", minPrice)
			return
		}
	}

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	openQty := qty
	if !isBatch {
		res, _, err = k.executeOrder(
			ctx, market, ordererAddr, isBuy, types.MemOrderBookSideOptions{
				IsBuy:         !isBuy,
				PriceLimit:    &price,
				QuantityLimit: &qty,
			}, false, false)
		if err != nil {
			return
		}
		openQty = openQty.Sub(res.ExecutedQuantity)
	}

	if isBatch || openQty.IsPositive() {
		deadline := ctx.BlockTime().Add(lifespan)
		depositDenom, _ := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, isBuy)
		depositCoin := sdk.NewCoin(depositDenom, types.DepositAmount(isBuy, price, openQty))
		order = types.NewOrder(
			orderId, typ, ordererAddr, market.Id, isBuy, price, qty,
			ctx.BlockHeight(), openQty, depositCoin.Amount, deadline)
		if err = k.EscrowCoins(ctx, market, ordererAddr, depositCoin); err != nil {
			return
		}
		k.SetOrder(ctx, order)
		k.SetOrderBookOrderIndex(ctx, order)
		k.SetOrdersByOrdererIndex(ctx, order)

		if typ == types.OrderTypeMM {
			// NOTE: NumMMOrders might have been changed in executeOrder if the
			// orderer completed own orders.
			numMMOrders, _ = k.GetNumMMOrders(ctx, ordererAddr, marketId)
			k.SetNumMMOrders(ctx, ordererAddr, marketId, numMMOrders+1)
		}
	}
	return
}

func (k Keeper) PlaceMarketOrder(
	ctx sdk.Context, marketId uint64, ordererAddr sdk.AccAddress,
	isBuy bool, qty sdk.Int) (orderId uint64, res types.ExecuteOrderResult, err error) {
	if !qty.IsPositive() { // sanity check
		panic("quantity must be positive")
	}

	market, found := k.GetMarket(ctx, marketId)
	if !found {
		err = sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
		return
	}
	if qty.LT(market.OrderQuantityLimits.Min) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity is less than the minimum order quantity allowed: %s < %s",
			qty, market.OrderQuantityLimits.Min)
		return
	} else if qty.GT(market.OrderQuantityLimits.Max) {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"quantity is greater than the maximum order quantity allowed: %s > %s",
			qty, market.OrderQuantityLimits.Max)
		return
	}
	marketState := k.MustGetMarketState(ctx, market.Id)
	if marketState.LastPrice == nil {
		err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market has no last price")
		return
	}
	maxPriceRatio := k.GetMaxOrderPriceRatio(ctx)
	minPrice, maxPrice := types.OrderPriceLimit(*marketState.LastPrice, maxPriceRatio)

	orderId = k.GetNextOrderIdWithUpdate(ctx)
	var (
		priceLimit sdk.Dec
	)
	spendable := k.bankKeeper.SpendableCoins(ctx, ordererAddr)
	if isBuy {
		priceLimit = maxPrice
	} else {
		base := spendable.AmountOf(market.BaseDenom)
		if qty.GT(base) {
			err = sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "%s%s is smaller than %s%s",
				base, market.BaseDenom, qty, market.BaseDenom)
			return
		}
		priceLimit = minPrice
	}
	res, _, err = k.executeOrder(
		ctx, market, ordererAddr, isBuy, types.MemOrderBookSideOptions{
			IsBuy:         !isBuy,
			PriceLimit:    &priceLimit,
			QuantityLimit: &qty,
		}, false, false)
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
		FeePaid:          res.FeePaid,
		FeeReceived:      res.FeeReceived,
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
	if err = k.cancelOrder(ctx, market, order); err != nil {
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
		orders = append(orders, order)
		cancelledOrderIds = append(cancelledOrderIds, order.Id)
		return false
	})
	for _, order := range orders {
		if err = k.cancelOrder(ctx, market, order); err != nil {
			return nil, err
		}
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

func (k Keeper) cancelOrder(ctx sdk.Context, market types.Market, order types.Order) error {
	ordererAddr := order.MustGetOrdererAddress()
	depositDenom, _ := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, order.IsBuy)
	refunded := sdk.NewCoin(depositDenom, order.RemainingDeposit)
	if err := k.ReleaseCoins(ctx, market, ordererAddr, refunded); err != nil {
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
	var markets []types.Market
	k.IterateAllMarkets(ctx, func(market types.Market) (stop bool) {
		markets = append(markets, market)
		return false
	})
	for _, market := range markets {
		// TODO: optimize by using timestamp queue
		var ordersToBeCanceled []types.Order
		k.IterateOrdersByMarket(ctx, market.Id, func(order types.Order) (stop bool) {
			if !blockTime.Before(order.Deadline) {
				ordersToBeCanceled = append(ordersToBeCanceled, order)
			}
			return false
		})

		for _, order := range ordersToBeCanceled {
			if err = k.cancelOrder(ctx, market, order); err != nil {
				return err
			}
			if err = ctx.EventManager().EmitTypedEvent(&types.EventOrderExpired{
				OrderId: order.Id,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) CollectFees(ctx sdk.Context, market types.Market) error {
	var deposits sdk.Coins
	k.IterateOrdersByMarket(ctx, market.Id, func(order types.Order) (stop bool) {
		payDenom, _ := types.PayReceiveDenoms(market.BaseDenom, market.QuoteDenom, order.IsBuy)
		deposit := sdk.NewCoin(payDenom, order.RemainingDeposit)
		deposits = deposits.Add(deposit)
		return false
	})
	escrowAddr := market.MustGetEscrowAddress()
	feeCollectorAddr := market.MustGetFeeCollectorAddress()
	escrowBalances := k.bankKeeper.SpendableCoins(ctx, escrowAddr)
	fees := escrowBalances.Sub(deposits)
	if fees.IsAllPositive() {
		if err := k.bankKeeper.SendCoins(ctx, escrowAddr, feeCollectorAddr, fees); err != nil {
			return err
		}
	}
	return nil
}
