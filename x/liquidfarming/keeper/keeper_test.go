package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v3/app"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/liquidfarming"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v3/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v3/x/lpfarm/types"
)

var (
	helperAddr      = utils.TestAddress(10000)
	sampleStartTime = utils.ParseTime("0001-01-01T00:00:00Z")
	sampleEndTime   = utils.ParseTime("9999-12-31T23:59:59Z")
)

type KeeperTestSuite struct {
	suite.Suite

	app       *chain.App
	ctx       sdk.Context
	keeper    keeper.Keeper
	querier   keeper.Querier
	msgServer types.MsgServer
	hdr       tmproto.Header
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidFarmingKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.ctx = s.ctx.WithBlockTime(time.Now()) // set to current time
	s.hdr = tmproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
}

//
// Below are just shortcuts to frequently-used functions.
//

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
	err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPrivatePlan(creator sdk.AccAddress, rewardAllocs []lpfarmtypes.RewardAllocation) lpfarmtypes.Plan {
	s.T().Helper()
	s.fundAddr(creator, s.app.LPFarmKeeper.GetPrivatePlanCreationFee(s.ctx))
	plan, err := s.app.LPFarmKeeper.CreatePrivatePlan(s.ctx, creator, "", rewardAllocs, sampleStartTime, sampleEndTime)
	s.Require().NoError(err)
	return plan
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string) liquiditytypes.Pair {
	s.T().Helper()
	s.fundAddr(creator, s.app.LiquidityKeeper.GetPairCreationFee(s.ctx))
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

// createPairWithLastPrice is a convenient method to create a pair with last price.
// it is needed as farming plan doesn't distribute farming rewards if the last price is not set.
func (s *KeeperTestSuite) createPairWithLastPrice(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, lastPrice sdk.Dec) liquiditytypes.Pair {
	s.T().Helper()
	pair := s.createPair(creator, baseCoinDenom, quoteCoinDenom)
	pair.LastPrice = &lastPrice
	s.app.LiquidityKeeper.SetPair(s.ctx, pair)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins) liquiditytypes.Pool {
	s.T().Helper()
	s.fundAddr(creator, s.app.LiquidityKeeper.GetPoolCreationFee(s.ctx).Add(depositCoins...))
	pool, err := s.app.LiquidityKeeper.CreatePool(s.ctx, liquiditytypes.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) deposit(depositor sdk.AccAddress, poolId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.DepositRequest {
	s.T().Helper()
	if fund {
		s.fundAddr(depositor, depositCoins)
	}
	req, err := s.app.LiquidityKeeper.Deposit(s.ctx, liquiditytypes.NewMsgDeposit(depositor, poolId, depositCoins))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) createLiquidFarm(poolId uint64, minFarmAmt, minBidAmt sdk.Int, feeRate sdk.Dec) types.LiquidFarm {
	s.T().Helper()
	liquidFarm := types.NewLiquidFarm(poolId, minFarmAmt, minBidAmt, feeRate)
	params := s.keeper.GetParams(s.ctx)
	params.LiquidFarms = append(params.LiquidFarms, liquidFarm)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.SetLiquidFarm(s.ctx, liquidFarm)
	return liquidFarm
}

func (s *KeeperTestSuite) createRewardsAuction(poolId uint64) {
	s.T().Helper()
	duration := s.keeper.GetRewardsAuctionDuration(s.ctx)
	s.keeper.CreateRewardsAuction(s.ctx, poolId, s.ctx.BlockTime().Add(duration*time.Hour))
}

func (s *KeeperTestSuite) liquidFarm(poolId uint64, farmer sdk.AccAddress, lpCoin sdk.Coin, fund bool) {
	s.T().Helper()
	if fund {
		s.fundAddr(farmer, sdk.NewCoins(lpCoin))
	}
	err := s.keeper.LiquidFarm(s.ctx, poolId, farmer, lpCoin)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) liquidUnfarm(poolId uint64, farmer sdk.AccAddress, lfCoin sdk.Coin, fund bool) {
	s.T().Helper()
	if fund {
		s.fundAddr(farmer, sdk.NewCoins(lfCoin))
	}
	_, err := s.keeper.LiquidUnfarm(s.ctx, poolId, farmer, lfCoin)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) placeBid(poolId uint64, bidder sdk.AccAddress, biddingCoin sdk.Coin, fund bool) types.Bid {
	s.T().Helper()
	if fund {
		s.fundAddr(bidder, sdk.NewCoins(biddingCoin))
	}

	auctionId := s.keeper.GetLastRewardsAuctionId(s.ctx, poolId)
	bid, err := s.keeper.PlaceBid(s.ctx, auctionId, poolId, bidder, biddingCoin)
	s.Require().NoError(err)

	return bid
}

//
// Below are helper functions to write test code easily
//

func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins { //nolint
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

// func (s *KeeperTestSuite) nextBlock() {
// 	s.T().Helper()
// 	s.app.EndBlock(abci.RequestEndBlock{})
// 	s.app.Commit()
// 	hdr := tmproto.Header{
// 		Height: s.app.LastBlockHeight() + 1,
// 		Time:   s.ctx.BlockTime().Add(5 * time.Second),
// 	}
// 	s.app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
// 	s.ctx = s.app.BaseApp.NewContext(false, hdr)
// 	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{Header: hdr})
// }

func (s *KeeperTestSuite) nextAuction() {
	s.T().Helper()
	endTime, found := s.keeper.GetLastRewardsAuctionEndTime(s.ctx)
	if !found {
		duration := s.keeper.GetRewardsAuctionDuration(s.ctx)
		endTime = s.ctx.BlockTime().Add(duration)
	}
	s.ctx = s.ctx.WithBlockTime(endTime)
	liquidfarming.BeginBlocker(s.ctx, s.keeper)
}
