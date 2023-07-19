package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestGetBestPrice() {
	market := s.CreateMarket("ucre", "uusd")

	_, found := s.keeper.GetBestPrice(s.Ctx, market.Id, true)
	s.Require().False(found)
	_, found = s.keeper.GetBestPrice(s.Ctx, market.Id, false)
	s.Require().False(found)

	ordererAddr := s.FundedAccount(1, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("0.99"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("0.98"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.01"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.02"), sdk.NewDec(10000), time.Hour)
	bestBuyPrice, found := s.keeper.GetBestPrice(s.Ctx, market.Id, true)
	s.Require().True(found)
	s.AssertEqual(utils.ParseDec("0.99"), bestBuyPrice)
	bestSellPrice, found := s.keeper.GetBestPrice(s.Ctx, market.Id, false)
	s.Require().True(found)
	s.AssertEqual(utils.ParseDec("1.01"), bestSellPrice)
}

func (s *KeeperTestSuite) TestIterateOrderBookSide() {
	market := s.CreateMarket("ucre", "uusd")
	ordererAddr := s.FundedAccount(1, enoughCoins)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("1.2"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("1.2"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("1.1"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, true, utils.ParseDec("1.0"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.3"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.3"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.4"), sdk.NewDec(10000), time.Hour)
	s.PlaceLimitOrder(
		market.Id, ordererAddr, false, utils.ParseDec("1.5"), sdk.NewDec(10000), time.Hour)

	type priceLevel struct {
		price    sdk.Dec
		orderIds []uint64
	}
	for _, tc := range []struct {
		name       string
		isBuy      bool
		priceLimit *sdk.Dec
		levels     []priceLevel
	}{
		{
			"buy side without price limit",
			true,
			nil,
			[]priceLevel{
				{utils.ParseDec("1.2"), []uint64{1, 2}},
				{utils.ParseDec("1.1"), []uint64{3}},
				{utils.ParseDec("1.0"), []uint64{4}},
			},
		},
		{
			"buy side with price limit 1",
			true,
			utils.ParseDecP("1.2"),
			[]priceLevel{
				{utils.ParseDec("1.2"), []uint64{1, 2}},
			},
		},
		{
			"buy side with price limit 2",
			true,
			utils.ParseDecP("1.1"),
			[]priceLevel{
				{utils.ParseDec("1.2"), []uint64{1, 2}},
				{utils.ParseDec("1.1"), []uint64{3}},
			},
		},
		{
			"buy side with price limit 3",
			true,
			utils.ParseDecP("1.0"),
			[]priceLevel{
				{utils.ParseDec("1.2"), []uint64{1, 2}},
				{utils.ParseDec("1.1"), []uint64{3}},
				{utils.ParseDec("1.0"), []uint64{4}},
			},
		},
		{
			"buy side with price limit 4",
			true,
			utils.ParseDecP("0.9"),
			[]priceLevel{
				{utils.ParseDec("1.2"), []uint64{1, 2}},
				{utils.ParseDec("1.1"), []uint64{3}},
				{utils.ParseDec("1.0"), []uint64{4}},
			},
		},
		{
			"buy side with price limit 5",
			true,
			utils.ParseDecP("1.3"),
			[]priceLevel{},
		},
		{
			"sell side without price limit",
			false,
			nil,
			[]priceLevel{
				{utils.ParseDec("1.3"), []uint64{5, 6}},
				{utils.ParseDec("1.4"), []uint64{7}},
				{utils.ParseDec("1.5"), []uint64{8}},
			},
		},
		{
			"sell side with price limit 1",
			false,
			utils.ParseDecP("1.3"),
			[]priceLevel{
				{utils.ParseDec("1.3"), []uint64{5, 6}},
			},
		},
		{
			"sell side with price limit 2",
			false,
			utils.ParseDecP("1.4"),
			[]priceLevel{
				{utils.ParseDec("1.3"), []uint64{5, 6}},
				{utils.ParseDec("1.4"), []uint64{7}},
			},
		},
		{
			"sell side with price limit 3",
			false,
			utils.ParseDecP("1.5"),
			[]priceLevel{
				{utils.ParseDec("1.3"), []uint64{5, 6}},
				{utils.ParseDec("1.4"), []uint64{7}},
				{utils.ParseDec("1.5"), []uint64{8}},
			},
		},
		{
			"sell side with price limit 4",
			false,
			utils.ParseDecP("1.6"),
			[]priceLevel{
				{utils.ParseDec("1.3"), []uint64{5, 6}},
				{utils.ParseDec("1.4"), []uint64{7}},
				{utils.ParseDec("1.5"), []uint64{8}},
			},
		},
		{
			"sell side with price limit 5",
			false,
			utils.ParseDecP("1.2"),
			[]priceLevel{},
		},
	} {
		s.Run(tc.name, func() {
			var levels []priceLevel
			s.keeper.IterateOrderBookSide(
				s.Ctx, market.Id, tc.isBuy, tc.priceLimit,
				func(price sdk.Dec, orders []types.Order) (stop bool) {
					var orderIds []uint64
					for _, order := range orders {
						orderIds = append(orderIds, order.Id)
					}
					levels = append(levels, priceLevel{price, orderIds})
					return false
				},
			)
			s.Require().Len(levels, len(tc.levels))
			for i, level := range tc.levels {
				s.AssertEqual(level.price, levels[i].price)
				s.Assert().Equal(level.orderIds, levels[i].orderIds)
			}
		})
	}
}
