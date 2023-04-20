package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	chain "github.com/crescent-network/crescent/v5/app"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/testutil"
)

func (s *KeeperTestSuite) TestPoolOrders() {
	creatorAddr := utils.TestAddress(1)
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, creatorAddr, utils.ParseCoins("100000_000000ucre,100000_000000uusd"))

	market, err := s.app.ExchangeKeeper.CreateSpotMarket(s.ctx, creatorAddr, "ucre", "uusd")
	s.Require().NoError(err)
	pool, err := s.k.CreatePool(s.ctx, creatorAddr, "ucre", "uusd", 500, sdk.NewDec(5))
	s.Require().NoError(err)

	_, _, _, _, err = s.k.AddLiquidity(
		s.ctx, creatorAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(1000_000000), sdk.NewInt(5000_000000), sdk.NewInt(100_000000), sdk.NewInt(500_000000))
	s.Require().NoError(err)

	ordererAddr := utils.TestAddress(2)
	_ = chain.FundAccount(s.app.BankKeeper, s.ctx, ordererAddr, utils.ParseCoins("1000000_000000ucre,1000000_000000uusd"))

	fmt.Println(s.app.ExchangeKeeper.TransientOrderBook(s.ctx, market.Id))

	balancesBefore := s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr)
	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		market.Id, ordererAddr, true, sdk.NewInt(110_000000))
	fmt.Println(s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr)
	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		market.Id, ordererAddr, false, sdk.NewInt(80_000000))
	fmt.Println(s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr)
	testutil.PlaceSpotMarketOrder(s.T(), s.ctx, s.app.ExchangeKeeper,
		market.Id, ordererAddr, false, sdk.NewInt(10_000000))
	fmt.Println(s.app.BankKeeper.SpendableCoins(s.ctx, ordererAddr).SafeSub(balancesBefore))
}
