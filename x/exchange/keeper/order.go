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
			orderId := k.GetNextOrderIdWithUpdate(ctx)
			deposit := types.DepositAmount(isBuy, *priceLimit, openQty)
			order = types.NewOrder(
				orderId, ordererAddr, market.Id, isBuy, *priceLimit, qty, openQty, deposit)
			if err = k.EscrowCoin(ctx, market, ordererAddr, market.DepositCoin(isBuy, deposit)); err != nil {
				return
			}
			k.SetOrder(ctx, order)
			k.SetOrderBookOrder(ctx, order)
		}
	}
	return
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
	k.constructTransientOrderBook(ctx, market, !isBuy, priceLimit, qtyLimit, quoteLimit)
	k.IterateTransientOrderBookSide(ctx, market.Id, !isBuy, func(order types.TransientOrder) (stop bool) {
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

		lastPrice = order.Order.Price

		if !simulate {
			var (
				paidCoin           sdk.Coin
				receivedCoin       sdk.Coin
				makerPaid          sdk.Int
				makerReceivedCoins sdk.Coins
			)
			if isBuy {
				paidCoin = sdk.NewCoin(market.QuoteDenom, execQuote)
				receivedCoin = sdk.NewCoin(
					market.BaseDenom,
					utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQty).TruncateInt())
				makerPaid = execQty
				if market.MakerFeeRate.IsPositive() {
					makerReceivedCoins = sdk.NewCoins(
						sdk.NewCoin(
							market.QuoteDenom,
							utils.OneDec.Sub(market.MakerFeeRate).MulInt(execQuote).TruncateInt()))
				} else {
					makerReceivedCoins = sdk.NewCoins(
						sdk.NewCoin(market.QuoteDenom, execQuote),
						sdk.NewCoin(market.BaseDenom, market.MakerFeeRate.Neg().MulInt(execQty).TruncateInt()))
				}
			} else {
				paidCoin = sdk.NewCoin(market.BaseDenom, execQty)
				receivedCoin = sdk.NewCoin(
					market.QuoteDenom,
					utils.OneDec.Sub(market.TakerFeeRate).MulInt(execQuote).TruncateInt())
				makerPaid = execQuote
				if market.MakerFeeRate.IsPositive() {
					makerReceivedCoins = sdk.NewCoins(
						sdk.NewCoin(
							market.BaseDenom,
							utils.OneDec.Sub(market.MakerFeeRate).MulInt(execQty).TruncateInt()))
				} else {
					makerReceivedCoins = sdk.NewCoins(
						sdk.NewCoin(market.BaseDenom, execQty),
						sdk.NewCoin(market.QuoteDenom, market.MakerFeeRate.Neg().MulInt(execQuote).TruncateInt()))
				}
			}
			if err := k.EscrowCoin(ctx, market, ordererAddr, paidCoin); err != nil {
				panic(err)
			}
			if err := k.ReleaseCoins(ctx, market, sdk.MustAccAddressFromBech32(order.Order.Orderer), makerReceivedCoins); err != nil {
				panic(err)
			}
			if err := k.ReleaseCoin(ctx, market, ordererAddr, receivedCoin); err != nil {
				panic(err)
			}
			order.Order.OpenQuantity = order.Order.OpenQuantity.Sub(execQty)
			order.Order.RemainingDeposit = order.Order.RemainingDeposit.Sub(makerPaid)
			order.Updated = true
			k.SetTransientOrderBookOrder(ctx, order)
			k.AfterOrderExecuted(ctx, order.Order, execQty)
		}
		return false
	})
	if !simulate {
		k.settleTransientOrderBook(ctx, market)
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
	if err := k.ReleaseCoin(ctx, market, senderAddr, refundedCoin); err != nil {
		return order, err
	}
	k.DeleteOrder(ctx, order)
	k.DeleteOrderBookOrder(ctx, order)
	return order, nil
}
