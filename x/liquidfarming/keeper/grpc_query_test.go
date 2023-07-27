package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/liquidfarming/types"
	liquiditytypes "github.com/crescent-network/crescent/v5/x/liquidity/types"
	lpfarmtypes "github.com/crescent-network/crescent/v5/x/lpfarm/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCLiquidFarms() {
	pair1 := s.createPair(s.addr(0), "denom1", "denom2")
	pool1 := s.createPool(s.addr(0), pair1.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	minFarmAmt1, minBidAmt1 := sdk.NewInt(10_000_000), sdk.NewInt(10_000_000)
	s.createLiquidFarm(pool1.Id, minFarmAmt1, minBidAmt1, sdk.ZeroDec())

	pair2 := s.createPair(s.addr(1), "denom3", "denom4")
	pool2 := s.createPool(s.addr(1), pair2.Id, utils.ParseCoins("100_000_000denom3, 100_000_000denom4"))
	minFarmAmt2, minBidAmt2 := sdk.NewInt(30_000_000), sdk.NewInt(30_000_000)
	s.createLiquidFarm(pool2.Id, minFarmAmt2, minBidAmt2, sdk.ZeroDec())

	for _, tc := range []struct {
		name      string
		req       *types.QueryLiquidFarmsRequest
		expectErr bool
		postRun   func(*types.QueryLiquidFarmsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmsRequest{},
			false,
			func(resp *types.QueryLiquidFarmsResponse) {
				s.Require().Len(resp.LiquidFarms, 2)

				for _, liquidFarm := range resp.LiquidFarms {
					switch liquidFarm.PoolId {
					case 1:
						s.Require().Equal(minFarmAmt1, liquidFarm.MinFarmAmount)
						s.Require().Equal(minBidAmt1, liquidFarm.MinBidAmount)
					case 2:
						s.Require().Equal(minFarmAmt2, liquidFarm.MinFarmAmount)
						s.Require().Equal(minBidAmt2, liquidFarm.MinBidAmount)
					}
					reserveAddr := types.LiquidFarmReserveAddress(liquidFarm.PoolId)
					lfCoinDenom := types.LiquidFarmCoinDenom(liquidFarm.PoolId)
					lfCoinSupplyAmt := s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount
					poolCoinDenom := liquiditytypes.PoolCoinDenom(liquidFarm.PoolId)
					position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, poolCoinDenom)
					if !found {
						position.FarmingAmount = sdk.ZeroInt()
					}
					s.Require().Equal(lfCoinDenom, liquidFarm.LFCoinDenom)
					s.Require().Equal(lfCoinSupplyAmt, liquidFarm.LFCoinSupply)
					s.Require().Equal(poolCoinDenom, liquidFarm.PoolCoinDenom)
					s.Require().Equal(position.FarmingAmount, liquidFarm.PoolCoinFarmingAmount)
				}
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarms(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCLiquidFarm() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	minFarmAmt, minBidAmt := sdk.NewInt(10_000_000), sdk.NewInt(10_000_000)
	s.createLiquidFarm(pool.Id, minFarmAmt, minBidAmt, sdk.ZeroDec())

	for _, tc := range []struct {
		name      string
		req       *types.QueryLiquidFarmRequest
		expectErr bool
		postRun   func(*types.QueryLiquidFarmResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"happy case",
			&types.QueryLiquidFarmRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryLiquidFarmResponse) {
				reserveAddr := types.LiquidFarmReserveAddress(resp.LiquidFarm.PoolId)
				lfCoinDenom := types.LiquidFarmCoinDenom(pool.Id)
				lfCoinSupplyAmt := s.app.BankKeeper.GetSupply(s.ctx, lfCoinDenom).Amount
				poolCoinDenom := liquiditytypes.PoolCoinDenom(resp.LiquidFarm.PoolId)
				position, found := s.app.LPFarmKeeper.GetPosition(s.ctx, reserveAddr, poolCoinDenom)
				if !found {
					position.FarmingAmount = sdk.ZeroInt()
				}
				s.Require().Equal(types.LiquidFarmReserveAddress(pool.Id).String(), resp.LiquidFarm.LiquidFarmReserveAddress)
				s.Require().Equal(lfCoinDenom, resp.LiquidFarm.LFCoinDenom)
				s.Require().Equal(lfCoinSupplyAmt, resp.LiquidFarm.LFCoinSupply)
				s.Require().Equal(poolCoinDenom, resp.LiquidFarm.PoolCoinDenom)
				s.Require().Equal(position.FarmingAmount, resp.LiquidFarm.PoolCoinFarmingAmount)
				s.Require().Equal(minFarmAmt, resp.LiquidFarm.MinFarmAmount)
				s.Require().Equal(minBidAmt, resp.LiquidFarm.MinBidAmount)
			},
		},
		{
			"query by invalid pool id",
			&types.QueryLiquidFarmRequest{
				PoolId: 5,
			},
			true,
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.LiquidFarm(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCRewardsAuctions() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsAuctionsRequest
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryRewardsAuctionsResponse) {},
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: 10,
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 0)
			},
		},
		{
			"query by pool id",
			&types.QueryRewardsAuctionsRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 3)
			},
		},
		{
			"query by pool id and AuctionStatusStarted status",
			&types.QueryRewardsAuctionsRequest{
				PoolId: pool.Id,
				Status: types.AuctionStatusStarted.String(),
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 1)
			},
		},
		{
			"query by pool id and AuctionStatusFinished status",
			&types.QueryRewardsAuctionsRequest{
				PoolId: pool.Id,
				Status: types.AuctionStatusFinished.String(),
			},
			false,
			func(resp *types.QueryRewardsAuctionsResponse) {
				s.Require().Len(resp.RewardsAuctions, 2)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuctions(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCRewardsAuction() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsAuctionRequest
		expectErr bool
		postRun   func(*types.QueryRewardsAuctionResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsAuctionRequest{
				PoolId:    10,
				AuctionId: 1,
			},
			true,
			func(resp *types.QueryRewardsAuctionResponse) {},
		},
		{
			"query by invalid auction id",
			&types.QueryRewardsAuctionRequest{
				PoolId:    1,
				AuctionId: 10,
			},
			true,
			func(resp *types.QueryRewardsAuctionResponse) {},
		},
		{
			"query finished auction",
			&types.QueryRewardsAuctionRequest{
				PoolId:    pool.Id,
				AuctionId: 1,
			},
			false,
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().Equal(pool.PoolCoinDenom, resp.RewardsAuction.BiddingCoinDenom)
				s.Require().Equal(types.PayingReserveAddress(pool.Id), resp.RewardsAuction.GetPayingReserveAddress())
				s.Require().Equal(types.AuctionStatusFinished, resp.RewardsAuction.Status)
			},
		},
		{
			"query started auction",
			&types.QueryRewardsAuctionRequest{
				PoolId:    pool.Id,
				AuctionId: 2,
			},
			false,
			func(resp *types.QueryRewardsAuctionResponse) {
				s.Require().Equal(pool.PoolCoinDenom, resp.RewardsAuction.BiddingCoinDenom)
				s.Require().Equal(types.PayingReserveAddress(pool.Id), resp.RewardsAuction.GetPayingReserveAddress())
				s.Require().Equal(types.AuctionStatusStarted, resp.RewardsAuction.Status)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.RewardsAuction(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCBids() {
	pair := s.createPair(s.addr(0), "denom1", "denom2")
	pool := s.createPool(s.addr(0), pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())

	s.deposit(s.addr(1), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(2), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.deposit(s.addr(3), pool.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), utils.ParseCoin("100_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(2), utils.ParseCoin("300_000pool1"), true)
	s.liquidFarm(pool.Id, s.addr(3), utils.ParseCoin("500_000pool1"), true)
	s.nextBlock()

	s.nextAuction()

	s.placeBid(pool.Id, s.addr(5), utils.ParseCoin("100_000pool1"), true)
	s.placeBid(pool.Id, s.addr(6), utils.ParseCoin("110_000pool1"), true)
	s.placeBid(pool.Id, s.addr(7), utils.ParseCoin("150_000pool1"), true)
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryBidsRequest
		expectErr bool
		postRun   func(*types.QueryBidsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryBidsRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryBidsResponse) {},
		},
		{
			"query by pool id",
			&types.QueryBidsRequest{
				PoolId: pool.Id,
			},
			false,
			func(resp *types.QueryBidsResponse) {
				s.Require().Len(resp.Bids, 3)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Bids(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRewards() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	err := s.keeper.LiquidFarm(s.ctx, pool.Id, s.addr(0), utils.ParseCoin("1_000_000pool1"))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(100_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(200_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(2), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(300_000_000)), true)
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryRewardsRequest
		expectErr bool
		postRun   func(*types.QueryRewardsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryRewardsRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryRewardsResponse) {},
		},
		{
			"query by valid pool id",
			&types.QueryRewardsRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryRewardsResponse) {
				s.Require().True(resp.Rewards.IsAllPositive())
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Rewards(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCExchangeRate() {
	pair := s.createPairWithLastPrice(helperAddr, "denom1", "denom2", sdk.NewDec(1))
	pool := s.createPool(helperAddr, pair.Id, utils.ParseCoins("100_000_000denom1, 100_000_000denom2"))
	plan := s.createPrivatePlan(helperAddr, []lpfarmtypes.RewardAllocation{
		{
			PairId:        pool.PairId,
			RewardsPerDay: utils.ParseCoins("100_000_000stake"),
		},
	})
	s.fundAddr(plan.GetFarmingPoolAddress(), utils.ParseCoins("500_000_000stake"))

	err := s.keeper.LiquidFarm(s.ctx, pool.Id, s.addr(0), utils.ParseCoin("1_000_000pool1"))
	s.Require().EqualError(err, "liquid farm by pool 1 not found: not found")

	s.createLiquidFarm(pool.Id, sdk.ZeroInt(), sdk.ZeroInt(), sdk.ZeroDec())
	s.Require().Len(s.keeper.GetParams(s.ctx).LiquidFarms, 1)

	s.liquidFarm(pool.Id, s.addr(0), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(100_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(1), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(200_000_000)), true)
	s.nextBlock()

	s.liquidFarm(pool.Id, s.addr(2), sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(300_000_000)), true)
	s.nextBlock()

	for _, tc := range []struct {
		name      string
		req       *types.QueryExchangeRateRequest
		expectErr bool
		postRun   func(*types.QueryExchangeRateResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query by invalid pool id",
			&types.QueryExchangeRateRequest{
				PoolId: 0,
			},
			true,
			func(resp *types.QueryExchangeRateResponse) {},
		},
		{
			"query by valid pool id",
			&types.QueryExchangeRateRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryExchangeRateResponse) {
				s.Require().True(resp.ExchangeRate.MintRate.Equal(sdk.OneDec()))
				s.Require().True(resp.ExchangeRate.BurnRate.Equal(sdk.OneDec()))
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.ExchangeRate(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
