package keeper_test

//import (
//	"testing"
//
//	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
//	"github.com/stretchr/testify/suite"
//	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
//
//	sdk "github.com/cosmos/cosmos-sdk/types"
//
//	chain "github.com/crescent-network/crescent/v4/app"
//
//	"github.com/crescent-network/crescent/v4/x/bootstrap"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/keeper"
//	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
//)
//
//const (
//	denom1 = "denom1"
//	denom2 = "denom2"
//	denom3 = "denom3"
//)
//
//var (
//	initialBalances = sdk.NewCoins(
//		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000_000),
//		sdk.NewInt64Coin(denom1, 1_000_000_000),
//		sdk.NewInt64Coin(denom2, 1_000_000_000),
//		sdk.NewInt64Coin(denom3, 1_000_000_000))
//)
//
//type KeeperTestSuite struct {
//	suite.Suite
//
//	app        *chain.App
//	ctx        sdk.Context
//	keeper     keeper.Keeper
//	querier    keeper.Querier
//	msgServer  types.MsgServer
//	govHandler govtypes.Handler
//	addrs      []sdk.AccAddress
//}
//
//func TestKeeperTestSuite(t *testing.T) {
//	suite.Run(t, new(KeeperTestSuite))
//}
//
//func (suite *KeeperTestSuite) SetupTest() {
//	app := chain.Setup(false)
//	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
//
//	suite.app = app
//	suite.ctx = ctx
//	suite.keeper = suite.app.BootstrapKeeper
//	suite.querier = keeper.Querier{Keeper: suite.keeper}
//	suite.msgServer = keeper.NewMsgServerImpl(suite.keeper)
//	suite.govHandler = bootstrap.NewBootstrapProposalHandler(suite.keeper)
//	suite.addrs = chain.AddTestAddrs(suite.app, suite.ctx, 30, sdk.ZeroInt())
//	for _, addr := range suite.addrs {
//		err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
//		suite.Require().NoError(err)
//	}
//	suite.SetIncentivePairs()
//}
//
//func (suite *KeeperTestSuite) AddTestAddrs(num int, coins sdk.Coins) []sdk.AccAddress {
//	addrs := chain.AddTestAddrs(suite.app, suite.ctx, num, sdk.ZeroInt())
//	for _, addr := range addrs {
//		err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
//		suite.Require().NoError(err)
//	}
//	return addrs
//}
//
//func (suite *KeeperTestSuite) SetIncentivePairs() {
//	params := suite.keeper.GetParams(suite.ctx)
//	params.IncentivePairs = []types.IncentivePair{
//		{
//			PairId: uint64(1),
//		},
//		{
//			PairId: uint64(2),
//		},
//		{
//			PairId: uint64(3),
//		},
//		{
//			PairId: uint64(4),
//		},
//		{
//			PairId: uint64(5),
//		},
//		{
//			PairId: uint64(6),
//		},
//		{
//			PairId: uint64(7),
//		},
//	}
//	suite.keeper.SetParams(suite.ctx, params)
//}
//
//func (suite *KeeperTestSuite) ResetIncentivePairs() {
//	params := suite.keeper.GetParams(suite.ctx)
//	params.IncentivePairs = []types.IncentivePair{}
//	suite.keeper.SetParams(suite.ctx, params)
//}
//
//func (suite *KeeperTestSuite) handleProposal(content govtypes.Content) {
//	suite.T().Helper()
//	err := content.ValidateBasic()
//	suite.Require().NoError(err)
//	err = suite.govHandler(suite.ctx, content)
//	suite.Require().NoError(err)
//}
