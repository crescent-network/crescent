package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/types"
)

func (s *KeeperTestSuite) TestFarm() {
	pair := s.createPair("denom1", "denom2")
	pool := s.createPool(pair.Id, utils.ParseCoins("1000_000000denom1,1000_000000denom2"))
	plan := s.createPrivatePlan([]types.RewardAllocation{
		{
			PairId:        pair.Id,
			RewardsPerDay: utils.ParseDecCoins("100_000000reward"),
		},
	})
	farmingPoolAddr, _ := sdk.AccAddressFromBech32(plan.FarmingPoolAddress)
	s.fundAddr(farmingPoolAddr, utils.ParseCoins("10000_000000reward"))

	farmerAddr := utils.TestAddress(0)
	farmerAddr2 := utils.TestAddress(1)
	s.assertEq(sdk.DecCoins{}, s.rewards(farmerAddr, pool.PoolCoinDenom))

	s.deposit(farmerAddr, pool.Id, utils.ParseCoins("1_000000denom1,1_000000denom2"))
	s.deposit(farmerAddr2, pool.Id, utils.ParseCoins("1_000000denom1,1_000000denom2"))
	s.nextBlock()

	poolCoin := s.getBalance(farmerAddr, pool.PoolCoinDenom)
	_, err := s.keeper.Farm(s.ctx, farmerAddr, poolCoin)
	s.Require().NoError(err)
	s.assertEq(sdk.DecCoins{}, s.rewards(farmerAddr, pool.PoolCoinDenom))

	poolCoin2 := s.getBalance(farmerAddr2, pool.PoolCoinDenom)
	_, err = s.keeper.Farm(s.ctx, farmerAddr2, poolCoin2)
	s.Require().NoError(err)

	s.nextBlock()
	fmt.Println(s.rewards(farmerAddr, pool.PoolCoinDenom))
	fmt.Println(s.rewards(farmerAddr2, pool.PoolCoinDenom))

	withdrawnRewards, err := s.keeper.Harvest(s.ctx, farmerAddr, pool.PoolCoinDenom)
	s.Require().NoError(err)
	fmt.Println("withdrawn", withdrawnRewards)

	s.nextBlock()
	fmt.Println(s.rewards(farmerAddr, pool.PoolCoinDenom))
	fmt.Println(s.rewards(farmerAddr2, pool.PoolCoinDenom))
}
