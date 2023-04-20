package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	minttypes "github.com/crescent-network/crescent/v5/x/mint/types"
)

type TestSuite struct {
	suite.Suite

	App *chain.App
	Ctx sdk.Context
}

func (s *TestSuite) SetupTest() {
	s.App = chain.Setup(false)
	s.Ctx = s.App.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *TestSuite) FundAccount(
	addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	s.Require().NoError(s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amt))
	s.Require().NoError(s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, addr, amt))
}
