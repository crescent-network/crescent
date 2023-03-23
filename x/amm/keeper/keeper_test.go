package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v5/app"
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
)

type KeeperTestSuite struct {
	suite.Suite

	app *chain.App
	ctx sdk.Context
	k   keeper.Keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.k = suite.app.AMMKeeper
}
