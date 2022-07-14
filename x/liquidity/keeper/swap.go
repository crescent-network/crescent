package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func (k Keeper) PriceLimits(ctx sdk.Context, lastPrice sdk.Dec) (lowest, highest sdk.Dec) {
	return types.PriceLimits(lastPrice, k.GetMaxPriceLimitRatio(ctx), int(k.GetTickPrecision(ctx)))
}

// ValidateMsgLimitOrder validates types.MsgLimitOrder with state and returns
// calculated offer coin and price that is fit into ticks.
func (k Keeper) ValidateMsgLimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	spendable := k.bankKeeper.SpendableCoins(ctx, msg.GetOrderer())
	if spendableAmt := spendable.AmountOf(msg.OfferCoin.Denom); spendableAmt.LT(msg.OfferCoin.Amount) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "%s is smaller than %s",
			sdk.NewCoin(msg.OfferCoin.Denom, spendableAmt), msg.OfferCoin)
	}

	tickPrec := k.GetTickPrecision(ctx)
	maxOrderLifespan := k.GetMaxOrderLifespan(ctx)

	if msg.OrderLifespan > maxOrderLifespan {
		return sdk.Coin{}, sdk.Dec{},
			sdkerrors.Wrapf(types.ErrTooLongOrderLifespan, "%s is longer than %s", msg.OrderLifespan, maxOrderLifespan)
	}

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", msg.PairId)
	}

	var upperPriceLimit, lowerPriceLimit sdk.Dec
	if pair.LastPrice != nil {
		lowerPriceLimit, upperPriceLimit = k.PriceLimits(ctx, *pair.LastPrice)
	} else {
		upperPriceLimit = amm.HighestTick(int(tickPrec))
		lowerPriceLimit = amm.LowestTick(int(tickPrec))
	}
	switch {
	case msg.Price.GT(upperPriceLimit):
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPriceOutOfRange, "%s is higher than %s", msg.Price, upperPriceLimit)
	case msg.Price.LT(lowerPriceLimit):
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPriceOutOfRange, "%s is lower than %s", msg.Price, lowerPriceLimit)
	}

	switch msg.Direction {
	case types.OrderDirectionBuy:
		if msg.OfferCoin.Denom != pair.QuoteCoinDenom || msg.DemandCoinDenom != pair.BaseCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToDownTick(msg.Price, int(tickPrec))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
		if msg.OfferCoin.IsLT(offerCoin) {
			return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
				types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, offerCoin)
		}
	case types.OrderDirectionSell:
		if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToUpTick(msg.Price, int(tickPrec))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
		if msg.OfferCoin.Amount.LT(msg.Amount) {
			return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
				types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount))
		}
	}
	if types.IsTooSmallOrderAmount(msg.Amount, price) {
		return sdk.Coin{}, sdk.Dec{}, types.ErrTooSmallOrder
	}

	return offerCoin, price, nil
}

// LimitOrder handles types.MsgLimitOrder and stores types.Order.
func (k Keeper) LimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (types.Order, error) {
	offerCoin, price, err := k.ValidateMsgLimitOrder(ctx, msg)
	if err != nil {
		return types.Order{}, err
	}

	refundedCoin := msg.OfferCoin.Sub(offerCoin)
	pair, _ := k.GetPair(ctx, msg.PairId)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(offerCoin)); err != nil {
		return types.Order{}, err
	}

	requestId := k.getNextOrderIdWithUpdate(ctx, pair)
	expireAt := ctx.BlockTime().Add(msg.OrderLifespan)
	order := types.NewOrderForLimitOrder(msg, requestId, pair, offerCoin, price, expireAt, ctx.BlockHeight())
	k.SetOrder(ctx, order)
	k.SetOrderIndex(ctx, order)

	ctx.GasMeter().ConsumeGas(k.GetOrderExtraGas(ctx), "OrderExtraGas")

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLimitOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, offerCoin.String()),
			sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, price.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(order.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(order.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyExpireAt, order.ExpireAt.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyRefundedCoins, refundedCoin.String()),
		),
	})

	return order, nil
}

