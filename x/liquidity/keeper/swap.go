package keeper

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmosquad-labs/squad/x/liquidity/amm"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

// ValidateMsgLimitOrder validates types.MsgLimitOrder with state and returns
// calculated offer coin and price that is fit into ticks.
func (k Keeper) ValidateMsgLimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	params := k.GetParams(ctx)

	if msg.OrderLifespan > params.MaxOrderLifespan {
		return sdk.Coin{}, sdk.Dec{},
			sdkerrors.Wrapf(types.ErrTooLongOrderLifespan, "%s is longer than %s", msg.OrderLifespan, params.MaxOrderLifespan)
	}

	pair, found := k.GetPair(ctx, msg.PairId)
	if !found {
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", msg.PairId)
	}

	var upperPriceLimit, lowerPriceLimit sdk.Dec
	if pair.LastPrice != nil {
		lastPrice := *pair.LastPrice
		upperPriceLimit = lastPrice.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio))
		lowerPriceLimit = lastPrice.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio))
	} else {
		upperPriceLimit = amm.HighestTick(int(params.TickPrecision))
		lowerPriceLimit = amm.LowestTick(int(params.TickPrecision))
	}
	switch {
	case msg.Price.GT(upperPriceLimit):
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPriceOutOfRange, "%s is higher than %s", msg.Price, upperPriceLimit)
	case msg.Price.LT(lowerPriceLimit):
		return sdk.Coin{}, sdk.Dec{}, sdkerrors.Wrapf(types.ErrPriceOutOfRange, "%s is lower than %s", msg.Price, lowerPriceLimit)
	}

	switch msg.Direction {
	case types.SwapDirectionBuy:
		if msg.OfferCoin.Denom != pair.QuoteCoinDenom || msg.DemandCoinDenom != pair.BaseCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToDownTick(msg.Price, int(params.TickPrecision))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
	case types.SwapDirectionSell:
		if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToUpTick(msg.Price, int(params.TickPrecision))
		offerCoin = msg.OfferCoin
	}
	return offerCoin, price, nil
}

// LimitOrder handles types.MsgLimitOrder and stores types.SwapRequest.
func (k Keeper) LimitOrder(ctx sdk.Context, msg *types.MsgLimitOrder) (types.SwapRequest, error) {
	offerCoin, price, err := k.ValidateMsgLimitOrder(ctx, msg)
	if err != nil {
		return types.SwapRequest{}, err
	}

	refundedCoin := msg.OfferCoin.Sub(offerCoin)
	pair, _ := k.GetPair(ctx, msg.PairId)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(offerCoin)); err != nil {
		return types.SwapRequest{}, err
	}

	requestId := k.GetNextSwapRequestIdWithUpdate(ctx, pair)
	expireAt := ctx.BlockTime().Add(msg.OrderLifespan)
	req := types.NewSwapRequestForLimitOrder(msg, requestId, pair, offerCoin, price, expireAt, ctx.BlockHeight())
	k.SetSwapRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLimitOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, offerCoin.String()),
			sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, price.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(req.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyExpireAt, req.ExpireAt.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyRefundedCoins, refundedCoin.String()),
		),
	})

	return req, nil
}

// ValidateMsgMarketOrder validates types.MsgMarketOrder with state and returns
// calculated offer coin and price.
func (k Keeper) ValidateMsgMarketOrder(ctx sdk.Context, msg *types.MsgMarketOrder) (offerCoin sdk.Coin, price sdk.Dec, err error) {
	params := k.GetParams(ctx)

	if msg.OrderLifespan > params.MaxOrderLifespan {
		return sdk.Coin{}, sdk.Dec{},
			sdkerrors.Wrapf(types.ErrTooLongOrderLifespan, "%s is longer than %s", msg.OrderLifespan, params.MaxOrderLifespan)
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
	case types.SwapDirectionBuy:
		if msg.OfferCoin.Denom != pair.QuoteCoinDenom || msg.DemandCoinDenom != pair.BaseCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.DemandCoinDenom, msg.OfferCoin.Denom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToDownTick(lastPrice.Mul(sdk.OneDec().Add(params.MaxPriceLimitRatio)), int(params.TickPrecision))
		offerCoin = sdk.NewCoin(msg.OfferCoin.Denom, amm.OfferCoinAmount(amm.Buy, price, msg.Amount))
	case types.SwapDirectionSell:
		if msg.OfferCoin.Denom != pair.BaseCoinDenom || msg.DemandCoinDenom != pair.QuoteCoinDenom {
			return sdk.Coin{}, sdk.Dec{},
				sdkerrors.Wrapf(types.ErrWrongPair, "denom pair (%s, %s) != (%s, %s)",
					msg.OfferCoin.Denom, msg.DemandCoinDenom, pair.BaseCoinDenom, pair.QuoteCoinDenom)
		}
		price = amm.PriceToUpTick(lastPrice.Mul(sdk.OneDec().Sub(params.MaxPriceLimitRatio)), int(params.TickPrecision))
		offerCoin = msg.OfferCoin
	}
	if msg.OfferCoin.IsLT(offerCoin) {
		return sdk.Coin{}, sdk.Dec{},
			sdkerrors.Wrapf(types.ErrInsufficientOfferCoin, "%s is smaller than %s", msg.OfferCoin, offerCoin)
	}

	return offerCoin, price, nil
}

