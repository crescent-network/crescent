package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	utils "github.com/crescent-network/crescent/v5/types"
)

// HERE! More tests for matching (the below one is just a copy from order_test.go)
// Including Fee test
func (s *KeeperTestSuite) TestOrderMatching_case1() {
	aliceAddr := s.FundedAccount(1, utils.ParseCoins("1000000ucre,1000000uusd"))
	bobAddr := s.FundedAccount(2, utils.ParseCoins("1000000ucre,1000000uusd"))

	market := s.CreateMarket("ucre", "uusd")

	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("100"), sdk.NewDec(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("99"), sdk.NewDec(1000), time.Hour)
	s.PlaceLimitOrder(market.Id, aliceAddr, true, utils.ParseDec("97"), sdk.NewDec(1000), time.Hour)

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("98"), sdk.NewDec(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1001497ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("998500ucre,1149051uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins("3ucre,146949uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))

	s.PlaceLimitOrder(
		market.Id, bobAddr, false, utils.ParseDec("96"), sdk.NewDec(1500), time.Hour)

	s.AssertEqual(utils.ParseCoins("1002994ucre,704000uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, aliceAddr))
	s.AssertEqual(utils.ParseCoins("997000ucre,1295111uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, bobAddr))
	s.AssertEqual(utils.ParseCoins("6ucre,889uusd"), s.App.BankKeeper.GetAllBalances(s.Ctx, market.MustGetEscrowAddress()))
}
