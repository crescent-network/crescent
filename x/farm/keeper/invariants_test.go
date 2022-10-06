package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/keeper"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestRewardsInvariants() {
	s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	s.createPool(1, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createPrivatePlan([]types.RewardAllocation{
		types.NewRewardAllocation(1, utils.ParseCoins("100_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	s.farm(farmerAddr, utils.ParseCoin("1_000000pool1"))

	s.nextBlock()

	// Send coins from the rewards pool to another address.
	s.Require().NoError(
		s.app.BankKeeper.SendCoins(
			s.ctx, types.RewardsPoolAddress, utils.TestAddress(1), utils.ParseCoins("100stake")))

	_, broken := keeper.OutstandingRewardsInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	_, broken = keeper.CanWithdrawInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)
}
