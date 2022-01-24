package keeper_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	crescentapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/keeper"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app       *crescentapp.CrescentApp
	ctx       sdk.Context
	keeper    keeper.Keeper
	querier   keeper.Querier
	msgServer types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = crescentapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
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
	err := s.app.BankKeeper.SendCoins(s.ctx, fromAddr, toAddr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) nextBlock() {
	liquidity.EndBlocker(s.ctx, s.keeper)
	liquidity.BeginBlocker(s.ctx, s.keeper)
}

// Below are useful helpers to write test code easily.
func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) types.Pair {
	params := s.keeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.keeper.CreatePair(s.ctx, types.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) types.Pool {
	params := s.keeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
	pool, err := s.keeper.CreatePool(s.ctx, types.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) depositBatch(depositor sdk.AccAddress, poolId uint64, depositCoins sdk.Coins, fund bool) types.DepositRequest {
	if fund {
		s.fundAddr(depositor, depositCoins)
	}
	req, err := s.keeper.DepositBatch(s.ctx, types.NewMsgDepositBatch(depositor, poolId, depositCoins))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) withdrawBatch(withdrawer sdk.AccAddress, poolId uint64, poolCoin sdk.Coin) types.WithdrawRequest {
	req, err := s.keeper.WithdrawBatch(s.ctx, types.NewMsgWithdrawBatch(withdrawer, poolId, poolCoin))
	s.Require().NoError(err)
	return req
}

//nolint
func parseCoin(s string) sdk.Coin {
	coin, err := sdk.ParseCoinNormalized(s)
	if err != nil {
		panic(err)
	}
	return coin
}

func parseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

//nolint
func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
