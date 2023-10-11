package keeper_test

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/app/testutil"
	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/keeper"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

var enoughCoins = sdk.NewCoins(
	sdk.NewCoin("ucre", sdk.NewIntWithDecimal(1, 60)),
	sdk.NewCoin("uatom", sdk.NewIntWithDecimal(1, 60)),
	sdk.NewCoin("uusd", sdk.NewIntWithDecimal(1, 60)),
	sdk.NewCoin("stake", sdk.NewIntWithDecimal(1, 60)))

type KeeperTestSuite struct {
	testutil.TestSuite
	keeper  keeper.Keeper
	querier keeper.Querier
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.TestSuite.SetupTest()
	s.keeper = s.App.ExchangeKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.FundAccount(utils.TestAddress(0), utils.ParseCoins("1ucre,1uusd,1uatom")) // make positive supplies
}

func (s *KeeperTestSuite) TestSetOrderSources() {
	// Same source name
	s.Require().PanicsWithValue("duplicate order source name: a", func() {
		k := keeper.Keeper{}
		k.SetOrderSources(types.NewMockOrderSource("a"), types.NewMockOrderSource("a"))
	})
	k := keeper.Keeper{}
	k.SetOrderSources(types.NewMockOrderSource("a"), types.NewMockOrderSource("b"))
	// Already set
	s.Require().PanicsWithValue("cannot set order sources twice", func() {
		s.keeper.SetOrderSources(types.NewMockOrderSource("b"), types.NewMockOrderSource("c"))
	})
}

func (s *KeeperTestSuite) createLiquidity(
	marketId uint64, ordererAddr sdk.AccAddress, centerPrice sdk.Dec, totalQty sdk.Int) {
	tick := types.TickAtPrice(centerPrice)
	interval := types.PriceIntervalAtTick(tick + 10*10)
	for i := 0; i < 10; i++ {
		sellPrice := centerPrice.Add(interval.MulInt64(int64(i+1) * 10))
		buyPrice := centerPrice.Sub(interval.MulInt64(int64(i+1) * 10))

		qty := totalQty.QuoRaw(200).Add(totalQty.QuoRaw(100).MulRaw(int64(i)))
		s.PlaceLimitOrder(marketId, ordererAddr, false, sellPrice, qty, time.Hour)
		s.PlaceLimitOrder(marketId, ordererAddr, true, buyPrice, qty, time.Hour)
	}
}

func (s *KeeperTestSuite) createLiquidity2(
	marketId uint64, ordererAddr sdk.AccAddress, centerPrice, maxOrderPriceRatio sdk.Dec, qtyPerTick sdk.Int) {
	minPrice, maxPrice := types.OrderPriceLimit(centerPrice, maxOrderPriceRatio)
	for i := 1; ; i++ {
		buyPrice := types.PriceAtTick(types.TickAtPrice(centerPrice) - 100*int32(i))
		if buyPrice.LT(minPrice) {
			break
		}
		s.PlaceLimitOrder(
			marketId, ordererAddr, true, buyPrice, qtyPerTick, time.Hour)
	}
	for i := 1; ; i++ {
		sellPrice := types.PriceAtTick(types.TickAtPrice(centerPrice) + 100*int32(i))
		if sellPrice.GT(maxPrice) {
			break
		}
		s.PlaceLimitOrder(
			marketId, ordererAddr, false, sellPrice, qtyPerTick, time.Hour)
	}
}

func (s *KeeperTestSuite) getEventOrderFilled(orderId uint64) (event *types.EventOrderFilled) {
	orderFilledEventType := proto.MessageName(event)
	for _, abciEvent := range s.Ctx.EventManager().ABCIEvents() {
		if abciEvent.Type == orderFilledEventType {
			e, err := sdk.ParseTypedEvent(abciEvent)
			s.Require().NoError(err)
			event = e.(*types.EventOrderFilled)
			if event.OrderId == orderId {
				return event
			}
		}
	}
	return nil
}

func (s *KeeperTestSuite) getEventOrderSourceOrdersFilled(sourceName string) (event *types.EventOrderSourceOrdersFilled) {
	orderSourceOrdersFilledEventType := proto.MessageName(event)
	for _, abciEvent := range s.Ctx.EventManager().ABCIEvents() {
		if abciEvent.Type == orderSourceOrdersFilledEventType {
			e, err := sdk.ParseTypedEvent(abciEvent)
			s.Require().NoError(err)
			event = e.(*types.EventOrderSourceOrdersFilled)
			if event.SourceName == sourceName {
				return event
			}
		}
	}
	return nil
}
