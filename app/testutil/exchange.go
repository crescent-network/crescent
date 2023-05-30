package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *TestSuite) CreateMarket(creatorAddr sdk.AccAddress, baseDenom, quoteDenom string, fundFee bool) exchangetypes.Market {
	s.T().Helper()
	if fundFee {
		s.FundAccount(creatorAddr, s.App.ExchangeKeeper.GetMarketCreationFee(s.Ctx))
	}
	market, err := s.App.ExchangeKeeper.CreateMarket(s.Ctx, creatorAddr, baseDenom, quoteDenom)
	s.Require().NoError(err)
	return market
}

func (s *TestSuite) PlaceLimitOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, price sdk.Dec, qty sdk.Int, lifespan time.Duration) (orderId uint64, order exchangetypes.Order, execQty sdk.Int, paid, received sdk.Coin) {
	s.T().Helper()
	var err error
	orderId, order, execQty, paid, received, err = s.App.ExchangeKeeper.PlaceLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty, lifespan)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceMarketOrder(
	marketId uint64, ordererAddr sdk.AccAddress, isBuy bool, qty sdk.Int) (orderId uint64, execQty sdk.Int, paid, received sdk.Coin) {
	s.T().Helper()
	var err error
	orderId, execQty, paid, received, err = s.App.ExchangeKeeper.PlaceMarketOrder(s.Ctx, marketId, ordererAddr, isBuy, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) CancelOrder(ordererAddr sdk.AccAddress, orderId uint64) (order exchangetypes.Order, refundedDeposit sdk.Coin) {
	s.T().Helper()
	var err error
	order, refundedDeposit, err = s.App.ExchangeKeeper.CancelOrder(s.Ctx, ordererAddr, orderId)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) SwapExactAmountIn(
	ordererAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.Coin, simulate bool) (output sdk.Coin, fees sdk.Coins) {
	s.T().Helper()
	var err error
	output, fees, err = s.App.ExchangeKeeper.SwapExactAmountIn(s.Ctx, ordererAddr, routes, input, minOutput, simulate)
	s.Require().NoError(err)
	return
}
