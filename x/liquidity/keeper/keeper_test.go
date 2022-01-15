package keeper_test

import (
	"encoding/binary"
	"testing"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidity/keeper"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
)

var (
	initialBalances = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000_000_000),
		sdk.NewInt64Coin(denom1, 100_000_000_000_000),
		sdk.NewInt64Coin(denom2, 100_000_000_000_000),
	)
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.CrescentApp
	ctx         sdk.Context
	keeper      keeper.Keeper
	querier     keeper.Querier
	srv         types.MsgServer
	addrs       []sdk.AccAddress
	samplePools []types.Pool
	samplePairs []types.Pair
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidityKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.srv = keeper.NewMsgServerImpl(s.keeper)
	s.addrs = simapp.AddTestAddrs(s.app, s.ctx, 6, sdk.ZeroInt())
	for _, addr := range s.addrs {
		err := simapp.FundAccount(s.app.BankKeeper, s.ctx, addr, initialBalances)
		s.Require().NoError(err)
	}
	s.samplePools = []types.Pool{
		{
			Id:                    1,
			PairId:                1,
			XCoinDenom:            denom1,
			YCoinDenom:            denom2,
			ReserveAddress:        types.PoolReserveAcc(1).String(),
			PoolCoinDenom:         types.PoolCoinDenom(1),
			LastDepositRequestId:  0,
			LastWithdrawRequestId: 0,
		},
	}
	s.samplePairs = []types.Pair{
		{
			Id:                      uint64(1),
			XCoinDenom:              denom1,
			YCoinDenom:              denom2,
			EscrowAddress:           types.PairEscrowAddr(1).String(),
			LastSwapRequestId:       0,
			LastCancelSwapRequestId: 0,
			LastPrice:               nil,
			CurrentBatchId:          1,
		},
	}
}

func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins {
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	err := s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, xCoin, yCoin sdk.Coin) types.Pool {
	k, ctx := s.keeper, s.ctx

	params := k.GetParams(ctx)
	depositCoins := sdk.NewCoins(xCoin, yCoin)
	s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	pool, err := k.CreatePool(ctx, types.NewMsgCreatePool(creator, xCoin, yCoin))
	s.Require().NoError(err)
	return pool
}

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

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
