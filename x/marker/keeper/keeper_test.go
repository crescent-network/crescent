package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v4/app"
	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/marker/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	app     *chain.App
	ctx     sdk.Context
	keeper  keeper.Keeper
	querier keeper.Querier
	hdr     tmproto.Header
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.keeper = s.app.MarkerKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.hdr = tmproto.Header{
		Height: 1,
		Time:   utils.ParseTime("2022-01-01T00:00:00Z"),
	}
	s.beginBlock()
}

func (s *KeeperTestSuite) beginBlock() {
	s.T().Helper()
	s.app.BeginBlock(abci.RequestBeginBlock{Header: s.hdr})
	s.ctx = s.app.BaseApp.NewContext(false, s.hdr)
}

func (s *KeeperTestSuite) endBlock() {
	s.T().Helper()
	s.app.EndBlock(abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
	s.app.Commit()
}

func (s *KeeperTestSuite) nextBlock() {
	s.T().Helper()
	s.endBlock()
	s.hdr.Height++
	s.hdr.Time = s.hdr.Time.Add(5 * time.Second)
	s.beginBlock()
}
