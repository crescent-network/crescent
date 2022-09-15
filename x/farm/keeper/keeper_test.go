package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v3/app"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/keeper"
	minttypes "github.com/crescent-network/crescent/v3/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app    *chain.App
	ctx    sdk.Context
	keeper keeper.Keeper
	hdr    tmproto.Header
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.hdr = tmproto.Header{
		Height: 1,
		Time:   utils.ParseTime("2022-01-01T00:00:00Z"),
	}
	s.beginBlock()
	s.keeper = s.app.FarmKeeper
}

func (s *KeeperTestSuite) beginBlock() {
	s.app.BeginBlock(abci.RequestBeginBlock{Header: s.hdr})
	s.ctx = s.app.BaseApp.NewContext(false, s.hdr)
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{Header: s.hdr})
}

func (s *KeeperTestSuite) endBlock() {
	s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.app.EndBlock(abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
}

func (s *KeeperTestSuite) nextBlock() {
	s.endBlock()
	s.hdr.Height++
	s.hdr.Time = s.hdr.Time.Add(5 * time.Second)
	s.beginBlock()
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) assertEq(exp, got interface{}) {
	s.T().Helper()
	var same bool
	switch exp := exp.(type) {
	case sdk.Int:
		same = exp.Equal(got.(sdk.Int))
	case sdk.Dec:
		same = exp.Equal(got.(sdk.Dec))
	case sdk.Coin:
		same = exp.IsEqual(got.(sdk.Coin))
	case sdk.Coins:
		same = exp.IsEqual(got.(sdk.Coins))
	}
	s.Require().True(same, "expected:\t%v\ngot:\t\t%v", exp, got)
}

func (s *KeeperTestSuite) farm(farmerAddr sdk.AccAddress, coin sdk.Coin, fund bool) (withdrawnRewards sdk.Coins, err error) {
	s.T().Helper()
	if fund {
		s.fundAddr(farmerAddr, sdk.NewCoins(coin))
	}
	return s.keeper.Farm(s.ctx, farmerAddr, coin)
}
