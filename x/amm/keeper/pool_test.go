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
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, creatorAddr, utils.ParseCoins("1000_000000ucre,1000_000000uusd"))

	market, err := s.app.ExchangeKeeper.CreateSpotMarket(s.ctx, creatorAddr, "ucre", "uusd")
	s.Require().NoError(err)

	pool, err := s.k.CreatePool(s.ctx, creatorAddr, "ucre", "uusd", 50, sdk.NewDec(10))
	s.Require().NoError(err)

	_, _, amt0, amt1, err := s.k.AddLiquidity(
		s.ctx, creatorAddr, pool.Id, utils.ParseDec("8"), utils.ParseDec("12.5"),
		sdk.NewInt(100_000000), sdk.NewInt(1000_000000), sdk.NewInt(99_000000), sdk.NewInt(999_000000))
	s.Require().NoError(err)
	fmt.Println("pool balances", amt0, amt1)

	printOrderBook := func() {
		s.app.ExchangeKeeper.IterateSpotOrderBook(s.ctx, market.Id, func(order types.SpotOrder) (stop bool) {
			if order.Price.GTE(utils.ParseDec("9.8")) && order.Price.LTE(utils.ParseDec("10.2")) {
				fmt.Println(order.IsBuy, order.Price, order.OpenQuantity)
			}
			return false
		})
	}
	fmt.Println("Initial:")
	printOrderBook()

	ordererAddr := utils.TestAddress(2)
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, ordererAddr, utils.ParseCoins("1000_000000ucre,1000_000000uusd"))

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market, false, sdk.NewInt(3_000000))

	fmt.Println("After sell:")
	printOrderBook()

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market, true, sdk.NewInt(1_000000))

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market, true, sdk.NewInt(2_000000))

	fmt.Println("After buy:")
	printOrderBook()
}
