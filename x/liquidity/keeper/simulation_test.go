package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/liquidity/types"
)

func (s *KeeperTestSuite) TestSimulation1() {
	r := rand.New(rand.NewSource(0))

	pair := s.createPair(s.addr(10000), "denom1", "denom2", true)

	const numUsers = 100
	for i := 0; i < numUsers; i++ {
		s.fundAddr(s.addr(i), utils.ParseCoins("1000000000000000denom1,1000000000000000denom2"))
	}

	dustCollector := s.keeper.GetDustCollector(s.ctx)

	const numBlocks, numOrders, numDeposits, numWithdraws = 10, 10, 2, 2
	fuzz := func() {
		for i := 0; i < numBlocks; i++ {
			pair, _ = s.keeper.GetPair(s.ctx, pair.Id)
			pools := s.keeper.GetAllPools(s.ctx)

			totalBalancesBefore := sdk.Coins{}
			for j := 0; j < numUsers; j++ {
				totalBalancesBefore = totalBalancesBefore.Add(s.getBalances(s.addr(j))...)
			}
			for _, pool := range pools {
				totalBalancesBefore = totalBalancesBefore.Add(s.getBalances(pool.GetReserveAddress())...)
				totalBalancesBefore = totalBalancesBefore.Sub(
					sdk.NewCoins(sdk.NewCoin(pool.PoolCoinDenom, totalBalancesBefore.AmountOf(pool.PoolCoinDenom))))
			}
			totalBalancesBefore = totalBalancesBefore.Add(s.getBalances(dustCollector)...)

			for j := 0; j < numOrders; j++ {
				orderer := s.addr(r.Intn(numUsers))
				var price sdk.Dec
				if pair.LastPrice == nil {
					price = utils.RandomDec(r, utils.ParseDec("0.1"), utils.ParseDec("10.0"))
				} else {
					price = utils.RandomDec(
						r,
						pair.LastPrice.Mul(utils.ParseDec("0.901")),
						pair.LastPrice.Mul(utils.ParseDec("1.099")))
				}
				if r.Intn(2) == 0 { // 50% chance
					// Buy
					amt := utils.RandomInt(
						r,
						sdk.NewInt(1000),
						s.getBalance(orderer, "denom2").Amount.ToDec().QuoTruncate(price).TruncateInt())
					s.buyLimitOrder(orderer, pair.Id, price, amt, 0, false)
				} else {
					// Sell
					amt := utils.RandomInt(
						r,
						sdk.NewInt(1000),
						s.getBalance(orderer, "denom1").Amount)
					s.sellLimitOrder(orderer, pair.Id, price, amt, 0, false)
				}
			}
			for j := 0; j < numDeposits; j++ {
				depositor := s.addr(r.Intn(numUsers))
				pool := pools[r.Intn(len(pools))]
				balances := s.getBalances(depositor)
				msg := types.NewMsgDeposit(depositor, pool.Id, simtypes.RandSubsetCoins(r, balances))
				_, _ = s.keeper.Deposit(s.ctx, msg)
			}
			for j := 0; j < numWithdraws; j++ {
				withdrawer := s.addr(r.Intn(numUsers))
				pool := pools[r.Intn(len(pools))]
				balance := s.getBalance(withdrawer, pool.PoolCoinDenom).Amount
				msg := types.NewMsgWithdraw(
					withdrawer, pool.Id, sdk.NewCoin(pool.PoolCoinDenom, utils.RandomInt(r, sdk.NewInt(1), balance)))
				_, _ = s.keeper.Withdraw(s.ctx, msg)
			}
			s.nextBlock()

			totalBalancesAfter := sdk.Coins{}
			for j := 0; j < numUsers; j++ {
				totalBalancesAfter = totalBalancesAfter.Add(s.getBalances(s.addr(j))...)
			}
			for _, pool := range pools {
				totalBalancesAfter = totalBalancesAfter.Add(s.getBalances(pool.GetReserveAddress())...)
				totalBalancesAfter = totalBalancesAfter.Sub(
					sdk.NewCoins(sdk.NewCoin(pool.PoolCoinDenom, totalBalancesAfter.AmountOf(pool.PoolCoinDenom))))
			}
			totalBalancesAfter = totalBalancesAfter.Add(s.getBalances(dustCollector)...)

			s.Require().True(coinsEq(sdk.Coins{}, s.getBalances(pair.GetEscrowAddress())))
			s.Require().True(coinsEq(totalBalancesBefore, totalBalancesAfter))
		}
	}

	// Add a basic pool with price 1.0.
	s.createPool(s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"), true)
	fuzz()

	// Add a ranged pool with price in range [0.9, 1.1].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		utils.ParseDec("0.9"), utils.ParseDec("1.1"), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [0.8, 1.2].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		utils.ParseDec("0.8"), utils.ParseDec("1.2"), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [0.95, 1.05].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		utils.ParseDec("0.95"), utils.ParseDec("1.05"), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [0.99, 1.01].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		utils.ParseDec("0.99"), utils.ParseDec("1.01"), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [0.999, 1.001].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		utils.ParseDec("0.999"), utils.ParseDec("1.001"), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [10^-14, 2-(10^-4)].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		sdk.NewDecWithPrec(1, 14), sdk.NewDec(2).Sub(sdk.NewDecWithPrec(1, 4)), utils.ParseDec("1.0"), true)
	fuzz()

	// Add a ranged pool with price in range [10^-8, 10^20].
	s.createRangedPool(
		s.addr(10000), pair.Id, utils.ParseCoins("1000000000denom1,1000000000denom2"),
		sdk.NewDecWithPrec(1, 8), sdk.NewIntWithDecimal(1, 20).ToDec(), utils.ParseDec("1.0"), true)
	fuzz()
}
