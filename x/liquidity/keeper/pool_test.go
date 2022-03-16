package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/types"
	"github.com/crescent-network/crescent/x/liquidity"
	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestCreatePool() {
	k, ctx := s.keeper, s.ctx

	// Create a pair.
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a normal pool.
	poolCreator := s.addr(1)
	s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	// Check if our pool is set correctly.
	pool, found := k.GetPool(ctx, 1)
	s.Require().True(found)
	s.Require().Equal(types.PoolCoinDenom(pool.Id), pool.PoolCoinDenom)
	s.Require().True(pool.GetReserveAddress().Equals(types.PoolReserveAddress(pool.Id)))
	s.Require().False(pool.Disabled)
}

func (s *KeeperTestSuite) TestPoolCreationFee() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	poolCreator := s.addr(1)
	depositCoins := utils.ParseCoins("1000000denom1,1000000denom2")
	s.fundAddr(poolCreator, depositCoins)

	// The pool creator doesn't have enough balance to pay the pool creation fee.
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, pair.Id, depositCoins))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
}

func (s *KeeperTestSuite) TestCreatePoolWithInsufficientDepositAmount() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// A user tries to create a pool with smaller amounts of coin
	// than the minimum initial deposit amount.
	// This should fail.
	poolCreator := s.addr(1)
	params := k.GetParams(ctx)
	minDepositAmount := params.MinInitialDepositAmount
	xCoin := sdk.NewCoin("denom1", minDepositAmount.Sub(sdk.OneInt()))
	yCoin := sdk.NewCoin("denom2", minDepositAmount)
	s.fundAddr(poolCreator, sdk.NewCoins(xCoin, yCoin).Add(params.PoolCreationFee...))
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, pair.Id, sdk.NewCoins(xCoin, yCoin)))
	s.Require().ErrorIs(err, types.ErrInsufficientDepositAmount)
}

func (s *KeeperTestSuite) TestCreateSamePool() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pair2 := s.createPair(s.addr(0), "denom2", "denom1", true)

	// Create a pool with denom1 and denom2.
	s.createPool(s.addr(1), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	// A user tries to create a pool with same denom pair that already exists,
	// this will fail.
	poolCreator := s.addr(2)
	depositCoins := utils.ParseCoins("1000000denom1,1000000denom2")
	params := k.GetParams(ctx)
	s.fundAddr(poolCreator, depositCoins.Add(params.PoolCreationFee...))
	_, err := k.CreatePool(ctx, types.NewMsgCreatePool(poolCreator, pair.Id, depositCoins))
	s.Require().ErrorIs(err, types.ErrPoolAlreadyExists)

	// Since the order of denom pair is important, it's ok to create a pool
	// with reversed denom pair:
	s.createPool(poolCreator, pair2.Id, utils.ParseCoins("1000000denom2,1000000denom1"), true)
}

func (s *KeeperTestSuite) TestDisabledPool() {
	// A disabled pool is:
	// 1. A pool with at least one side of its x/y coin's balance is 0.
	// 2. A pool with 0 pool coin supply(all investors has withdrawn their coins)

	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pair2 := s.createPair(s.addr(0), "denom3", "denom4", true)

	poolCreator := s.addr(1)
	// Create a pool.
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	// Send the pool's balances to somewhere else.
	s.sendCoins(pool.GetReserveAddress(), s.addr(2), s.getBalances(pool.GetReserveAddress()))

	// By now, the pool is not marked as disabled automatically.
	// When someone sends a deposit/withdraw request to the pool or
	// the pool tries to participate in matching, then the pool
	// is marked as disabled.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().False(pool.Disabled)

	// A depositor tries to deposit to the pool.
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	// Now, the pool is disabled.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().True(pool.Disabled)

	// Here's the second example.
	// This time, the pool creator withdraws all his coins.
	pool = s.createPool(poolCreator, pair2.Id, utils.ParseCoins("1000000denom3,1000000denom4"), true)
	s.withdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	s.nextBlock()

	// The pool is disabled again.
	pool, _ = k.GetPool(ctx, pool.Id)
	s.Require().True(pool.Disabled)
}

func (s *KeeperTestSuite) TestCreatePoolAfterDisabled() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a disabled pool.
	poolCreator := s.addr(1)
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.withdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	s.nextBlock()

	// Now a new pool can be created with same denom pair because
	// all pools with same denom pair are disabled.
	s.createPool(s.addr(2), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
}

