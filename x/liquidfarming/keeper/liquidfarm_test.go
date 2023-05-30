package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/keeper"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
)

func (s *KeeperTestSuite) TestMintShare() {
	liquidFarm := s.CreateSampleLiquidFarm()
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
				liquidFarm.Id,
				utils.ParseCoins("100_000000ucre,500_000000uusd"),
			),
			func(resp *types.MsgMintShareResponse) {
				s.Require().Equal(utils.ParseCoin("4357388321lfshare1"), resp.MintedShare)
				s.Require().Equal(sdk.NewInt(4357388321), resp.Liquidity)
				s.Require().Equal(utils.ParseCoins("90686676ucre,500000000uusd"), resp.Amount)
			},
			"",
		},
		{
			"liquid farm not found",
			types.NewMsgMintShare(minterAddr, 2, utils.ParseCoins("100_000000ucre,500_000000uusd")),
			nil,
			"liquid farm not found: not found",
		},
		{
			"invalid desired amount",
			types.NewMsgMintShare(minterAddr, 1, utils.ParseCoins("100_000000uatom")),
			nil,
			"pool has no uatom in its reserve: invalid request",
		},
		{
			"invalid desired amount 2",
			types.NewMsgMintShare(minterAddr, 1, utils.ParseCoins("100_000000ucre")),
			nil,
			"minted liquidity is zero: insufficient funds",
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

func (s *KeeperTestSuite) TestLiquidFarmShare() {
	liquidFarm := s.CreateSampleLiquidFarm()

	senderAddr := utils.TestAddress(1)
	mintedShare, _, liquidity, origAmt := s.MintShare(
		senderAddr, liquidFarm.Id, utils.ParseCoins("10_0000000ucre,50_000000uusd"), true)
	s.Require().Equal("9068668ucre,50000000uusd", origAmt.String())

	s.Require().Equal(mintedShare.Amount, liquidity)

	burnedLiquidity, _, amt := s.BurnShare(senderAddr, liquidFarm.Id, mintedShare)
	s.Require().Equal("9068667ucre,49999999uusd", amt.String()) // Decimal loss
	s.Require().Equal(liquidity, burnedLiquidity)
}

func (s *KeeperTestSuite) TestBurnShare() {
	liquidFarm := s.CreateSampleLiquidFarm()
	minterAddr := utils.TestAddress(1)
	s.MintShare(minterAddr, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	for _, tc := range []struct {
		name        string
		msg         *types.MsgBurnShare
		postRun     func(resp *types.MsgBurnShareResponse)
		expectedErr string
	}{
		{
			"happy case",
			types.NewMsgBurnShare(minterAddr, liquidFarm.Id, utils.ParseCoin("100000lfshare1")),
			func(resp *types.MsgBurnShareResponse) {
				s.Require().Equal(sdk.NewInt(100000), resp.RemovedLiquidity)
				s.Require().Equal(utils.ParseCoins("2081ucre,11474uusd"), resp.Amount)
			},
			"",
		},
		{
			"liquid farm not found",
			types.NewMsgBurnShare(
				minterAddr,
				2,
				utils.ParseCoin("100000lfshare2"),
			),
			nil,
			"liquid farm not found: not found",
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

func (s *KeeperTestSuite) TestLiquidUnfarm_Complex_WithRewards() {
	liquidFarm := s.CreateSampleLiquidFarm()

	minterAddr1 := utils.TestAddress(1)
	minterAddr2 := utils.TestAddress(2)
	s.MintShare(minterAddr1, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(minterAddr2, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	s.MintShare(minterAddr2, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.NextBlock()

	// Ensure that the minters have received the minted share
	s.Require().Equal("4357388321lfshare1", s.GetBalance(minterAddr1, "lfshare1").String())
	s.Require().Equal("8714776642lfshare1", s.GetBalance(minterAddr2, "lfshare1").String())

	s.AdvanceRewardsAuctions()

	// Ensure rewards auction is created
	auction, found := s.keeper.GetLastRewardsAuction(s.Ctx, liquidFarm.Id)
	s.Require().True(found)

	bidderAddr1 := utils.TestAddress(3)
	bidderAddr2 := utils.TestAddress(4)
	s.MintShare(bidderAddr1, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr2, liquidFarm.Id, utils.ParseCoins("200_000000ucre,1000_000000uusd"), true)
	s.PlaceBid(bidderAddr1, liquidFarm.Id, auction.Id, utils.ParseCoin("100000lfshare1"))
	s.PlaceBid(bidderAddr2, liquidFarm.Id, auction.Id, utils.ParseCoin("200000lfshare1"))
	s.NextBlock()

	s.AdvanceRewardsAuctions()

	// Ensure compounding rewards are set in the store
	liquidFarm, _ = s.keeper.GetLiquidFarm(s.Ctx, liquidFarm.Id)
	auction, found = s.keeper.GetPreviousRewardsAuction(s.Ctx, liquidFarm)
	s.Require().True(found)
	s.Require().Equal("200000lfshare1", auction.WinningBid.Share.String())

	// Ensure bidderAddr2 has received rewards
	s.Require().True(s.GetBalance(bidderAddr2, "uatom").Amount.GT(sdk.NewInt(1)))

	// Ensure the next rewards auction is created
	s.Require().Len(s.keeper.GetAllRewardsAuctions(s.Ctx), 2)

	bidderAddr3 := utils.TestAddress(3)
	bidderAddr4 := utils.TestAddress(4)
	auction, _ = s.keeper.GetLastRewardsAuction(s.Ctx, liquidFarm.Id)
	s.MintShare(bidderAddr3, liquidFarm.Id, utils.ParseCoins("100_000000ucre,500_000000uusd"), true)
	s.MintShare(bidderAddr4, liquidFarm.Id, utils.ParseCoins("300_000000ucre,1500_000000uusd"), true)
	s.PlaceBid(bidderAddr3, liquidFarm.Id, auction.Id, utils.ParseCoin("100000lfshare1"))
	s.PlaceBid(bidderAddr4, liquidFarm.Id, auction.Id, utils.ParseCoin("300000lfshare1"))

	s.AdvanceRewardsAuctions()

	// Ensure compounding rewards are updated with the new bidding amount in the store
	liquidFarm, _ = s.keeper.GetLiquidFarm(s.Ctx, liquidFarm.Id)
	auction, found = s.keeper.GetPreviousRewardsAuction(s.Ctx, liquidFarm)
	s.Require().True(found)
	s.Require().Equal("300000lfshare1", auction.WinningBid.Share.String())

	// Ensure bidderAddr4 has received farming rewards
	s.Require().True(s.GetBalance(bidderAddr4, "uatom").Amount.GT(sdk.NewInt(1)))

	// Burn all shares
	s.BurnShare(minterAddr1, liquidFarm.Id, s.GetBalance(minterAddr1, "lfshare1"))
	s.BurnShare(minterAddr2, liquidFarm.Id, s.GetBalance(minterAddr2, "lfshare1"))
}
