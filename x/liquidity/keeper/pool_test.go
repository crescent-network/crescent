package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestCreatePool() {
	// Create a pair.
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a normal pool.
	poolCreator := s.addr(1)
	s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	// Check if our pool is set correctly.
	pool, found := s.keeper.GetPool(s.ctx, 1)
	s.Require().True(found)
	s.Require().Equal(types.PoolCoinDenom(pool.Id), pool.PoolCoinDenom)
	s.Require().True(pool.GetReserveAddress().Equals(types.PoolReserveAddress(pool.Id)))
	s.Require().False(pool.Disabled)
}

func (s *KeeperTestSuite) TestCreateRangedPool() {
	params := s.keeper.GetParams(s.ctx)
	params.TickPrecision = 5 // to test too small min price case
	s.keeper.SetParams(s.ctx, params)

	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	poolCreator := s.addr(1)
	s.fundAddr(poolCreator, utils.ParseCoins("1000000denom1,1000000denom2,1000000denom3,1000000stake"))
	poolCreatorWithNoFee := s.addr(2)
	s.fundAddr(poolCreatorWithNoFee, utils.ParseCoins("1000000denom1,1000000denom2"))

	validDepositCoins := utils.ParseCoins("1000000denom1,1000000denom2")

	for _, tc := range []struct {
		name        string
		msg         *types.MsgCreateRangedPool
		postRun     func(ctx sdk.Context, pool types.Pool)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgCreateRangedPool(
				poolCreator, pair.Id, validDepositCoins,
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			func(ctx sdk.Context, pool types.Pool) {
				s.Require().Equal(types.PoolTypeRanged, pool.Type)
				s.Require().NotNil(pool.MinPrice)
				s.Require().True(decEq(utils.ParseDec("0.9"), *pool.MinPrice))
				s.Require().NotNil(pool.MaxPrice)
				s.Require().True(decEq(utils.ParseDec("1.1"), *pool.MaxPrice))
				s.Require().Equal(poolCreator.String(), pool.Creator)
				s.Require().Equal(types.PoolCoinDenom(pool.Id), pool.PoolCoinDenom)
				s.Require().Equal(types.PoolReserveAddress(pool.Id).String(), pool.ReserveAddress)
				s.Require().False(pool.Disabled)
				s.Require().True(coinsEq(
					utils.ParseCoins("906867denom1,1000000denom2"),
					s.app.BankKeeper.GetAllBalances(ctx, pool.GetReserveAddress())))
				s.Require().True(coinEq(
					utils.ParseCoin("1000000000000pool1"),
					s.app.BankKeeper.GetBalance(ctx, pool.GetCreator(), pool.PoolCoinDenom)))
			},
			"",
		},
		{
			"pair not found",
			types.NewMsgCreateRangedPool(
				poolCreator, 2, validDepositCoins,
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			nil,
			"pair 2 not found: not found",
		},
		{
			"wrong deposit coin denoms",
			types.NewMsgCreateRangedPool(
				poolCreator, pair.Id, utils.ParseCoins("1000000denom2,1000000denom3"),
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			nil,
			"coin denom denom3 is not in the pair: invalid coin denom",
		},
		{
			"insufficient deposit amount",
			types.NewMsgCreateRangedPool(
				poolCreator, pair.Id, utils.ParseCoins("999999denom1,999999denom2"),
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			nil,
			"insufficient deposit amount",
		},
		{
			"insufficient pool creation fee",
			types.NewMsgCreateRangedPool(
				poolCreatorWithNoFee, pair.Id, validDepositCoins,
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			nil,
			"insufficient pool creation fee: 0stake is smaller than 1000000stake: insufficient funds",
		},
		{
			"too small min price",
			types.NewMsgCreateRangedPool(
				poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
				sdk.NewDecWithPrec(1, 15), utils.ParseDec("1.1"), utils.ParseDec("1.0")),
			nil,
			"min price must not be less than 0.000000000000100000: invalid request",
		},
		{
			"too small deposit amount",
			types.NewMsgCreateRangedPool(
				poolCreator, pair.Id, utils.ParseCoins("1000000denom1,10denom2"),
				utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.1")),
			nil,
			"insufficient deposit amount",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.ctx.CacheContext()
			pool, err := s.keeper.CreateRangedPool(cacheCtx, tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				pool, found := s.keeper.GetPool(cacheCtx, pool.Id)
				s.Require().True(found)
				tc.postRun(cacheCtx, pool)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPoolCreationFee() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	poolCreator := s.addr(1)
	depositCoins := utils.ParseCoins("1000000denom1,1000000denom2")
	s.fundAddr(poolCreator, depositCoins)

	// The pool creator doesn't have enough balance to pay the pool creation fee.
	_, err := s.keeper.CreatePool(s.ctx, types.NewMsgCreatePool(poolCreator, pair.Id, depositCoins))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
}

func (s *KeeperTestSuite) TestCreatePoolWithInsufficientDepositAmount() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// A user tries to create a pool with smaller amounts of coin
	// than the minimum initial deposit amount.
	// This should fail.
	poolCreator := s.addr(1)
	minDepositAmount := s.keeper.GetMinInitialDepositAmount(s.ctx)
	xCoin := sdk.NewCoin("denom1", minDepositAmount.Sub(sdk.OneInt()))
	yCoin := sdk.NewCoin("denom2", minDepositAmount)
	s.fundAddr(poolCreator, sdk.NewCoins(xCoin, yCoin).Add(s.keeper.GetPoolCreationFee(s.ctx)...))
	_, err := s.keeper.CreatePool(s.ctx, types.NewMsgCreatePool(poolCreator, pair.Id, sdk.NewCoins(xCoin, yCoin)))
	s.Require().ErrorIs(err, types.ErrInsufficientDepositAmount)
}

func (s *KeeperTestSuite) TestCreateSamePool() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a pool with denom1 and denom2.
	s.createPool(s.addr(1), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	// This will fail since there's already a basic pool.
	depositCoins := utils.ParseCoins("2000000denom1,1000000denom2")
	s.fundAddr(s.addr(2), depositCoins.Add(s.keeper.GetPoolCreationFee(s.ctx)...))
	msg := types.NewMsgCreatePool(s.addr(2), pair.Id, depositCoins)
	s.Require().NoError(msg.ValidateBasic())
	_, err := s.keeper.CreatePool(s.ctx, msg)
	s.Require().ErrorIs(err, types.ErrPoolAlreadyExists)

	// However, this will not fail since it's creating a ranged pool.
	s.createRangedPool(
		s.addr(3), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
		utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0"), true)

	// Creation of multiple ranged pools with same parameters is allowed.
	s.createRangedPool(
		s.addr(4), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
		utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0"), true)
}

func (s *KeeperTestSuite) TestDisabledPool() {
	// A disabled pool is:
	// 1. A pool with at least one side of its x/y coin's balance is 0.
	// 2. A pool with 0 pool coin supply(all investors has withdrawn their coins)

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
	pool, _ = s.keeper.GetPool(s.ctx, pool.Id)
	s.Require().False(pool.Disabled)

	// A depositor tries to deposit to the pool.
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	// Now, the pool is disabled.
	pool, _ = s.keeper.GetPool(s.ctx, pool.Id)
	s.Require().True(pool.Disabled)

	// Here's the second example.
	// This time, the pool creator withdraws all his coins.
	pool = s.createPool(poolCreator, pair2.Id, utils.ParseCoins("1000000denom3,1000000denom4"), true)
	s.withdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	s.nextBlock()

	// The pool is disabled again.
	pool, _ = s.keeper.GetPool(s.ctx, pool.Id)
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
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	poolCreator := s.addr(1)
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.Require().True(intEq(s.keeper.GetMinInitialPoolCoinSupply(s.ctx), s.getBalance(poolCreator, pool.PoolCoinDenom).Amount))

	pair = s.createPair(s.addr(0), "denom2", "denom3", true)

	pool = s.createPool(poolCreator, pair.Id, utils.ParseCoins("2000000000000denom2,500000000000000denom3"), true)
	s.Require().True(intEq(sdk.NewInt(100000000000000), s.getBalance(poolCreator, pool.PoolCoinDenom).Amount))
}

func (s *KeeperTestSuite) TestPoolIndexes() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	pool := s.createPool(s.addr(1), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	pool2, found := s.keeper.GetPoolByReserveAddress(s.ctx, pool.GetReserveAddress())
	s.Require().True(found)
	s.Require().Equal(pool.Id, pool2.Id)

	pools := s.keeper.GetPoolsByPair(s.ctx, pair.Id)
	s.Require().Len(pools, 1)
	s.Require().Equal(pool.Id, pools[0].Id)
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

func (s *KeeperTestSuite) TestTooLargePool() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)

	_, err := s.keeper.Deposit(s.ctx, types.NewMsgDeposit(s.addr(1), pool.Id, utils.ParseCoins("10000000000000000000000000000000000000000denom1,10000000000000000000000000000000000000000denom2")))
	s.Require().ErrorIs(err, types.ErrTooLargePool)
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
	err := s.keeper.ExecuteDepositRequest(s.ctx, req)
	s.Require().NoError(err)
	req, _ = s.keeper.GetDepositRequest(s.ctx, pool.Id, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	// Delete the previous request and refund coins to the depositor.
	liquidity.BeginBlocker(s.ctx, s.keeper)

	// Now any deposits will result in an error.
	_, err = s.keeper.Deposit(s.ctx, types.NewMsgDeposit(depositor, pool.Id, depositCoins))
	s.Require().ErrorIs(err, types.ErrDisabledPool)
}

func (s *KeeperTestSuite) TestWithdrawFromDisabledPool() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	// Create a disabled pool by sending the pool's balances to somewhere else.
	poolCreator := s.addr(1)
	pool := s.createPool(poolCreator, pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	poolReserveAddr := pool.GetReserveAddress()
	s.sendCoins(poolReserveAddr, s.addr(1), s.getBalances(poolReserveAddr))

	// The pool creator tries to withdraw his coins, but this will fail.
	req := s.withdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom))
	err := s.keeper.ExecuteWithdrawRequest(s.ctx, req)
	s.Require().NoError(err)
	req, _ = s.keeper.GetWithdrawRequest(s.ctx, pool.Id, req.Id)
	s.Require().Equal(types.RequestStatusFailed, req.Status)

	// Delete the previous request and refund coins to the withdrawer.
	liquidity.BeginBlocker(s.ctx, s.keeper)

	// Now any withdrawals will result in an error.
	_, err = s.keeper.Withdraw(s.ctx, types.NewMsgWithdraw(poolCreator, pool.Id, s.getBalance(poolCreator, pool.PoolCoinDenom)))
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

func (s *KeeperTestSuite) TestPoolOrderOverflow_ExternalFunds() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	externalFunds := utils.ParseCoins("10000000000000000000000000000000000000000000denom2")
	s.fundAddr(s.addr(1), externalFunds)
	s.sendCoins(s.addr(1), pool.GetReserveAddress(), externalFunds)
	s.sellLimitOrder(s.addr(2), pair.Id, utils.ParseDec("0.0000000001"), sdk.NewInt(1e18), 0, true)
	s.Require().NotPanics(func() {
		liquidity.EndBlocker(s.ctx, s.keeper)
	})
	liquidity.BeginBlocker(s.ctx, s.keeper)
}

func (s *KeeperTestSuite) TestRangedPoolDepositWithdraw() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createRangedPool(
		s.addr(1), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
		utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"), true)
	rx, ry := s.keeper.GetPoolBalances(s.ctx, pool)
	ammPool := pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{})
	s.Require().True(utils.DecApproxEqual(ammPool.Price(), utils.ParseDec("1.0")))

	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("400000denom1,1000000denom2"), true)
	liquidity.EndBlocker(s.ctx, s.keeper)
	rx, ry = s.keeper.GetPoolBalances(s.ctx, pool)
	ammPool = pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{})
	s.Require().True(utils.DecApproxEqual(ammPool.Price(), utils.ParseDec("1.0")))

	poolCoin := s.getBalance(s.addr(2), pool.PoolCoinDenom)
	s.withdraw(s.addr(2), pool.Id, poolCoin.SubAmount(poolCoin.Amount.QuoRaw(3))) // withdraw 2/3 pool coin
	liquidity.EndBlocker(s.ctx, s.keeper)
	rx, ry = s.keeper.GetPoolBalances(s.ctx, pool)
	ammPool = pool.AMMPool(rx.Amount, ry.Amount, sdk.Int{})
	s.Require().True(utils.DecApproxEqual(ammPool.Price(), utils.ParseDec("1.0")))
}

func (s *KeeperTestSuite) TestRangedPoolDepositWithdraw_single_side() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createRangedPool(
		s.addr(1), pair.Id, utils.ParseCoins("1000000denom1"),
		utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("0.5"), true)

	rx, ry := s.keeper.GetPoolBalances(s.ctx, pool)
	s.Require().True(intEq(sdk.ZeroInt(), rx.Amount))
	s.Require().True(intEq(sdk.NewInt(1000000), ry.Amount))
	ps := s.keeper.GetPoolCoinSupply(s.ctx, pool)

	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("50000denom1"), true)
	s.nextBlock()

	pc := s.getBalance(s.addr(2), pool.PoolCoinDenom)

	rx, ry = s.keeper.GetPoolBalances(s.ctx, pool)
	s.Require().True(intEq(sdk.ZeroInt(), rx.Amount))
	s.Require().True(intEq(sdk.NewInt(1050000), ry.Amount))
	s.Require().True(intEq(ps.QuoRaw(20), pc.Amount))

	balanceBefore := s.getBalance(s.addr(2), "denom1")
	s.withdraw(s.addr(2), pool.Id, sdk.NewCoin(pool.PoolCoinDenom, pc.Amount))
	s.nextBlock()
	balanceAfter := s.getBalance(s.addr(2), "denom1")

	s.Require().True(balanceAfter.Sub(balanceBefore).Amount.Sub(sdk.NewInt(50000)).LTE(sdk.OneInt()))

	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	s.Require().True(intEq(sdk.ZeroInt(), s.getBalance(s.addr(3), "denom1").Amount))
	s.Require().True(intEq(sdk.NewInt(1000000), s.getBalance(s.addr(3), "denom2").Amount))
}

func (s *KeeperTestSuite) TestRangedPoolDepositWithdraw_single_side2() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)
	pool := s.createRangedPool(
		s.addr(1), pair.Id, utils.ParseCoins("1000000denom2"),
		utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("2.0"), true)

	rx, ry := s.keeper.GetPoolBalances(s.ctx, pool)
	s.Require().True(intEq(sdk.NewInt(1000000), rx.Amount))
	s.Require().True(intEq(sdk.ZeroInt(), ry.Amount))
	ps := s.keeper.GetPoolCoinSupply(s.ctx, pool)

	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("50000denom2"), true)
	s.nextBlock()

	pc := s.getBalance(s.addr(2), pool.PoolCoinDenom)

	rx, ry = s.keeper.GetPoolBalances(s.ctx, pool)
	s.Require().True(intEq(sdk.NewInt(1050000), rx.Amount))
	s.Require().True(intEq(sdk.ZeroInt(), ry.Amount))
	s.Require().True(intEq(ps.QuoRaw(20), pc.Amount))

	balanceBefore := s.getBalance(s.addr(2), "denom2")
	s.withdraw(s.addr(2), pool.Id, sdk.NewCoin(pool.PoolCoinDenom, pc.Amount))
	s.nextBlock()
	balanceAfter := s.getBalance(s.addr(2), "denom2")

	s.Require().True(balanceAfter.Sub(balanceBefore).Amount.Sub(sdk.NewInt(50000)).LTE(sdk.OneInt()))

	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
	s.nextBlock()

	s.Require().True(intEq(sdk.ZeroInt(), s.getBalance(s.addr(3), "denom2").Amount))
	s.Require().True(intEq(sdk.NewInt(1000000), s.getBalance(s.addr(3), "denom1").Amount))
}

