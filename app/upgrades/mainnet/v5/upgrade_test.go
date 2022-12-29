package v5_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	chain "github.com/crescent-network/crescent/v4/app"
	v5 "github.com/crescent-network/crescent/v4/app/upgrades/mainnet/v5"
	"github.com/crescent-network/crescent/v4/cmd/crescentd/cmd"
)

type UpgradeTestSuite struct {
	suite.Suite
	ctx sdk.Context
	app *chain.App
}

func (s *UpgradeTestSuite) SetupTest() {
	cmd.GetConfig()
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Height: 1})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

const testUpgradeHeight = 10

func (s *UpgradeTestSuite) TestUpgradeV5() {
	testCases := []struct {
		title   string
		before  func()
		after   func()
		expPass bool
	}{
		{
			"v5 add permissioned wasmd module",
			func() {
				wasmParams := s.app.WasmKeeper.GetParams(s.ctx)
				s.Require().EqualValues(wasmParams.CodeUploadAccess.Permission, wasmtypes.AccessTypeEverybody)
				s.Require().EqualValues(wasmParams.InstantiateDefaultPermission, wasmtypes.AccessTypeEverybody)
			},
			func() {
				wasmParams := s.app.WasmKeeper.GetParams(s.ctx)
				s.Require().EqualValues(wasmParams.CodeUploadAccess.Permission, wasmtypes.AccessTypeNobody)
				s.Require().EqualValues(wasmParams.InstantiateDefaultPermission, wasmtypes.AccessTypeEverybody)
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.title, func() {
			s.SetupTest()

			tc.before()

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight - 1)
			plan := upgradetypes.Plan{Name: v5.UpgradeName, Height: testUpgradeHeight}
			err := s.app.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
			s.Require().NoError(err)

			_, exists := s.app.UpgradeKeeper.GetUpgradePlan(s.ctx)
			s.Require().True(exists)

			s.ctx = s.ctx.WithBlockHeight(testUpgradeHeight)
			s.Require().NotPanics(func() {
				s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{})
			})

			tc.after()
		})
	}
}
