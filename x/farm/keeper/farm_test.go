package keeper_test

import (
	utils "github.com/crescent-network/crescent/v3/types"
)

func (s *KeeperTestSuite) TestFarm() {
	farmerAddr := utils.TestAddress(0)
	_, err := s.farm(farmerAddr, utils.ParseCoin("1000000denom1"), true)
	s.Require().NoError(err)
}
