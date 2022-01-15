package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestSwapBatch() {
	// TODO: Refactor this to test proper case. This is for a simple test.
	createMsg := &types.MsgCreatePool{
		Creator: s.addrs[0].String(),
		XCoin:   sdk.NewInt64Coin(denom1, 100000000),
		YCoin:   sdk.NewInt64Coin(denom2, 100000000),
	}
	_, err := s.keeper.CreatePool(s.ctx, createMsg)
	s.Require().NoError(err)

	_, found := s.keeper.GetPairByDenoms(s.ctx, denom1, denom2)
	s.Require().True(found)

	swapMsg := &types.MsgSwapBatch{
		Orderer:         s.addrs[0].String(),
		XCoinDenom:      denom1,
		YCoinDenom:      denom2,
		OfferCoin:       sdk.NewInt64Coin(denom2, 10000),
		DemandCoinDenom: denom1,
		Price:           sdk.MustNewDecFromStr("1.0"),
		OrderLifespan:   10 * time.Second,
	}
	_, err = s.keeper.SwapBatch(s.ctx, swapMsg)
	s.Require().NoError(err)

}
