package keeper_test

import (
	"testing"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v4/app"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"

	"github.com/crescent-network/crescent/v4/x/bootstrap"
	"github.com/crescent-network/crescent/v4/x/bootstrap/keeper"
	"github.com/crescent-network/crescent/v4/x/bootstrap/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
	denom4 = "denom4"
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000_000),
		sdk.NewInt64Coin(denom1, 1_000_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000_000),
		sdk.NewInt64Coin(denom3, 1_000_000_000),
		sdk.NewInt64Coin(denom4, 1_000_000_000))
)

type KeeperTestSuite struct {
	suite.Suite

	app        *chain.App
	ctx        sdk.Context
	keeper     keeper.Keeper
	querier    keeper.Querier
	msgServer  types.MsgServer
	govHandler govtypes.Handler
	addrs      []sdk.AccAddress
	pairs      []liquiditytypes.Pair
	pools      []liquiditytypes.Pool
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := chain.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.app = app
	s.ctx = ctx
	s.keeper = s.app.BootstrapKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.govHandler = bootstrap.NewBootstrapProposalHandler(s.keeper)
	s.addrs = chain.AddTestAddrs(s.app, s.ctx, 30, sdk.ZeroInt())
	for _, addr := range s.addrs {
		err := chain.FundAccount(s.app.BankKeeper, s.ctx, addr, initialBalances)
		s.Require().NoError(err)
	}

	// set testing params
	params := s.keeper.GetParams(ctx)
	params.QuoteCoinWhitelist = []string{denom2, denom3}
	s.keeper.SetParams(ctx, params)

	pair1 := s.createPair(s.addrs[0], denom1, denom2, true)
	pool1 := s.createPool(s.addrs[0], pair1.Id, sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(10000000)), sdk.NewCoin(denom2, sdk.NewInt(10000000))), true)
	s.pairs = append(s.pairs, pair1)
	s.pools = append(s.pools, pool1)

	pair2 := s.createPair(s.addrs[0], denom1, denom3, true)
	pool2 := s.createPool(s.addrs[0], pair2.Id, sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(10000000)), sdk.NewCoin(denom3, sdk.NewInt(10000000))), true)
	s.pairs = append(s.pairs, pair2)
	s.pools = append(s.pools, pool2)

	pair3 := s.createPair(s.addrs[0], denom3, denom1, true)
	pool3 := s.createPool(s.addrs[0], pair3.Id, sdk.NewCoins(sdk.NewCoin(denom3, sdk.NewInt(10000000)), sdk.NewCoin(denom1, sdk.NewInt(10000000))), true)
	s.pairs = append(s.pairs, pair3)
	s.pools = append(s.pools, pool3)

	pair4 := s.createPair(s.addrs[0], denom3, denom4, true)
	pool4 := s.createPool(s.addrs[0], pair4.Id, sdk.NewCoins(sdk.NewCoin(denom3, sdk.NewInt(10000000)), sdk.NewCoin(denom4, sdk.NewInt(10000000))), true)
	s.pairs = append(s.pairs, pair4)
	s.pools = append(s.pools, pool4)

}

func (suite *KeeperTestSuite) AddTestAddrs(num int, coins sdk.Coins) []sdk.AccAddress {
	addrs := chain.AddTestAddrs(suite.app, suite.ctx, num, sdk.ZeroInt())
	for _, addr := range addrs {
		err := chain.FundAccount(suite.app.BankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)
	}
	return addrs
}

func (s *KeeperTestSuite) handleProposal(content govtypes.Content) {
	s.T().Helper()
	err := content.ValidateBasic()
	s.Require().NoError(err)
	err = s.govHandler(s.ctx, content)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	err := s.app.BankKeeper.MintCoins(s.ctx, liquiditytypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, liquiditytypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liquiditytypes.Pair {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.Pool {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
	pool, err := s.app.LiquidityKeeper.CreatePool(s.ctx, liquiditytypes.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}