// ValidateMsgMarketOrder validates types.MsgMarketOrder with state and returns
// calculated offer coin and price.
func (k Keeper) ValidateMsgMarketOrder(ctx sdk.Context, msg *types.MsgMarketOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	spendable := k.bankKeeper.SpendableCoins(ctx, msg.GetOrderer())
	if spendableAmt := spendable.AmountOf(msg.OfferCoin.Denom); spendableAmt.LT(msg.OfferCoin.Amount) {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds, "%s is smaller than %s",
			sdk.NewCoin(msg.OfferCoin.Denom, spendableAmt), msg.OfferCoin)
	}

	maxOrderLifespan := k.GetMaxOrderLifespan(ctx)
	maxPriceLimitRatio := k.GetMaxPriceLimitRatio(ctx)
	tickPrec := k.GetTickPrecision(ctx)

	if msg.OrderLifespan > maxOrderLifespan {
		return sdk.Coin{}, sdk.Dec{},
			sdkerrors.Wrapf(types.ErrTooLongOrderLifespan, "%s is longer than %s", msg.OrderLifespan, maxOrderLifespan)
	}

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", msg.PairId)
	}

	if pair.LastPrice == nil {
		return sdk.Coin{}, sdk.Dec{}, types.ErrNoLastPrice
	}
	lastPrice := *pair.LastPrice

	switch msg.Direction {
	case types.OrderDirectionBuy:
		if msg.OfferCoin.Denom != pair.QuoteCoinDenom || msg.DemandCoinDenom != pair.BaseCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToDownTick(lastPrice.Mul(sdk.OneDec().Add(maxPriceLimitRatio)), int(tickPrec))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
		if msg.OfferCoin.IsLT(offerCoin) {
			return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
				types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, offerCoin)
		}
	case types.OrderDirectionSell:
		if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToUpTick(lastPrice.Mul(sdk.OneDec().Sub(maxPriceLimitRatio)), int(tickPrec))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount)
		if msg.OfferCoin.Amount.LT(msg.Amount) {
			return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(
				types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, sdk.NewCoin(msg.OfferCoin.Denom, msg.Amount))
		}
	}
	if types.IsTooSmallOrderAmount(msg.Amount, price) {
		return sdk.Coin{}, sdk.Dec{}, types.ErrTooSmallOrder
	}

	return offerCoin, price, nil
}

// MarketOrder handles types.MsgMarketOrder and stores types.Order.
func (k Keeper) MarketOrder(ctx sdk.Context, msg *types.MsgMarketOrder) (types.Order, error) {
	offerCoin, price, err := k.ValidateMsgMarketOrder(ctx, msg)
	if err != nil {
		return types.Order{}, err
	}

	refundedCoin := msg.OfferCoin.Sub(offerCoin)
	pair, _ := k.GetPair(ctx, msg.PairId)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(offerCoin)); err != nil {
		return types.Order{}, err
	}

	requestId := k.getNextOrderIdWithUpdate(ctx, pair)
	expireAt := ctx.BlockTime().Add(msg.OrderLifespan)
	order := types.NewOrderForMarketOrder(msg, requestId, pair, offerCoin, price, expireAt, ctx.BlockHeight())
	k.SetOrder(ctx, order)
	k.SetOrderIndex(ctx, order)

	ctx.GasMeter().ConsumeGas(k.GetOrderExtraGas(ctx), "OrderExtraGas")

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMarketOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, offerCoin.String()),
			sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, price.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(order.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(order.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyExpireAt, order.ExpireAt.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyRefundedCoins, refundedCoin.String()),
		),
	})

	return order, nil
}

// ValidateMsgCancelOrder validates types.MsgCancelOrder and returns the order.
func (k Keeper) ValidateMsgCancelOrder(ctx sdk.Context, msg *types.MsgCancelOrder) (order types.Order, err error) {
	var found bool
	order, found = k.GetOrder(ctx, msg.PairId, msg.OrderId)
	if !found {
		return types.Order{},
			sdkerrors.Wrapf(sdkerrors.ErrNotFound, "order %d not found in pair %d", msg.OrderId, msg.PairId)
	}
	if msg.Orderer != order.Orderer {
		return types.Order{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "mismatching orderer")
	}
	if order.Status == types.OrderStatusCanceled {
		return types.Order{}, types.ErrAlreadyCanceled
	}
	pair, _ := k.GetPair(ctx, msg.PairId)
	if order.BatchId == pair.CurrentBatchId {
		return types.Order{}, types.ErrSameBatch
	}
	return order, nil
}

