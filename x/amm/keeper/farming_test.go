package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestFarming() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	lpAddr1 := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	lpAddr2 := s.FundedAccount(2, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	position1, liquidity1, _ := s.AddLiquidity(
		lpAddr1, lpAddr1, pool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	position2, liquidity2, _ := s.AddLiquidity(
		lpAddr2, lpAddr2, pool.Id, utils.ParseDec("4.8"), utils.ParseDec("5.2"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	fmt.Println(liquidity1)
	fmt.Println(liquidity2)

	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1uatom")) // make initial supply
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(0), "", utils.TestAddress(0), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("1000000uatom")),
		},
		utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T23:59:59Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	s.NextBlock()

	s.Collect(lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("9uatom"))
	s.Collect(lpAddr2, lpAddr2, position2.Id, utils.ParseCoins("47uatom"))

	ordererAddr := s.FundedAccount(3, utils.ParseCoins("10000_000000uusd"))
	s.PlaceMarketOrder(pool.MarketId, ordererAddr, true, sdk.NewInt(120_000000))

	poolState := s.App.AMMKeeper.MustGetPoolState(s.Ctx, pool.Id)
	fmt.Println(poolState.CurrentPrice)

	s.NextBlock()

	s.Collect(lpAddr1, lpAddr1, position1.Id, utils.ParseCoins("56uatom"))
	s.Collect(lpAddr2, lpAddr2, position2.Id, utils.ParseCoins(""))
}

func (s *KeeperTestSuite) TestTerminatePrivateFarmingPlan() {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	pool := s.CreatePool(utils.TestAddress(0), market.Id, utils.ParseDec("5"), true)
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
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
	farmingPoolAddr := sdk.MustAccAddressFromBech32(farmingPlan.FarmingPoolAddress)
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
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	pool := s.CreatePool(utils.TestAddress(0), market.Id, utils.ParseDec("5"), true)
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, lpAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
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
		lpAddr, lpAddr, pool.Id, utils.ParseDec("4.9999"), utils.ParseDec("5.0001"),
		utils.ParseCoins("10000_000000000000000000ucre,50000_000000000000000000uusd"))
	s.CreatePrivateFarmingPlan(
		utils.TestAddress(2), "Farming plan", utils.TestAddress(2), []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("1_000000uatom")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000uatom"), true)
	s.NextBlock()
	s.NextBlock()
	s.Require().Equal("113uatom", s.CollectibleCoins(position.Id).String())
}

func (s *KeeperTestSuite) TestFarmingTooSmallLiquidity() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	pool.TickSpacing = 1
	s.keeper.SetPool(s.Ctx, pool)
	lpAddr := s.FundedAccount(1, enoughCoins)
	position, _, _ := s.AddLiquidity(
		lpAddr, lpAddr, pool.Id, utils.ParseDec("0.0000001"), utils.ParseDec("10000000"),
		utils.ParseCoins("10ucre,50uusd"))
	creatorAddr := s.FundedAccount(100, utils.ParseCoins("1uibc1")) // Create supply.
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("10_000000000000000000uibc1")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("1000_000000000000000000uibc1"), true)
	s.NextBlock()
	s.NextBlock()
	s.Require().Equal("1157407407407405uibc1", s.CollectibleCoins(position.Id).String())
}