// MarketOrder handles types.MsgMarketOrder and stores types.SwapRequest.
func (k Keeper) MarketOrder(ctx sdk.Context, msg *types.MsgMarketOrder) (types.SwapRequest, error) {
	offerCoin, price, err := k.ValidateMsgMarketOrder(ctx, msg)
	if err != nil {
		return types.SwapRequest{}, err
	}

	refundedCoin := msg.OfferCoin.Sub(offerCoin)
	pair, _ := k.GetPair(ctx, msg.PairId)
	if err := k.bankKeeper.SendCoins(ctx, msg.GetOrderer(), pair.GetEscrowAddress(), sdk.NewCoins(offerCoin)); err != nil {
		return types.SwapRequest{}, err
	}

	requestId := k.GetNextSwapRequestIdWithUpdate(ctx, pair)
	expireAt := ctx.BlockTime().Add(msg.OrderLifespan)
	req := types.NewSwapRequestForMarketOrder(msg, requestId, pair, offerCoin, price, expireAt, ctx.BlockHeight())
	k.SetSwapRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMarketOrder,
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapDirection, msg.Direction.String()),
			sdk.NewAttribute(types.AttributeKeyOfferCoin, msg.OfferCoin.String()),
			sdk.NewAttribute(types.AttributeKeyDemandCoinDenom, msg.DemandCoinDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, price.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyBatchId, strconv.FormatUint(req.BatchId, 10)),
			sdk.NewAttribute(types.AttributeKeyExpireAt, req.ExpireAt.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyRefundedCoins, refundedCoin.String()),
		),
	})

	return req, nil
}

// ValidateMsgCancelOrder validates types.MsgCancelOrder and returns the
// swap request.
func (k Keeper) ValidateMsgCancelOrder(ctx sdk.Context, msg *types.MsgCancelOrder) (swapReq types.SwapRequest, err error) {
	var found bool
	swapReq, found = k.GetSwapRequest(ctx, msg.PairId, msg.SwapRequestId)
	if !found {
		return types.SwapRequest{},
			sdkerrors.Wrapf(sdkerrors.ErrNotFound, "swap request %d not found in pair %d", msg.SwapRequestId, msg.PairId)
	}
	if msg.Orderer != swapReq.Orderer {
		return types.SwapRequest{}, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "mismatching orderer")
	}
	if swapReq.Status == types.SwapRequestStatusCanceled {
		return types.SwapRequest{}, types.ErrAlreadyCanceled
	}
	pair, _ := k.GetPair(ctx, msg.PairId)
	if swapReq.BatchId == pair.CurrentBatchId {
		return types.SwapRequest{}, types.ErrSameBatch
	}
	return swapReq, nil
}

// CancelOrder handles types.MsgCancelOrder and cancels an order.
func (k Keeper) CancelOrder(ctx sdk.Context, msg *types.MsgCancelOrder) error {
	swapReq, err := k.ValidateMsgCancelOrder(ctx, msg)
	if err != nil {
		return err
	}

	if err := k.FinishSwapRequest(ctx, swapReq, types.SwapRequestStatusCanceled); err != nil {
		return err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelOrder,
			sdk.NewAttribute(types.AttributeKeyOrderer, msg.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(msg.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapRequestId, strconv.FormatUint(msg.SwapRequestId, 10)),
		),
	})

	return nil
}

