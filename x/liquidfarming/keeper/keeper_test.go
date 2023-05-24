package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

type KeeperTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supply
}

func (s *KeeperTestSuite) CreateSampleLiquidFarm() types.LiquidFarm {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	pool := s.CreatePool(utils.TestAddress(0), market.Id, utils.ParseDec("5"), true)
	s.CreatePrivateFarmingPlan(utils.TestAddress(0), "", utils.TestAddress(0), []ammtypes.FarmingRewardAllocation{
		ammtypes.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000uatom")),
	}, utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T00:00:00Z"),
		utils.ParseCoins("100000_000000uatom"), true)
	return s.CreateLiquidFarm(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))
}
