package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestSoleFarmer() {
	pair := s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	plan := s.createPrivatePlan([]types.RewardAllocation{
		{
			PairId:        pair.Id,
			RewardsPerDay: utils.ParseCoins("100_000000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("10000_000000stake"))

	farmerAddr := utils.TestAddress(0)
	pool := s.createPool(farmerAddr, pair.Id, utils.ParseCoins("1000_000000denom1,1000_000000denom2"))
	_, err := s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	s.Require().NoError(err)

	s.nextBlock()

	farm, _ := s.keeper.GetFarm(s.ctx, pool.PoolCoinDenom)
	fmt.Println(farm.CurrentRewards, farm.OutstandingRewards)

	_, err = s.keeper.Farm(s.ctx, farmerAddr, utils.ParseCoin("1_000000pool1"))
	s.Require().NoError(err)

	farm, _ = s.keeper.GetFarm(s.ctx, pool.PoolCoinDenom)
	fmt.Println(farm.CurrentRewards, farm.OutstandingRewards)
}
