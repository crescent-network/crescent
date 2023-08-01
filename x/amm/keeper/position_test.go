package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestAddLiquidity() {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, sdk.NewDec(1))

	senderAddr := s.FundedAccount(1, enoughCoins)
	position, liquidity, amt := s.AddLiquidity(
		senderAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		utils.ParseCoins("1000000ucre,1000000uusd"))
	fmt.Println(position, liquidity, amt)

	_, amt = s.RemoveLiquidity(senderAddr, position.Id, sdk.NewInt(9472135))
	fmt.Println(amt)
}

func (s *KeeperTestSuite) TestReinitializePosition() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	ownerAddr := s.FundedAccount(1, enoughCoins)
	lowerPrice, upperPrice := utils.ParseDec("4.5"), utils.ParseDec("5.5")
	desiredAmt := utils.ParseCoins("100_000000ucre,500_000000uusd")
	position, liquidity, _ := s.AddLiquidity(
		ownerAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewDec(1000000))
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewDec(1000000))

	s.RemoveLiquidity(ownerAddr, position.Id, liquidity)
	position, _ = s.keeper.GetPosition(s.Ctx, position.Id)
	fmt.Println(position.Liquidity)
	s.AddLiquidity(
		ownerAddr, pool.Id, lowerPrice, upperPrice, desiredAmt)
}

func (s *KeeperTestSuite) TestRemoveAllAndCollect() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	// Accrue fees.
	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewDec(10_000000))
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewDec(10_000000))

	s.RemoveLiquidity(lpAddr, position.Id, position.Liquidity)

	fee, farmingRewards, err := s.keeper.CollectibleCoins(s.Ctx, position.Id)
	s.Require().NoError(err)
	s.Collect(lpAddr, position.Id, fee.Add(farmingRewards...))
}

func (s *KeeperTestSuite) TestNegativeFarmingRewardsGrowthInside() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1.1366"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("1.1"), utils.ParseDec("1.2"),
		utils.ParseCoins("1000_000000ucre,1000_000000uusd"))
	creatorAddr := s.FundedAccount(2, enoughCoins)
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.NextBlock()
	s.NextBlock()
	_, farmingRewards := s.CollectibleCoins(1)
	s.Require().Equal("11573ucre", farmingRewards.String())
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.9"), utils.ParseDec("1.1"),
		utils.ParseCoins("1000_000000uusd"))
	_, farmingRewards = s.CollectibleCoins(1)
	s.Require().Equal("11573ucre", farmingRewards.String())
	_, farmingRewards = s.CollectibleCoins(2)
	s.Require().Equal("", farmingRewards.String())
}

func (s *KeeperTestSuite) TestNarrowPosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("20"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("19.999"), utils.ParseDec("20.001"),
		utils.ParseCoins("500ucre,10000uusd"))
	s.AssertEqual(sdk.NewInt(89441601), liquidity)

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("19.998"), utils.ParseDec("20.001"),
		utils.ParseCoins("250ucre,10000uusd"))
	s.AssertEqual(sdk.NewInt(44720241), liquidity)

	s.FundAccount(lpAddr, utils.ParseCoins("1_000_000_000_000000000000000000ufoo,1_000_000_000_000000000000000000ubar"))
	_, pool2 := s.CreateMarketAndPool("ufoo", "ubar", utils.ParseDec("1"))

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("0.99999"), utils.ParseDec("1.0001"),
		utils.ParseCoins("1_000_000_000_000000000000000000ufoo,1_000_000_000_000000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("20001499987500662572187476524684"), liquidity)
}

func (s *KeeperTestSuite) TestExtremePrice() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1000000000"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("900000000"), utils.ParseDec("1100000000"),
		utils.ParseCoins("46ucre,50000000000uusd"))
	s.AssertEqual(sdk.NewInt(30811388), liquidity)

	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("899900000"), utils.ParseDec("1100000000"),
		utils.ParseCoins("46ucre,50000000000uusd"))
	s.AssertEqual(sdk.NewInt(30779775), liquidity)

	_, pool2 := s.CreateMarketAndPool("uusd", "ucre", utils.ParseDec("0.00000001"))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("0.000000009"), utils.ParseDec("0.000000011"),
		utils.ParseCoins("90686675071631uusd,1000000ucre"))
	s.AssertEqual(sdk.NewInt(194868329792), liquidity)

	s.FundAccount(
		lpAddr, utils.ParseCoins("1000000000000000000000000000000000000000000ufoo,1000000000000000000000000000000000000000000ubar"))
	_, pool3 := s.CreateMarketAndPool("ufoo", "ubar", exchangetypes.PriceAtTick(types.MaxTick-1))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool3.Id, exchangetypes.PriceAtTick(types.MaxTick-2), exchangetypes.PriceAtTick(types.MaxTick),
		utils.ParseCoins("10ufoo,100000000000000000000000000000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("199998499993749943749335928"), liquidity)

	_, pool4 := s.CreateMarketAndPool("ubar", "ufoo", exchangetypes.PriceAtTick(types.MinTick+1))
	_, liquidity, _ = s.AddLiquidity(
		lpAddr, pool4.Id, exchangetypes.PriceAtTick(types.MinTick), exchangetypes.PriceAtTick(types.MinTick+2),
		utils.ParseCoins("10ufoo,1000000000000000ubar"))
	s.AssertEqual(utils.ParseInt("2000050001250"), liquidity)
}

func (s *KeeperTestSuite) TestSmallPosition() {
	s.keeper.SetDefaultTickSpacing(s.Ctx, 1)
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	_, liquidity, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("0.5"), utils.ParseDec("1.5"),
		utils.ParseCoins("7ucre,10uusd"))
	s.AssertEqual(sdk.NewInt(34), liquidity)

	// TODO: write more code
}
