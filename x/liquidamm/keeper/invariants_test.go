package keeper_test

import (
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
)

func (s *KeeperTestSuite) TestShareSupplyInvariant() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr := utils.TestAddress(1)
	s.MintShare(
		minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)

	s.NextBlock()
	_, broken := keeper.ShareSupplyInvariant(s.keeper)(s.Ctx)
	s.Require().False(broken)

	// mint sb token manually.
	s.FundAccount(minterAddr, utils.ParseCoins("1000000sb1"))

	_, broken = keeper.ShareSupplyInvariant(s.keeper)(s.Ctx)
	s.Require().True(broken)
}
