package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	chain "github.com/crescent-network/crescent/v5/app"
	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/simulation"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
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
	s.keeper = s.App.ExchangeKeeper
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

func (s *SimTestSuite) TestSimulateMsgCreateMarket() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	op := simulation.SimulateMsgCreateMarket(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreateMarket
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreateMarket, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1u45vv4pl5zchwgvd3pt9qamrflf7ykv52kn5u3", msg.Sender)
	s.Require().Equal("stake", msg.BaseDenom)
	s.Require().Equal("denom2", msg.QuoteDenom)
}

func (s *SimTestSuite) TestSimulateMsgPlaceLimitOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	s.CreateMarket("denom1", "denom2")

	op := simulation.SimulateMsgPlaceLimitOrder(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgPlaceLimitOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgPlaceLimitOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1se3z6fgrp0hy7feh2dsntqjwnfy8myg3l6h99r", msg.Sender)
	s.Require().EqualValues(1, msg.MarketId)
	s.Require().Equal(false, msg.IsBuy)
	s.AssertEqual(utils.ParseDec("7.6119"), msg.Price)
	s.AssertEqual(sdk.NewInt(79000424), msg.Quantity)
	s.Require().Equal(2*time.Hour, msg.Lifespan)
}

func (s *SimTestSuite) TestSimulateMsgPlaceMMLimitOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	s.CreateMarket("denom1", "denom2")

	op := simulation.SimulateMsgPlaceMMLimitOrder(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgPlaceMMLimitOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgPlaceMMLimitOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1se3z6fgrp0hy7feh2dsntqjwnfy8myg3l6h99r", msg.Sender)
	s.Require().EqualValues(1, msg.MarketId)
	s.Require().Equal(false, msg.IsBuy)
	s.AssertEqual(utils.ParseDec("7.6119"), msg.Price)
	s.AssertEqual(sdk.NewInt(79000424), msg.Quantity)
	s.Require().Equal(2*time.Hour, msg.Lifespan)
}

func (s *SimTestSuite) TestSimulateMsgPlaceMarketOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	market := s.CreateMarket("denom1", "denom2")
	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	marketState.LastPrice = utils.ParseDecP("1.2")
	s.keeper.SetMarketState(s.Ctx, market.Id, marketState)

	op := simulation.SimulateMsgPlaceMarketOrder(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgPlaceMarketOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgPlaceMarketOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1r6vgn9cwpvja7448fg0fgglj63rcs6y84p8egu", msg.Sender)
	s.Require().EqualValues(1, msg.MarketId)
	s.Require().Equal(false, msg.IsBuy)
	s.AssertEqual(sdk.NewInt(113892), msg.Quantity)
}

func (s *SimTestSuite) TestSimulateMsgCancelOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 2)

	market := s.CreateMarket("denom1", "denom2")
	s.PlaceLimitOrder(
		market.Id, accs[0].Address, true, utils.ParseDec("1.2"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, accs[0].Address, true, utils.ParseDec("1.21"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, accs[0].Address, true, utils.ParseDec("1.22"), sdk.NewInt(10_000000), time.Hour)
	s.NextBlock()

	op := simulation.SimulateMsgCancelOrder(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCancelOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCancelOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Sender)
	s.Require().EqualValues(2, msg.OrderId)
}

func (s *SimTestSuite) TestSimulateMsgCancelAllOrders() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 2)

	// The order state doesn't matter but...
	market1 := s.CreateMarket("denom1", "denom2")
	s.PlaceLimitOrder(
		market1.Id, accs[0].Address, true, utils.ParseDec("1.2"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(
		market1.Id, accs[0].Address, true, utils.ParseDec("1.21"), sdk.NewInt(10_000000), time.Hour)
	market2 := s.CreateMarket("denom2", "denom3")
	s.PlaceLimitOrder(
		market2.Id, accs[0].Address, false, utils.ParseDec("5"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(
		market2.Id, accs[0].Address, false, utils.ParseDec("4.99"), sdk.NewInt(10_000000), time.Hour)
	s.NextBlock()
	s.PlaceLimitOrder(
		market1.Id, accs[0].Address, true, utils.ParseDec("1.22"), sdk.NewInt(10_000000), time.Hour)
	s.PlaceLimitOrder(
		market2.Id, accs[0].Address, false, utils.ParseDec("4.98"), sdk.NewInt(10_000000), time.Hour)

	op := simulation.SimulateMsgCancelAllOrders(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCancelAllOrders
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCancelAllOrders, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Sender)
	s.Require().EqualValues(2, msg.MarketId)
}

func (s *SimTestSuite) TestSimulateMsgSwapExactAmountIn() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 2)

	mmAddr := accs[0].Address
	market1 := s.CreateMarket("denom1", "denom2")
	s.MakeLastPrice(market1.Id, mmAddr, utils.ParseDec("1"))
	pool1 := s.CreatePool(market1.Id, utils.ParseDec("1"))
	s.AddLiquidity(
		accs[1].Address, pool1.Id, utils.ParseDec("0.5"), utils.ParseDec("2"),
		utils.ParseCoins("10_000000denom1,10_0000000denom2"))
	market2 := s.CreateMarket("denom2", "denom3")
	s.MakeLastPrice(market2.Id, mmAddr, utils.ParseDec("5"))
	pool2 := s.CreatePool(market2.Id, utils.ParseDec("5"))
	s.AddLiquidity(
		accs[1].Address, pool2.Id, utils.ParseDec("1"), utils.ParseDec("20"),
		utils.ParseCoins("10_000000denom2,50_000000denom3"))
	market3 := s.CreateMarket("denom3", "denom1")
	s.MakeLastPrice(market3.Id, mmAddr, utils.ParseDec("0.05"))
	pool3 := s.CreatePool(market3.Id, utils.ParseDec("0.05"))
	s.AddLiquidity(
		accs[1].Address, pool3.Id, utils.ParseDec("0.0001"), utils.ParseDec("1000"),
		utils.ParseCoins("10_000000denom3,50_000000denom1"))

	op := simulation.SimulateMsgSwapExactAmountIn(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgSwapExactAmountIn
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgSwapExactAmountIn, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos12jszjrc0qhjt0ugt2uh4ptwu0h55pq6qfp9ecl", msg.Sender)
	s.Require().Equal([]uint64{2, 1, 3}, msg.Routes)
	s.AssertEqual(utils.ParseCoin("82823denom3"), msg.Input)
	s.AssertEqual(utils.ParseCoin("315661denom3"), msg.MinOutput) // Arbitrage!
}
