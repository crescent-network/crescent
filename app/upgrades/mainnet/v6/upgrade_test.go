package v6_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	v6 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v6"
	utils "github.com/crescent-network/crescent/v5/types"
)

type UpgradeTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradev6() {
	enoughCoins := utils.ParseCoins(
		"1000000000000000ucre,1000000000000000uusd,1000000000000000uatom,1000000000000000stake")
	creatorAddr := s.FundedAccount(1, enoughCoins)
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, creatorAddr)
	_ = acc.SetSequence(1)
	_ = acc.SetPubKey(ed25519.GenPrivKey().PubKey())

	// TODO: add logics ...

	s.NextBlock()

	// Set the upgrade plan.
	upgradeHeight := s.Ctx.BlockHeight() + 1
	upgradePlan := upgradetypes.Plan{Name: v6.UpgradeName, Height: upgradeHeight}
	s.Require().NoError(s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, upgradePlan))
	_, havePlan := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(havePlan)

	// Let the upgrade happen.
	s.NextBlock()

	// TODO: check new module's param
}
