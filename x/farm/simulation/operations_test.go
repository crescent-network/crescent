package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v3/app"
	utils "github.com/crescent-network/crescent/v3/types"
	"github.com/crescent-network/crescent/v3/x/farm/keeper"
	"github.com/crescent-network/crescent/v3/x/farm/simulation"
	"github.com/crescent-network/crescent/v3/x/farm/types"
	liquiditytypes "github.com/crescent-network/crescent/v3/x/liquidity/types"
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
	s.keeper = s.app.FarmKeeper
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
		s.app.AccountKeeper, s.app.BankKeeper, s.app.LiquidityKeeper, s.app.FarmKeeper)
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
		types.NewRewardAllocation(1, utils.ParseCoins("169567170stake")),
	}, msg.RewardAllocations)
	s.Require().Equal(utils.ParseTime("2022-01-02T00:00:00Z"), msg.StartTime)
	s.Require().Equal(utils.ParseTime("2022-01-05T00:00:00Z"), msg.EndTime)
}
