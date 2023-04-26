package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) SwapExactIn(
	ctx sdk.Context, ordererAddr sdk.AccAddress,
	routes []uint64, input, minOutput sdk.Coin, simulate bool) (output sdk.Coin, err error) {
	currentIn := input
	for _, marketId := range routes {
		if !currentIn.Amount.IsPositive() {
			err = types.ErrInsufficientOutput // TODO: use different error
			return
		}
		market, found := k.GetMarket(ctx, marketId)
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
		totalExecQty, totalExecQuote := k.executeOrder(
			ctx, market, ordererAddr, isBuy, nil, qtyLimit, quoteLimit, simulate)
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
	if output.Amount.LT(minOutput.Amount) {
		err = sdkerrors.Wrapf(
			types.ErrInsufficientOutput, "output %s < wanted %s", output, minOutput)
		return
	}
	return
}

func (k Keeper) FindAllRoutes(ctx sdk.Context, fromDenom, toDenom string, maxRoutesLen int) (allRoutes [][]uint64) {
	denomMap := map[string]map[string]uint64{}
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.MarketByDenomsIndexKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		baseDenom, quoteDenom := types.ParseMarketByDenomsIndexKey(iter.Key())
		marketId := sdk.BigEndianToUint64(iter.Value())
		if _, ok := denomMap[baseDenom]; !ok {
			denomMap[baseDenom] = map[string]uint64{}
		}
		if _, ok := denomMap[quoteDenom]; !ok {
			denomMap[quoteDenom] = map[string]uint64{}
		}
		denomMap[baseDenom][quoteDenom] = marketId
		denomMap[quoteDenom][baseDenom] = marketId
	}
	return types.FindAllRoutes(denomMap, fromDenom, toDenom, maxRoutesLen)
}
