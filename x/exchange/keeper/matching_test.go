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

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("15001ucre"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-9985001)},
		utils.ParseCoin("25_860003uusd")}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
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
	expectedOsBalancesDiff := sdk.Coins{}

	s.AssertEqual(utils.ParseDec("2.1010"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(expectedFee, feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)

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

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMakerNearMaxPrice() {
	// Order book looks like:
	// (os)     ###### | 10^24          |  --> new last price
	// (os)         ## | 10^24-10^20    |
	//                 | 10^24-2*10^20  | market order --> prev last price

	firstPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	secondPrice := types.MaxPrice
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, secondPrice, sdk.NewInt(6_000000_000000)),
		types.NewMockOrderSourceOrder(false, firstPrice, sdk.NewInt(2_000000_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	market.OrderQuoteLimits.Max = sdk.NewIntWithDecimal(1, 60)
	s.keeper.SetMarket(s.Ctx, market)
	mmAddr := s.FundedAccount(1, enoughCoins)
	lastPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20))).Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	s.MakeLastPrice(market.Id, mmAddr, lastPrice)

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(5_000000_000001))

	s.AssertEqual(types.MaxPrice, *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewInt(5_000000_000001), res.ExecutedQuantity)
	firstPaid := sdk.NewInt(2_000000_000000).ToDec().Mul(firstPrice).Ceil().TruncateInt()
	secondPaid := sdk.NewInt(3_000000_000001).ToDec().Mul(secondPrice).Ceil().TruncateInt()
	s.AssertEqual(firstPaid.Add(secondPaid), res.Paid.Amount)
	s.AssertEqual(utils.ParseCoin("4_985000_000000ucre"), res.Received)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("7500_000001ucre"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-4992500000001)},
		sdk.Coin{Denom: "uusd", Amount: firstPaid.Add(secondPaid)}}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMakerNearMaxPrice2() {
	// Order book looks like:
	// (os)     ###### | 10^24          |  --> new last price
	// (os)         ## | 10^24-10^20    |
	//                 | 10^24-2*10^20  | market order --> prev last price

	firstPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	secondPrice := types.MaxPrice
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, secondPrice, sdk.NewInt(6_000000_000000)),
		types.NewMockOrderSourceOrder(false, firstPrice, sdk.NewInt(2_000000_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	market.OrderQuoteLimits.Max = sdk.NewIntWithDecimal(1, 60)
	s.keeper.SetMarket(s.Ctx, market)
	mmAddr := s.FundedAccount(1, enoughCoins)
	lastPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20))).Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	s.MakeLastPrice(market.Id, mmAddr, lastPrice)

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewInt(10_000000_000000)) // sweep all the order source orders

	s.AssertEqual(types.MaxPrice, *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewInt(8_000000_000000), res.ExecutedQuantity)
	firstPaid := sdk.NewInt(2_000000_000000).ToDec().Mul(firstPrice).Ceil().TruncateInt()
	secondPaid := sdk.NewInt(6_000000_000000).ToDec().Mul(secondPrice).Ceil().TruncateInt()
	s.AssertEqual(firstPaid.Add(secondPaid), res.Paid.Amount)
	s.AssertEqual(utils.ParseCoin("7_976000_000000ucre"), res.Received)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("12000_000000ucre"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-7988000000000)},
		sdk.Coin{Denom: "uusd", Amount: firstPaid.Add(secondPaid)}}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMakerNearMinPrice() {
	// Order book looks like:
	//  market order #####.1    | 10^-14+2*10^-17  |                --> prev. last price
	//                          | 10^-14+10^-17    |  ## (os)
	//                          | 10^-14           |  ###### (os)   --> new last price

	firstPrice := types.MinPrice.Add(sdk.NewDecWithPrec(1, 17))
	secondPrice := types.MinPrice

	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, secondPrice, sdk.NewInt(6000_00000_00000_00000)),
		types.NewMockOrderSourceOrder(true, firstPrice, sdk.NewInt(2000_00000_00000_00000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	lastPrice := types.MinPrice.Add(sdk.NewDecWithPrec(1, 17)).Add(sdk.NewDecWithPrec(1, 17))

	// make last TestOrderSourceMatchingAsMakerNearMinPrice
	s.PlaceLimitOrder(market.Id, mmAddr, true, lastPrice, sdk.NewInt(1000_00000_00000_00000), time.Hour)
	s.PlaceLimitOrder(market.Id, mmAddr, false, lastPrice, sdk.NewInt(1000_00000_00000_00000), time.Hour)

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(5000_00000_00000_00001))

	s.AssertEqual(types.MinPrice, *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewInt(5000_00000_00000_00001), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5000_00000_00000_00001ucre"), res.Paid)
	s.AssertEqual(utils.ParseCoin("49869uusd"), res.Received)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("76uusd"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(5000_00000_00000_00001)},
		sdk.Coin{Denom: "uusd", Amount: sdk.NewInt(-49945)}}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMakerNearMinPrice2() {
	// Order book looks like:
	//  market order #######    | 10^-14+2*10^-17  |                --> prev. last price
	//                          | 10^-14+10^-17    |  ## (os)
	//                          | 10^-14           |  ###### (os)   --> new last price

	firstPrice := types.MinPrice.Add(sdk.NewDecWithPrec(1, 17))
	secondPrice := types.MinPrice

	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(true, secondPrice, sdk.NewInt(6000_00000_00000_00000)),
		types.NewMockOrderSourceOrder(true, firstPrice, sdk.NewInt(2000_00000_00000_00000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	lastPrice := types.MinPrice.Add(sdk.NewDecWithPrec(1, 17)).Add(sdk.NewDecWithPrec(1, 17))

	// make last TestOrderSourceMatchingAsMakerNearMinPrice
	s.PlaceLimitOrder(market.Id, mmAddr, true, lastPrice, sdk.NewInt(1000_00000_00000_00000), time.Hour)
	s.PlaceLimitOrder(market.Id, mmAddr, false, lastPrice, sdk.NewInt(1000_00000_00000_00000), time.Hour)

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, false, sdk.NewInt(9000_00000_00000_00000)) // sweep all the order source orders

	s.AssertEqual(types.MinPrice, *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewInt(8000_00000_00000_00000), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("8000_00000_00000_00000ucre"), res.Paid)
	s.AssertEqual(utils.ParseCoin("79779uusd"), res.Received)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("121uusd"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(8000_00000_00000_00000)},
		sdk.Coin{Denom: "uusd", Amount: sdk.NewInt(-79900)}}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsMakerNearMaxPriceWithMaxQty() {
	// Order book looks like:
	// (os) ########## | 10^24          |  --> new last price
	// (os)  ######### | 10^24-10^20    |
	//                 | 10^24-2*10^20  | market order --> prev last price

	firstPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	secondPrice := types.MaxPrice
	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, secondPrice, sdk.NewIntWithDecimal(1, 30)),
		types.NewMockOrderSourceOrder(false, firstPrice, sdk.NewIntWithDecimal(1, 29).Mul(sdk.NewInt(9))))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	market := s.CreateMarket("ucre", "uusd")
	market.OrderQuoteLimits.Max = sdk.NewIntWithDecimal(1, 60)
	s.keeper.SetMarket(s.Ctx, market)
	mmAddr := s.FundedAccount(1, enoughCoins)
	lastPrice := types.MaxPrice.Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20))).Sub(sdk.NewDecFromInt(sdk.NewIntWithDecimal(1, 20)))
	s.MakeLastPrice(market.Id, mmAddr, lastPrice)

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)
	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, res := s.PlaceMarketOrder(market.Id, ordererAddr, true, sdk.NewIntWithDecimal(1, 30))

	s.AssertEqual(types.MaxPrice, *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.NewIntWithDecimal(1, 30), res.ExecutedQuantity)
	firstPaid := sdk.NewIntWithDecimal(1, 29).Mul(sdk.NewInt(9)).ToDec().Mul(firstPrice).Ceil().TruncateInt()
	secondPaid := sdk.NewIntWithDecimal(1, 29).ToDec().Mul(secondPrice).Ceil().TruncateInt()
	s.AssertEqual(firstPaid.Add(secondPaid), res.Paid.Amount)
	s.AssertEqual(sdk.NewIntWithDecimal(1, 30).Sub(sdk.NewIntWithDecimal(1, 27).Mul(sdk.NewInt(3))), res.Received.Amount)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(sdk.NewIntWithDecimal(1, 26).Mul(sdk.NewInt(15)), feeAmount.AmountOf("ucre"))

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewIntWithDecimal(1, 30).Sub(sdk.NewIntWithDecimal(1, 26).Mul(sdk.NewInt(15))).Neg()},
		sdk.Coin{Denom: "uusd", Amount: firstPaid.Add(secondPaid)}}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsTakerAndMaker() {
	// Order book looks like:
	//                 | 2.7200 |
	//                 | 2.7100 |
	// (os) ########## | 2.6080 |
	// (os)     ###### | 2.6060 |     limit order  --> new last price: taker
	//                 | 2.5050 |--> prev. last price
	// (os)         ## | 2.5040 |
	//                 | 2.4010 |
	//                 | 2.4000 |

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.505"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, order, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("2.6060"), sdk.NewInt(7_999999), time.Hour)
	s.AssertEqual(utils.ParseDec("2.5050"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.ZeroInt(), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("0uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("0ucre"), res.Received)

	s.NextBlock()

	order, found := s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().True(found)

	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.608"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.606"), sdk.NewInt(6_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	ev := s.getEventOrderFilled(order.Id)

	s.AssertEqual(sdk.NewInt(7_999999), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_645998uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("7_975999ucre"), ev.Received)

	order, found = s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().False(found)

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(7_999999), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_645998uusd"), ev_os.Received)
	s.AssertEqual(utils.ParseCoin("7_991000ucre"), ev_os.Paid)

	s.AssertEqual(utils.ParseDec("2.6060"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("15001ucre"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-7991000)},
		utils.ParseCoin("20_645998uusd")}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsTakerAndMakerWithPriority() {
	// Order book looks like:
	//                                    | 2.7200 |
	//                                    | 2.7100 |
	//                    (os) ########## | 2.6080 |
	// limit order ###  + (os)     ###### | 2.6060 |     limit order  --> new last price: taker
	//                                    | 2.5050 |   --> prev. last price
	//                    (os)         ## | 2.5040 |
	//                                    | 2.4010 |
	//                                    | 2.4000 |

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.505"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, order, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("2.6060"), sdk.NewInt(7_999999), time.Hour)
	s.AssertEqual(utils.ParseDec("2.5050"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.ZeroInt(), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("0uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("0ucre"), res.Received)

	s.NextBlock()

	order, found := s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().True(found)

	ordererAddr2 := s.FundedAccount(3, enoughCoins)
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("2.6060"), sdk.NewInt(3_000000), time.Hour)
	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)

	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.608"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.606"), sdk.NewInt(6_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	ev := s.getEventOrderFilled(order.Id)

	s.AssertEqual(sdk.NewInt(7_999999), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_645998uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("7_975999ucre"), ev.Received)

	ev2 := s.getEventOrderFilled(order2.Id)
	s.Require().Nil(ev2)

	order, found = s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(3_000000), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(3_000000), order2.RemainingDeposit)

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(7_999999), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_645998uusd"), ev_os.Received)
	s.AssertEqual(utils.ParseCoin("7_991000ucre"), ev_os.Paid)

	s.AssertEqual(utils.ParseDec("2.6060"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("15001ucre"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-7991000)},
		utils.ParseCoin("20_645998uusd")}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}

func (s *KeeperTestSuite) TestOrderSourceMatchingAsTakerAndMakerWithSamePriority() {
	// Order book looks like:
	//                                      | 2.7200 |
	//                                      | 2.7100 |
	//                      (os) ########## | 2.6080 |
	// limit order ### + limit order ###### | 2.6060 |     limit order  --> new last price: taker
	//                                      | 2.5050 |   --> prev. last price
	//                      (os)         ## | 2.5040 |
	//                                      | 2.4010 |
	//                                      | 2.4000 |

	market := s.CreateMarket("ucre", "uusd")
	mmAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, mmAddr, utils.ParseDec("2.505"))

	feeCollector := market.MustGetFeeCollectorAddress()
	feeAmountBeforeMatching := s.GetAllBalances(feeCollector)

	ordererAddr := s.FundedAccount(2, enoughCoins)
	_, order, res := s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("2.6060"), sdk.NewInt(7_999999), time.Hour)
	s.AssertEqual(utils.ParseDec("2.5050"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	s.AssertEqual(sdk.ZeroInt(), res.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("0uusd"), res.Paid)
	s.AssertEqual(utils.ParseCoin("0ucre"), res.Received)

	s.NextBlock()

	order, found := s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().True(found)

	ordererAddr2 := s.FundedAccount(3, enoughCoins)
	order2 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr2, false, utils.ParseDec("2.6060"), sdk.NewInt(3_000000), time.Hour)
	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)

	ordererAddr3 := s.FundedAccount(4, enoughCoins)
	order3 := s.PlaceBatchLimitOrder(
		market.Id, ordererAddr3, false, utils.ParseDec("2.6060"), sdk.NewInt(6_000000), time.Hour)
	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)

	os := types.NewMockOrderSource(
		"mockOrderSource",
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.608"), sdk.NewInt(10_000000)),
		types.NewMockOrderSourceOrder(false, utils.ParseDec("2.504"), sdk.NewInt(2_000000)))
	s.FundAccount(os.Address, enoughCoins)
	s.App.ExchangeKeeper = *s.App.ExchangeKeeper.SetOrderSources(os)
	s.keeper = s.App.ExchangeKeeper

	osBalanceBeforeMatching := s.GetAllBalances(os.Address)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	ev := s.getEventOrderFilled(order.Id)

	s.AssertEqual(sdk.NewInt(7_999999), ev.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("20_645998uusd"), ev.Paid)
	s.AssertEqual(utils.ParseCoin("7_975999ucre"), ev.Received)

	ev2 := s.getEventOrderFilled(order2.Id)
	s.AssertEqual(sdk.NewInt(1_999999), ev2.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_204179uusd"), ev2.Received)
	s.AssertEqual(utils.ParseCoin("1_999999ucre"), ev2.Paid)

	ev3 := s.getEventOrderFilled(order3.Id)
	s.AssertEqual(sdk.NewInt(4_000000), ev3.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("10_408364uusd"), ev3.Received)
	s.AssertEqual(utils.ParseCoin("4_000000ucre"), ev3.Paid)

	order, found = s.keeper.GetOrder(s.Ctx, order.Id)
	s.Require().False(found)

	order2, found = s.keeper.GetOrder(s.Ctx, order2.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(1_000001), order2.OpenQuantity)
	s.AssertEqual(sdk.NewInt(1_000001), order2.RemainingDeposit)

	order3, found = s.keeper.GetOrder(s.Ctx, order3.Id)
	s.Require().True(found)
	s.AssertEqual(sdk.NewInt(2_000000), order3.OpenQuantity)
	s.AssertEqual(sdk.NewInt(2_000000), order3.RemainingDeposit)

	ev_os := s.getEventOrderSourceOrdersFilled("mockOrderSource")
	s.AssertEqual(sdk.NewInt(2_000000), ev_os.ExecutedQuantity)
	s.AssertEqual(utils.ParseCoin("5_010001uusd"), ev_os.Received)
	s.AssertEqual(utils.ParseCoin("2_000000ucre"), ev_os.Paid)

	s.AssertEqual(utils.ParseDec("2.6060"), *s.App.ExchangeKeeper.MustGetMarketState(s.Ctx, market.Id).LastPrice)

	feeAmountAfterMatching := s.GetAllBalances(feeCollector)
	feeAmount := feeAmountAfterMatching.Sub(feeAmountBeforeMatching)
	s.AssertEqual(utils.ParseCoins("24000ucre,23454uusd"), feeAmount)

	osBalanceAfterMatching := s.GetAllBalances(os.Address)
	osBalanceDiff, _ := osBalanceAfterMatching.SafeSub(osBalanceBeforeMatching)
	expectedOsBalancesDiff := sdk.Coins{
		sdk.Coin{Denom: "ucre", Amount: sdk.NewInt(-2000000)},
		utils.ParseCoin("5_010001uusd")}
	s.AssertEqual(expectedOsBalancesDiff, osBalanceDiff)
}
