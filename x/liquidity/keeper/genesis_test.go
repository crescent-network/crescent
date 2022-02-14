package keeper_test

import (
	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(squad.ParseTime("2022-01-01T00:00:00Z"))
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)

	s.deposit(s.addr(1), pool.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	poolCoin := s.getBalance(s.addr(1), pool.PoolCoinDenom)
	poolCoin.Amount = poolCoin.Amount.QuoRaw(2)
	s.withdraw(s.addr(1), pool.Id, poolCoin)
	s.nextBlock()

	s.buyLimitOrder(s.addr(2), pair.Id, squad.ParseDec("1.0"), newInt(10000), 0, true)
	s.nextBlock()

	depositReq := s.deposit(s.addr(3), pool.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)
	withdrawReq := s.withdraw(s.addr(1), pool.Id, poolCoin)
	order := s.sellLimitOrder(s.addr(3), pair.Id, squad.ParseDec("1.0"), newInt(1000), 0, true)

	genState := k.ExportGenesis(ctx)

	bz := s.app.AppCodec().MustMarshalJSON(genState)

	s.SetupTest()
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(squad.ParseTime("2022-01-01T00:00:00Z"))
	k, ctx = s.keeper, s.ctx

	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	k.InitGenesis(ctx, genState2)
	genState3 := k.ExportGenesis(ctx)

	s.Require().Equal(*genState, *genState3)

	depositReq2, found := k.GetDepositRequest(ctx, depositReq.PoolId, depositReq.Id)
	s.Require().True(found)
	s.Require().Equal(depositReq, depositReq2)
	withdrawReq2, found := k.GetWithdrawRequest(ctx, withdrawReq.PoolId, withdrawReq.Id)
	s.Require().True(found)
	s.Require().Equal(withdrawReq, withdrawReq2)
	order2, found := k.GetOrder(ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(order, order2)
}
