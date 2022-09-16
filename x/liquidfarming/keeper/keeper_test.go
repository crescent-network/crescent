package keeper_test

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming"
	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
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
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidFarmingKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

//
// Below are just shortcuts to frequently-used functions.
//

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	err := s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, addr, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createPrivateFixedAmountPlan(
	creatorAcc sdk.AccAddress,
	stakingCoinWeightsMap map[string]string,
	epochAmountMap map[string]int64,
	fund bool,
) {
	s.T().Helper()
	stakingCoinWeights := sdk.NewDecCoins()
	for denom, weight := range stakingCoinWeightsMap {
		stakingCoinWeights = stakingCoinWeights.Add(sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(weight)))
	}

	epochAmount := sdk.NewCoins()
	for denom, amount := range epochAmountMap {
		epochAmount = epochAmount.Add(sdk.NewInt64Coin(denom, amount))
	}

	if fund {
		fees := s.app.FarmingKeeper.GetParams(s.ctx).PrivatePlanCreationFee
		s.fundAddr(creatorAcc, epochAmount.Add(fees...))
	}

	msg := farmingtypes.NewMsgCreateFixedAmountPlan(
		fmt.Sprintf("plan%d", s.app.FarmingKeeper.GetGlobalPlanId(s.ctx)+1),
		creatorAcc,
		stakingCoinWeights,
		utils.ParseTime("0001-01-01T00:00:00Z"),
		utils.ParseTime("9999-12-31T00:00:00Z"),
		epochAmount,
	)
	_, err := s.app.FarmingKeeper.CreateFixedAmountPlan(s.ctx, msg, creatorAcc, creatorAcc, farmingtypes.PlanTypePrivate)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) advanceEpochDays() {
	currentEpochDays := s.app.FarmingKeeper.GetCurrentEpochDays(s.ctx)
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Duration(currentEpochDays) * farmingtypes.Day))
	farming.EndBlocker(s.ctx, s.app.FarmingKeeper)
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liquiditytypes.Pair {
	s.T().Helper()
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.Pool {
	s.T().Helper()
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
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
	s.keeper.CreateRewardsAuction(s.ctx, poolId)
}

func (s *KeeperTestSuite) farm(poolId uint64, farmer sdk.AccAddress, farmingCoin sdk.Coin, fund bool) {
	s.T().Helper()
	if fund {
		s.fundAddr(farmer, sdk.NewCoins(farmingCoin))
	}

	err := s.keeper.Farm(s.ctx, poolId, farmer, farmingCoin)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) unfarm(poolId uint64, farmer sdk.AccAddress, lfCoin sdk.Coin, fund bool) {
	s.T().Helper()
	if fund {
		s.fundAddr(farmer, sdk.NewCoins(lfCoin))
	}

	_, err := s.keeper.Unfarm(s.ctx, poolId, farmer, lfCoin)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) placeBid(poolId uint64, bidder sdk.AccAddress, biddingCoin sdk.Coin, fund bool) types.Bid {
	s.T().Helper()
	if fund {
		s.fundAddr(bidder, sdk.NewCoins(biddingCoin))
	}

	bid, err := s.keeper.PlaceBid(s.ctx, poolId, bidder, biddingCoin)
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

func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins {
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
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
