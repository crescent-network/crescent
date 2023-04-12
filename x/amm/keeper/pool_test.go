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

	pool, err := s.k.CreatePool(s.ctx, creatorAddr, "ucre", "uusd", 100, sdk.NewDec(1))
	s.Require().NoError(err)

	_, _, _, _, err = s.k.AddLiquidity(
		s.ctx, creatorAddr, pool.Id, utils.ParseDec("0.8"), utils.ParseDec("1.25"),
		sdk.NewInt(100_000000), sdk.NewInt(100_000000), sdk.NewInt(100_000000), sdk.NewInt(100_000000))
	s.Require().NoError(err)

	printOrderBook := func() {
		s.app.ExchangeKeeper.IterateSpotOrderBook(s.ctx, market.Id, func(order types.SpotOrder) (stop bool) {
			if order.Price.GTE(utils.ParseDec("0.98")) && order.Price.LTE(utils.ParseDec("1.03")) {
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
		ordererAddr, market, false, sdk.NewInt(30_000000))

	fmt.Println("After sell:")
	printOrderBook()

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market, true, sdk.NewInt(10_000000))

	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		ordererAddr, market, true, sdk.NewInt(20_000000))

	fmt.Println("After buy:")
	printOrderBook()
}