// CancelOrder handles types.MsgCancelOrder and cancels an order.
func (k Keeper) CancelOrder(ctx sdk.Context, msg *types.MsgCancelOrder) error {
	order, err := k.ValidateMsgCancelOrder(ctx, msg)
	if err != nil {
		return err
	}

	if err := k.FinishOrder(ctx, order, types.OrderStatusCanceled); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(msg.OrderId, 10)),
		),
	})

	return nil
}

// CancelAllOrders handles types.MsgCancelAllOrders and cancels all orders.
func (k Keeper) CancelAllOrders(ctx sdk.Context, msg *types.MsgCancelAllOrders) error {
	orderPairCache := map[uint64]types.Pair{} // maps order's pair id to pair, to cache the result
	pairIdSet := map[uint64]struct{}{}        // set of pairs where to cancel orders
	var pairIds []string                      // needed to emit an event
	for _, pairId := range msg.PairIds {
		pair, found := k.GetPair(ctx, pairId)
		if !found { // check if the pair exists
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", pairId)
		}
		pairIdSet[pairId] = struct{}{} // add pair id to the set
		pairIds = append(pairIds, strconv.FormatUint(pairId, 10))
		orderPairCache[pairId] = pair // also cache the pair to use at below
	}

	var canceledOrderIds []string
	if err := k.IterateOrdersByOrderer(ctx, msg.GetOrderer(), func(order types.Order) (stop bool, err error) {
		_, ok := pairIdSet[order.PairId] // is the pair included in the pair set?
		if len(pairIdSet) == 0 || ok {   // pair ids not specified(cancel all), or the pair is in the set
			pair, ok := orderPairCache[order.PairId]
			if !ok {
				pair, _ = k.GetPair(ctx, order.PairId)
				orderPairCache[order.PairId] = pair
			}
			if order.Status != types.OrderStatusCanceled && order.BatchId < pair.CurrentBatchId {
				if err := k.FinishOrder(ctx, order, types.OrderStatusCanceled); err != nil {
					return false, err
				}
				canceledOrderIds = append(canceledOrderIds, strconv.FormatUint(order.Id, 10))
			}
		}
		return false, nil
	}); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelAllOrders,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairIds, strings.Join(pairIds, ",")),
			sdk.NewAttribute(types.AttributeKeyCanceledOrderIds, strings.Join(canceledOrderIds, ",")),
		),
	})

	return nil
}

func (k Keeper) ExecuteMatching(ctx sdk.Context, pair types.Pair) error {
	ob := amm.NewOrderBook()

	if err := k.IterateOrdersByPair(ctx, pair.Id, func(order types.Order) (stop bool, err error) {
		switch order.Status {
		case types.OrderStatusNotExecuted,
			types.OrderStatusNotMatched,
			types.OrderStatusPartiallyMatched:
			if order.Status != types.OrderStatusNotExecuted && order.ExpiredAt(ctx.BlockTime()) {
				if err := k.FinishOrder(ctx, order, types.OrderStatusExpired); err != nil {
					return false, err
				}
				return false, nil
			}
			// TODO: add orders only when price is in the range?
			ob.AddOrder(types.NewUserOrder(order))
			if order.Status == types.OrderStatusNotExecuted {
				order.SetStatus(types.OrderStatusNotMatched)
				k.SetOrder(ctx, order)
			}
		case types.OrderStatusCanceled:
		default:
			return false, fmt.Errorf("invalid order status: %s", order.Status)
		}
		return false, nil
	}); err != nil {
		return err
	}

	var pools []*types.PoolOrderer
	_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
		if pool.Disabled {
			return false, nil
		}
		rx, ry := k.getPoolBalances(ctx, pool, pair)
		ps := k.GetPoolCoinSupply(ctx, pool)
		ammPool := types.NewPoolOrderer(
			pool.AMMPool(rx.Amount, ry.Amount, ps),
			pool.Id, pool.GetReserveAddress(), pair.BaseCoinDenom, pair.QuoteCoinDenom)
		if ammPool.IsDepleted() {
			k.MarkPoolAsDisabled(ctx, pool)
			return false, nil
		}
		pools = append(pools, ammPool)
		return false, nil
	})

	matchPrice, quoteCoinDiff, matched := k.Match(ctx, ob, pools, pair.LastPrice)
	if matched {
		orders := ob.Orders()
		if err := k.ApplyMatchResult(ctx, pair, orders, quoteCoinDiff); err != nil {
			return err
		}
		pair.LastPrice = &matchPrice
	}

	pair.CurrentBatchId++
	k.SetPair(ctx, pair)

	return nil
}

