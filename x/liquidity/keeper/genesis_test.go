package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func (s *KeeperTestSuite) TestDefaultGenesis() {
	genState := *types.DefaultGenesis()

	s.keeper.InitGenesis(s.ctx, genState)
	got := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState, *got)
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	poolCoin := s.getBalance(s.addr(1), pool.PoolCoinDenom)
	poolCoin.Amount = poolCoin.Amount.QuoRaw(2)
	s.withdraw(s.addr(1), pool.Id, poolCoin)
	s.nextBlock()

	s.buyLimitOrder(s.addr(2), pair.Id, utils.ParseDec("1.0"), newInt(10000), 0, true)
	s.nextBlock()

	depositReq := s.deposit(s.addr(3), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	withdrawReq := s.withdraw(s.addr(1), pool.Id, poolCoin)
	order := s.sellLimitOrder(s.addr(3), pair.Id, utils.ParseDec("1.0"), newInt(1000), 0, true)

	genState := s.keeper.ExportGenesis(s.ctx)

	bz := s.app.AppCodec().MustMarshalJSON(genState)

	s.SetupTest()
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(utils.ParseTime("2022-01-01T00:00:00Z"))

	var genState2 types.GenesisState
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)
	genState3 := s.keeper.ExportGenesis(s.ctx)

	s.Require().Equal(*genState, *genState3)

	depositReq2, found := s.keeper.GetDepositRequest(s.ctx, depositReq.PoolId, depositReq.Id)
	s.Require().True(found)
	s.Require().Equal(depositReq, depositReq2)
	withdrawReq2, found := s.keeper.GetWithdrawRequest(s.ctx, withdrawReq.PoolId, withdrawReq.Id)
	s.Require().True(found)
	s.Require().Equal(withdrawReq, withdrawReq2)
	order2, found := s.keeper.GetOrder(s.ctx, order.PairId, order.Id)
	s.Require().True(found)
	s.Require().Equal(order, order2)
}

func (s *KeeperTestSuite) TestImportExportGenesisEmpty() {
	genState := s.keeper.ExportGenesis(s.ctx)

	var genState2 types.GenesisState
	bz := s.app.AppCodec().MustMarshalJSON(genState)
	s.app.AppCodec().MustUnmarshalJSON(bz, &genState2)
	s.keeper.InitGenesis(s.ctx, genState2)

	genState3 := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(*genState, genState2)
	s.Require().Equal(genState2, *genState3)
}

func (s *KeeperTestSuite) TestIndexesAfterImport() {
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(utils.ParseTime("2022-03-01T00:00:00Z"))

	pair1 := s.createPair(s.addr(0), "denom1", "denom2", true)
	pair2 := s.createPair(s.addr(1), "denom2", "denom3", true)

	pool1 := s.createPool(s.addr(2), pair1.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	pool2 := s.createPool(s.addr(3), pair2.Id, utils.ParseCoins("1000000denom2,1000000denom3"), true)

	s.deposit(s.addr(4), pool1.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.deposit(s.addr(5), pool2.Id, utils.ParseCoins("1000000denom2,1000000denom3"), true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	liquidity.BeginBlocker(s.ctx, s.keeper)

	depositReq1 := s.deposit(s.addr(4), pool1.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	depositReq2 := s.deposit(s.addr(5), pool2.Id, utils.ParseCoins("1000000denom2,1000000denom3"), true)

	withdrawReq1 := s.withdraw(s.addr(4), pool1.Id, utils.ParseCoin("1000000pool1"))
	withdrawReq2 := s.withdraw(s.addr(5), pool2.Id, utils.ParseCoin("1000000pool2"))

	order1 := s.limitOrder(s.addr(6), pair1.Id, types.OrderDirectionBuy, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Minute, true)
	order2 := s.limitOrder(s.addr(7), pair2.Id, types.OrderDirectionSell, utils.ParseDec("1.0"), sdk.NewInt(10000), time.Minute, true)

	liquidity.EndBlocker(s.ctx, s.keeper)

	genState := s.keeper.ExportGenesis(s.ctx)
	s.SetupTest()
	s.ctx = s.ctx.WithBlockHeight(1).WithBlockTime(utils.ParseTime("2022-03-02T00:00:00Z"))
	s.keeper.InitGenesis(s.ctx, *genState)

	// Check pair indexes.
	pair, found := s.keeper.GetPairByDenoms(s.ctx, "denom1", "denom2")
	s.Require().True(found)
	s.Require().Equal(pair1.Id, pair.Id)

	resp1, err := s.querier.Pairs(sdk.WrapSDKContext(s.ctx), &types.QueryPairsRequest{
		Denoms: []string{"denom2", "denom1"},
	})
	s.Require().NoError(err)
	s.Require().Len(resp1.Pairs, 1)
	s.Require().Equal(pair1.Id, resp1.Pairs[0].Id)

	resp2, err := s.querier.Pairs(sdk.WrapSDKContext(s.ctx), &types.QueryPairsRequest{
		Denoms: []string{"denom2", "denom3"},
	})
	s.Require().NoError(err)
	s.Require().Len(resp2.Pairs, 1)
	s.Require().Equal(pair2.Id, resp2.Pairs[0].Id)

	// Check pool indexes.
	pools := s.keeper.GetPoolsByPair(s.ctx, pair2.Id)
	s.Require().Len(pools, 1)
	s.Require().Equal(pool2.Id, pools[0].Id)

	pool, found := s.keeper.GetPoolByReserveAddress(s.ctx, pool1.GetReserveAddress())
	s.Require().True(found)
	s.Require().Equal(pool1.Id, pool.Id)

	// Check deposit request indexes.
	depositReqs := s.keeper.GetDepositRequestsByDepositor(s.ctx, s.addr(4))
	s.Require().Len(depositReqs, 1)
	s.Require().Equal(depositReq1.Id, depositReqs[0].Id)

	depositReqs = s.keeper.GetDepositRequestsByDepositor(s.ctx, s.addr(5))
	s.Require().Len(depositReqs, 1)
	s.Require().Equal(depositReq2.Id, depositReqs[0].Id)

	// Check withdraw request indexes
	withdrawReqs := s.keeper.GetWithdrawRequestsByWithdrawer(s.ctx, s.addr(4))
	s.Require().Len(withdrawReqs, 1)
	s.Require().Equal(withdrawReq1.Id, withdrawReqs[0].Id)

	withdrawReqs = s.keeper.GetWithdrawRequestsByWithdrawer(s.ctx, s.addr(5))
	s.Require().Len(withdrawReqs, 1)
	s.Require().Equal(withdrawReq2.Id, withdrawReqs[0].Id)

	// Check order indexes
	orders := s.keeper.GetOrdersByOrderer(s.ctx, s.addr(6))
	s.Require().Len(orders, 1)
	s.Require().Equal(order1.Id, orders[0].Id)

	orders = s.keeper.GetOrdersByOrderer(s.ctx, s.addr(7))
	s.Require().Len(orders, 1)
	s.Require().Equal(order2.Id, orders[0].Id)
}
