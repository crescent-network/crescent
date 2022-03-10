package keeper_test

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/claim/keeper"
	"github.com/cosmosquad-labs/squad/x/claim/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
	liqtypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
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
	s.keeper = s.app.ClaimKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
}

//
// Below are just shortcuts to internal functions.
//

func (s *KeeperTestSuite) createAirdrop(
	id uint64,
	sourceAddr sdk.AccAddress,
	sourceCoins sdk.Coins,
	conditions []types.ConditionType,
	startTime time.Time,
	endTime time.Time,
	fund bool,
) types.Airdrop {
	if fund {
		s.fundAddr(sourceAddr, sourceCoins)
	}

	s.keeper.SetAirdrop(s.ctx, types.Airdrop{
		Id:            id,
		SourceAddress: sourceAddr.String(),
		Conditions:    conditions,
		StartTime:     startTime,
		EndTime:       endTime,
	})

	airdrop, found := s.keeper.GetAirdrop(s.ctx, id)
	s.Require().True(found)

	return airdrop
}

func (s *KeeperTestSuite) createClaimRecord(
	airdropId uint64,
	recipient sdk.AccAddress,
	initialClaimableCoins sdk.Coins,
	claimableCoins sdk.Coins,
	claimedConditions []types.ConditionType,
) types.ClaimRecord {
	s.keeper.SetClaimRecord(s.ctx, types.ClaimRecord{
		AirdropId:             airdropId,
		Recipient:             recipient.String(),
		InitialClaimableCoins: initialClaimableCoins,
		ClaimableCoins:        claimableCoins,
		ClaimedConditions:     claimedConditions,
	})

	r, found := s.keeper.GetClaimRecordByRecipient(s.ctx, airdropId, recipient)
	s.Require().True(found)

	return r
}

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liqtypes.Pair {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liqtypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) liqtypes.Pool {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
	pool, err := s.app.LiquidityKeeper.CreatePool(s.ctx, liqtypes.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

func (s *KeeperTestSuite) deposit(depositor sdk.AccAddress, poolId uint64, depositCoins sdk.Coins, fund bool) liqtypes.DepositRequest {
	if fund {
		s.fundAddr(depositor, depositCoins)
	}
	req, err := s.app.LiquidityKeeper.Deposit(s.ctx, liqtypes.NewMsgDeposit(depositor, poolId, depositCoins))
	s.Require().NoError(err)
	return req
}

func (s *KeeperTestSuite) limitOrder(
	orderer sdk.AccAddress, pairId uint64, dir liqtypes.OrderDirection,
	price sdk.Dec, amt sdk.Int, orderLifespan time.Duration, fund bool) liqtypes.Order {
	pair, found := s.app.LiquidityKeeper.GetPair(s.ctx, pairId)
	s.Require().True(found)

	var offerCoin sdk.Coin
	var demandCoinDenom string
	switch dir {
	case liqtypes.OrderDirectionBuy:
		offerCoin = sdk.NewCoin(pair.QuoteCoinDenom, price.MulInt(amt).Ceil().TruncateInt())
		demandCoinDenom = pair.BaseCoinDenom
	case liqtypes.OrderDirectionSell:
		offerCoin = sdk.NewCoin(pair.BaseCoinDenom, amt)
		demandCoinDenom = pair.QuoteCoinDenom
	}

	if fund {
		s.fundAddr(orderer, sdk.NewCoins(offerCoin))
	}

	req, err := s.app.LiquidityKeeper.LimitOrder(s.ctx, liqtypes.NewMsgLimitOrder(
		orderer, pairId, dir, offerCoin, demandCoinDenom,
		price, amt, orderLifespan),
	)
	s.Require().NoError(err)

	return req
}

func (s *KeeperTestSuite) sellLimitOrder(
	orderer sdk.AccAddress, pairId uint64, price sdk.Dec,
	amt sdk.Int, orderLifespan time.Duration, fund bool) liqtypes.Order {
	return s.limitOrder(
		orderer, pairId, liqtypes.OrderDirectionSell, price, amt, orderLifespan, fund)
}

func (s *KeeperTestSuite) createFixedAmountPlan(
	farmingPoolAcc sdk.AccAddress,
	stakingCoinWeightsMap map[string]string,
	epochAmountMap map[string]int64,
	fund bool,
) {
	stakingCoinWeights := sdk.NewDecCoins()
	for denom, weight := range stakingCoinWeightsMap {
		stakingCoinWeights = stakingCoinWeights.Add(sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(weight)))
	}

	epochAmount := sdk.NewCoins()
	for denom, amount := range epochAmountMap {
		epochAmount = epochAmount.Add(sdk.NewInt64Coin(denom, amount))
	}

	if fund {
		s.fundAddr(farmingPoolAcc, epochAmount)
	}

	msg := farmingtypes.NewMsgCreateFixedAmountPlan(
		fmt.Sprintf("plan%d", s.app.FarmingKeeper.GetGlobalPlanId(s.ctx)+1),
		farmingPoolAcc,
		stakingCoinWeights,
		s.ctx.BlockTime(),
		s.ctx.BlockTime().AddDate(0, 6, 0),
		epochAmount,
	)
	_, err := s.app.FarmingKeeper.CreateFixedAmountPlan(s.ctx, msg, farmingPoolAcc, farmingPoolAcc, farmingtypes.PlanTypePublic)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) stake(farmerAcc sdk.AccAddress, amt sdk.Coins, fund bool) {
	if fund {
		s.fundAddr(farmerAcc, amt)
	}

	err := s.app.FarmingKeeper.Stake(s.ctx, farmerAcc, amt)
	s.Require().NoError(err)
}

//
// Below are useful helpers to write test code easily.
//

func (s *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	return s.app.BankKeeper.GetBalance(s.ctx, addr, denom)
}

func (s *KeeperTestSuite) getAllBalances(addr sdk.AccAddress) sdk.Coins {
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) addr(addrNum int) sdk.AccAddress {
	addr := make(sdk.AccAddress, 20)
	binary.PutVarint(addr, int64(addrNum))
	return addr
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, coins sdk.Coins) {
	err := chain.FundAccount(s.app.BankKeeper, s.ctx, addr, coins)
	s.Require().NoError(err)
}

func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
