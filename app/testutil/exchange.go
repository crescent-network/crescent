package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *TestSuite) CreateMarket(baseDenom, quoteDenom string) exchangetypes.Market {
	s.T().Helper()
	creatorAddr := utils.TestAddress(1000000)
	creationFee := s.App.ExchangeKeeper.GetMarketCreationFee(s.Ctx)
	if !s.GetAllBalances(creatorAddr).IsAllGTE(creationFee) {
		s.FundAccount(creatorAddr, creationFee)
	}
	market, err := s.App.ExchangeKeeper.CreateMarket(s.Ctx, creatorAddr, baseDenom, quoteDenom)
	s.Require().NoError(err)
	return market
}

func (s *TestSuite) PlaceLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (orderId uint64, order exchangetypes.Order, res exchangetypes.ExecuteOrderResult) {
	s.T().Helper()
	var err error
	orderId, order, res, err = s.App.ExchangeKeeper.PlaceLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty, lifespan)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceBatchLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (order exchangetypes.Order) {
	s.T().Helper()
	var err error
	order, err = s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty, lifespan)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceMMLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (orderId uint64, order exchangetypes.Order, res exchangetypes.ExecuteOrderResult) {
	s.T().Helper()
	var err error
	orderId, order, res, err = s.App.ExchangeKeeper.PlaceMMLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty, lifespan)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceMMBatchLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price, qty sdk.Dec, lifespan time.Duration) (order exchangetypes.Order) {
	s.T().Helper()
	var err error
	order, err = s.App.ExchangeKeeper.PlaceMMBatchLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty, lifespan)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceMarketOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, qty sdk.Dec) (orderId uint64, res exchangetypes.ExecuteOrderResult) {
	s.T().Helper()
	var err error
	orderId, res, err = s.App.ExchangeKeeper.PlaceMarketOrder(s.Ctx, marketId, ordererAddr, isBuy, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) CancelOrder(ordererAddr sdk.AccAddress, orderId uint64) (order exchangetypes.Order) {
	s.T().Helper()
	var err error
	order, err = s.App.ExchangeKeeper.CancelOrder(s.Ctx, ordererAddr, orderId)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) CancelAllOrders(ordererAddr sdk.AccAddress, marketId uint64) (orders []exchangetypes.Order) {
	s.T().Helper()
	var err error
	orders, err = s.App.ExchangeKeeper.CancelAllOrders(s.Ctx, ordererAddr, marketId)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) SwapExactAmountIn(
	ordererAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.DecCoin, simulate bool) (output sdk.DecCoin, results []exchangetypes.SwapRouteResult) {
	s.T().Helper()
	var err error
	output, results, err = s.App.ExchangeKeeper.SwapExactAmountIn(s.Ctx, ordererAddr, routes, input, minOutput, simulate)
	s.Require().NoError(err)
	return
}
