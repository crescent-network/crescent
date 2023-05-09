package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	minttypes "github.com/crescent-network/crescent/v5/x/mint/types"
)

type TestSuite struct {
	suite.Suite

	App *chain.App
	Ctx sdk.Context
}

func (s *TestSuite) SetupTest() {
	s.App = chain.Setup(false)
	s.Ctx = s.App.NewContext(false, tmproto.Header{
		Height: 0,
		Time:   utils.ParseTime("2023-01-01T00:00:00Z"),
	})
	s.BeginBlock()
}

func (s *TestSuite) BeginBlock() {
	s.T().Helper()
	newHeader := s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).
		WithBlockTime(s.Ctx.BlockTime().Add(5 * time.Second)).
		BlockHeader()
	s.App.BeginBlock(abci.RequestBeginBlock{Header: newHeader})
	s.Ctx = s.App.BaseApp.NewContext(false, newHeader)
}

func (s *TestSuite) EndBlock() {
	s.T().Helper()
	s.App.EndBlock(abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})
	s.App.Commit()
}

func (s *TestSuite) NextBlock() {
	s.T().Helper()
	s.EndBlock()
	s.BeginBlock()
}

func (s *TestSuite) FundAccount(
	addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	if amt.IsAllPositive() {
		s.Require().NoError(s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amt))
		s.Require().NoError(s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, addr, amt))
	}
}

func (s *TestSuite) FundedAccount(addrNum int, amt sdk.Coins) sdk.AccAddress {
	s.T().Helper()
	addr := utils.TestAddress(addrNum)
	s.FundAccount(addr, amt)
	return addr
}
