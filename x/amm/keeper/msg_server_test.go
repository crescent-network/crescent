package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
)

func (s *KeeperTestSuite) TestMsgServer_TerminatePrivateFarmingPlan() {
	_, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("5"))
	creatorAddr := s.FundedAccount(1, enoughCoins)
	termAddr1 := utils.TestAddress(2)
	privPlan := s.CreatePrivateFarmingPlan(
		creatorAddr, "Farming plan", termAddr1, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		},
		utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)

	farmingPoolAddr := s.FundedAccount(3, utils.ParseCoins("10000_000000ucre"))
	termAddr2 := utils.TestAddress(4)
	publicPlan := s.CreatePublicFarmingPlan(
		"Farming plan", farmingPoolAddr, termAddr2, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))

	for _, tc := range []struct {
		name        string
		msg         *types.MsgTerminatePrivateFarmingPlan
		expectedErr string
		postRun     func(resp *types.MsgTerminatePrivateFarmingPlanResponse)
	}{
		{
			"happy case",
			types.NewMsgTerminatePrivateFarmingPlan(termAddr1, privPlan.Id),
			"",
			func(resp *types.MsgTerminatePrivateFarmingPlanResponse) {
				s.Require().Equal("10000000000ucre", s.GetAllBalances(termAddr1).String())
			},
		},
		{
			"wrong sender",
			types.NewMsgTerminatePrivateFarmingPlan(creatorAddr, privPlan.Id),
			"plan's termination address must be same with the sender's address: unauthorized",
			nil,
		},
		{
			"farming plan not found",
			types.NewMsgTerminatePrivateFarmingPlan(creatorAddr, 3),
			"farming plan not found: not found",
			nil,
		},
		{
			"public plan",
			types.NewMsgTerminatePrivateFarmingPlan(termAddr2, publicPlan.Id),
			"cannot terminate public plan: invalid request",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			oldCtx := s.Ctx
			s.Ctx, _ = s.Ctx.CacheContext()
			s.Require().NoError(tc.msg.ValidateBasic())
			resp, err := s.msgServer.TerminatePrivateFarmingPlan(sdk.WrapSDKContext(s.Ctx), tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
			s.Ctx = oldCtx
		})
	}
}
