package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidamm/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidamm/types"
)

func (s *KeeperTestSuite) TestCreateDuplicatePublicPosition() {
	market := s.CreateMarket("ucre", "uusd")
	pool := s.CreatePool(market.Id, utils.ParseDec("5"))
	s.CreatePublicPosition(
		pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"), utils.ParseDec("0.003"))

	_, err := s.keeper.CreatePublicPosition(
		s.Ctx, pool.Id, utils.ParseDec("4.5"), utils.ParseDec("5.5"),
		utils.ParseDec("0.001"))
	s.Require().EqualError(err, "public position with same parameters already exists")
}

func (s *KeeperTestSuite) TestMintShare() {
	publicPosition := s.CreateSamplePublicPosition()
	minterAddr := s.FundedAccount(1, utils.ParseCoins("10000_000000ucre,10000_000000uusd"))
	for _, tc := range []struct {
		name        string
		msg         *types.MsgMintShare
		postRun     func(resp *types.MsgMintShareResponse)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgMintShare(
				minterAddr,
				publicPosition.Id,
				utils.ParseCoins("100_000000ucre,500_000000uusd"),
			),
			func(resp *types.MsgMintShareResponse) {
				s.Require().Equal(utils.ParseCoin("4357388321sb1"), resp.MintedShare)
				s.Require().Equal(sdk.NewInt(4357388321), resp.Liquidity)
				s.Require().Equal(utils.ParseCoins("90686676ucre,500000000uusd"), resp.Amount)
			},
			"",
		},
		{
			"public position not found",
			types.NewMsgMintShare(minterAddr, 2, utils.ParseCoins("100_000000ucre,500_000000uusd")),
			nil,
			"public position not found: not found",
		},
		{
			"invalid desired amount",
			types.NewMsgMintShare(minterAddr, 1, utils.ParseCoins("100_000000uatom")),
			nil,
			"pool doesn't have denom uatom: invalid request",
		},
		{
			"invalid desired amount 2",
			types.NewMsgMintShare(minterAddr, 1, utils.ParseCoins("100_000000ucre")),
			nil,
			"added liquidity is zero",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.Ctx.CacheContext()
			resp, err := keeper.NewMsgServerImpl(s.keeper).MintShare(sdk.WrapSDKContext(cacheCtx), tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPublicPositionShare() {
	publicPosition := s.CreateSamplePublicPosition()

	senderAddr := utils.TestAddress(1)
	mintedShare, _, liquidity, origAmt := s.MintShare(
		senderAddr, publicPosition.Id, utils.ParseCoins("10_0000000ucre,50_000000uusd"), true)
	s.Require().Equal("9068668ucre,50000000uusd", origAmt.String())

	s.Require().Equal(mintedShare.Amount, liquidity)

	burnedLiquidity, _, amt := s.BurnShare(senderAddr, publicPosition.Id, mintedShare)
	s.Require().Equal("9068668ucre,50000000uusd", amt.String())
	s.Require().Equal(liquidity, burnedLiquidity)
}

func (s *KeeperTestSuite) TestBurnShare() {
	publicPosition := s.CreateSamplePublicPosition()
	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	for _, tc := range []struct {
		name        string
		msg         *types.MsgBurnShare
		postRun     func(resp *types.MsgBurnShareResponse)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgBurnShare(minterAddr, publicPosition.Id, utils.ParseCoin("100000sb1")),
			func(resp *types.MsgBurnShareResponse) {
				s.Require().Equal(sdk.NewInt(100000), resp.RemovedLiquidity)
				s.Require().Equal(utils.ParseCoins("2081ucre,11474uusd"), resp.Amount)
			},
			"",
		},
		{
			"public position not found",
			types.NewMsgBurnShare(
				minterAddr,
				2,
				utils.ParseCoin("100000sb2"),
			),
			nil,
			"public position not found: not found",
		},
	} {
		s.Run(tc.name, func() {
			s.Require().NoError(tc.msg.ValidateBasic())
			cacheCtx, _ := s.Ctx.CacheContext()
			resp, err := keeper.NewMsgServerImpl(s.keeper).BurnShare(sdk.WrapSDKContext(cacheCtx), tc.msg)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBurnShare_Complex_WithRewards() {
	publicPosition := s.CreateSamplePublicPosition()

	minterAddr1 := utils.TestAddress(1)
	minterAddr2 := utils.TestAddress(2)
	s.MintShare(minterAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(minterAddr2, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	s.MintShare(minterAddr2, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	// Ensure that the minters have received the minted share
	s.Require().Equal("4357388321sb1", s.GetBalance(minterAddr1, "sb1").String())
	s.Require().Equal("8714776642sb1", s.GetBalance(minterAddr2, "sb1").String())

	s.AdvanceRewardsAuctions()

	// Ensure rewards auction is created
	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.Require().True(found)

	bidderAddr1 := utils.TestAddress(3)
	bidderAddr2 := utils.TestAddress(4)
	s.MintShare(bidderAddr1, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr2, publicPosition.Id, utils.ParseCoins("200_000000ucre,1000_000000uusd"), true)
	s.PlaceBid(bidderAddr1, publicPosition.Id, auction.Id, utils.ParseCoin("100000sb1"))
	s.PlaceBid(bidderAddr2, publicPosition.Id, auction.Id, utils.ParseCoin("200000sb1"))
	s.NextBlock()

	s.AdvanceRewardsAuctions()

	// Ensure compounding rewards are set in the store
	publicPosition, _ = s.keeper.GetPublicPosition(s.Ctx, publicPosition.Id)
	auction, found = s.keeper.GetPreviousRewardsAuction(s.Ctx, publicPosition)
	s.Require().True(found)
	s.Require().Equal("200000sb1", auction.WinningBid.Share.String())

	// Ensure bidderAddr2 has received rewards
	s.Require().True(s.GetBalance(bidderAddr2, "uatom").Amount.GT(sdk.NewInt(1)))

	// Ensure the next rewards auction is created
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 2)

	bidderAddr3 := utils.TestAddress(3)
	bidderAddr4 := utils.TestAddress(4)
	auction, _ = s.keeper.GetLastRewardsAuction(s.Ctx, publicPosition.Id)
	s.MintShare(bidderAddr3, publicPosition.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr4, publicPosition.Id, utils.ParseCoins("300_000000ucre,1500_000000uusd"), true)
	s.PlaceBid(bidderAddr3, publicPosition.Id, auction.Id, utils.ParseCoin("100000sb1"))
	s.PlaceBid(bidderAddr4, publicPosition.Id, auction.Id, utils.ParseCoin("300000sb1"))

	s.AdvanceRewardsAuctions()

	// Ensure compounding rewards are updated with the new bidding amount in the store
	publicPosition, _ = s.keeper.GetPublicPosition(s.Ctx, publicPosition.Id)
	auction, found = s.keeper.GetPreviousRewardsAuction(s.Ctx, publicPosition)
	s.Require().True(found)
	s.Require().Equal("300000sb1", auction.WinningBid.Share.String())

	// Ensure bidderAddr4 has received farming rewards
	s.Require().True(s.GetBalance(bidderAddr4, "uatom").Amount.GT(sdk.NewInt(1)))

	// Burn all shares
	s.BurnShare(minterAddr1, publicPosition.Id, s.GetBalance(minterAddr1, "sb1"))
	s.BurnShare(minterAddr2, publicPosition.Id, s.GetBalance(minterAddr2, "sb1"))
}
