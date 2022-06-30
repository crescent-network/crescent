package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/amm"
	"github.com/crescent-network/crescent/v2/x/liquidity/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidity/simulation"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

type SimTestSuite struct {
	suite.Suite

	app    *chain.App
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (s *SimTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.keeper = s.app.LiquidityKeeper
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) TestSimulateMsgCreatePair() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCreatePair(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreatePair
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreatePair, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Creator)
	s.Require().Equal("denom3", msg.BaseCoinDenom)
	s.Require().Equal("stake", msg.QuoteCoinDenom)
}

func (s *SimTestSuite) TestSimulateMsgCreatePool() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCreatePool(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreatePool
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreatePool, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Creator)
	s.Require().Equal(pair.Id, msg.PairId)
	s.Require().Equal("170567169denom1,131275595stake", msg.DepositCoins.String())
}

func (s *SimTestSuite) TestSimulateMsgCreateRangedPool() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCreateRangedPool(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreateRangedPool
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreateRangedPool, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Creator)
	s.Require().Equal(pair.Id, msg.PairId)
	s.Require().Equal("130275595denom1,169567169stake", msg.DepositCoins.String())
	s.Require().Equal("0.030928000000000000", msg.MinPrice.String())
	s.Require().Equal("92.378000000000000000", msg.MaxPrice.String())
	s.Require().Equal("0.040475000000000000", msg.InitialPrice.String())
}

func (s *SimTestSuite) TestSimulateMsgDeposit() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")
	pool := s.createPool(accs[0].Address, pair.Id, utils.ParseCoins("1000000denom1,1000000stake"))

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgDeposit(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgDeposit
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgDeposit, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Depositor)
	s.Require().Equal(pool.Id, msg.PoolId)
	s.Require().Equal("169567170denom1,130275596stake", msg.DepositCoins.String())
}

func (s *SimTestSuite) TestSimulateMsgWithdraw() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")
	pool := s.createPool(accs[0].Address, pair.Id, utils.ParseCoins("1000000denom1,1000000stake"))

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgWithdraw(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgWithdraw
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgWithdraw, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Withdrawer)
	s.Require().Equal(pool.Id, msg.PoolId)
	s.Require().Equal("134387295170pool1", msg.PoolCoin.String())
}

func (s *SimTestSuite) TestSimulateMsgLimitOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgLimitOrder(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgLimitOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgLimitOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Orderer)
	s.Require().Equal(pair.Id, msg.PairId)
	s.Require().Equal(types.OrderDirectionSell, msg.Direction)
	s.Require().Equal("6010denom1", msg.OfferCoin.String())
	s.Require().Equal("stake", msg.DemandCoinDenom)
	s.Require().Equal("1.228200000000000000", msg.Price.String())
	s.Require().Equal("6010", msg.Amount.String())
	s.Require().Equal("9h14m25.122290029s", msg.OrderLifespan.String())
}

func (s *SimTestSuite) TestSimulateMsgMarketOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")
	p := utils.ParseDec("1.0")
	pair.LastPrice = &p
	s.keeper.SetPair(s.ctx, pair)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgMarketOrder(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgMarketOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgMarketOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Orderer)
	s.Require().Equal(pair.Id, msg.PairId)
	s.Require().Equal(types.OrderDirectionSell, msg.Direction)
	s.Require().Equal("10383denom1", msg.OfferCoin.String())
	s.Require().Equal("stake", msg.DemandCoinDenom)
	s.Require().Equal("10383", msg.Amount.String())
	s.Require().Equal("15h40m44.578894929s", msg.OrderLifespan.String())
}

func (s *SimTestSuite) TestSimulateMsgCancelOrder() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	pair := s.createPair(accs[0].Address, "denom1", "stake")
	order := s.limitOrder(accs[0].Address, pair.Id, types.OrderDirectionBuy, utils.ParseDec("1.0"), sdk.NewInt(1000000), time.Hour)
	// Increment the pair's current batch id to simulate next block.
	pair.CurrentBatchId++
	s.keeper.SetPair(s.ctx, pair)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCancelOrder(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCancelOrder
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCancelOrder, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Orderer)
	s.Require().Equal(pair.Id, msg.PairId)
	s.Require().Equal(order.Id, msg.OrderId)
}

func (s *SimTestSuite) TestSimulateMsgCancelAllOrders() {
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 1)

	pair1 := s.createPair(accs[0].Address, "denom1", "stake")
	pair2 := s.createPair(accs[0].Address, "stake", "denom1")
	pair3 := s.createPair(accs[0].Address, "denom2", "stake")
	s.limitOrder(accs[0].Address, pair1.Id, types.OrderDirectionBuy, utils.ParseDec("1.0"), sdk.NewInt(1000000), time.Hour)
	s.limitOrder(accs[0].Address, pair2.Id, types.OrderDirectionSell, utils.ParseDec("1.0"), sdk.NewInt(1000000), time.Hour)
	s.limitOrder(accs[0].Address, pair3.Id, types.OrderDirectionSell, utils.ParseDec("1.0"), sdk.NewInt(1000000), time.Hour)
	pair1.CurrentBatchId++
	s.keeper.SetPair(s.ctx, pair1)
	pair2.CurrentBatchId++
	s.keeper.SetPair(s.ctx, pair2)
	pair3.CurrentBatchId++
	s.keeper.SetPair(s.ctx, pair3)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCancelAllOrders(s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCancelAllOrders
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCancelAllOrders, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3", msg.Orderer)
	s.Require().Equal([]uint64{pair2.Id}, msg.PairIds)
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accs := simtypes.RandomAccounts(r, n)

	initAmt := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 200)
	initCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, initAmt),
		sdk.NewCoin("denom1", initAmt),
		sdk.NewCoin("denom2", initAmt),
		sdk.NewCoin("denom3", initAmt))

	// add coins to the accounts
	for _, acc := range accs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, acc.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(chain.FundAccount(s.app.BankKeeper, s.ctx, acc.GetAddress(), initCoins))
	}

	return accs
}

func (s *SimTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string) types.Pair {
	pair, err := s.keeper.CreatePair(s.ctx, types.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *SimTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins) types.Pool {
	pool, err := s.keeper.CreatePool(s.ctx, types.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

func (s *SimTestSuite) limitOrder(
	orderer sdk.AccAddress, pairId uint64, dir types.OrderDirection,
	price sdk.Dec, amt sdk.Int, orderLifespan time.Duration) types.Order {
	pair, found := s.keeper.GetPair(s.ctx, pairId)
	s.Require().True(found)
	var ammDir amm.OrderDirection
	var offerCoinDenom, demandCoinDenom string
	switch dir {
	case types.OrderDirectionBuy:
		ammDir = amm.Buy
		offerCoinDenom, demandCoinDenom = pair.QuoteCoinDenom, pair.BaseCoinDenom
	case types.OrderDirectionSell:
		ammDir = amm.Sell
		offerCoinDenom, demandCoinDenom = pair.BaseCoinDenom, pair.QuoteCoinDenom
	}
	offerCoin := sdk.NewCoin(offerCoinDenom, amm.OfferCoinAmount(ammDir, price, amt))
	msg := types.NewMsgLimitOrder(
		orderer, pairId, dir, offerCoin, demandCoinDenom,
		price, amt, orderLifespan)
	req, err := s.keeper.LimitOrder(s.ctx, msg)
	s.Require().NoError(err)
	return req
}
