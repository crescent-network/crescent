package keeper_test

import (
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

func (s *KeeperTestSuite) TestFarmingPlanProposalHandler() {
	farmingPoolAddr := utils.TestAddress(0)
	s.fundAddr(farmingPoolAddr, utils.ParseCoins("100_000000stake"))

	pair := s.createPair("denom1", "denom2")
	createPlanReq := types.NewCreatePlanRequest(
		"Farming Plan #1", farmingPoolAddr,
		[]types.RewardAllocation{
			types.NewPairRewardAllocation(pair.Id, utils.ParseCoins("100_000000stake")),
		}, sampleStartTime, sampleEndTime)
	proposal := types.NewFarmingPlanProposal(
		"Create a new public farming plan", "Description",
		[]types.CreatePlanRequest{createPlanReq}, nil)
	s.handleProposal(proposal)

	plan, found := s.keeper.GetPlan(s.ctx, 1)
	s.Require().True(found)
	s.Require().Equal(farmingPoolAddr.String(), plan.FarmingPoolAddress)
	s.Require().Equal(farmingPoolAddr.String(), plan.TerminationAddress)
	s.Require().False(plan.IsPrivate)

	terminatePlanReq := types.NewTerminatePlanRequest(1)
	proposal = types.NewFarmingPlanProposal(
		"Terminate the public farming plan", "Description",
		nil, []types.TerminatePlanRequest{terminatePlanReq})
	s.handleProposal(proposal)

	plan, found = s.keeper.GetPlan(s.ctx, 1)
	s.Require().True(found)
	s.Require().True(plan.IsTerminated)
	// Balances not changed.
	s.assertEq(utils.ParseCoins("100_000000stake"), s.getBalances(farmingPoolAddr))

	privPlan := s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(pair.Id, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))
	terminatePlanReq = types.NewTerminatePlanRequest(privPlan.Id)
	proposal = types.NewFarmingPlanProposal(
		"Terminate a private farming plan", "Description",
		nil, []types.TerminatePlanRequest{terminatePlanReq})
	s.Require().NoError(proposal.ValidateBasic())
	// It isn't possible to terminate private plans via FarmingPlanProposal.
	s.Require().Error(s.govHandler(s.ctx, proposal))
}
