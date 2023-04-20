package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *TestSuite) CreateSpotMarket(creatorAddr sdk.AccAddress, baseDenom, quoteDenom string, fundFee bool) exchangetypes.SpotMarket {
	s.T().Helper()
	if fundFee {
		s.FundAccount(creatorAddr, s.App.ExchangeKeeper.GetSpotMarketCreationFee(s.Ctx))
	}
	market, err := s.App.ExchangeKeeper.CreateSpotMarket(s.Ctx, creatorAddr, baseDenom, quoteDenom)
	s.Require().NoError(err)
	return market
}

func (s *TestSuite) PlaceSpotLimitOrder(
	marketId string, ordererAddr sdk.AccAddress, isBuy bool, price sdk.Dec, qty sdk.Int) (order exchangetypes.SpotOrder, execQty, execQuote sdk.Int) {
	s.T().Helper()
	var err error
	order, execQty, execQuote, err = s.App.ExchangeKeeper.PlaceSpotLimitOrder(s.Ctx, marketId, ordererAddr, isBuy, price, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) PlaceSpotMarketOrder(
	marketId string, ordererAddr sdk.AccAddress, isBuy bool, qty sdk.Int) (execQty, execQuote sdk.Int) {
	s.T().Helper()
	var err error
	execQty, execQuote, err = s.App.ExchangeKeeper.PlaceSpotMarketOrder(s.Ctx, marketId, ordererAddr, isBuy, qty)
	s.Require().NoError(err)
	return
}

func (s *TestSuite) SwapExactIn(
	ordererAddr sdk.AccAddress, routes []string, input, minOutput sdk.Coin) (output sdk.Coin) {
	s.T().Helper()
	var err error
	output, err = s.App.ExchangeKeeper.SwapExactIn(s.Ctx, ordererAddr, routes, input, minOutput)
	s.Require().NoError(err)
	return output
}
