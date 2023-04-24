package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) SwapExactIn(
	ctx sdk.Context, ordererAddr sdk.AccAddress,
	routes []uint64, input, minOutput sdk.Coin) (output sdk.Coin, err error) {
	currentIn := input
	for _, marketId := range routes {
		if !currentIn.Amount.IsPositive() {
			break
		}
		market, found := k.GetSpotMarket(ctx, marketId)
		if !found {
			err = sdkerrors.Wrapf(sdkerrors.ErrNotFound, "market %d not found", marketId)
			return
		}
		var (
			isBuy                bool
			qtyLimit, quoteLimit *sdk.Int
		)
		if market.BaseDenom == currentIn.Denom {
			isBuy = false
			qtyLimit = &currentIn.Amount
			output.Denom = market.QuoteDenom
		} else {
			isBuy = true
			quoteLimit = &currentIn.Amount
			output.Denom = market.BaseDenom
		}
		totalExecQty, totalExecQuote := k.executeSpotOrder(ctx, market, ordererAddr, isBuy, nil, qtyLimit, quoteLimit)
		if isBuy {
			output.Amount = totalExecQty
		} else {
			output.Amount = totalExecQuote
		}
		currentIn = output
	}
	if output.Denom != minOutput.Denom {
		err = sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "output denom %s != wanted %s", output.Denom, minOutput.Denom)
		return
	}
	return
}