func (s *KeeperTestSuite) TestMaxNumActivePoolsPerPair() {
	s.Run("basic pool and ranged pools", func() {
		s.SetupTest()
		pair := s.createPair(s.addr(0), "denom1", "denom2", true)

		pool1 := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"), true)
		for i := 0; i < types.MaxNumActivePoolsPerPair-1; i++ {
			s.createRangedPool(
				s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
				utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"), true)
		}

		s.fundAddr(s.addr(0), utils.ParseCoins("1000000denom1,1000000denom2"))
		_, err := s.keeper.CreateRangedPool(s.ctx, types.NewMsgCreateRangedPool(
			s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
			utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.5")))
		s.Require().ErrorIs(err, types.ErrTooManyPools)

		s.withdraw(s.addr(0), pool1.Id, sdk.NewCoin(pool1.PoolCoinDenom, s.keeper.GetPoolCoinSupply(s.ctx, pool1)))
		s.nextBlock()

		s.createRangedPool(
			s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
			utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"), true)
	})

	s.Run("ranged pools only", func() {
		s.SetupTest()
		pair := s.createPair(s.addr(0), "denom1", "denom2", true)

		for i := 0; i < types.MaxNumActivePoolsPerPair; i++ {
			s.createRangedPool(
				s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
				utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"), true)
		}
		pool1, _ := s.keeper.GetPool(s.ctx, 1)

		s.fundAddr(s.addr(0), utils.ParseCoins("1000000denom1,1000000denom2"))
		_, err := s.keeper.CreateRangedPool(s.ctx, types.NewMsgCreateRangedPool(
			s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
			utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.5")))
		s.Require().ErrorIs(err, types.ErrTooManyPools)

		s.withdraw(s.addr(0), pool1.Id, sdk.NewCoin(pool1.PoolCoinDenom, s.keeper.GetPoolCoinSupply(s.ctx, pool1)))
		s.nextBlock()

		s.createRangedPool(
			s.addr(0), pair.Id, utils.ParseCoins("1000000denom1,1000000denom2"),
			utils.ParseDec("0.5"), utils.ParseDec("2.0"), utils.ParseDec("1.0"), true)
	})
}
