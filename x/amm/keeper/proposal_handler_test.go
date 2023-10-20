package keeper_test

import (
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestPoolParameterChangeProposal() {
	handler := amm.NewProposalHandler(s.keeper)
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	// Change tick spacing only
	proposal := types.NewPoolParameterChangeProposal(
		"Title", "Description", []types.PoolParameterChange{
			types.NewPoolParameterChange(pool.Id, 10),
		})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().NoError(handler(s.Ctx, proposal))

	pool, _ = s.keeper.GetPool(s.Ctx, pool.Id)
	s.Require().EqualValues(10, pool.TickSpacing)

	// Failing cases
	proposal = types.NewPoolParameterChangeProposal(
		"Title", "Description", []types.PoolParameterChange{
			types.NewPoolParameterChange(pool.Id, 10),
		})
	s.Require().NoError(proposal.ValidateBasic())
	// Same tick spacing
	s.Require().EqualError(handler(s.Ctx, proposal), "tick spacing is not changed: 10: invalid request")
}

func (s *KeeperTestSuite) TestPublicFarmingPlanProposal() {
	handler := amm.NewProposalHandler(s.keeper)
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))

	farmingPoolAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre"))
	createPlanReq := types.NewCreatePublicFarmingPlanRequest(
		"Farming Plan", farmingPoolAddr, farmingPoolAddr,
		[]types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	proposal := types.NewPublicFarmingPlanProposal(
		"Title", "Description",
		[]types.CreatePublicFarmingPlanRequest{createPlanReq}, nil)
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().NoError(handler(s.Ctx, proposal))

	publicPlan, found := s.keeper.GetFarmingPlan(s.Ctx, 1)
	s.Require().True(found)
	s.Require().Equal(farmingPoolAddr.String(), publicPlan.FarmingPoolAddress)
	s.Require().Equal(farmingPoolAddr.String(), publicPlan.TerminationAddress)
	s.Require().False(publicPlan.IsPrivate)

	terminatePlanReq := types.NewTerminateFarmingPlanRequest(publicPlan.Id)
	proposal = types.NewPublicFarmingPlanProposal(
		"Title", "Description",
		nil, []types.TerminateFarmingPlanRequest{terminatePlanReq})
	s.Require().NoError(handler(s.Ctx, proposal))

	publicPlan, found = s.keeper.GetFarmingPlan(s.Ctx, publicPlan.Id)
	s.Require().True(found)
	s.Require().True(publicPlan.IsTerminated)
	// Balances not changed.
	s.Require().Equal("10000000000ucre", s.GetAllBalances(farmingPoolAddr).String())

	creatorAddr := s.FundedAccount(2, enoughCoins)
	privPlan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	terminatePlanReq = types.NewTerminateFarmingPlanRequest(privPlan.Id)
	proposal = types.NewPublicFarmingPlanProposal(
		"Title", "Description",
		nil, []types.TerminateFarmingPlanRequest{terminatePlanReq})
	s.Require().NoError(proposal.ValidateBasic())
	// It is possible to terminate private plans via PublicFarmingPlanProposal.
	s.Require().NoError(handler(s.Ctx, proposal))

	privPlan, found = s.keeper.GetFarmingPlan(s.Ctx, privPlan.Id)
	s.Require().True(found)
	s.Require().True(privPlan.IsTerminated)

	// Failing cases
	terminatePlanReq = types.NewTerminateFarmingPlanRequest(privPlan.Id)
	proposal = types.NewPublicFarmingPlanProposal(
		"Title", "Description",
		nil, []types.TerminateFarmingPlanRequest{terminatePlanReq})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().EqualError(handler(s.Ctx, proposal), "plan is already terminated: invalid request")

	terminatePlanReq = types.NewTerminateFarmingPlanRequest(3)
	proposal = types.NewPublicFarmingPlanProposal(
		"Title", "Description",
		nil, []types.TerminateFarmingPlanRequest{terminatePlanReq})
	s.Require().NoError(proposal.ValidateBasic())
	s.Require().EqualError(handler(s.Ctx, proposal), "farming plan 3 not found: not found")
}
