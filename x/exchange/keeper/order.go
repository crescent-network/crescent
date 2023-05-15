package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
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
	if !isTemp {
		k.SetOrder(ctx, order)
		k.SetOrderBookOrder(ctx, order)
	}
	return order, nil
}

func (k Keeper) executeOrder(
	ctx sdk.Context, market types.Market, ordererAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qtyLimit, quoteLimit *sdk.Int, simulate bool) (totalExecQty, totalExecQuote sdk.Int) {
	if qtyLimit == nil && quoteLimit == nil { // sanity check
		panic("quantity limit and quote limit cannot be set to nil at the same time")
	}
	if simulate {
		ctx, _ = ctx.CacheContext()
	}
	var lastPrice sdk.Dec
	totalExecQty = utils.ZeroInt
	totalExecQuote = utils.ZeroInt
	obs := k.ConstructTempOrderBookSide(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit)
	var sourceNames []string
	resultsBySourceName := map[string][]types.TempOrder{}
	for _, level := range obs.Levels {
		if priceLimit != nil &&
			((isBuy && level.Price.GT(*priceLimit)) ||
				(!isBuy && level.Price.LT(*priceLimit))) {
			break
		}
		if qtyLimit != nil && !qtyLimit.Sub(totalExecQty).IsPositive() {
			break
		}
		if quoteLimit != nil && !quoteLimit.Sub(totalExecQuote).IsPositive() {
			break
		}

		executableQty := types.TotalExecutableQuantity(level.Orders, level.Price)
		execQty := executableQty
		if qtyLimit != nil {
			execQty = utils.MinInt(execQty, qtyLimit.Sub(totalExecQty))
		}
		if quoteLimit != nil {
			execQty = utils.MinInt(
				execQty,
				quoteLimit.Sub(totalExecQuote).ToDec().QuoTruncate(level.Price).TruncateInt())
		}

		market.FillTempOrderBookLevel(level, execQty, level.Price, true)
		execQuote := types.QuoteAmount(isBuy, level.Price, execQty)
		// TODO: refactor code
		if isBuy {
			if err := k.EscrowCoin(ctx, market, ordererAddr, sdk.NewCoin(market.QuoteDenom, execQuote), true); err != nil {
				panic(err)
			}
			receive := utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQty).TruncateInt()
			if err := k.ReleaseCoin(ctx, market, ordererAddr, sdk.NewCoin(market.BaseDenom, receive), true); err != nil {
				panic(err)
			}
		} else {
			if err := k.EscrowCoin(ctx, market, ordererAddr, sdk.NewCoin(market.BaseDenom, execQty), true); err != nil {
				panic(err)
			}
			receive := utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQuote).TruncateInt()
			if err := k.ReleaseCoin(ctx, market, ordererAddr, sdk.NewCoin(market.QuoteDenom, receive), true); err != nil {
				panic(err)
			}
		}
		totalExecQty = totalExecQty.Add(execQty)
		totalExecQuote = totalExecQuote.Add(execQuote)

		if !simulate {
			for _, order := range level.Orders {
				if order.IsUpdated && order.Source != nil {
					sourceName := order.Source.Name()
					results, ok := resultsBySourceName[sourceName]
					if !ok {
						sourceNames = append(sourceNames, sourceName)
					}
					resultsBySourceName[sourceName] = append(results, *order)
				}
			}
			lastPrice = level.Price
		}
	}
	if !simulate {
		if err := k.ApplyTempOrderBookSideChanges(ctx, market, obs); err != nil {
			panic(err)
		}
		if err := k.ExecuteSendCoins(ctx); err != nil {
			panic(err) // TODO: return error
		}
		for _, sourceName := range sourceNames {
			results := resultsBySourceName[sourceName]
			if len(results) > 0 {
				source := k.sources[sourceName]
				source.AfterOrdersExecuted(ctx, market, results)
			}
		}
		if !lastPrice.IsNil() {
			state := k.MustGetMarketState(ctx, market.Id)
			state.LastPrice = &lastPrice
			k.SetMarketState(ctx, market.Id, state)
		}
	}
	return
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
