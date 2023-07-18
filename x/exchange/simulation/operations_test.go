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
	s.Require().Equal("cosmos1r6vgn9cwpvja7448fg0fgglj63rcs6y84p8egu", msg.Sender)
	s.Require().EqualValues(1, msg.MarketId)
	s.Require().Equal(false, msg.IsBuy)
	s.AssertEqual(utils.ParseDec("157.11"), msg.Price)
	s.AssertEqual(utils.ParseDec("75503769"), msg.Quantity)
	s.Require().Equal(2*time.Hour, msg.Lifespan)
}

func (s *SimTestSuite) TestSimulateMsgPlaceMarketOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	s.CreateMarket("denom1", "denom2")

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
	s.AssertEqual(utils.ParseDec("118906"), msg.Quantity)
}
