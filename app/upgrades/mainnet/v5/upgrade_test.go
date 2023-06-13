package v5_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	v5 "github.com/crescent-network/crescent/v5/app/upgrades/mainnet/v5"
)

type UpgradeTestSuite struct {
	testutil.TestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgradeV5() {
	// Set the upgrade plan.
	upgradeHeight := s.Ctx.BlockHeight() + 1
	plan := upgradetypes.Plan{Name: v5.UpgradeName, Height: upgradeHeight}
	s.Require().NoError(s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan))
	_, havePlan := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(havePlan)

	// TODO: write test

	// Let the upgrade happen.
	s.NextBlock()

	// TODO: write test
}
