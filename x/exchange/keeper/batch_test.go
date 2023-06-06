package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/exchange/types"
)

func (s *KeeperTestSuite) TestSimulateMidBlockBatchMatching() {
	market := s.CreateMarket(utils.TestAddress(0), "ucre", "uusd", true)
	marketState := s.keeper.MustGetMarketState(s.Ctx, market.Id)
	marketState.LastPrice = utils.ParseDecP("5")
	s.keeper.SetMarketState(s.Ctx, market.Id, marketState)

	ordererAddr1 := s.FundedAccount(1, enoughCoins)
	ordererAddr2 := s.FundedAccount(2, enoughCoins)
	ordererAddr3 := s.FundedAccount(3, enoughCoins)

	var msgs []sdk.Msg
	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr1, market.Id, true, utils.ParseDec("4.99"), sdk.NewInt(50_000000), 0))
	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr1, market.Id, true, utils.ParseDec("4.98"), sdk.NewInt(100_000000), 0))

	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr2, market.Id, true, utils.ParseDec("5.01"), sdk.NewInt(100_000000), 0))
	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr2, market.Id, true, utils.ParseDec("5.02"), sdk.NewInt(50_000000), 0))

	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr3, market.Id, false, utils.ParseDec("5.01"), sdk.NewInt(100_000000), 0))
	msgs = append(msgs, types.NewMsgPlaceBatchLimitOrder(
		ordererAddr3, market.Id, false, utils.ParseDec("5.02"), sdk.NewInt(100_000000), 0))

	// MidBlock
	midBlockCtx := s.Ctx.WithContext(utils.MidBlockContext(s.Ctx.Context()))
	for _, msg := range msgs {
		switch msg.(type) {
		case *types.MsgPlaceBatchLimitOrder, *types.MsgPlaceMMBatchLimitOrder:
			handler := s.App.MsgServiceRouter().Handler(msg)
			if handler != nil {
				_, err := handler(midBlockCtx, msg)
				s.Require().NoError(err)
			}
		}
	}
	s.Require().NoError(s.keeper.RunBatchMatching(s.Ctx, market))

	// DeliverTx
	for _, msg := range msgs {
		handler := s.App.MsgServiceRouter().Handler(msg)
		if handler != nil {
			_, err := handler(s.Ctx, msg)
			s.Require().NoError(err)
		}
	}

	s.NextBlock() // cancel and refund expired orders
}