func (s *KeeperTestSuite) TestCreatePoolInitialPoolCoinSupply() {
	params := s.keeper.GetParams(s.ctx)

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	poolCreator := s.addr(1)
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.Require().True(intEq(params.MinInitialPoolCoinSupply, s.getBalance(poolCreator, pool.PoolCoinDenom).Amount))

	pair = s.createPair(s.addr(0), "denom2", "denom3", true)

	pool = s.createPool(poolCreator, pair.Id, utils.ParseCoins("2000000000000denom2,500000000000000denom3"), true)
	s.Require().True(intEq(sdk.NewInt(100000000000000), s.getBalance(poolCreator, pool.PoolCoinDenom).Amount))
}

func (s *KeeperTestSuite) TestDeposit() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	// The depositor now has the same pool coin as the pool creator.
	expectedPoolCoin := s.getBalance(s.addr(0), pool.PoolCoinDenom)
	s.Require().True(coinsEq(sdk.NewCoins(expectedPoolCoin), s.getBalances(s.addr(1))))

	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("500000denom1,500000denom2"), true)
	s.nextBlock()

	// The next depositor has 1/2 pool coin of the pool creator.
	expectedPoolCoin = sdk.NewCoin(pool.PoolCoinDenom, s.getBalance(s.addr(0), pool.PoolCoinDenom).Amount.QuoRaw(2))
	s.Require().True(coinsEq(sdk.NewCoins(expectedPoolCoin), s.getBalances(s.addr(2))))
}

func (s *KeeperTestSuite) TestDepositRefund() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1500000denom2"), true)

	depositor := s.addr(1)
	depositCoins := utils.ParseCoins("20000denom1,15000denom2")
	s.fundAddr(depositor, depositCoins)
	req := s.deposit(depositor, pool.Id, depositCoins, false)
	liquidity.EndBlocker(s.ctx, s.keeper)
	req, _ = s.keeper.GetDepositRequest(s.ctx, req.PoolId, req.Id)
	s.Require().Equal(types.RequestStatusSucceeded, req.Status)

	s.Require().True(coinEq(utils.ParseCoin("10000denom1"), s.getBalance(depositor, "denom1")))
	s.Require().True(coinEq(utils.ParseCoin("0denom2"), s.getBalance(depositor, "denom2")))
	liquidity.BeginBlocker(s.ctx, s.keeper)

	pair = s.createPair(s.addr(0), "denom2", "denom1", true)
	pool = s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000000denom2,1000000000000000denom1"), true)

	depositor = s.addr(2)
	depositCoins = utils.ParseCoins("1denom1,1denom2")
	s.fundAddr(depositor, depositCoins)
	req = s.deposit(depositor, pool.Id, depositCoins, false)
	liquidity.EndBlocker(s.ctx, s.keeper)
	req, _ = s.keeper.GetDepositRequest(s.ctx, req.PoolId, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	s.Require().True(coinsEq(depositCoins, s.getBalances(depositor)))
}

func (s *KeeperTestSuite) TestDepositRefundTooSmallMintedPoolCoin() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1500000denom2"), true)

	depositor := s.addr(1)
	depositCoins := utils.ParseCoins("20000denom1,15000denom2")
	s.fundAddr(depositor, depositCoins)
	req := s.deposit(depositor, pool.Id, depositCoins, false)
	liquidity.EndBlocker(s.ctx, s.keeper)
	req, _ = s.keeper.GetDepositRequest(s.ctx, req.PoolId, req.Id)
	s.Require().Equal(types.RequestStatusSucceeded, req.Status)

	s.Require().True(coinEq(utils.ParseCoin("10000denom1"), s.getBalance(depositor, "denom1")))
	s.Require().True(coinEq(utils.ParseCoin("0denom2"), s.getBalance(depositor, "denom2")))
	liquidity.BeginBlocker(s.ctx, s.keeper)

	pair = s.createPair(s.addr(0), "denom2", "denom1", true)
	pool = s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000000denom2,1000000000000000denom1"), true)

	depositor = s.addr(2)
	depositCoins = utils.ParseCoins("1denom1,1denom2")
	s.fundAddr(depositor, depositCoins)
	req = s.deposit(depositor, pool.Id, depositCoins, false)
	liquidity.EndBlocker(s.ctx, s.keeper)
	req, _ = s.keeper.GetDepositRequest(s.ctx, req.PoolId, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	s.Require().True(coinsEq(depositCoins, s.getBalances(depositor)))
}

func (s *KeeperTestSuite) TestWithdraw() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	depositor := s.addr(1)
	depositCoins := utils.ParseCoins("1000000denom1,1000000denom2")
	s.deposit(depositor, pool.Id, depositCoins, true)
	s.nextBlock()

	poolCoin := s.getBalance(depositor, pool.PoolCoinDenom)
	s.withdraw(depositor, pool.Id, poolCoin)
	s.nextBlock()

	s.Require().True(coinsEq(depositCoins, s.getBalances(depositor)))
}

