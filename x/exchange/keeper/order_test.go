package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestFoo() {
	aliceAddr := utils.TestAddress(1)
	bobAddr := utils.TestAddress(2)

	s.FundAccount(aliceAddr, utils.ParseCoins("1000000ucre,1000000uusd"))
	s.FundAccount(bobAddr, utils.ParseCoins("1000000ucre,1000000uusd"))

	market := s.CreateMarket(utils.TestAddress(3), "ucre", "uusd", true)

	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("100"), sdk.NewInt(1000))
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("99"), sdk.NewInt(1000))
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("97"), sdk.NewInt(1000))

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("98"), sdk.NewInt(1500))

	s.Require().Equal("1001500ucre,704224uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr).String())
	s.Require().Equal("998500ucre,1149051uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr).String())
	s.Require().Equal("146725uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(market.EscrowAddress)).String())

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("96"), sdk.NewInt(1500))

	s.Require().Equal("1003000ucre,704443uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr).String())
	s.Require().Equal("997000ucre,1295111uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr).String())
	s.Require().Equal("446uusd", s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.MustAccAddressFromBech32(market.EscrowAddress)).String())
}
