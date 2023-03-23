package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) PlaceSpotOrder(
	ctx sdk.Context, senderAddr sdk.AccAddress, marketId string,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder, rested bool, err error) {
	market, found := k.GetSpotMarket(ctx, marketId)
	if !found {
		return types.SpotLimitOrder{}, false, sdkerrors.Wrap(sdkerrors.ErrNotFound, "market not found")
	}

	executedQty, executedQuoteAmt, outputs := k.executeSpotOrder(ctx, market, senderAddr, isBuy, priceLimit, qty)
	fmt.Printf("Order result - executedQty=%v executedQuoteAmt=%v\n", executedQty, executedQuoteAmt)

	if executedQty.IsPositive() {
		executedQuoteCoin := sdk.Coins{sdk.NewCoin(market.QuoteDenom, executedQuoteAmt)}
		executedBaseCoin := sdk.Coins{sdk.NewCoin(market.BaseDenom, executedQty)}
		var inputs []banktypes.Input
		if isBuy {
			inputs = append(inputs,
				banktypes.NewInput(senderAddr, executedQuoteCoin),
				banktypes.NewInput(k.accountKeeper.GetModuleAddress(types.ModuleName), executedBaseCoin))
		} else {
			inputs = append(inputs,
				banktypes.NewInput(senderAddr, executedBaseCoin),
				banktypes.NewInput(k.accountKeeper.GetModuleAddress(types.ModuleName), executedQuoteCoin))
		}

		if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
			return types.SpotLimitOrder{}, false, err
		}
	}

	if priceLimit != nil { // limit order
		if remainingQty := qty.Sub(executedQty); remainingQty.IsPositive() {
			order = k.restSpotLimitOrder(ctx, senderAddr, marketId, isBuy, *priceLimit, remainingQty)
			offerCoin := sdk.Coins{market.OfferCoin(isBuy, *priceLimit, remainingQty)}
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, offerCoin); err != nil {
				return types.SpotLimitOrder{}, false, err
			}
			rested = true
		}
	}
	return
}

func (k Keeper) executeSpotOrder(
	ctx sdk.Context, market types.SpotMarket, senderAddr sdk.AccAddress,
	isBuy bool, priceLimit *sdk.Dec, qty sdk.Int) (executedQty, executedQuoteAmt sdk.Int, outputs []banktypes.Output) {
	remainingQty := qty
	executedQuoteAmt = types.ZeroInt
	k.IterateSpotOrderBook(ctx, market.Id, !isBuy, priceLimit, func(order types.SpotLimitOrder) (stop bool) {
		executableQty := types.MinInt(order.OpenQuantity, remainingQty)
		quoteAmt := types.QuoteAmount(isBuy, order.Price, executableQty)
		executedQuoteAmt = executedQuoteAmt.Add(quoteAmt)

		order.OpenQuantity = order.OpenQuantity.Sub(executableQty)
		var paidCoin, receivedCoin sdk.Coins
		if isBuy {
			paidCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
		} else {
			paidCoin = sdk.Coins{sdk.NewCoin(market.BaseDenom, executableQty)}
			receivedCoin = sdk.Coins{sdk.NewCoin(market.QuoteDenom, quoteAmt)}
		}
		outputs = append(outputs,
			banktypes.Output{Address: order.Orderer, Coins: paidCoin},
			banktypes.NewOutput(senderAddr, receivedCoin))

		if order.OpenQuantity.IsZero() {
			k.DeleteSpotOrderBookOrder(ctx, order)
			k.DeleteSpotLimitOrder(ctx, order)
			// TODO: emit event
			fmt.Println("Deleting order -", order.Price, order.Sequence)
		} else {
			k.SetSpotLimitOrder(ctx, order)
			fmt.Println("Updating order -", order.Price, order.Sequence, "->", order.OpenQuantity)
		}

		remainingQty = remainingQty.Sub(executableQty)
		return remainingQty.IsZero()
	})
	executedQty = qty.Sub(remainingQty)
	return
}

func (k Keeper) restSpotLimitOrder(
	ctx sdk.Context, senderAddr sdk.AccAddress, marketId string,
	isBuy bool, price sdk.Dec, qty sdk.Int) (order types.SpotLimitOrder) {
	seq := k.GetNextOrderSequence(ctx)
	order = types.NewSpotLimitOrder(senderAddr, marketId, isBuy, price, qty, seq)
	k.SetSpotLimitOrder(ctx, order)
	k.SetSpotOrderBookOrder(ctx, order)
	return
}
