package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v4/app"
	utils "github.com/crescent-network/crescent/v4/types"
	liquiditytypes "github.com/crescent-network/crescent/v4/x/liquidity/types"
	"github.com/crescent-network/crescent/v4/x/lpfarm/keeper"
	"github.com/crescent-network/crescent/v4/x/lpfarm/simulation"
	"github.com/crescent-network/crescent/v4/x/lpfarm/types"
)

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

type SimTestSuite struct {
	suite.Suite

	app    *chain.App
	ctx    sdk.Context
	keeper keeper.Keeper
}

func (s *SimTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	hdr := tmproto.Header{
		Height: 1,
		Time:   utils.ParseTime("2022-01-01T00:00:00Z"),
	}
	s.app.BeginBlock(abci.RequestBeginBlock{Header: hdr})
	s.ctx = s.app.BaseApp.NewContext(false, hdr)
	s.keeper = s.app.LPFarmKeeper
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accs := simtypes.RandomAccounts(r, n)
	initAmt := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, acc := range accs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, acc.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(chain.FundAccount(s.app.BankKeeper, s.ctx, acc.GetAddress(), initCoins))
	}

	return accs
}

func (s *SimTestSuite) TestSimulateMsgCreatePrivatePlan() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	s.app.LiquidityKeeper.SetPair(s.ctx, liquiditytypes.NewPair(1, "denom1", "denom2"))

	op := simulation.SimulateMsgCreatePrivatePlan(
		s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper, s.app.LPFarmKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreatePrivatePlan
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreatePrivatePlan, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Creator)
	s.Require().Equal("Farming Plan", msg.Description)
	s.Require().Equal([]types.RewardAllocation{
		types.NewPairRewardAllocation(1, utils.ParseCoins("169567170stake")),
	}, msg.RewardAllocations)
	s.Require().Equal(utils.ParseTime("2022-01-02T00:00:00Z"), msg.StartTime)
	s.Require().Equal(utils.ParseTime("2022-01-06T00:00:00Z"), msg.EndTime)
}

func (s *SimTestSuite) TestSimulateMsgFarm() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, accs[0].Address, utils.ParseCoins("1000_000000pool1")))
	op := simulation.SimulateMsgFarm(s.app.AccountKeeper, s.app.BankKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgFarm
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgFarm, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Farmer)
	s.Require().Equal("505033207pool1", msg.Coin.String())
}

func (s *SimTestSuite) TestSimulateMsgUnfarm() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, accs[0].Address, utils.ParseCoins("100_000000pool1")))
	_, _ = s.keeper.Farm(s.ctx, accs[0].Address, utils.ParseCoin("100_000000pool1"))

	op := simulation.SimulateMsgUnfarm(
		s.app.AccountKeeper, s.app.BankKeeper, s.app.LPFarmKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgUnfarm
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgUnfarm, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Farmer)
	s.Require().Equal("21211168pool1", msg.Coin.String())
}

func (s *SimTestSuite) TestSimulateMsgHarvest() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 1)

	s.Require().NoError(
		chain.FundAccount(
			s.app.BankKeeper, s.ctx, accs[0].Address, utils.ParseCoins("100_000000pool1")))
	_, _ = s.keeper.Farm(s.ctx, accs[0].Address, utils.ParseCoin("100_000000pool1"))

	op := simulation.SimulateMsgHarvest(
		s.app.AccountKeeper, s.app.BankKeeper, s.app.LPFarmKeeper)
	opMsg, futureOps, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgHarvest
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgHarvest, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Farmer)
	s.Require().Equal("pool1", msg.Denom)
}