func (k Keeper) Match(ctx sdk.Context, ob *amm.OrderBook, pools []*types.PoolOrderer, lastPrice *sdk.Dec) (matchPrice sdk.Dec, quoteCoinDiff sdk.Int, matched bool) {
	tickPrec := int(k.GetTickPrecision(ctx))
	if lastPrice == nil {
		ov := amm.MultipleOrderViews{ob.MakeView()}
		for _, pool := range pools {
			ov = append(ov, pool)
		}
		var found bool
		matchPrice, found = amm.FindMatchPrice(ov, tickPrec)
		if !found {
			return sdk.Dec{}, sdk.Int{}, false
		}
		for _, pool := range pools {
			buyAmt := pool.BuyAmountOver(matchPrice, true)
			if buyAmt.IsPositive() {
				ob.AddOrder(pool.Order(amm.Buy, matchPrice, buyAmt))
			}
			sellAmt := pool.SellAmountUnder(matchPrice, true)
			if sellAmt.IsPositive() {
				ob.AddOrder(pool.Order(amm.Sell, matchPrice, sellAmt))
			}
		}
		quoteCoinDiff, matched = ob.MatchAtSinglePrice(matchPrice)
	} else {
		lowestPrice, highestPrice := k.PriceLimits(ctx, *lastPrice)
		for _, pool := range pools {
			poolOrders := amm.PoolOrders(pool, pool, lowestPrice, highestPrice, tickPrec)
			ob.AddOrder(poolOrders...)
		}
		matchPrice, quoteCoinDiff, matched = ob.Match(*lastPrice)
	}
	return
}

