package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidity/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidity/types"
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

	orderState, _ := s.keeper.GetOrderState(s.ctx, order.PairId, order.Id)
	oldOrderState := orderState
	orderState.RemainingOfferCoinAmount = sdk.NewInt(2000000)
	s.keeper.SetOrderState(s.ctx, order.PairId, order.Id, orderState)
	_, broken = keeper.RemainingOfferCoinEscrowInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)

	s.keeper.SetOrderState(s.ctx, order.PairId, order.Id, oldOrderState)
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

func (s *KeeperTestSuite) TestNumMMOrdersInvariant() {
	pair := s.createPair(s.addr(0), "denom1", "denom2", true)

	orderer := s.addr(1)
	// Place random MM orders
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionBuy,
		utils.ParseDec("0.99"), sdk.NewInt(1000000), time.Hour, true)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionBuy,
		utils.ParseDec("0.98"), sdk.NewInt(1000000), time.Hour, true)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionBuy,
		utils.ParseDec("0.97"), sdk.NewInt(1000000), time.Hour, true)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionSell,
		utils.ParseDec("1.01"), sdk.NewInt(1000000), time.Hour, true)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionSell,
		utils.ParseDec("1.02"), sdk.NewInt(1000000), time.Hour, true)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionSell,
		utils.ParseDec("1.03"), sdk.NewInt(1000000), time.Hour, true)

	_, broken := keeper.NumMMOrdersInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	s.nextBlock()

	// Cancel some MM orders and place another order
	s.cancelOrder(orderer, pair.Id, 1)
	s.cancelOrder(orderer, pair.Id, 2)
	s.mmOrder(
		orderer, pair.Id, types.OrderDirectionSell,
		utils.ParseDec("1.04"), sdk.NewInt(1000000), time.Hour, true)

	_, broken = keeper.NumMMOrdersInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	// After deleting canceled orders, the invariant must not be broken
	s.nextBlock()
	_, broken = keeper.NumMMOrdersInvariant(s.keeper)(s.ctx)
	s.Require().False(broken)

	// Break it
	s.keeper.SetNumMMOrders(s.ctx, orderer, pair.Id, 3)
	_, broken = keeper.NumMMOrdersInvariant(s.keeper)(s.ctx)
	s.Require().True(broken)
}
