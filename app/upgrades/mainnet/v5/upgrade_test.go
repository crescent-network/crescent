package v5_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	v5 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v5"
	utils "github.com/crescent-network/crescent/v5/types"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"
)

type UpgradeTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradeV5() {
	enoughCoins := utils.ParseCoins(
		"1000000000000000ucre,1000000000000000uusd,1000000000000000uatom,1000000000000000stake")
	creatorAddr := s.FundedAccount(1, enoughCoins)
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, creatorAddr)
	_ = acc.SetSequence(1)
	_ = acc.SetPubKey(ed25519.GenPrivKey().PubKey())
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)
	pair, err := s.App.LiquidityKeeper.CreatePair(s.Ctx, liquiditytypes.NewMsgCreatePair(
		creatorAddr, "ucre", "uusd"))
	s.Require().NoError(err)
	_, err = s.App.LiquidityKeeper.CreatePool(s.Ctx, liquiditytypes.NewMsgCreatePool(
		creatorAddr, pair.Id, utils.ParseCoins("100_000000ucre,500_000000uusd")))
	s.Require().NoError(err)
	_, err = s.App.LiquidityKeeper.CreateRangedPool(s.Ctx, liquiditytypes.NewMsgCreateRangedPool(
		creatorAddr, pair.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"),
		utils.ParseDec("4"), utils.ParseDec("6"), utils.ParseDec("5")))
	s.Require().NoError(err)
	lpfarmPlan, err := s.App.LPFarmKeeper.CreatePrivatePlan(s.Ctx, creatorAddr, "", []lpfarmtypes.RewardAllocation{
		lpfarmtypes.NewPairRewardAllocation(pair.Id, utils.ParseCoins("100_000000uatom")),
	}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	s.Require().NoError(err)
	s.FundAccount(lpfarmPlan.GetFarmingPoolAddress(), enoughCoins)
	s.NextBlock()

	// Set the upgrade plan.
	upgradeHeight := s.Ctx.BlockHeight() + 1
	upgradePlan := upgradetypes.Plan{Name: v5.UpgradeName, Height: upgradeHeight}
	s.Require().NoError(s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, upgradePlan))
	_, havePlan := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(havePlan)

	// Let the upgrade happen.
	s.NextBlock()

	_, found := s.App.AMMKeeper.GetPool(s.Ctx, 1)
	s.Require().True(found)
}
