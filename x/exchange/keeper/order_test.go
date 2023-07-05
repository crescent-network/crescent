package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestPlaceLimitOrder() {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	msgServer := keeper.NewMsgServerImpl(s.keeper)
	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	resp, err := msgServer.PlaceLimitOrder(
		sdk.WrapSDKContext(s.Ctx), types.NewMsgPlaceLimitOrder(
			ordererAddr1, market.Id, true, utils.ParseDec("5.1"), sdk.NewInt(10_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(1, resp.OrderId)
	s.Require().Equal(sdk.NewInt(0), resp.ExecutedQuantity)
	s.Require().Equal(sdk.NewInt64Coin("uusd", 0), resp.Paid)
	s.Require().Equal(sdk.NewInt64Coin("ucre", 0), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"0"`),
		"paid":              []byte(`{"denom":"uusd","amount":"0"}`),
		"received":          []byte(`{"denom":"ucre","amount":"0"}`),
	})

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	resp, err = msgServer.PlaceLimitOrder(
		sdk.WrapSDKContext(s.Ctx), types.NewMsgPlaceLimitOrder(
			ordererAddr2, market.Id, false, utils.ParseDec("5"), sdk.NewInt(5_000000), time.Hour))
	s.Require().NoError(err)
	s.Require().EqualValues(2, resp.OrderId)
	s.Require().Equal(sdk.NewInt(5_000000), resp.ExecutedQuantity)
	s.Require().Equal(sdk.NewInt64Coin("ucre", 5_000000), resp.Paid)
	// Matched at 5.1
	s.Require().Equal(sdk.NewInt64Coin("uusd", 25_423500), resp.Received)
	s.CheckEvent(&types.EventPlaceLimitOrder{}, map[string][]byte{
		"executed_quantity": []byte(`"5000000"`),
		"paid":              []byte(`{"denom":"ucre","amount":"5000000"}`),
		"received":          []byte(`{"denom":"uusd","amount":"25423500"}`),
	})
}

func (s *KeeperTestSuite) TestPlaceBatchLimitOrder() {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)

	ordererBalances1Before := s.GetAllBalances(ordererAddr1)
	ordererBalances2Before := s.GetAllBalances(ordererAddr2)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("5.1"), sdk.NewInt(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5"), sdk.NewInt(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.Require().Equal("5.050000000000000000", marketState.LastPrice.String())
	ordererBalances1After := s.GetAllBalances(ordererAddr1)
	ordererBalances2After := s.GetAllBalances(ordererAddr2)
	ordererBalances1Diff, _ := ordererBalances1After.SafeSub(ordererBalances1Before)
	ordererBalances2Diff, _ := ordererBalances2After.SafeSub(ordererBalances2Before)
	// Both taker
	s.Require().Equal("9970000ucre,-50500000uusd", ordererBalances1Diff.String())
	s.Require().Equal("-10000000ucre,50348500uusd", ordererBalances2Diff.String())

	ordererBalances1Before = ordererBalances1After
	ordererBalances2Before = ordererBalances2After
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr1, true, utils.ParseDec("5.2"), sdk.NewInt(10_000000), 0)
	s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("5.07"), sdk.NewInt(10_000000), 0)
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))
	s.NextBlock()

	marketState = s.keeper.MustGetMarketState(s.Ctx, market.Id)
	s.Require().NotNil(marketState.LastPrice)
	s.Require().Equal("5.070000000000000000", marketState.LastPrice.String())
	ordererBalances1After = s.GetAllBalances(ordererAddr1)
	ordererBalances2After = s.GetAllBalances(ordererAddr2)
	ordererBalances1Diff, _ = ordererBalances1After.SafeSub(ordererBalances1Before)
	ordererBalances2Diff, _ = ordererBalances2After.SafeSub(ordererBalances2Before)
	s.Require().Equal("9970000ucre,-50700000uusd", ordererBalances1Diff.String())
	// Maker
	s.Require().Equal("-9985000ucre,50700000uusd", ordererBalances2Diff.String())
}

func (s *KeeperTestSuite) TestOrderMatching() {
	aliceAddr := s.FundedAccount(1, enoughCoins)
	bobAddr := s.FundedAccount(2, enoughCoins)

	market := s.CreateMarket(utils.TestAddress(3), "ucre", "uusd", true)

	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("100"), sdk.NewInt(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("99"), sdk.NewInt(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("97"), sdk.NewInt(1000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("98"), sdk.NewInt(1500), time.Hour)

	s.Require().Equal("1001500ucre,704224uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr).String())
	s.Require().Equal("998500ucre,1149051uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr).String())
	s.Require().Equal("146725uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()).String())

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("96"), sdk.NewInt(1500), time.Hour)

	s.Require().Equal("1003000ucre,704443uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr).String())
	s.Require().Equal("997000ucre,1295111uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr).String())
	s.Require().Equal("446uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()).String())
}
