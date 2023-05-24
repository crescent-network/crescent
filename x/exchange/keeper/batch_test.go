package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestBatch() {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	market.MakerFeeRate = sdk.ZeroDec()
	market.TakerFeeRate = sdk.ZeroDec()
	s.App.ExchangeKeeper.SetMarket(s.Ctx, market)
	// Manually set the last price to test batch matching.
	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	marketState.LastPrice = utils.ParseDecP("0.098")
	s.App.ExchangeKeeper.SetMarketState(s.Ctx, market.Id, marketState)

	aliceAddr := s.FundedAccount(1, utils.ParseCoins("100_000000ucre,100_000000uusd"))
	bobAddr := s.FundedAccount(2, utils.ParseCoins("100_000000ucre,100_000000uusd"))

	order1, err := s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, market.Id, aliceAddr, false, utils.ParseDec("0.1"), sdk.NewInt(10000), time.Hour)
	s.Require().NoError(err)
	order2, err := s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, market.Id, aliceAddr, false, utils.ParseDec("0.099"), sdk.NewInt(9995), time.Hour)
	s.Require().NoError(err)
	order3, err := s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, market.Id, bobAddr, true, utils.ParseDec("0.101"), sdk.NewInt(10000), time.Hour)
	s.Require().NoError(err)
	order4, err := s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, market.Id, bobAddr, true, utils.ParseDec("0.1"), sdk.NewInt(5000), time.Hour)
	s.Require().NoError(err)

	s.App.ExchangeKeeper.RunBatch(s.Ctx, market, []types.Order{order1, order2, order3, order4})

	fmt.Println(s.GetAllBalances(aliceAddr))
	fmt.Println(s.GetAllBalances(bobAddr))
}
