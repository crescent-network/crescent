package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app       *chain.App
	ctx       sdk.Context
	keeper    keeper.Keeper
	querier   keeper.Querier
	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	hdr := tmproto.Header{
		Height: 1,
		Time:   utils.ParseTime("2022-01-01T00:00:00Z"),
	}
	s.app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
	s.ctx = s.app.BaseApp.NewContext(false, hdr)
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{Header: hdr})
	s.keeper = s.app.LiquidityKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

// Below are just shortcuts to frequently-used functions.
func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins {
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

func (s *KeeperTestSuite) sendCoins(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.SendCoins(s.ctx, fromAddr, toAddr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) nextBlock() {
	s.T().Helper()
	s.app.EndBlock(abci.RequestEndBlock{})
	s.app.Commit()
	hdr := tmproto.Header{
		Height: s.app.LastBlockHeight() + 1,
		Time:   s.ctx.BlockTime().Add(5 * time.Second),
	}
	s.app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
	s.ctx = s.app.BaseApp.NewContext(false, hdr)
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{Header: hdr})
}

// Below are useful helpers to write test code easily.
func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) types.Pair {
	s.T().Helper()
	if fund {
		s.fundAddr(creator, s.keeper.GetPairCreationFee(s.ctx))
	}
	msg := types.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom)
	s.Require().NoError(msg.ValidateBasic())
	pair, err := s.keeper.CreatePair(s.ctx, msg)
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) types.Pool {
	s.T().Helper()
	if fund {
		s.fundAddr(creator, depositCoins.Add(s.keeper.GetPoolCreationFee(s.ctx)...))
	}
	msg := types.NewMsgCreatePool(creator, pairId, depositCoins)
	s.Require().NoError(msg.ValidateBasic())
	pool, err := s.keeper.CreatePool(s.ctx, msg)
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) createRangedPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, minPrice, maxPrice, initialPrice sdk.Dec, fund bool) types.Pool {
	s.T().Helper()
	if fund {
		s.fundAddr(creator, depositCoins.Add(s.keeper.GetPoolCreationFee(s.ctx)...))
	}
	msg := types.NewMsgCreateRangedPool(creator, pairId, depositCoins, minPrice, maxPrice, initialPrice)
	s.Require().NoError(msg.ValidateBasic())
	pool, err := s.keeper.CreateRangedPool(s.ctx, msg)
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) deposit(depositor sdk.AccAddress, poolId uint64, depositCoins sdk.Coins, fund bool) types.DepositRequest {
	s.T().Helper()
	if fund {
		s.fundAddr(depositor, depositCoins)
	}
	req, err := s.keeper.Deposit(s.ctx, types.NewMsgDeposit(depositor, poolId, depositCoins))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) withdraw(withdrawer sdk.AccAddress, poolId uint64, poolCoin sdk.Coin) types.WithdrawRequest {
	s.T().Helper()
	req, err := s.keeper.Withdraw(s.ctx, types.NewMsgWithdraw(withdrawer, poolId, poolCoin))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) limitOrder(
	orderer sdk.AccAddress, pairId uint64, dir types.OrderDirection,
	price sdk.Dec, amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	pair, found := s.keeper.GetPair(s.ctx, pairId)
	s.Require().True(found)
	var ammDir amm.OrderDirection
	var offerCoinDenom, demandCoinDenom string
	switch dir {
	case types.OrderDirectionBuy:
		ammDir = amm.Buy
		offerCoinDenom, demandCoinDenom = pair.QuoteCoinDenom, pair.BaseCoinDenom
	case types.OrderDirectionSell:
		ammDir = amm.Sell
		offerCoinDenom, demandCoinDenom = pair.BaseCoinDenom, pair.QuoteCoinDenom
	}
	offerCoin := sdk.NewCoin(offerCoinDenom, amm.OfferCoinAmount(ammDir, price, amt))
	if fund {
		s.fundAddr(orderer, sdk.NewCoins(offerCoin))
	}
	msg := types.NewMsgLimitOrder(
		orderer, pairId, dir, offerCoin, demandCoinDenom,
		price, amt, orderLifespan)
	s.Require().NoError(msg.ValidateBasic())
	req, err := s.keeper.LimitOrder(s.ctx, msg)
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) buyLimitOrder(
	orderer sdk.AccAddress, pairId uint64, price sdk.Dec,
	amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	return s.limitOrder(
		orderer, pairId, types.OrderDirectionBuy, price, amt, orderLifespan, fund)
}

func (s *KeeperTestSuite) sellLimitOrder(
	orderer sdk.AccAddress, pairId uint64, price sdk.Dec,
	amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	return s.limitOrder(
		orderer, pairId, types.OrderDirectionSell, price, amt, orderLifespan, fund)
}

func (s *KeeperTestSuite) marketOrder(
	orderer sdk.AccAddress, pairId uint64, dir types.OrderDirection,
	amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	pair, found := s.keeper.GetPair(s.ctx, pairId)
	s.Require().True(found)
	s.Require().NotNil(pair.LastPrice)
	lastPrice := *pair.LastPrice
	var offerCoin sdk.Coin
	var demandCoinDenom string
	switch dir {
	case types.OrderDirectionBuy:
		maxPrice := lastPrice.Mul(sdk.OneDec().Add(s.keeper.GetMaxPriceLimitRatio(s.ctx)))
		offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, amm.OfferCoinAmount(amm.Buy, maxPrice, amt))
		demandCoinDenom = pair.BaseCoinDenom
	case types.OrderDirectionSell:
		offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
		demandCoinDenom = pair.QuoteCoinDenom
	}
	if fund {
		s.fundAddr(orderer, sdk.NewCoins(offerCoin))
	}
	msg := types.NewMsgMarketOrder(
		orderer, pairId, dir, offerCoin, demandCoinDenom,
		amt, orderLifespan)
	s.Require().NoError(msg.ValidateBasic())
	req, err := s.keeper.MarketOrder(s.ctx, msg)
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) buyMarketOrder(
	orderer sdk.AccAddress, pairId uint64,
	amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	return s.marketOrder(
		orderer, pairId, types.OrderDirectionBuy, amt, orderLifespan, fund)
}

//nolint
func (s *KeeperTestSuite) sellMarketOrder(
	orderer sdk.AccAddress, pairId uint64,
	amt sdk.Int, orderLifespan time.Duration, fund bool) types.Order {
	s.T().Helper()
	return s.marketOrder(
		orderer, pairId, types.OrderDirectionSell, amt, orderLifespan, fund)
}

//nolint
func (s *KeeperTestSuite) cancelOrder(orderer sdk.AccAddress, pairId, orderId uint64) {
	s.T().Helper()
	err := s.keeper.CancelOrder(s.ctx, types.NewMsgCancelOrder(orderer, pairId, orderId))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) cancelAllOrders(orderer sdk.AccAddress, pairIds []uint64) {
	s.T().Helper()
	err := s.keeper.CancelAllOrders(s.ctx, types.NewMsgCancelAllOrders(orderer, pairIds))
	s.Require().NoError(err)
}

func coinEq(exp, got sdk.Coin) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func intEq(exp, got sdk.Int) (bool, string, string, string) {
	return exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func decEq(exp, got sdk.Dec) (bool, string, string, string) {
	return exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func newInt(i int64) sdk.Int {
	return sdk.NewInt(i)
}
