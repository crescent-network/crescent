package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestRewardsGrowthInvariant() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.5"))
	farmingPoolAddr := s.FundedAccount(0, enoughCoins)
	s.CreatePublicFarmingPlan(
		"Farming plan", farmingPoolAddr, farmingPoolAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.4"), utils.ParseDec("0.6"),
		utils.ParseCoins("100_000000ucre,50_000000uusd"))
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.001"), utils.ParseDec("1000"),
		utils.ParseCoins("1000_000000ucre,500_000000uusd"))
	s.NextBlock()
	s.NextBlock()

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(200_000000))
	s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(500_000000))

	_, broken := keeper.RewardsGrowthInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	// Halve the rewards.
	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	poolState.FeeGrowthGlobal = poolState.FeeGrowthGlobal.QuoDec(sdk.NewDec(2))
	poolState.FarmingRewardsGrowthGlobal = poolState.FarmingRewardsGrowthGlobal.QuoDec(sdk.NewDec(2))
	s.keeper.SetPoolState(s.Ctx, pool.Id, poolState)

	_, broken = keeper.RewardsGrowthInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestCanCollectInvariant() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.5"))
	farmingPoolAddr := s.FundedAccount(0, enoughCoins)
	s.CreatePublicFarmingPlan(
		"Farming plan", farmingPoolAddr, farmingPoolAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	prices := []sdk.Dec{
		utils.ParseDec("0.0001"),
		utils.ParseDec("0.001"),
		utils.ParseDec("0.01"),
		utils.ParseDec("0.1"),
		utils.ParseDec("1"),
		utils.ParseDec("10"),
		utils.ParseDec("100"),
		utils.ParseDec("1000"),
		utils.ParseDec("10000"),
	}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10; i++ {
		j := r.Intn(len(prices) - 1)
		lowerPrice := prices[j]
		upperPrice := prices[j+1+r.Intn(len(prices)-(j+1))]
		s.AddLiquidity(
			lpAddr, pool.Id, lowerPrice, upperPrice,
			utils.ParseCoins("100_000000ucre,50_000000uusd"))
	}

	s.NextBlock()
	s.NextBlock()
	s.NextBlock()

	ordererAddr := s.FundedAccount(2, enoughCoins)
	for i := 0; i < 100; i++ {
		s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(50_000000))
	}
	for i := 0; i < 150; i++ {
		s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(50_000000))
	}

	_, broken := keeper.CanCollectInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	position := s.keeper.MustGetPosition(s.Ctx, 1)
	position.LastFeeGrowthInside = position.LastFeeGrowthInside.MulDec(sdk.NewDec(2))
	position.LastFarmingRewardsGrowthInside = position.LastFarmingRewardsGrowthInside.MulDec(sdk.NewDec(2))
	s.keeper.SetPosition(s.Ctx, position)

	_, broken = keeper.CanCollectInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}

func (s *KeeperTestSuite) TestPoolCurrentLiquidityInvariant() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.5"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.4"), utils.ParseDec("0.6"),
		utils.ParseCoins("100_000000ucre,50_000000uusd"))
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.001"), utils.ParseDec("1000"),
		utils.ParseCoins("1000_000000ucre,500_000000uusd"))

	ordererAddr := s.FundedAccount(2, enoughCoins)
	s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(200_000000))

	_, broken := keeper.PoolCurrentLiquidityInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	poolState.CurrentLiquidity = poolState.CurrentLiquidity.Add(sdk.NewInt(100))
	s.keeper.SetPoolState(s.Ctx, pool.Id, poolState)

	_, broken = keeper.PoolCurrentLiquidityInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}
