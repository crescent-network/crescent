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
	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	minttypes "github.com/crescent-network/crescent/v3/x/mint/types"
)

var (
	helperAddr      = utils.TestAddress(10000)
	sampleStartTime = utils.ParseTime("0001-01-01T00:00:00Z")
	sampleEndTime   = utils.ParseTime("9999-12-31T23:59:59Z")
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
	s.keeper = s.app.FarmKeeper
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

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) assertEq(exp, got interface{}) {
	s.T().Helper()
	var equal bool
	switch exp := exp.(type) {
	case sdk.Int:
		equal = exp.Equal(got.(sdk.Int))
	case sdk.Dec:
		equal = exp.Equal(got.(sdk.Dec))
	case sdk.Coin:
		equal = exp.IsEqual(got.(sdk.Coin))
	case sdk.Coins:
		equal = exp.IsEqual(got.(sdk.Coins))
	case sdk.DecCoins:
		equal = exp.IsEqual(got.(sdk.DecCoins))
	}
	s.Require().True(equal, "expected:\t%v\ngot:\t\t%v", exp, got)
}

//nolint
func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins {
	s.T().Helper()
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	s.T().Helper()
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

func (s *KeeperTestSuite) createPair(baseCoinDenom, quoteCoinDenom string) liquiditytypes.Pair {
	s.T().Helper()
	s.fundAddr(helperAddr, s.app.LiquidityKeeper.GetPairCreationFee(s.ctx))
	pair, err := s.app.LiquidityKeeper.CreatePair(
		s.ctx, liquiditytypes.NewMsgCreatePair(helperAddr, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(pairId uint64, depositCoins sdk.Coins) liquiditytypes.Pool {
	s.T().Helper()
	s.fundAddr(helperAddr, s.app.LiquidityKeeper.GetPoolCreationFee(s.ctx).Add(depositCoins...))
	pool, err := s.app.LiquidityKeeper.CreatePool(
		s.ctx, liquiditytypes.NewMsgCreatePool(helperAddr, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

//nolint
func (s *KeeperTestSuite) createRangedPool(
	pairId uint64, depositCoins sdk.Coins, minPrice, maxPrice, initialPrice sdk.Dec,
) liquiditytypes.Pool {
	s.T().Helper()
	s.fundAddr(helperAddr, s.app.LiquidityKeeper.GetPoolCreationFee(s.ctx).Add(depositCoins...))
	pool, err := s.app.LiquidityKeeper.CreateRangedPool(
		s.ctx, liquiditytypes.NewMsgCreateRangedPool(
			helperAddr, pairId, depositCoins, minPrice, maxPrice, initialPrice))
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) deposit(depositorAddr sdk.AccAddress, poolId uint64, depositCoins sdk.Coins) {
	s.T().Helper()
	s.fundAddr(depositorAddr, depositCoins)
	_, err := s.app.LiquidityKeeper.Deposit(
		s.ctx, liquiditytypes.NewMsgDeposit(depositorAddr, poolId, depositCoins))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPrivatePlan(rewardAllocs []types.RewardAllocation) types.Plan {
	s.T().Helper()
	s.fundAddr(helperAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	plan, err := s.keeper.CreatePrivatePlan(
		s.ctx, helperAddr, "", rewardAllocs, sampleStartTime, sampleEndTime)
	s.Require().NoError(err)
	return plan
}

func (s *KeeperTestSuite) rewards(farmerAddr sdk.AccAddress, denom string) sdk.DecCoins {
	cacheCtx, _ := s.ctx.CacheContext()
	_, found := s.keeper.GetFarm(s.ctx, denom)
	if !found {
		return nil
	}
	position, found := s.keeper.GetPosition(s.ctx, farmerAddr, denom)
	if !found {
		return nil
	}
	endPeriod := s.keeper.IncrementFarmPeriod(cacheCtx, denom)
	return s.keeper.Rewards(cacheCtx, position, endPeriod)
}
