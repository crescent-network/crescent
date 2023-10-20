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
	s.AssertEqual(utils.ParseDec("101"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMaker() {
	// Order book looks like:
	//                 | 2.7200 |
	//                 | 2.7100 |
	// (os) ########## | 2.6080 | --> last price
	// (os)     ###### | 2.6060 |
	// (os)         ## | 2.5040 |
	//                 | 2.4020 |  market order
	//                 | 2.4010 |
	//                 | 2.4000 |
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.608"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.606"), sdk.NewInt(6_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.401"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(10_000001))

	s.AssertEqual(utils.ParseDec("2.6080"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewInt(10_000001), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("25_860003uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("9_970000ucre"), res.Received)
	expectedFee := sdk.Coins{utils.ParseCoin("15001ucre"), utils.ParseCoin("1uusd")}
	expectedOsBalancDiff := sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-9985001)},
		utils.ParseCoin("25_860002uusd")}

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(expectedFee, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
}

// With having the 10% price change limit, there could be no matching even though there are enough orders.
func (s *KeeperTestSuite) TestMatchingByMaxPriceLimit() {
	// Order book looks like:
	//                 | 2.7200 |
	//                 | 2.7100 |
	// (os) ########## | 2.6080 |
	// (os)     ###### | 2.6060 |
	// (os)         ## | 2.5040 |
	//                 | 2.4020 |  market order
	//                 | 2.1010 | --> last price
	//                 | 2.1000 |
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.608"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.606"), sdk.NewInt(6_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.101"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(10_000001))
	s.AssertEqual(sdk.NewInt(0), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("0uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("0ucre"), res.Received)
	expectedFee := sdk.Coins{}
	expectedOsBalancDiff := sdk.Coins{}

	s.AssertEqual(utils.ParseDec("2.1010"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(expectedFee, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)

	_, _, _, err := s.App.ExchangeKeeper.PlaceLimitOrder(s.Ctx, market.Id, ordererAddr, true, utils.ParseDec("2.5040"), sdk.NewInt(2_000001), time.Hour)
	s.Require().EqualError(err, "price is higher than the limit 2.311100000000000000: order price out of range")
}

func (s *KeeperTestSuite) TestMatchingViaMaxPriceLimit() {
	// Order book looks like:
	//                 | 2.7200 |
	//                 | 2.7100 |
	// (os) ########## | 2.7080 |
	// (os)     ###### | 2.6010 | --> last price
	// (os)         ## | 2.5040 |
	//                 | 2.4020 |  market order
	//                 | 2.4010 |
	//                 | 2.1000 |
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.708"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.601"), sdk.NewInt(6_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.401"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(10_000001))
	s.AssertEqual(sdk.NewInt(8_000000), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_614000uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("7_976000ucre"), res.Received)
	expectedFee := sdk.Coins{utils.ParseCoin("12000ucre")}
	expectedOsBalancDiff := sdk.Coins{sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-7988000)},
		utils.ParseCoin("20_614000uusd")}

	s.AssertEqual(utils.ParseDec("2.6010"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(expectedFee, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	s.AssertEqual(expectedOsBalancDiff, osBalanceDiff)
}
