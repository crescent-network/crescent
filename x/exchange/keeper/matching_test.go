package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

// An example to test order sources.
func (s *KeeperTestSuite) TestOrderSourceMatching() {
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, utils.ParseDec("101"), sdk.NewInt(10_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")

	ordererAddr := s.FundedAccount(1, enoughCoins)
	_, _, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("100"), sdk.NewInt(5_000000), time.Hour)
	s.AssertEqual(sdk.NewInt(5_000000), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_000000ucre"), res.Paid)
	s.AssertEqual(utils.ParseCoin("503_485000uusd"), res.Received)
}
