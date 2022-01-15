package keeper_test

import (
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (s *KeeperTestSuite) TestDisabledPool() {
	k, ctx := s.keeper, s.ctx
	// A disabled pool is:
	// 1. A pool with at least one side of its x/y coin's balance is 0.
	// 2. A pool with 0 pool coin supply(all investors has withdrawn their coins)

	poolCreator := s.addr(0)
	// Create a pool.
	pool := s.createPool(poolCreator, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	// Send the pool's balances to somewhere else.
	s.sendCoins(pool.GetReserveAddress(), s.addr(0), s.getBalances(pool.GetReserveAddress()))

	// By now, the pool is nor marked as disabled automatically.
	// When someone sends a deposit/withdraw request to the pool or
	// the pool tries to participate in matching, then the pool
	// is marked as disabled.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().False(pool.Disabled)

	// A depositor tries to deposit to the pool.
	s.depositBatch(s.addr(1), pool.Id, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	s.nextBlock()

	// Now, the pool is disabled.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().True(pool.Disabled)

	// Here's the second example.
	// This time, the pool creator withdraws all his coins.
	pool = s.createPool(poolCreator, parseCoin("1000000denom3"), parseCoin("1000000denom4"), true)
	s.withdrawBatch(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	s.nextBlock()

	// The pool is disabled again.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().True(pool.Disabled)
}

func (s *KeeperTestSuite) TestDepositToDisabledPool() {
	k, ctx := s.keeper, s.ctx

	// Create a disabled pool by sending the pool's balances to somewhere else.
	pool := s.createPool(s.addr(0), parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	poolReserveAddr := pool.GetReserveAddress()
	s.sendCoins(poolReserveAddr, s.addr(1), s.getBalances(poolReserveAddr))

	// The depositor deposits coins but this will fail because the pool
	// is treated as disabled.
	depositor := s.addr(2)
	depositXCoin, depositYCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	req := s.depositBatch(depositor, pool.Id, depositXCoin, depositYCoin, true)
	err := k.ExecuteDepositRequest(ctx, req)
	s.Require().NoError(err)
	req, _ = k.GetDepositRequest(ctx, pool.Id, req.Id)
	s.Require().False(req.Succeeded)

	// Delete the previous request and refund coins to the depositor.
	liquidity.BeginBlocker(ctx, k)

	// Now any deposits will result in an error.
	_, err = k.DepositBatch(ctx, types.NewMsgDepositBatch(depositor, pool.Id, depositXCoin, depositYCoin))
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestWithdrawFromDisabledPool() {
	k, ctx := s.keeper, s.ctx

	// Create a disabled pool by sending the pool's balances to somewhere else.
	poolCreator := s.addr(0)
	pool := s.createPool(poolCreator, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	poolReserveAddr := pool.GetReserveAddress()
	s.sendCoins(poolReserveAddr, s.addr(1), s.getBalances(poolReserveAddr))

	// The pool creator tries to withdraw his coins, but this will fail.
	req := s.withdrawBatch(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	err := k.ExecuteWithdrawRequest(ctx, req)
	s.Require().NoError(err)
	req, _ = k.GetWithdrawRequest(ctx, pool.Id, req.Id)
	s.Require().False(req.Succeeded)

	// Delete the previous request and refund coins to the withdrawer.
	liquidity.BeginBlocker(ctx, k)

	// Now any withdrawals will result in an error.
	_, err = k.WithdrawBatch(ctx, types.NewMsgWithdrawBatch(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom)))
	s.Require().Error(err)
}