// CancelAllOrders handles types.MsgCancelAllOrders and cancels all orders.
func (k Keeper) CancelAllOrders(ctx sdk.Context, msg *types.MsgCancelAllOrders) error {
	cb := func(pair types.Pair, req types.SwapRequest) (stop bool, err error) {
		if req.Orderer == msg.Orderer && req.Status != types.SwapRequestStatusCanceled && req.BatchId < pair.CurrentBatchId {
			if err := k.FinishSwapRequest(ctx, req, types.SwapRequestStatusCanceled); err != nil {
				return false, err
			}
		}
		return false, nil
	}

	if len(msg.PairIds) == 0 {
		pairMap := map[uint64]types.Pair{}
		if err := k.IterateAllSwapRequests(ctx, func(req types.SwapRequest) (stop bool, err error) {
			pair, ok := pairMap[req.PairId]
			if !ok {
				pair, _ = k.GetPair(ctx, req.PairId)
				pairMap[req.PairId] = pair
			}
			return cb(pair, req)
		}); err != nil {
			return err
		}

		return nil
	}

	for _, pairId := range msg.PairIds {
		pair, found := k.GetPair(ctx, pairId)
		if !found {
			return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "pair %d not found", pairId)
		}
		if err := k.IterateSwapRequestsByPair(ctx, pairId, func(req types.SwapRequest) (stop bool, err error) {
			return cb(pair, req)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) ExecuteMatching(ctx sdk.Context, pair types.Pair) error {
	ob := amm.NewOrderBook()
	skip := true // Whether to skip the matching since there is no orders.
	if err := k.IterateSwapRequestsByPair(ctx, pair.Id, func(req types.SwapRequest) (stop bool, err error) {
		switch req.Status {
		case types.SwapRequestStatusNotExecuted,
			types.SwapRequestStatusNotMatched,
			types.SwapRequestStatusPartiallyMatched:
			if req.Status != types.SwapRequestStatusNotExecuted && !ctx.BlockTime().Before(req.ExpireAt) {
				if err := k.FinishSwapRequest(ctx, req, types.SwapRequestStatusExpired); err != nil {
					return false, err
				}
				return false, nil
			}
			ob.Add(types.NewUserOrder(req))
			if req.Status == types.SwapRequestStatusNotExecuted {
				req.Status = types.SwapRequestStatusNotMatched
				k.SetSwapRequest(ctx, req)
			}
			skip = false
		case types.SwapRequestStatusCanceled:
		default:
			return false, fmt.Errorf("invalid swap request status: %s", req.Status)
		}
		return false, nil
	}); err != nil {
		return err
	}

	if skip { // TODO: update this when there are more than one pools
		return nil
	}

	var poolOrderSources []amm.OrderSource
	_ = k.IteratePoolsByPair(ctx, pair.Id, func(pool types.Pool) (stop bool, err error) {
		rx, ry := k.GetPoolBalance(ctx, pool, pair)
		ps := k.GetPoolCoinSupply(ctx, pool)
		ammPool := amm.NewBasicPool(rx, ry, ps)
		if ammPool.IsDepleted() {
			k.MarkPoolAsDisabled(ctx, pool)
			return false, nil
		}
		poolOrderSource := types.NewPoolOrderSource(ammPool, pool.Id, pool.GetReserveAddress(), pair.BaseCoinDenom, pair.QuoteCoinDenom)
		poolOrderSources = append(poolOrderSources, poolOrderSource)
		return false, nil
	})

	os := amm.MergeOrderSources(append(poolOrderSources, ob)...)

	params := k.GetParams(ctx)
	matchPrice, found := amm.FindMatchPrice(os, int(params.TickPrecision))
	if found {
		buyOrders := os.BuyOrdersOver(matchPrice)
		sellOrders := os.SellOrdersUnder(matchPrice)

		types.SortOrders(buyOrders, types.DescendingPrice)
		types.SortOrders(sellOrders, types.AscendingPrice)

		quoteCoinDust, matched := amm.MatchOrders(buyOrders, sellOrders, matchPrice)
		if matched {
			if err := k.ApplyMatchResult(ctx, pair, append(buyOrders, sellOrders...), quoteCoinDust); err != nil {
				return err
			}
			pair.LastPrice = &matchPrice
		}
	}

	pair.CurrentBatchId++
	k.SetPair(ctx, pair)

	// TODO: emit an event?
	return nil
}

func (k Keeper) ApplyMatchResult(ctx sdk.Context, pair types.Pair, orders []amm.Order, quoteCoinDust sdk.Int) error {
	bulkOp := types.NewBulkSendCoinsOperation()
	for _, order := range orders {
		if !order.IsMatched() {
			continue
		}
		if order, ok := order.(*types.PoolOrder); ok {
			paidCoin := order.OfferCoin.Sub(order.RemainingOfferCoin)
			bulkOp.SendCoins(order.ReserveAddress, pair.GetEscrowAddress(), sdk.NewCoins(paidCoin))
		}
	}
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}
	bulkOp = types.NewBulkSendCoinsOperation()
	for _, order := range orders {
		if !order.IsMatched() {
			continue
		}
		switch order := order.(type) {
		case *types.UserOrder:
			// TODO: optimize read/write (can there be only one write?)
			req, _ := k.GetSwapRequest(ctx, pair.Id, order.RequestId)
			req.OpenAmount = order.OpenAmount
			req.RemainingOfferCoin = order.RemainingOfferCoin
			req.ReceivedCoin = req.ReceivedCoin.AddAmount(order.ReceivedDemandCoin.Amount)
			if order.OpenAmount.IsZero() {
				if err := k.FinishSwapRequest(ctx, req, types.SwapRequestStatusCompleted); err != nil {
					return err
				}
			} else {
				req.Status = types.SwapRequestStatusPartiallyMatched
				k.SetSwapRequest(ctx, req)
				// TODO: emit an event?
			}
			bulkOp.SendCoins(pair.GetEscrowAddress(), order.Orderer, sdk.NewCoins(order.ReceivedDemandCoin))
		case *types.PoolOrder:
			bulkOp.SendCoins(pair.GetEscrowAddress(), order.ReserveAddress, sdk.NewCoins(order.ReceivedDemandCoin))
		}
	}
	bulkOp.SendCoins(pair.GetEscrowAddress(), types.DustCollectorAddress, sdk.NewCoins(sdk.NewCoin(pair.QuoteCoinDenom, quoteCoinDust)))
	if err := bulkOp.Run(ctx, k.bankKeeper); err != nil {
		return err
	}
	return nil
}

func (k Keeper) FinishSwapRequest(ctx sdk.Context, req types.SwapRequest, status types.SwapRequestStatus) error {
	if req.Status == types.SwapRequestStatusCompleted || req.Status.IsCanceledOrExpired() { // sanity check
		return nil
	}

	if req.RemainingOfferCoin.IsPositive() {
		pair, _ := k.GetPair(ctx, req.PairId)
		if err := k.bankKeeper.SendCoins(ctx, pair.GetEscrowAddress(), req.GetOrderer(), sdk.NewCoins(req.RemainingOfferCoin)); err != nil {
			return err
		}
	}

	req.Status = status
	k.SetSwapRequest(ctx, req)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeOrderResult,
			sdk.NewAttribute(types.AttributeKeyRequestId, strconv.FormatUint(req.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyOrderer, req.Orderer),
			sdk.NewAttribute(types.AttributeKeyPairId, strconv.FormatUint(req.PairId, 10)),
			sdk.NewAttribute(types.AttributeKeySwapDirection, req.Direction.String()),
			// TODO: include these attributes?
			//sdk.NewAttribute(types.AttributeKeyOfferCoin, req.OfferCoin.String()),
			//sdk.NewAttribute(types.AttributeKeyAmount, req.Amount.String()),
			//sdk.NewAttribute(types.AttributeKeyOpenAmount, req.OpenAmount.String()),
			//sdk.NewAttribute(types.AttributeKeyPrice, req.Price.String()),
			sdk.NewAttribute(types.AttributeKeyRemainingOfferCoin, req.RemainingOfferCoin.String()),
			sdk.NewAttribute(types.AttributeKeyReceivedCoin, req.ReceivedCoin.String()),
			sdk.NewAttribute(types.AttributeKeyStatus, req.Status.String()),
		),
	})

	return nil
}
