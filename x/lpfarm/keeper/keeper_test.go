package keeper_test

import (
	"testing"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v4/app"
	utils "github.com/crescent-network/crescent/v4/types"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
	"github.com/crescent-network/crescent/v4/x/lpfarm"
	"github.com/crescent-network/crescent/v4/x/lpfarm/keeper"
	"github.com/crescent-network/crescent/v4/x/lpfarm/types"
	minttypes "github.com/crescent-network/crescent/v4/x/mint/types"
)

var (
	helperAddr      = utils.TestAddress(10000)
	sampleStartTime = utils.ParseTime("0001-01-01T00:00:00Z")
	sampleEndTime   = utils.ParseTime("9999-12-31T23:59:59Z")
)

type KeeperTestSuite struct {
	suite.Suite

	app        *chain.App
	ctx        sdk.Context
	keeper     keeper.Keeper
	querier    keeper.Querier
	govHandler govtypes.Handler
	hdr        tmproto.Header
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.keeper = s.app.LPFarmKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.govHandler = lpfarm.NewFarmingPlanProposalHandler(s.keeper)
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

func (s *KeeperTestSuite) handleProposal(content govtypes.Content) {
	s.T().Helper()
	s.Require().NoError(content.ValidateBasic())
	s.Require().NoError(s.govHandler(s.ctx, content))
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	s.T().Helper()
	s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, amt))
	s.Require().NoError(
		s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, addr, amt))
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

func (s *KeeperTestSuite) getBalances(addr sdk.AccAddress) sdk.Coins {
	s.T().Helper()
	return s.app.BankKeeper.GetAllBalances(s.ctx, addr)
}

func (s *KeeperTestSuite) createPair(baseCoinDenom, quoteCoinDenom string) liquiditytypes.Pair {
	s.T().Helper()
	s.fundAddr(helperAddr, s.app.LiquidityKeeper.GetPairCreationFee(s.ctx))
	pair, err := s.app.LiquidityKeeper.CreatePair(
		s.ctx, liquiditytypes.NewMsgCreatePair(helperAddr, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPairWithLastPrice(baseCoinDenom, quoteCoinDenom string, lastPrice sdk.Dec) liquiditytypes.Pair {
	s.T().Helper()
	pair := s.createPair(baseCoinDenom, quoteCoinDenom)
	pair.LastPrice = &lastPrice
	s.app.LiquidityKeeper.SetPair(s.ctx, pair)
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

func (s *KeeperTestSuite) createPrivatePlan(rewardAllocs []types.RewardAllocation, initialFunds sdk.Coins) types.Plan {
	s.T().Helper()
	s.fundAddr(helperAddr, s.keeper.GetPrivatePlanCreationFee(s.ctx))
	plan, err := s.keeper.CreatePrivatePlan(
		s.ctx, helperAddr, "", rewardAllocs, sampleStartTime, sampleEndTime)
	s.Require().NoError(err)
	s.fundAddr(plan.GetFarmingPoolAddress(), initialFunds)
	return plan
}

func (s *KeeperTestSuite) createPublicPlan(farmingPoolAddr sdk.AccAddress, rewardAllocs []types.RewardAllocation) types.Plan {
	s.T().Helper()
	plan, err := s.keeper.CreatePublicPlan(
		s.ctx, "", farmingPoolAddr, rewardAllocs, sampleStartTime, sampleEndTime)
	s.Require().NoError(err)
	return plan
}

func (s *KeeperTestSuite) farm(farmerAddr sdk.AccAddress, coin sdk.Coin) sdk.Coins {
	s.T().Helper()
	s.fundAddr(farmerAddr, sdk.NewCoins(coin))
	withdrawnRewards, err := s.keeper.Farm(s.ctx, farmerAddr, coin)
	s.Require().NoError(err)
	return withdrawnRewards
}

func (s *KeeperTestSuite) unfarm(farmerAddr sdk.AccAddress, coin sdk.Coin) sdk.Coins {
	s.T().Helper()
	withdrawnRewards, err := s.keeper.Unfarm(s.ctx, farmerAddr, coin)
	s.Require().NoError(err)
	return withdrawnRewards
}

func (s *KeeperTestSuite) harvest(farmerAddr sdk.AccAddress, denom string) sdk.Coins {
	s.T().Helper()
	withdrawnRewards, err := s.keeper.Harvest(s.ctx, farmerAddr, denom)
	s.Require().NoError(err)
	return withdrawnRewards
}

func (s *KeeperTestSuite) rewards(farmerAddr sdk.AccAddress, denom string) sdk.DecCoins {
	_, found := s.keeper.GetFarm(s.ctx, denom)
	if !found {
		return nil
	}
	return s.keeper.Rewards(s.ctx, farmerAddr, denom)
}

func (s *KeeperTestSuite) assertHistoricalRewards(exp map[string]map[uint64]types.HistoricalRewards) {
	s.T().Helper()
	got := map[string]map[uint64]types.HistoricalRewards{}
	s.keeper.IterateAllHistoricalRewards(s.ctx, func(denom string, period uint64, hist types.HistoricalRewards) (stop bool) {
		histsByPeriod, ok := got[denom]
		if !ok {
			histsByPeriod = map[uint64]types.HistoricalRewards{}
			got[denom] = histsByPeriod
		}
		histsByPeriod[period] = hist
		return false
	})
	s.Require().Len(got, len(exp))
	for denom := range exp {
		s.Require().Len(got[denom], len(exp[denom]))
		for period := range exp[denom] {
			_, ok := got[denom][period]
			s.Require().True(ok)
			exp, got := exp[denom][period], got[denom][period]
			s.assertEq(exp.CumulativeUnitRewards, got.CumulativeUnitRewards)
			s.Require().EqualValues(exp.ReferenceCount, got.ReferenceCount)
		}
	}
}

func (s *KeeperTestSuite) createSamplePlans() (privPlan, pubPlan types.Plan) {
	s.T().Helper()
	pair1 := s.createPairWithLastPrice("denom1", "denom2", sdk.NewDec(1))
	pair2 := s.createPairWithLastPrice("denom2", "denom3", sdk.NewDec(1))
	s.createPool(pair1.Id, utils.ParseCoins("100_000000denom1,100_000000denom2"))
	s.createPool(pair2.Id, utils.ParseCoins("100_000000denom2,100_000000denom3"))

	privPlan = s.createPrivatePlan([]types.RewardAllocation{
		types.NewPairRewardAllocation(pair1.Id, utils.ParseCoins("100_000000stake")),
		types.NewPairRewardAllocation(pair2.Id, utils.ParseCoins("200_000000stake")),
	}, utils.ParseCoins("10000_000000stake"))
	farmingPoolAddr := utils.TestAddress(100)
	proposal := types.NewFarmingPlanProposal(
		"Title", "Description",
		[]types.CreatePlanRequest{
			types.NewCreatePlanRequest(
				"Farming Plan", farmingPoolAddr,
				[]types.RewardAllocation{
					types.NewPairRewardAllocation(pair1.Id, utils.ParseCoins("300_000000stake")),
					types.NewPairRewardAllocation(pair2.Id, utils.ParseCoins("400_000000stake")),
				}, sampleStartTime, sampleEndTime),
		}, nil)
	s.handleProposal(proposal)
	pubPlan, _ = s.keeper.GetPlan(s.ctx, 2)
	return
}
