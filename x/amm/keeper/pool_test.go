package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
)

func (s *KeeperTestSuite) TestPoolOrders() {
	creatorAddr := utils.TestAddress(1)
	s.FundAccount(creatorAddr, utils.ParseCoins("100000_000000ucre,100000_000000uusd"))

	market := s.CreateSpotMarket(creatorAddr, "ucre", "uusd", true)
	pool := s.CreatePool(creatorAddr, "ucre", "uusd", 500, sdk.NewDec(5), true)

	s.AddLiquidity(
		creatorAddr, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		sdk.NewInt(1000_000000), sdk.NewInt(5000_000000), sdk.NewInt(100_000000), sdk.NewInt(500_000000))

	ordererAddr := utils.TestAddress(2)
	s.FundAccount(ordererAddr, utils.ParseCoins("1000000_000000ucre,1000000_000000uusd"))

	balancesBefore := s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceSpotMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(110_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceSpotMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(80_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))

	balancesBefore = s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr)
	s.PlaceSpotMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(10_000000))
	fmt.Println(s.App.BankKeeper.SpendableCoins(s.Ctx, ordererAddr).SafeSub(balancesBefore))
}
