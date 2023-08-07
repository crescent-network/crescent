package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// Querier is used as Keeper will have duplicate methods if used directly,
// and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Params queries the parameters of the module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Querier) AllMarkets(c context.Context, req *types.QueryAllMarketsRequest) (*types.QueryAllMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	marketStore := prefix.NewStore(store, types.MarketKeyPrefix)
	var marketResps []types.MarketResponse
	pageRes, err := query.Paginate(marketStore, req.Pagination, func(key, value []byte) error {
		var market types.Market
		k.cdc.MustUnmarshal(value, &market)
		marketResps = append(marketResps, k.MakeMarketResponse(ctx, market))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllMarketsResponse{
		Markets:    marketResps,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Market(c context.Context, req *types.QueryMarketRequest) (*types.QueryMarketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	market, found := k.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "market not found")
	}
	return &types.QueryMarketResponse{Market: k.MakeMarketResponse(ctx, market)}, nil
}

func (k Querier) AllOrders(c context.Context, req *types.QueryAllOrdersRequest) (*types.QueryAllOrdersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	var ordererAddr sdk.AccAddress
	if req.Orderer != "" {
		var err error
		ordererAddr, err = sdk.AccAddressFromBech32(req.Orderer)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid orderer: %v", err)
		}
	}
	if req.MarketId > 0 {
		if found := k.LookupMarket(ctx, req.MarketId); !found {
			return nil, status.Error(codes.NotFound, "market not found")
		}
	}
	var (
		keyPrefix   []byte
		orderGetter func(key, value []byte) types.Order
	)
	getOrderFromOrdersByOrdererIndexKey := func(key, _ []byte) types.Order {
		orderId := types.ParseOrderIdFromOrdersByOrdererIndexKey(utils.Key(keyPrefix, key))
		return k.MustGetOrder(ctx, orderId)
	}
	if req.Orderer != "" && req.MarketId > 0 {
		keyPrefix = types.GetOrdersByOrdererAndMarketIteratorPrefix(ordererAddr, req.MarketId)
		orderGetter = getOrderFromOrdersByOrdererIndexKey
	} else if req.Orderer != "" {
		keyPrefix = types.GetOrdersByOrdererIteratorPrefix(ordererAddr)
		orderGetter = getOrderFromOrdersByOrdererIndexKey
	} else if req.MarketId > 0 {
		keyPrefix = types.GetOrdersByMarketIteratorPrefix(req.MarketId)
		orderGetter = func(_, value []byte) types.Order {
			return k.MustGetOrder(ctx, sdk.BigEndianToUint64(value))
		}
	} else {
		keyPrefix = types.OrderKeyPrefix
		orderGetter = func(_, value []byte) types.Order {
			var order types.Order
			k.cdc.MustUnmarshal(value, &order)
			return order
		}
	}
	store := ctx.KVStore(k.storeKey)
	orderStore := prefix.NewStore(store, keyPrefix)
	var orders []types.Order
	pageRes, err := query.Paginate(orderStore, req.Pagination, func(key, value []byte) error {
		order := orderGetter(key, value)
		orders = append(orders, order)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAllOrdersResponse{
		Orders:     orders,
		Pagination: pageRes,
	}, nil
}

func (k Querier) Order(c context.Context, req *types.QueryOrderRequest) (*types.QueryOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	order, found := k.GetOrder(ctx, req.OrderId)
	if !found {
		return nil, status.Error(codes.NotFound, "order not found")
	}
	return &types.QueryOrderResponse{Order: order}, nil
}

func (k Querier) BestSwapExactAmountInRoutes(c context.Context, req *types.QueryBestSwapExactAmountInRoutesRequest) (*types.QueryBestSwapExactAmountInRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	maxRoutesLen := int(k.GetMaxSwapRoutesLen(ctx))
	input, err := sdk.ParseDecCoin(req.Input)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
	}
	if err := sdk.ValidateDenom(req.OutputDenom); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid output denom: %v", err)
	}
	allRoutes := k.FindAllRoutes(ctx, input.Denom, req.OutputDenom, maxRoutesLen)
	if len(allRoutes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no routes")
	}
	var (
		bestRoutes  []uint64
		bestOutput  = sdk.NewDecCoin(req.OutputDenom, utils.ZeroInt)
		bestResults []types.SwapRouteResult
	)
	// TODO: cache (begin, end, input) <-> output
	for _, routes := range allRoutes {
		output, results, err := k.SwapExactAmountIn(
			ctx, sdk.AccAddress{}, routes, input, sdk.NewDecCoin(req.OutputDenom, utils.ZeroInt), true)
		if err != nil && !errors.Is(err, types.ErrSwapNotEnoughInput) && !errors.Is(err, types.ErrSwapNotEnoughLiquidity) { // sanity check
			panic(err)
		}
		if err == nil {
			if output.Amount.GT(bestOutput.Amount) {
				bestRoutes = routes
				bestOutput = output
				bestResults = results
			}
		}
	}
	if len(bestRoutes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no possible routes for positive output")
	}
	return &types.QueryBestSwapExactAmountInRoutesResponse{
		Routes:  bestRoutes,
		Output:  bestOutput,
		Results: bestResults,
	}, nil
}

func (k Querier) OrderBook(c context.Context, req *types.QueryOrderBookRequest) (*types.QueryOrderBookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	market, found := k.GetMarket(ctx, req.MarketId)
	if !found {
		return nil, status.Error(codes.NotFound, "market not found")
	}
	orderBooks := k.MakeOrderBooks(ctx, market, 30)
	return &types.QueryOrderBookResponse{
		OrderBooks: orderBooks,
	}, nil
}