func (s *KeeperTestSuite) TestWithdrawRefund() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	depositor := s.addr(1)
	s.deposit(depositor, pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	// Make the pool depleted.
	s.sendCoins(pool.GetReserveAddress(), s.addr(2), s.getBalances(pool.GetReserveAddress()))

	poolCoin := s.getBalance(depositor, pool.PoolCoinDenom)
	s.withdraw(depositor, pool.Id, poolCoin)
	s.nextBlock()

	s.Require().True(coinsEq(sdk.NewCoins(poolCoin), s.getBalances(depositor)))
}

func (s *KeeperTestSuite) TestWithdrawRefundTooSmallWithdrawCoins() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	depositor := s.addr(1)
	s.deposit(depositor, pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()
	poolCoin := s.getBalance(depositor, pool.PoolCoinDenom)

	// Withdrawing too small amount of pool coin.
	s.withdraw(depositor, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 100))
	s.nextBlock()

	s.Require().True(coinsEq(sdk.NewCoins(poolCoin), s.getBalances(depositor)))
}

func (s *KeeperTestSuite) TestDepositToDisabledPool() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a disabled pool by sending the pool's balances to somewhere else.
	pool := s.createPool(s.addr(1), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	poolReserveAddr := pool.GetReserveAddress()
	s.sendCoins(poolReserveAddr, s.addr(2), s.getBalances(poolReserveAddr))

	// The depositor deposits coins but this will fail because the pool
	// is treated as disabled.
	depositor := s.addr(3)
	depositCoins := utils.ParseCoins("1000000denom1,1000000denom2")
	req := s.deposit(depositor, pool.Id, depositCoins, true)
	err := k.ExecuteDepositRequest(ctx, req)
	s.Require().NoError(err)
	req, _ = k.GetDepositRequest(ctx, pool.Id, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	// Delete the previous request and refund coins to the depositor.
	liquidity.BeginBlocker(ctx, k)

	// Now any deposits will result in an error.
	_, err = k.Deposit(ctx, types.NewMsgDeposit(depositor, pool.Id, depositCoins))
	s.Require().ErrorIs(err, types.ErrDisabledPool)
}

func (s *KeeperTestSuite) TestWithdrawFromDisabledPool() {
	k, ctx := s.keeper, s.ctx

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a disabled pool by sending the pool's balances to somewhere else.
	poolCreator := s.addr(1)
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	poolReserveAddr := pool.GetReserveAddress()
	s.sendCoins(poolReserveAddr, s.addr(1), s.getBalances(poolReserveAddr))

	// The pool creator tries to withdraw his coins, but this will fail.
	req := s.withdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	err := k.ExecuteWithdrawRequest(ctx, req)
	s.Require().NoError(err)
	req, _ = k.GetWithdrawRequest(ctx, pool.Id, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	// Delete the previous request and refund coins to the withdrawer.
	liquidity.BeginBlocker(ctx, k)

	// Now any withdrawals will result in an error.
	_, err = k.Withdraw(ctx, types.NewMsgWithdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom)))
	s.Require().ErrorIs(err, types.ErrDisabledPool)
}

func (s *KeeperTestSuite) TestGetDepositRequestsByDepositor() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	req1 := s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	req2 := s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	reqs := s.keeper.GetDepositRequestsByDepositor(s.ctx, s.addr(1))
	s.Require().Len(reqs, 2)
	s.Require().Equal(req1.PoolId, reqs[0].PoolId)
	s.Require().Equal(req1.Id, reqs[0].Id)
	s.Require().Equal(req2.PoolId, reqs[1].PoolId)
	s.Require().Equal(req2.Id, reqs[1].Id)
}

func (s *KeeperTestSuite) TestWithdrawRequestsByWithdrawer() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()
	req1 := s.withdraw(s.addr(1), pool.Id, utils.ParseCoin("10000pool1"))
	req2 := s.withdraw(s.addr(1), pool.Id, utils.ParseCoin("10000pool1"))
	reqs := s.keeper.GetWithdrawRequestsByWithdrawer(s.ctx, s.addr(1))
	s.Require().Len(reqs, 2)
	s.Require().Equal(req1.PoolId, reqs[0].PoolId)
	s.Require().Equal(req1.Id, reqs[0].Id)
	s.Require().Equal(req2.PoolId, reqs[1].PoolId)
	s.Require().Equal(req2.Id, reqs[1].Id)
}
