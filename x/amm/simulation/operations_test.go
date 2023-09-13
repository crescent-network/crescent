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
	"github.com/crescent-network/crescent/v5/x/amm/keeper"
	"github.com/crescent-network/crescent/v5/x/amm/simulation"
	"github.com/crescent-network/crescent/v5/x/amm/types"
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
	s.keeper = s.App.AMMKeeper
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

func (s *SimTestSuite) TestSimulateMsgCreatePool() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	// Create all possible markets
	var denoms []string
	s.App.BankKeeper.IterateTotalSupply(s.Ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	for _, denomA := range denoms {
		for _, denomB := range denoms {
			if denomA != denomB {
				s.CreateMarket(denomA, denomB)
			}
		}
	}

	op := simulation.SimulateMsgCreatePool(
		s.App.AccountKeeper, s.App.BankKeeper, s.App.ExchangeKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgCreatePool
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgCreatePool, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1dj7jl7a84qaj7566qaa4uxj7d28vvr304204ha", msg.Sender)
	s.Require().EqualValues(5, msg.MarketId)
	s.AssertEqual(utils.ParseDec("6.585228379443441869"), msg.Price)
}

func (s *SimTestSuite) TestSimulateMsgAddLiquidity() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	// Create all possible markets and pools
	var denoms []string
	s.App.BankKeeper.IterateTotalSupply(s.Ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	for _, denomA := range denoms {
		for _, denomB := range denoms {
			if denomA != denomB {
				market := s.CreateMarket(denomA, denomB)
				price := utils.SimRandomDec(r, utils.ParseDec("0.05"), utils.ParseDec("500"))
				s.CreatePool(market.Id, price)
			}
		}
	}

	op := simulation.SimulateMsgAddLiquidity(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgAddLiquidity
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgAddLiquidity, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1ea66m0hr892xhmtjzq7s76vq8cdcmnqgewvtsy", msg.Sender)
	s.Require().EqualValues(2, msg.PoolId)
	s.AssertEqual(utils.ParseDec("144"), msg.LowerPrice)
	s.AssertEqual(utils.ParseDec("584.5"), msg.UpperPrice)
	s.AssertEqual(utils.ParseCoins("569259denom1,111578257denom3"), msg.DesiredAmount)
}

func (s *SimTestSuite) TestSimulateMsgRemoveLiquidity() {
	r := rand.New(rand.NewSource(0))
	accs := s.getTestingAccounts(r, 50)

	var denoms []string
	s.App.BankKeeper.IterateTotalSupply(s.Ctx, func(coin sdk.Coin) bool {
		denoms = append(denoms, coin.Denom)
		return false
	})
	market := s.CreateMarket(denoms[0], denoms[1])
	pool := s.CreatePool(market.Id, utils.ParseDec("12.345"))
	s.AddLiquidity(
		accs[0].Address, pool.Id,
		utils.ParseDec("10"), utils.ParseDec("15"),
		sdk.NewCoins(sdk.NewInt64Coin(denoms[0], 100_000000), sdk.NewInt64Coin(denoms[1], 100_000000)))

	op := simulation.SimulateMsgRemoveLiquidity(
		s.App.AccountKeeper, s.App.BankKeeper, s.keeper)
	opMsg, futureOps, err := op(r, s.App.BaseApp, s.Ctx, accs, "")
	s.Require().NoError(err)
	s.Require().True(opMsg.OK)
	s.Require().Len(futureOps, 0)

	var msg types.MsgRemoveLiquidity
	types.ModuleCdc.MustUnmarshalJSON(opMsg.Msg, &msg)

	s.Require().Equal(types.TypeMsgRemoveLiquidity, msg.Type())
	s.Require().Equal(types.ModuleName, msg.Route())
	s.Require().Equal("cosmos1tp4es44j4vv8m59za3z0tm64dkmlnm8wg2frhc", msg.Sender)
	s.Require().EqualValues(1, msg.PositionId)
	s.Require().Equal(sdk.NewInt(261585745), msg.Liquidity)
}
