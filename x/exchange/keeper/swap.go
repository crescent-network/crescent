package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context, ordererAddr sdk.AccAddress,
	routes []uint64, input, minOutput sdk.Coin, simulate bool) (output sdk.Coin, results []types.SwapRouteResult, err error) {
	if maxRoutesLen := int(k.GetMaxSwapRoutesLen(ctx)); len(routes) > maxRoutesLen {
		return output, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "routes length exceeded the limit %d", maxRoutesLen)
	}
	halveFees := len(routes) > 1
	currentIn := input
	for _, marketId := range routes {
		if !currentIn.Amount.IsPositive() {
			return output, nil, sdkerrors.Wrap(types.ErrSwapNotEnoughInput, currentIn.String())
		}
		if !simulate {
			balances := k.bankKeeper.SpendableCoins(ctx, ordererAddr)
			if balance := balances.AmountOf(currentIn.Denom); balance.LT(currentIn.Amount) {
				return output, nil, sdkerrors.Wrapf(
					sdkerrors.ErrInsufficientFunds, "%s%s < %s", balance, currentIn.Denom, currentIn)
			}
		}
		market, found := k.GetMarket(ctx, marketId)
		if !found {
			return output, nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "market %d not found", marketId)
		}
		var (
			isBuy                bool
			qtyLimit, quoteLimit *sdk.Int
		)
		switch currentIn.Denom {
		case market.BaseDenom:
			isBuy = false
			qtyLimit = &currentIn.Amount
		case market.QuoteDenom:
			isBuy = true
			quoteLimit = &currentIn.Amount
		default:
			return output, nil, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest, "denom %s not in market %d", currentIn.Denom, market.Id)
		}
		res, err := k.executeOrder(
			ctx, market, ordererAddr, isBuy, nil, qtyLimit, quoteLimit, halveFees, simulate)
		if err != nil {
			return output, nil, err
		}
		output = res.Received
		if !res.FullyExecuted {
			if res.Paid.IsLT(currentIn) {
				return output, nil, sdkerrors.Wrapf(
					types.ErrSwapNotEnoughLiquidity, "in market %d; paid %s < input %s", marketId, res.Paid, currentIn)
			}
		}
		results = append(results, types.SwapRouteResult{
			MarketId: marketId,
			Input:    currentIn,
			Output:   output,
			Fee:      res.Fee,
		})
		currentIn = output
	}
	if output.Denom != minOutput.Denom {
		return output, nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "output denom %s != min output denom %s", output.Denom, minOutput.Denom)
	}
	if output.Amount.LT(minOutput.Amount) {
		return output, nil, sdkerrors.Wrapf(
			types.ErrSwapNotEnoughOutput, "output %s < min output %s", output, minOutput)
	}
	if err = ctx.EventManager().EmitTypedEvent(&types.EventSwapExactAmountIn{
		Orderer: ordererAddr.String(),
		Routes:  routes,
		Input:   input,
		Output:  output,
		Results: results,
	}); err != nil {
		return output, nil, err
	}
	return output, results, nil
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
	var currentRoutes []uint64
	visited := map[uint64]struct{}{}
	var backtrack func(currentDenom string)
	// TODO: prevent stack overflow?
	backtrack = func(currentDenom string) {
		denoms := maps.Keys(denomMap[currentDenom])
		slices.Sort(denoms)
		for _, denom := range denoms {
			marketId := denomMap[currentDenom][denom]
			if _, ok := visited[marketId]; !ok {
				if denom == toDenom {
					routes := make([]uint64, len(currentRoutes), len(currentRoutes)+1)
					copy(routes[:len(currentRoutes)], currentRoutes)
					routes = append(routes, marketId)
					allRoutes = append(allRoutes, routes)
				} else {
					visited[marketId] = struct{}{}
					currentRoutes = append(currentRoutes, marketId)
					if len(currentRoutes) < maxRoutesLen {
						backtrack(denom)
					}
					currentRoutes = currentRoutes[:len(currentRoutes)-1]
					delete(visited, marketId)
				}
			}
		}
	}
	backtrack(fromDenom)
	return
}
