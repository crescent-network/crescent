package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context, ordererAddr sdk.AccAddress,
	routes []uint64, input, minOutput sdk.Coin, simulate bool) (output sdk.Coin, err error) {
	if maxRoutesLen := int(k.GetMaxSwapRoutesLen(ctx)); len(routes) > maxRoutesLen {
		return output, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "routes len exceeded the limit %d", maxRoutesLen)
	}
	halveFees := len(routes) > 1
	currentIn := input
	for _, marketId := range routes {
		if !currentIn.Amount.IsPositive() {
			return output, types.ErrInsufficientOutput // TODO: use different error
		}
		market, found := k.GetMarket(ctx, marketId)
		if !found {
			return output, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "market %d not found", marketId)
		}
		var (
			isBuy                bool
			qtyLimit, quoteLimit *sdk.Int
		)
		switch currentIn.Denom {
		case market.BaseDenom:
			isBuy = false
			qtyLimit = &currentIn.Amount
			output.Denom = market.QuoteDenom
		case market.QuoteDenom:
			isBuy = true
			quoteLimit = &currentIn.Amount
			output.Denom = market.BaseDenom
		default:
			return output, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "denom %s not in market %d", currentIn.Denom, market.Id)
		}
		_, _, output = k.executeOrder(
			ctx, market, ordererAddr, isBuy, nil, qtyLimit, quoteLimit, halveFees, simulate)
		currentIn = output
	}
	if output.Denom != minOutput.Denom {
		return output, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "output denom %s != wanted %s", output.Denom, minOutput.Denom)
	}
	if output.Amount.LT(minOutput.Amount) {
		return output, sdkerrors.Wrapf(
			types.ErrInsufficientOutput, "output %s < wanted %s", output, minOutput)
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventSwapExactAmountIn{
		Orderer: ordererAddr.String(),
		Routes:  routes,
		Input:   input,
		Output:  output,
	}); err != nil {
		return output, err
	}
	return output, nil
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
