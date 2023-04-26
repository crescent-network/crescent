package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestPoolOrders() {
	creatorAddr := utils.TestAddress(1)
	s.FundAccount(creatorAddr, utils.ParseCoins("100000_000000ucre,100000_000000uusd"))

	market := s.CreateMarket(creatorAddr, "ucre", "uusd", true)
	pool := s.CreatePool(creatorAddr, market.Id, 500, sdk.NewDec(5), true)

	s.AddLiquidity(
		creatorAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(1000_000000), sdk.NewInt(5000_000000), sdk.NewInt(100_000000), sdk.NewInt(500_000000))

	ordererAddr := utils.TestAddress(2)
	s.FundAccount(ordererAddr, utils.ParseCoins("1000000_000000ucre,1000000_000000uusd"))

	balancesBefore := s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(110_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(80_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(10_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))
}

func (s *KeeperTestSuite) TestPoolBenefits() {
	aliceAddr := utils.TestAddress(1)
	bobAddr := utils.TestAddress(2)
	ordererAddr := utils.TestAddress(3)
	initialBalances := utils.ParseCoins("10000_000000ucre,10000_000000uusd")
	s.FundAccount(aliceAddr, initialBalances)
	s.FundAccount(bobAddr, initialBalances)
	s.FundAccount(ordererAddr, initialBalances)

	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	pool := s.CreatePool(utils.TestAddress(0), market.Id, 50, utils.ParseDec("5"), true)

	alicePosition, aliceLiquidity, _, _ := s.AddLiquidity(
		aliceAddr, pool.Id, utils.ParseDec("4.98"), utils.ParseDec("5.02"),
		sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.OneInt(), sdk.OneInt())
	bobPosition, bobLiquidity, _, _ := s.AddLiquidity(
		bobAddr, pool.Id, utils.ParseDec("4.99"), utils.ParseDec("5.05"),
		sdk.NewInt(1000_000000), sdk.NewInt(1000_000000), sdk.OneInt(), sdk.OneInt())

	fmt.Println("Alice provides", aliceLiquidity)
	fmt.Println("Bob provides", bobLiquidity)

	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(1000_000000))
	fmt.Println(s.Collect(aliceAddr, alicePosition.Id, sdk.NewInt(100_000000), sdk.NewInt(100_000000)))
	fmt.Println(s.Collect(bobAddr, bobPosition.Id, sdk.NewInt(100_000000), sdk.NewInt(100_000000)))
}