func (k Querier) MakeMarketResponse(ctx sdk.Context, market types.Market) types.MarketResponse {
	marketState := k.MustGetMarketState(ctx, market.Id)
	return types.NewMarketResponse(market, marketState)
}

func (k Querier) MakeOrderBooks(ctx sdk.Context, market types.Market, maxNumPriceLevels int) []types.OrderBook {
	// Use cache context so that we can ignore escrow from order sources while
	// constructing temporary order books.
	cacheCtx, _ := ctx.CacheContext()

	// Since price intervals among all price levels in an order book must be
	// consistent, we have to group price levels together below the price
	// where the tick interval changes.
	// Because of this, we read N sell price levels to find out the highest
	// price in the order book and use the price interval at that price as
	// the smallest possible price interval.
	sellObs := k.ConstructMemOrderBookSide(cacheCtx, market, types.MemOrderBookSideOptions{
		IsBuy:             false,
		MaxNumPriceLevels: maxNumPriceLevels,
	}, nil)
	// Read at most one level to check if any buy order exists.
	buyObs := k.ConstructMemOrderBookSide(cacheCtx, market, types.MemOrderBookSideOptions{
		IsBuy:             true,
		MaxNumPriceLevels: 1,
	}, nil)
	var highestPrice sdk.Dec
	if len(sellObs.Levels()) > 0 {
		highestPrice = sellObs.Levels()[len(sellObs.Levels())-1].Price()
	} else if len(buyObs.Levels()) > 0 {
		highestPrice = buyObs.Levels()[0].Price()
	} else {
		return nil // No orders
	}
	smallestPriceInterval := types.PriceIntervalAtTick(types.TickAtPrice(highestPrice))

	cacheCtx, _ = ctx.CacheContext()
	var priceLimit *sdk.Dec
	if len(sellObs.Levels()) > 0 {
		p := sellObs.Levels()[0].Price().Add(smallestPriceInterval.MulInt64(int64(100 * maxNumPriceLevels)))
		priceLimit = &p
	}
	sellObs = k.ConstructMemOrderBookSide(cacheCtx, market, types.MemOrderBookSideOptions{
		IsBuy:      false,
		PriceLimit: priceLimit,
	}, nil)
	priceLimit = nil
	if len(buyObs.Levels()) > 0 {
		p := buyObs.Levels()[0].Price().Sub(smallestPriceInterval.MulInt64(int64(100 * maxNumPriceLevels)))
		priceLimit = &p
	}
	buyObs = k.ConstructMemOrderBookSide(cacheCtx, market, types.MemOrderBookSideOptions{
		IsBuy:      true,
		PriceLimit: priceLimit,
	}, nil)

	var orderBooks []types.OrderBook
	for _, p := range []int{1, 10, 100} {
		priceInterval := smallestPriceInterval.MulInt64(int64(p))
		ob := types.OrderBook{
			PriceInterval: priceInterval,
			Sells:         nil,
			Buys:          nil,
		}
		if len(sellObs.Levels()) > 0 {
			levelIdx := 0
			currentPrice := FitPriceToPriceInterval(sellObs.Levels()[levelIdx].Price(), priceInterval, false)
			for i := 0; i < maxNumPriceLevels && levelIdx < len(sellObs.Levels()); {
				qty := utils.ZeroDec
				for levelIdx < len(sellObs.Levels()) && sellObs.Levels()[levelIdx].Price().LTE(currentPrice) {
					qty = qty.Add(types.TotalExecutableQuantity(sellObs.Levels()[levelIdx].Orders()))
					levelIdx++
				}
				if qty.IsPositive() {
					ob.Sells = append(ob.Sells, types.OrderBookPriceLevel{
						P: currentPrice,
						Q: qty,
					})
					i++
				}
				currentPrice = currentPrice.Add(priceInterval)
			}
		}
		if len(buyObs.Levels()) > 0 {
			levelIdx := 0
			currentPrice := FitPriceToPriceInterval(buyObs.Levels()[levelIdx].Price(), priceInterval, true)
			for i := 0; i < maxNumPriceLevels && levelIdx < len(buyObs.Levels()); {
				qty := utils.ZeroDec
				for levelIdx < len(buyObs.Levels()) && buyObs.Levels()[levelIdx].Price().GTE(currentPrice) {
					qty = qty.Add(types.TotalExecutableQuantity(buyObs.Levels()[levelIdx].Orders()))
					levelIdx++
				}
				if qty.IsPositive() {
					ob.Buys = append(ob.Buys, types.OrderBookPriceLevel{
						P: currentPrice,
						Q: qty,
					})
					i++
				}
				currentPrice = currentPrice.Sub(priceInterval)
			}
		}
		orderBooks = append(orderBooks, ob)
	}
	return orderBooks
}

func FitPriceToPriceInterval(price, interval sdk.Dec, roundUp bool) sdk.Dec {
	b := price.BigInt()
	b.Quo(b, interval.BigInt()).Mul(b, interval.BigInt())
	priceRoundedDown := sdk.NewDecFromBigIntWithPrec(b, sdk.Precision)
	if roundUp && !priceRoundedDown.Equal(price) {
		priceRoundedDown = priceRoundedDown.Add(interval)
	}
	return priceRoundedDown
}