func (k Keeper) ApplyMatchResult(ctx sdk.Context, pair types.Pair, orders []amm.Order, quoteCoinDiff sdk.Int) error {
	bulkOp := types.NewBulkSendCoinsOperation()
	for _, order := range orders { // TODO: need optimization to filter matched orders only
		order, ok := order.(*types.PoolOrder)
		if !ok {
			continue
		}
		if !order.IsMatched() {
			continue
		}
		paidCoin := sdk.NewCoin(order.OfferCoinDenom, order.PaidOfferCoinAmount)
		bulkOp.QueueSendCoins(order.ReserveAddress, pair.GetEscrowAddress(), sdk.NewCoins(paidCoin))
	}
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}
	bulkOp = types.NewBulkSendCoinsOperation()
	type PoolMatchResult struct {
		PoolId         uint64
		OrderDirection types.OrderDirection
		PaidCoin       sdk.Coin
		ReceivedCoin   sdk.Coin
		MatchedAmount  sdk.Int
	}
	poolMatchResultById := map[uint64]*PoolMatchResult{}
	var poolMatchResults []*PoolMatchResult
	for _, order := range orders {
		if !order.IsMatched() {
			continue
		}

		matchedAmt := order.GetAmount().Sub(order.GetOpenAmount())

		switch order := order.(type) {
		case *types.UserOrder:
			paidCoin := sdk.NewCoin(order.OfferCoinDenom, order.PaidOfferCoinAmount)
			receivedCoin := sdk.NewCoin(order.DemandCoinDenom, order.ReceivedDemandCoinAmount)

			o, _ := k.GetOrder(ctx, pair.Id, order.OrderId)
			o.OpenAmount = o.OpenAmount.Sub(matchedAmt)
			o.RemainingOfferCoin = o.RemainingOfferCoin.Sub(paidCoin)
			o.ReceivedCoin = o.ReceivedCoin.Add(receivedCoin)

			if o.OpenAmount.IsZero() {
				if err := k.FinishOrder(ctx, o, types.OrderStatusCompleted); err != nil {
					return err
				}
			} else {
				o.SetStatus(types.OrderStatusPartiallyMatched)
				k.SetOrder(ctx, o)
			}
			bulkOp.QueueSendCoins(pair.GetEscrowAddress(), order.Orderer, sdk.NewCoins(receivedCoin))

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeUserOrderMatched,
					sdk.NewAttribute(types.AttributeKeyOrderDirection, types.OrderDirectionFromAMM(order.Direction).String()),
					sdk.NewAttribute(types.AttributeKeyOrderer, order.Orderer.String()),
					sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(pair.Id, 10)),
					sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(order.OrderId, 10)),
					sdk.NewAttribute(types.AttributeKeyMatchedAmount, matchedAmt.String()),
					sdk.NewAttribute(types.AttributeKeyPaidCoin, paidCoin.String()),
					sdk.NewAttribute(types.AttributeKeyReceivedCoin, receivedCoin.String()),
				),
			})
		case *types.PoolOrder:
			paidCoin := sdk.NewCoin(order.OfferCoinDenom, order.PaidOfferCoinAmount)
			receivedCoin := sdk.NewCoin(order.DemandCoinDenom, order.ReceivedDemandCoinAmount)

			bulkOp.QueueSendCoins(pair.GetEscrowAddress(), order.ReserveAddress, sdk.NewCoins(receivedCoin))

			r, ok := poolMatchResultById[order.PoolId]
			if !ok {
				r = &PoolMatchResult{
					PoolId:         order.PoolId,
					OrderDirection: types.OrderDirectionFromAMM(order.Direction),
					PaidCoin:       sdk.NewCoin(paidCoin.Denom, sdk.ZeroInt()),
					ReceivedCoin:   sdk.NewCoin(receivedCoin.Denom, sdk.ZeroInt()),
					MatchedAmount:  sdk.ZeroInt(),
				}
				poolMatchResultById[order.PoolId] = r
				poolMatchResults = append(poolMatchResults, r)
			}
			dir := types.OrderDirectionFromAMM(order.Direction)
			if r.OrderDirection != dir {
				panic(fmt.Errorf("wrong order direction: %s != %s", dir, r.OrderDirection))
			}
			r.PaidCoin = r.PaidCoin.Add(paidCoin)
			r.ReceivedCoin = r.ReceivedCoin.Add(receivedCoin)
			r.MatchedAmount = r.MatchedAmount.Add(matchedAmt)
		default:
			panic(fmt.Errorf("invalid order type: %T", order))
		}
	}
	bulkOp.QueueSendCoins(pair.GetEscrowAddress(), k.GetDustCollector(ctx), sdk.NewCoins(sdk.NewCoin(pair.QuoteCoinDenom, quoteCoinDiff)))
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}
	for _, r := range poolMatchResults {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypePoolOrderMatched,
				sdk.NewAttribute(types.AttributeKeyOrderDirection, r.OrderDirection.String()),
				sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(pair.Id, 10)),
				sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(r.PoolId, 10)),
				sdk.NewAttribute(types.AttributeKeyMatchedAmount, r.MatchedAmount.String()),
				sdk.NewAttribute(types.AttributeKeyPaidCoin, r.PaidCoin.String()),
				sdk.NewAttribute(types.AttributeKeyReceivedCoin, r.ReceivedCoin.String()),
			),
		})
	}
	return nil
}

func (k Keeper) FinishOrder(ctx sdk.Context, order types.Order, status types.OrderStatus) error {
	if order.Status == types.OrderStatusCompleted || order.Status.IsCanceledOrExpired() { // sanity check
		return nil
	}

	if order.RemainingOfferCoin.IsPositive() {
		pair, _ := k.GetPair(ctx, order.PairId)
		if err := k.bankKeeper.SendCoins(ctx, pair.GetEscrowAddress(), order.GetOrderer(), sdk.NewCoins(order.RemainingOfferCoin)); err != nil {
			return err
		}
	}

	order.SetStatus(status)
	k.SetOrder(ctx, order)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOrderResult,
			sdk.NewAttribute(types.AttributeKeyOrderDirection, order.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOrderer, order.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(order.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderId, strconv.FormatUint(order.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyAmount, order.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyOpenAmount, order.OpenAmount.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, order.OfferCoin.String()),
			sdk.NewAttribute(types.AttributeKeyRemainingOfferCoin, order.RemainingOfferCoin.String()),
			sdk.NewAttribute(types.AttributeKeyReceivedCoin, order.ReceivedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyStatus, order.Status.String()),
		),
	})

	return nil
}
