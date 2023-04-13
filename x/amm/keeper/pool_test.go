package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/testutil"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestPoolOrders() {
	creatorAddr := utils.TestAddress(1)
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, creatorAddr, utils.ParseCoins("100000_000000ucre,100000_000000uusd"))

	market, err := s.app.ExchangeKeeper.CreateSpotMarket(s.ctx, creatorAddr, "ucre", "uusd")
	s.Require().NoError(err)

	pool, err := s.k.CreatePool(s.ctx, creatorAddr, "ucre", "uusd", 50, sdk.NewDec(10))
	s.Require().NoError(err)

	_, _, amt0, amt1, err := s.k.AddLiquidity(
		s.ctx, creatorAddr, pool.Id, utils.ParseDec("8"), utils.ParseDec("12.5"),
		sdk.NewInt(1000_000000), sdk.NewInt(10000_000000), sdk.NewInt(999_000000), sdk.NewInt(9999_000000))
	s.Require().NoError(err)
	fmt.Println("input", amt0, amt1)
	reserveAddr := sdk.MustAccAddressFromBech32(pool.ReserveAddress)
	fmt.Println("pool balances", s.app.BankKeeper.GetAllBalances(s.ctx, reserveAddr))

	printOrderBook := func() {
		s.app.ExchangeKeeper.IterateSpotOrderBook(s.ctx, market.Id, func(order types.SpotOrder) (stop bool) {
			if order.Price.GTE(utils.ParseDec("9.9")) && order.Price.LTE(utils.ParseDec("10.2")) {
				fmt.Println(order.IsBuy, order.Price, order.OpenQuantity)
			}
			return false
		})
	}

	fmt.Println("Initial")
	printOrderBook()

	ordererAddr := utils.TestAddress(2)
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, ordererAddr, utils.ParseCoins("1000_000000ucre,1000_000000uusd"))

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market.Id, false, sdk.NewInt(100_000000))
	fmt.Println("After sell")
	printOrderBook()

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market.Id, true, sdk.NewInt(100_000000))
	fmt.Println("After buy")
	printOrderBook()
}
