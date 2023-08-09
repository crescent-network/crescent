package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

type KeeperTestSuite struct {
	testutil.TestSuite
	keeper  keeper.Keeper
	querier keeper.Querier
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supply
	s.keeper = s.App.LiquidAMMKeeper
	s.querier = keeper.Querier{Keeper: s.App.LiquidAMMKeeper}
}

func (s *KeeperTestSuite) CreateSamplePublicPosition() types.PublicPosition {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	s.CreatePrivateFarmingPlan(utils.TestAddress(0), "", utils.TestAddress(0), []ammtypes.FarmingRewardAllocation{
		ammtypes.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000uatom")),
	}, utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T00:00:00Z"),
		utils.ParseCoins("100000_000000uatom"), true)
	return s.CreatePublicPosition(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))
}

func (s *KeeperTestSuite) SetupSampleScenario() {
	market1 := s.CreateMarket("ucre", "uusd")
	pool1 := s.CreatePool(market1.Id, utils.ParseDec("5"))
	s.CreatePrivateFarmingPlan(utils.TestAddress(0), "", utils.TestAddress(0), []ammtypes.FarmingRewardAllocation{
		ammtypes.NewFarmingRewardAllocation(1, utils.ParseCoins("100_000000uatom")),
	}, utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T00:00:00Z"),
		utils.ParseCoins("100000_000000uatom"), true)
	enoughCoins := utils.ParseCoins("100000_000000uatom,100000_000000ucre,100000_000000uusd")
	lpAddr := s.FundedAccount(1, enoughCoins)
	s.AddLiquidity(
		lpAddr, pool1.Id, utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseCoins("1000_000000ucre,5000_000000uusd"))

	market2 := s.CreateMarket("uatom", "uusd")
	pool2 := s.CreatePool(market2.Id, utils.ParseDec("10"))

	s.NextBlock()
	s.NextBlock()

	publicPosition1 := s.CreatePublicPosition(
		pool1.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))
	s.CreatePublicPosition(
		pool2.Id, utils.ParseDec("5"), utils.ParseDec("20"),
		sdk.NewInt(10000), utils.ParseDec("0.003"))

	// Two account mints share.
	minterAddr1 := utils.TestAddress(2)
	minterAddr2 := utils.TestAddress(3)
	s.MintShare(minterAddr1, publicPosition1.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(minterAddr2, publicPosition1.Id, utils.ParseCoins("300_000000ucre,1500_000000uusd"), true)

	// Auction starts and rewards are accrued
	s.AdvanceRewardsAuctions()
	s.NextBlock()
	s.NextBlock()

	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition1.Id)
	s.Require().True(found)

	bidderAddr1 := utils.TestAddress(4)
	bidderAddr2 := utils.TestAddress(5)
	bidderAddr3 := utils.TestAddress(6)
	bidderShare1, _, _, _ := s.MintShare(bidderAddr1, publicPosition1.Id, utils.ParseCoins("10_000000ucre,50_000000uusd"), true)
	bidderShare2, _, _, _ := s.MintShare(bidderAddr2, publicPosition1.Id, utils.ParseCoins("20_000000ucre,100_000000uusd"), true)
	bidderShare3, _, _, _ := s.MintShare(bidderAddr3, publicPosition1.Id, utils.ParseCoins("30_000000ucre,150_000000uusd"), true)
	s.PlaceBid(bidderAddr1, publicPosition1.Id, auction.Id, bidderShare1)
	s.PlaceBid(bidderAddr2, publicPosition1.Id, auction.Id, bidderShare2)
	s.PlaceBid(bidderAddr3, publicPosition1.Id, auction.Id, bidderShare3)

	s.AdvanceRewardsAuctions()
	s.NextBlock()
	s.NextBlock()

	minterAddr3 := utils.TestAddress(7)
	s.MintShare(minterAddr3, publicPosition1.Id, utils.ParseCoins("500_000000ucre,2500_000000uusd"), true)

	auction, _ = s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition1.Id)
	bidderShare1, _, _, _ = s.MintShare(bidderAddr1, publicPosition1.Id, utils.ParseCoins("10_000000ucre,50_000000uusd"), true)
	bidderShare2, _, _, _ = s.MintShare(bidderAddr2, publicPosition1.Id, utils.ParseCoins("20_000000ucre,100_000000uusd"), true)
	s.PlaceBid(bidderAddr1, publicPosition1.Id, auction.Id, bidderShare1)
	s.PlaceBid(bidderAddr2, publicPosition1.Id, auction.Id, bidderShare2)
}
