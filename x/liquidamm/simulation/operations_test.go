package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	chain "github.com/crescent-network/crescent/v5/app"
	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	ammtypes "github.com/crescent-network/crescent/v5/x/amm/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/simulation"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

type SimTestSuite struct {
	testutil.TestSuite
	keeper keeper.Keeper
}

func (s *SimTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.keeper = s.App.LiquidAMMKeeper
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accs := simtypes.RandomAccounts(r, n)

	initAmt := s.App.StakingKeeper.TokensFromConsensusPower(s.Ctx, 200)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewCoin("denom1", initAmt),
		sdk.NewCoin("denom2", initAmt),
		sdk.NewCoin("denom3", initAmt))

	// add coins to the accounts
	for _, acc := range accs {
		acc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, acc.Address)
		s.App.AccountKeeper.SetAccount(s.Ctx, acc)
		s.Require().NoError(chain.FundAccount(s.App.BankKeeper, s.Ctx, acc.GetAddress(), initCoins))
	}

	return accs
}

func (s *SimTestSuite) TestSimulateMsgMintShare() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	market := s.CreateMarket("denom1", "denom2")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	s.CreatePublicPosition(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"), utils.ParseDec("0.003"))

	op := simulation.SimulateMsgMintShare(s.App.AccountKeeper, s.App.BankKeeper, s.App.AMMKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgMintShare
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Assert().Equal(types.TypeMsgMintShare, msg.Type())
	s.Assert().Equal(types.ModuleName, msg.Route())
	s.Assert().Equal("cosmos1r6vgn9cwpvja7448fg0fgglj63rcs6y84p8egu", msg.Sender)
	s.Assert().Equal(uint64(1), msg.PublicPositionId)
	s.AssertEqual(utils.ParseCoins("116930denom1,644687denom2"), msg.DesiredAmount)
}

func (s *SimTestSuite) TestSimulateMsgBurnShare() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	market := s.CreateMarket("denom1", "denom2")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	publicPosition := s.CreatePublicPosition(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"), utils.ParseDec("0.003"))
	s.MintShare(accs[0].Address, publicPosition.Id, utils.ParseCoins("100_000000denom1,500_000000denom2"), true)

	op := simulation.SimulateMsgBurnShare(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgBurnShare
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Assert().Equal(types.TypeMsgBurnShare, msg.Type())
	s.Assert().Equal(types.ModuleName, msg.Route())
	s.Assert().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Sender)
	s.Assert().Equal(uint64(1), msg.PublicPositionId)
	s.AssertEqual(utils.ParseCoin("672658281sb1"), msg.Share)
}

func (s *SimTestSuite) TestSimulateMsgPlaceBid() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	market := s.CreateMarket("denom1", "denom2")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	publicPosition := s.CreatePublicPosition(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"), utils.ParseDec("0.003"))
	farmingPoolAddr := s.FundedAccount(100, utils.ParseCoins("10000_000000stake"))
	s.CreatePublicFarmingPlan(
		"Farming plan", farmingPoolAddr, farmingPoolAddr, []ammtypes.FarmingRewardAllocation{
			ammtypes.NewFarmingRewardAllocation(pool.Id, utils.ParseCoins("100_000000stake")),
		}, utils.ParseTime("0001-01-01T00:00:00Z"), utils.ParseTime("9999-12-31T00:00:00Z"))
	s.MintShare(accs[0].Address, publicPosition.Id, utils.ParseCoins("100_000000denom1,500_000000denom2"), true)
	s.NextBlock()
	s.AdvanceRewardsAuctions()

	op := simulation.SimulateMsgPlaceBid(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgPlaceBid
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Assert().Equal(types.TypeMsgPlaceBid, msg.Type())
	s.Assert().Equal(types.ModuleName, msg.Route())
	s.Assert().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Sender)
	s.Assert().Equal(uint64(1), msg.PublicPositionId)
	s.AssertEqual(utils.ParseCoin("70093sb1"), msg.Share)
}
