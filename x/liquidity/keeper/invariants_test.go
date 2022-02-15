package keeper_test

import (
	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity/keeper"
)

func (s *KeeperTestSuite) TestDepositCoinsEscrowInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)

	req := s.deposit(s.addr(1), pool.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)
	_, broken := keeper.DepositCoinsEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldReq := req
	req.DepositCoins = squad.ParseCoins("2000000denom1,2000000denom2")
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
	pool := s.createPool(s.addr(0), pair.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)

	s.deposit(s.addr(1), pool.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	req := s.withdraw(s.addr(1), pool.Id, squad.ParseCoin("1000000pool1"))
	_, broken := keeper.PoolCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldReq := req
	req.PoolCoin = squad.ParseCoin("2000000pool1")
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

	req := s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), newInt(1000000), 0, true)
	_, broken := keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	oldReq := req
	req.RemainingOfferCoin = squad.ParseCoin("2000000denom1")
	s.keeper.SetOrder(s.ctx, req)
	_, broken = keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	req = oldReq
	s.keeper.SetOrder(s.ctx, req)
	s.nextBlock()
	_, broken = keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)
}

func (s *KeeperTestSuite) TestPoolStatusInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)

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
