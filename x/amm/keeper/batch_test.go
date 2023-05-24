package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestBatch() {
	market, pool := s.CreateSampleMarketAndPool()
	marketState := s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id)
	marketState.LastPrice = utils.ParseDecP("5")
	s.App.ExchangeKeeper.SetMarketState(s.Ctx, market.Id, marketState)

	creatorAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	s.AddLiquidity(
		creatorAddr, creatorAddr, pool.Id, utils.ParseDec("4.8"), utils.ParseDec("5.2"),
		utils.ParseCoins("1000_000000ucre,5000_000000uusd"))

	ordererAddr := s.FundedAccount(2, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	order, err := s.App.ExchangeKeeper.PlaceBatchLimitOrder(s.Ctx, 1, ordererAddr, true, utils.ParseDec("5.05"), sdk.NewInt(10_000000))
	s.Require().NoError(err)

	fmt.Println(s.GetAllBalances(ordererAddr))
	s.App.ExchangeKeeper.RunBatch(s.Ctx, market, []exchangetypes.Order{order})
	fmt.Println(s.GetAllBalances(ordererAddr))
}
