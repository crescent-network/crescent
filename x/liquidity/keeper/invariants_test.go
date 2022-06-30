package keeper_test

import (
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/keeper"
)

func (s *KeeperTestSuite) TestDepositCoinsEscrowInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	req := s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	_, broken := keeper.DepositCoinsEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldReq := req
	req.DepositCoins = utils.ParseCoins("2000000denom1,2000000denom2")
	s.keeper.SetDepositRequest(s.ctx, req)
	_, broken = keeper.DepositCoinsEscrowInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	req = oldReq
	s.keeper.SetDepositRequest(s.ctx, req)
	s.nextBlock()
	_, broken = keeper.DepositCoinsEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)
}

func (s *KeeperTestSuite) TestPoolCoinEscrowInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	req := s.withdraw(s.addr(1), pool.Id, utils.ParseCoin("1000000pool1"))
	_, broken := keeper.PoolCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldReq := req
	req.PoolCoin = utils.ParseCoin("2000000pool1")
	s.keeper.SetWithdrawRequest(s.ctx, req)
	_, broken = keeper.PoolCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	req = oldReq
	s.keeper.SetWithdrawRequest(s.ctx, req)
	s.nextBlock()
	_, broken = keeper.PoolCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)
}

func (s *KeeperTestSuite) TestRemainingOfferCoinEscrowInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	order := s.buyLimitOrder(s.addr(1), pair.Id, utils.ParseDec("1.0"), newInt(1000000), 0, true)
	_, broken := keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldOrder := order
	order.RemainingOfferCoin = utils.ParseCoin("2000000denom1")
	s.keeper.SetOrder(s.ctx, order)
	_, broken = keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	order = oldOrder
	s.keeper.SetOrder(s.ctx, order)
	s.nextBlock()
	_, broken = keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)
}

func (s *KeeperTestSuite) TestPoolStatusInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	_, broken := keeper.PoolStatusInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	s.withdraw(s.addr(0), pool.Id, s.getBalance(s.addr(0), pool.PoolCoinDenom))
	s.nextBlock()

	_, broken = keeper.PoolStatusInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	pool, _ = s.keeper.GetPool(s.ctx, pool.Id)
	pool.Disabled = false
	s.keeper.SetPool(s.ctx, pool)
	_, broken = keeper.PoolStatusInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)
}
