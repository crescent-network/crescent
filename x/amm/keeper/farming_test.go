package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestMaxNumPrivateFarmingPlans() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	s.keeper.SetMaxNumPrivateFarmingPlans(s.Ctx, 1)

	creatorAddr := s.FundedAccount(1, enoughCoins)
	s.CreatePublicFarmingPlan(
		"Farming plan", creatorAddr, creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.Require().EqualValues(0, s.keeper.GetNumPrivateFarmingPlans(s.Ctx))

	s.Require().True(s.keeper.CanCreatePrivateFarmingPlan(s.Ctx))
	plan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan 1", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.Require().EqualValues(1, s.keeper.GetNumPrivateFarmingPlans(s.Ctx))
	s.Require().False(s.keeper.CanCreatePrivateFarmingPlan(s.Ctx))

	_, err := s.keeper.CreatePrivateFarmingPlan(
		s.Ctx, creatorAddr, "Farming plan 2", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("50_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.Require().EqualError(err, "maximum number of active private farming plans reached: invalid request")

	s.Require().NoError(s.keeper.TerminateFarmingPlan(s.Ctx, plan))

	s.Require().EqualValues(0, s.keeper.GetNumPrivateFarmingPlans(s.Ctx))
	s.Require().True(s.keeper.CanCreatePrivateFarmingPlan(s.Ctx))

	_, err = s.keeper.CreatePrivateFarmingPlan(
		s.Ctx, creatorAddr, "Farming plan 2", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("50_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestCreatePrivateFarmingPlan() {
	creatorAddr := utils.TestAddress(1)
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	for _, tc := range []struct {
		name        string
		preRun      func()
		msg         *types.MsgCreatePrivateFarmingPlan
		expectedErr string
	}{
		{
			"happy case",
			func() {
				s.FundAccount(creatorAddr, enoughCoins)
			},
			types.NewMsgCreatePrivateFarmingPlan(
				creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
			"",
		},
		{
			"not enough fee",
			func() {},
			types.NewMsgCreatePrivateFarmingPlan(
				creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
			"0stake is smaller than 1000000stake: insufficient funds",
		},
		{
			"past end time",
			func() {
				s.FundAccount(creatorAddr, enoughCoins)
			},
			types.NewMsgCreatePrivateFarmingPlan(
				creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
				}, utils.ParseTime("2021-01-01T00:00:00Z"), utils.ParseTime("2022-01-01T00:00:00Z")),
			"end time is past: invalid request",
		},
		{
			"pool not found",
			func() {
				s.FundAccount(creatorAddr, enoughCoins)
			},
			types.NewMsgCreatePrivateFarmingPlan(
				creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(2, utils.ParseCoins("100_000000ucre")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
			"pool 2 not found: not found",
		},
		{
			"rewards has no supply",
			func() {
				s.FundAccount(creatorAddr, enoughCoins)
			},
			types.NewMsgCreatePrivateFarmingPlan(
				creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
					types.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000ueur")),
				}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z")),
			"denom ueur has no supply: invalid request",
		},
	} {
		s.Run(tc.name, func() {
			oldCtx := s.Ctx
			s.Ctx, _ = s.Ctx.CacheContext()
			tc.preRun()
			_, err := keeper.NewMsgServerImpl(s.keeper).CreatePrivateFarmingPlan(sdk.WrapSDKContext(s.Ctx), tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
			s.Ctx = oldCtx
		})
	}
}

func (s *KeeperTestSuite) TestTerminateFarmingPlan() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	creatorAddr := s.FundedAccount(1, enoughCoins)
	balancesBefore := s.GetAllBalances(creatorAddr)
	plan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan 1", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.Require().NoError(s.keeper.TerminateFarmingPlan(s.Ctx, plan))
	balancesAfter := s.GetAllBalances(creatorAddr)
	s.Require().Equal("10000000000ucre", balancesAfter.Sub(balancesBefore).String())
	plan, _ = s.keeper.GetFarmingPlan(s.Ctx, plan.Id)
	err := s.keeper.TerminateFarmingPlan(s.Ctx, plan)
	s.Require().EqualError(err, "plan is already terminated: invalid request")
}

func (s *KeeperTestSuite) TestAllocateFarmingRewards() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	creatorAddr := s.FundedAccount(1, enoughCoins)
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan 1", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.NextBlock()
	s.NextBlock()
	poolState := s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	s.Require().Equal("", poolState.FarmingRewardsGrowthGlobal.String())

	lpAddr := s.FundedAccount(2, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	s.EndBlock()
	s.BeginBlock(0)
	// Elapsed 0.
	poolState = s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	s.Require().Equal("", poolState.FarmingRewardsGrowthGlobal.String())

	s.NextBlock()
	poolState = s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	rewardsGrowthGlobalBefore := poolState.FarmingRewardsGrowthGlobal

	s.EndBlock()
	s.BeginBlock(s.keeper.GetMaxFarmingBlockTime(s.Ctx))
	poolState = s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	rewardsGrowthGlobalDiff1 := poolState.FarmingRewardsGrowthGlobal.Sub(rewardsGrowthGlobalBefore)
	rewardsGrowthGlobalBefore = poolState.FarmingRewardsGrowthGlobal

	// Block time is clipped to the maximum.
	s.EndBlock()
	s.BeginBlock(s.keeper.GetMaxFarmingBlockTime(s.Ctx) + 10*time.Second)
	poolState = s.keeper.MustGetPoolState(s.Ctx, pool.Id)
	rewardsGrowthGlobalDiff2 := poolState.FarmingRewardsGrowthGlobal.Sub(rewardsGrowthGlobalBefore)
	s.Require().Equal(rewardsGrowthGlobalDiff1, rewardsGrowthGlobalDiff2)
}

func (s *KeeperTestSuite) TestFarming() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	lpAddr1 := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	lpAddr2 := s.FundedAccount(2, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	position1, _, _ := s.AddLiquidity(
		lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, _, _ := s.AddLiquidity(
		lpAddr2, pool.Id, utils.ParseDec("4.8"), utils.ParseDec("5.2"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	// fmt.Println(liquidity1)
	// fmt.Println(liquidity2)

	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1uatom")) // make initial supply
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(0), "", utils.TestAddress(0), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("1000000uatom")),
		},
		utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T23:59:59Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	s.NextBlock()

	s.Collect(lpAddr1, position1.Id, utils.ParseCoins("9uatom"))
	s.Collect(lpAddr2, position2.Id, utils.ParseCoins("47uatom"))

	ordererAddr := s.FundedAccount(3, utils.ParseCoins("10000_000000uusd"))
	s.PlaceLimitOrder(
		pool.MarketId, ordererAddr, true, utils.ParseDec("6"), sdk.NewInt(120_000000), 0)

	// poolState := s.App.AMMKeeper.MustGetPoolState(s.Ctx, pool.Id)
	// fmt.Println(poolState.CurrentSqrtPrice)

	s.NextBlock()

	s.Collect(lpAddr1, position1.Id, utils.ParseCoins("56uatom"))
	s.Collect(lpAddr2, position2.Id, utils.ParseCoins(""))
}

func (s *KeeperTestSuite) TestTerminatePrivateFarmingPlan() {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	creatorAddr := s.FundedAccount(2, enoughCoins)
	termAddr := utils.TestAddress(3)
	farmingPlan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", termAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)

	s.NextBlock()
	s.NextBlock()
	s.NextBlock()

	balancesBefore := s.GetAllBalances(termAddr)
	farmingPoolAddr := farmingPlan.MustGetFarmingPoolAddress()
	remainingFarmingRewards := s.GetAllBalances(farmingPoolAddr)

	msgServer := keeper.NewMsgServerImpl(s.keeper)
	msg := types.NewMsgTerminatePrivateFarmingPlan(termAddr, 1)
	_, err := msgServer.TerminatePrivateFarmingPlan(sdk.WrapSDKContext(s.Ctx), msg)
	s.Require().NoError(err)

	s.Require().Equal("", s.GetAllBalances(farmingPoolAddr).String())
	s.Require().Equal(
		balancesBefore.Add(remainingFarmingRewards...).String(),
		s.GetAllBalances(termAddr).String())
}

func (s *KeeperTestSuite) TestTerminatePrivatePlan_Unauthorized() {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))

	creatorAddr := s.FundedAccount(2, enoughCoins)
	termAddr := utils.TestAddress(3)
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", termAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)

	msgServer := keeper.NewMsgServerImpl(s.keeper)
	msg := types.NewMsgTerminatePrivateFarmingPlan(creatorAddr, 1)
	_, err := msgServer.TerminatePrivateFarmingPlan(sdk.WrapSDKContext(s.Ctx), msg)
	s.Require().ErrorIs(err, sdkerrors.ErrUnauthorized)
}

func (s *KeeperTestSuite) TestFarmingTooMuchLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	pool.TickSpacing = 1
	s.keeper.SetPool(s.Ctx, pool)
	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("4.9999"), utils.ParseDec("5.0001"),
		utils.ParseCoins("10000_000000000000000000ucre,50000_000000000000000000uusd"))
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(2), "Farming plan", utils.TestAddress(2), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("1_000000uatom")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000uatom"), true)
	s.NextBlock()
	s.NextBlock()
	_, farmingRewards := s.CollectibleCoins(position.Id)
	s.Require().Equal("113uatom", farmingRewards.String())
}

func (s *KeeperTestSuite) TestFarmingTooSmallLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	pool.TickSpacing = 1
	s.keeper.SetPool(s.Ctx, pool)
	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, pool.Id, utils.ParseDec("0.0000001"), utils.ParseDec("10000000"),
		utils.ParseCoins("10ucre,50uusd"))
	creatorAddr := s.FundedAccount(100, utils.ParseCoins("1uibc1")) // Create supply.
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("10_000000000000000000uibc1")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("1000_000000000000000000uibc1"), true)
	s.NextBlock()
	s.NextBlock()
	_, farmingRewards := s.CollectibleCoins(position.Id)
	s.Require().Equal("1157407407407405uibc1", farmingRewards.String())
}

func (s *KeeperTestSuite) TestFarmingInsufficientFarmingRewards() {
	_, pool1 := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	_, pool2 := s.CreateMarketAndPool("uatom", "uusd", utils.ParseDec("10"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	position1, _, _ := s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, _, _ := s.AddLiquidity(
		lpAddr, pool2.Id, utils.ParseDec("9.5"), utils.ParseDec("10.5"),
		utils.ParseCoins("100_000000uatom,1000_000000uusd"))

	creatorAddr := utils.TestAddress(2)
	plan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr,
		[]types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool1.Id, utils.ParseCoins("20_000000ucre")),
			types.NewFarmingRewardAllocation(pool2.Id, utils.ParseCoins("10_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		nil, true)
	s.FundAccount(plan.MustGetFarmingPoolAddress(), utils.ParseCoins("3000ucre"))

	s.NextBlock()
	_, farmingRewards := s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("1156ucre"), farmingRewards)
	_, farmingRewards = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("577ucre"), farmingRewards)
	// There can be a bit more funds in the farming rewards pool than the sum of
	// collectible farming rewards due to truncation.
	s.AssertEqual(
		utils.ParseCoins("1735ucre"), s.GetAllBalances(types.FarmingRewardsPoolAddress))

	// The farming pool doesn't have enough funds to distribute rewards
	// for this block, so the farming pool is entirely ignored from the
	// reward allocation.
	s.AssertEqual(utils.ParseCoins("1265ucre"), s.GetAllBalances(plan.MustGetFarmingPoolAddress()))
	s.NextBlock()

	// Check that the rewards haven't changed.
	_, farmingRewards = s.CollectibleCoins(position1.Id)
	s.AssertEqual(utils.ParseCoins("1156ucre"), farmingRewards)
	_, farmingRewards = s.CollectibleCoins(position2.Id)
	s.AssertEqual(utils.ParseCoins("577ucre"), farmingRewards)
	s.AssertEqual(
		utils.ParseCoins("1735ucre"), s.GetAllBalances(types.FarmingRewardsPoolAddress))
}
