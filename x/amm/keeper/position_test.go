package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
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

func (s *KeeperTestSuite) TestRewardsPool() {
	market1, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	market2, pool2 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseCoins("100_000000ucre,500_000000uusd"))
	s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("9"), utils.ParseDec("12"), utils.ParseCoins("100_000000uatom,1000_000000uusd"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market1.Id, ordererAddr, true, sdk.NewDec(1_000000))
	s.PlaceMarketOrder(market2.Id, ordererAddr, false, sdk.NewDec(1_000000))

	s.AssertEqual(utils.ParseCoins("1499ucre,2620uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("268uatom,14982uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))

	fee, _ := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("1498ucre,2619uusd"), fee)
	s.Collect(lpAddr, position1.Id, utils.ParseCoins("1497ucre,2618uusd"))
	s.AssertEqual(utils.ParseCoins("2ucre,2uusd"), s.GetAllBalances(pool1.MustGetRewardsPoolAddress()))
	s.AssertEqual(utils.ParseCoins("268uatom,14982uusd"), s.GetAllBalances(pool2.MustGetRewardsPoolAddress()))
}

func (s *KeeperTestSuite) TestLastRemoveLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	lpAddr1 := s.FundedAccount(1, enoughCoins)
	lpAddr2 := s.FundedAccount(2, enoughCoins)
	lpAddr3 := s.FundedAccount(3, enoughCoins)

	// Three position parameters are same.
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position3, _, _ := s.AddLiquidity(
		lpAddr3, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	_, amt := s.RemoveLiquidity(lpAddr1, position1.Id, position1.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529840ucre,499999999uusd"), amt)
	_, amt = s.RemoveLiquidity(lpAddr2, position2.Id, position2.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529840ucre,499999999uusd"), amt)
	// The last liquidity remover takes all remaining reserve balances in the pool.
	_, amt = s.RemoveLiquidity(lpAddr3, position3.Id, position3.Liquidity)
	s.AssertEqual(utils.ParseCoins("82529843ucre,500000002uusd"), amt)

	// No balances left in the pool.
	s.AssertEqual(sdk.Coins{}, s.GetAllBalances(pool.MustGetReserveAddress()))
}
