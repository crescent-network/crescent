package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"
)

func (s *KeeperTestSuite) TestCreatePool() {
	k, ctx := s.keeper, s.ctx

	// Create a normal pool.
	poolCreator := s.addr(0)
	xCoin, yCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	s.createPool(poolCreator, xCoin, yCoin, true)

	// Check if our pool is set correctly.
	pool, found := k.GetPool(ctx, 1)
	s.Require().True(found)
	s.Require().Equal("denom1", pool.XCoinDenom)
	s.Require().Equal("denom2", pool.YCoinDenom)
	s.Require().Equal(types.PoolCoinDenom(pool.Id), pool.PoolCoinDenom)
	s.Require().True(pool.GetReserveAddress().Equals(types.PoolReserveAcc(pool.Id)))
	s.Require().False(pool.Disabled)

	// A new pair has been created automatically, too.
	pair, found := k.GetPairByDenoms(ctx, "denom1", "denom2")
	s.Require().True(found)
	s.Require().Equal(pair.Id, pool.PairId)
	s.Require().Equal("denom1", pair.XCoinDenom)
	s.Require().Equal("denom2", pair.YCoinDenom)
	s.Require().True(pair.GetEscrowAddress().Equals(types.PairEscrowAddr(pair.Id)))
	s.Require().Equal(uint64(1), pair.CurrentBatchId)
	s.Require().Nil(pair.LastPrice)
}

func (s *KeeperTestSuite) TestPoolCreationFee() {
	k, ctx := s.keeper, s.ctx

	poolCreator := s.addr(0)
	xCoin, yCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	s.fundAddr(poolCreator, sdk.NewCoins(xCoin, yCoin))

	// The pool creator doesn't have enough balance to pay the pool creation fee.
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, xCoin, yCoin))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
}

func (s *KeeperTestSuite) TestCreatePoolWithInsufficientDepositAmount() {
	k, ctx := s.keeper, s.ctx

	// A user tries to create a pool with smaller amounts of coin
	// than the minimum initial deposit amount.
	// This should fail.
	params := k.GetParams(ctx)
	minDepositAmount := params.MinInitialDepositAmount
	creator := s.addr(0)
	xCoin := sdk.NewCoin("denom1", minDepositAmount.Sub(sdk.OneInt()))
	yCoin := sdk.NewCoin("denom2", minDepositAmount)
	s.fundAddr(creator, sdk.NewCoins(xCoin, yCoin).Add(params.PoolCreationFee...))
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(creator, xCoin, yCoin))
	s.Require().ErrorIs(err, types.ErrInsufficientDepositAmount)
}

func (s *KeeperTestSuite) TestCreateSamePool() {
	k, ctx := s.keeper, s.ctx

	// Create a pool with denom1 and denom2.
	s.createPool(s.addr(0), parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)

	// A user tries to create a pool with same denom pair that already exists,
	// this will fail.
	poolCreator := s.addr(1)
	xCoin, yCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	params := k.GetParams(ctx)
	s.fundAddr(poolCreator, sdk.NewCoins(xCoin, yCoin).Add(params.PoolCreationFee...))
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, xCoin, yCoin))
	s.Require().ErrorIs(err, types.ErrPoolAlreadyExists)

	// Since the order of denom pair is important, it's ok to create a pool
	// with reversed denom pair:
	s.createPool(poolCreator, parseCoin("1000000denom2"), parseCoin("1000000denom1"), true)
}

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
	s.Require().ErrorIs(err, types.ErrDisabledPool)
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
	s.Require().ErrorIs(err, types.ErrDisabledPool)
}

func (s *KeeperTestSuite) TestCreatePoolAfterDisabled() {
	// Create a disabled pool.
	poolCreator := s.addr(0)
	pool := s.createPool(poolCreator, parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
	s.withdrawBatch(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	s.nextBlock()

	// Now a new pool can be created with same denom pair because
	// all pools with same denom pair are disabled.
	s.createPool(s.addr(1), parseCoin("1000000denom1"), parseCoin("1000000denom2"), true)
}
