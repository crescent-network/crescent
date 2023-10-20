package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utils "github.com/crescent-network/crescent/v5/types"
	"github.com/crescent-network/crescent/v5/x/amm/types"
	exchangekeeper "github.com/crescent-network/crescent/v5/x/exchange/keeper"
	exchangetypes "github.com/crescent-network/crescent/v5/x/exchange/types"
)

// SetupSampleScenario creates markets, pools and positions for query tests.
func (s *KeeperTestSuite) SetupSampleScenario() {
	s.T().Helper()

	mmAddr := s.FundedAccount(100, enoughCoins)
	creUsdMarket := s.CreateMarket("ucre", "uusd")
	atomUsdMarket := s.CreateMarket("uatom", "uusd")
	s.MakeLastPrice(creUsdMarket.Id, mmAddr, utils.ParseDec("5"))
	s.MakeLastPrice(atomUsdMarket.Id, mmAddr, utils.ParseDec("10"))
	// pool id != market id
	atomUsdPool := s.CreatePool(atomUsdMarket.Id, utils.ParseDec("10"))
	creUsdPool := s.CreatePool(creUsdMarket.Id, utils.ParseDec("5"))

	aliceAddr := s.FundedAccount(1, enoughCoins)
	bobAddr := s.FundedAccount(2, enoughCoins)
	s.AddLiquidity(
		aliceAddr, creUsdPool.Id, utils.ParseDec("4"), utils.ParseDec("6"),
		utils.ParseCoins("100_000000ucre,500_000000uusd"))
	s.AddLiquidity(
		aliceAddr, atomUsdPool.Id, utils.ParseDec("9"), utils.ParseDec("11"),
		utils.ParseCoins("100_000000uatom,1000_000000uusd"))
	s.AddLiquidity(
		bobAddr, atomUsdPool.Id, utils.ParseDec("9.9"), utils.ParseDec("10.1"),
		utils.ParseCoins("10_000000uatom,100_000000uusd"))
	s.AddLiquidity(
		bobAddr, creUsdPool.Id, utils.ParseDec("4.9"), utils.ParseDec("5.1"),
		utils.ParseCoins("10_000000ucre,50_000000uusd"))

	creatorAddr := s.FundedAccount(3, enoughCoins)
	s.CreatePrivateFarmingPlan(
		creatorAddr, "CRE/USD Farming", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(creUsdPool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"),
		utils.ParseCoins("10000_000000ucre"), true)
	s.CreatePublicFarmingPlan(
		"ATOM/USD Farming", creatorAddr, creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(atomUsdPool.Id, utils.ParseCoins("100_000000uatom")),
		}, utils.ParseTime("2023-01-01T00:00:00Z"), utils.ParseTime("2024-01-01T00:00:00Z"))
	// These farming plans will be terminated immediately.
	s.CreatePublicFarmingPlan(
		"Old CRE/USD Farming", creatorAddr, creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(creUsdPool.Id, utils.ParseCoins("100_000000ucre")),
		}, utils.ParseTime("2022-01-01T00:00:00Z"), utils.ParseTime("2023-01-01T00:00:05Z"))
	s.CreatePrivateFarmingPlan(
		creatorAddr, "Old ATOM/USD Farming", creatorAddr, []types.FarmingRewardAllocation{
			types.NewFarmingRewardAllocation(atomUsdPool.Id, utils.ParseCoins("100_000000uatom")),
		}, utils.ParseTime("2022-01-01T00:00:00Z"), utils.ParseTime("2023-01-01T00:00:05Z"),
		utils.ParseCoins("10000_000000uatom"), true)

	// Distribute farming rewards for some blocks.
	s.NextBlock()
	s.NextBlock()
	s.NextBlock()

	ordererAddr := s.FundedAccount(4, enoughCoins)
	s.SwapExactAmountIn(
		ordererAddr, []uint64{atomUsdMarket.Id, creUsdMarket.Id},
		utils.ParseCoin("20_000000uatom"), utils.ParseCoin("38_000000ucre"), false)
	s.SwapExactAmountIn(
		ordererAddr, []uint64{creUsdMarket.Id, atomUsdMarket.Id},
		utils.ParseCoin("50_000000ucre"), utils.ParseCoin("20_000000uatom"), false)
	s.SwapExactAmountIn(
		ordererAddr, []uint64{creUsdMarket.Id, atomUsdMarket.Id},
		utils.ParseCoin("50_000000ucre"), utils.ParseCoin("20_000000uatom"), false)
}

func (s *KeeperTestSuite) TestQueryParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.Ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.Ctx), resp.Params)
}

func (s *KeeperTestSuite) TestQueryAllPools() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllPoolsRequest
		expectedErr string
		postRun     func(resp *types.QueryAllPoolsResponse)
	}{
		{
			"happy case",
			&types.QueryAllPoolsRequest{},
			"",
			func(resp *types.QueryAllPoolsResponse) {
				s.Require().Len(resp.Pools, 2)
				s.Require().EqualValues(1, resp.Pools[0].Id)
				s.Require().EqualValues(2, resp.Pools[1].Id)
			},
		},
		{
			"query by market id",
			&types.QueryAllPoolsRequest{
				MarketId: 1,
			},
			"",
			func(resp *types.QueryAllPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
				pool := resp.Pools[0]
				s.Require().EqualValues(1, pool.MarketId)
				s.Require().EqualValues(2, pool.Id)
				s.AssertEqual(utils.ParseCoin("153630723ucre"), pool.Balance0)
				s.AssertEqual(utils.ParseCoin("257144189uusd"), pool.Balance1)
			},
		},
		{
			"no market",
			&types.QueryAllPoolsRequest{
				MarketId: 3,
			},
			"rpc error: code = NotFound desc = market not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllPools(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryPool() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryPoolRequest
		expectedErr string
		postRun     func(resp *types.QueryPoolResponse)
	}{
		{
			"happy case",
			&types.QueryPoolRequest{
				PoolId: 1,
			},
			"",
			func(resp *types.QueryPoolResponse) {
				s.Require().EqualValues(1, resp.Pool.Id)
				s.AssertEqual(utils.ParseCoin("71757892uatom"), resp.Pool.Balance0)
				s.AssertEqual(utils.ParseCoin("1390645818uusd"), resp.Pool.Balance1)
				s.Require().Equal("cosmos1srphgsfqllr85ndknjme24txux8m0sz0hhpnnksn2339d3a788rsawjx77", resp.Pool.RewardsPool)
				s.AssertEqual(sdk.NewInt(12470981864), resp.Pool.TotalLiquidity)
			},
		},
		{
			"not found",
			&types.QueryPoolRequest{
				PoolId: 0,
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"not found 2",
			&types.QueryPoolRequest{
				PoolId: 4,
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pool(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryAllPositions() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllPositionsRequest
		expectedErr string
		postRun     func(resp *types.QueryAllPositionsResponse)
	}{
		{
			"happy case",
			&types.QueryAllPositionsRequest{},
			"",
			func(resp *types.QueryAllPositionsResponse) {
				s.Require().Len(resp.Positions, 4)
				s.Require().EqualValues(1, resp.Positions[0].Id)
				s.Require().EqualValues(2, resp.Positions[1].Id)
				s.Require().EqualValues(3, resp.Positions[2].Id)
				s.Require().EqualValues(4, resp.Positions[3].Id)
			},
		},
		{
			"query by pool",
			&types.QueryAllPositionsRequest{
				PoolId: 1,
			},
			"",
			func(resp *types.QueryAllPositionsResponse) {
				s.Require().Len(resp.Positions, 2)
				s.Require().EqualValues(2, resp.Positions[0].Id)
				s.Require().EqualValues(3, resp.Positions[1].Id)
			},
		},
		{
			"pool not found",
			&types.QueryAllPositionsRequest{
				PoolId: 3,
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"query by owner",
			&types.QueryAllPositionsRequest{
				Owner: utils.TestAddress(1).String(), // Alice
			},
			"",
			func(resp *types.QueryAllPositionsResponse) {
				s.Require().Len(resp.Positions, 2)
				s.Require().EqualValues(2, resp.Positions[0].Id)
				s.Require().EqualValues(1, resp.Positions[1].Id)
			},
		},
		{
			"invalid owner",
			&types.QueryAllPositionsRequest{
				Owner: "invalidaddr",
			},
			"rpc error: code = InvalidArgument desc = invalid owner: decoding bech32 failed: invalid separator index -1",
			nil,
		},
		{
			"query by pool id and owner",
			&types.QueryAllPositionsRequest{
				PoolId: 2,
				Owner:  utils.TestAddress(2).String(), // Bob
			},
			"",
			func(resp *types.QueryAllPositionsResponse) {
				s.Require().Len(resp.Positions, 1)
				s.Require().EqualValues(4, resp.Positions[0].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllPositions(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryPosition() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryPositionRequest
		expectedErr string
		postRun     func(resp *types.QueryPositionResponse)
	}{
		{
			"happy case",
			&types.QueryPositionRequest{
				PositionId: 1,
			},
			"",
			func(resp *types.QueryPositionResponse) {
				s.Require().EqualValues(1, resp.Position.Id)
				s.Require().Equal("4.000000000000000000", resp.Position.LowerPrice.String())
				s.Require().Equal("6.000000000000000000", resp.Position.UpperPrice.String())
			},
		},
		{
			"position not found",
			&types.QueryPositionRequest{
				PositionId: 0,
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
		{
			"position not found 2",
			&types.QueryPositionRequest{
				PositionId: 5,
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Position(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryPositionAssets() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryPositionAssetsRequest
		expectedErr string
		postRun     func(resp *types.QueryPositionAssetsResponse)
	}{
		{
			"happy case",
			&types.QueryPositionAssetsRequest{
				PositionId: 1,
			},
			"",
			func(resp *types.QueryPositionAssetsResponse) {
				s.AssertEqual(utils.ParseCoin("133726249ucre"), resp.Coin0)
				s.AssertEqual(utils.ParseCoin("257144184uusd"), resp.Coin1)
			},
		},
		{
			"position not found",
			&types.QueryPositionAssetsRequest{
				PositionId: 0,
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
		{
			"position not found 2",
			&types.QueryPositionAssetsRequest{
				PositionId: 5,
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			ctx, _ := s.Ctx.CacheContext()
			resp, err := s.querier.PositionAssets(sdk.WrapSDKContext(ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQuerySimulateAddLiquidity() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QuerySimulateAddLiquidityRequest
		expectedErr string
		postRun     func(resp *types.QuerySimulateAddLiquidityResponse)
	}{
		{
			"happy case",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        2,
				LowerPrice:    "4.5",
				UpperPrice:    "5.5",
				DesiredAmount: "100000000ucre,500000000uusd",
			},
			"",
			func(resp *types.QuerySimulateAddLiquidityResponse) {
				fmt.Println(s.keeper.MustGetPoolState(s.Ctx, 2).CurrentSqrtPrice)
				s.AssertEqual(sdk.NewInt(2223021101), resp.Liquidity)
				s.AssertEqual(utils.ParseCoins("100000000ucre,192673uusd"), resp.Amount)
			},
		},
		{
			"pool not found",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        0,
				LowerPrice:    "4.5",
				UpperPrice:    "5.5",
				DesiredAmount: "100000000ucre,500000000uusd",
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"pool not found 2",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        3,
				LowerPrice:    "4.5",
				UpperPrice:    "5.5",
				DesiredAmount: "100000000ucre,500000000uusd",
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"invalid lower price",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        2,
				LowerPrice:    "4.5?",
				UpperPrice:    "5.5",
				DesiredAmount: "100000000ucre,500000000uusd",
			},
			"rpc error: code = InvalidArgument desc = invalid lower price: failed to set decimal string with base 10: 45?0000000000000000",
			nil,
		},
		{
			"invalid upper price",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        2,
				LowerPrice:    "4.5",
				UpperPrice:    "5.5?",
				DesiredAmount: "100000000ucre,500000000uusd",
			},
			"rpc error: code = InvalidArgument desc = invalid upper price: failed to set decimal string with base 10: 55?0000000000000000",
			nil,
		},
		{
			"invalid desired amount",
			&types.QuerySimulateAddLiquidityRequest{
				PoolId:        2,
				LowerPrice:    "4.5",
				UpperPrice:    "5.5",
				DesiredAmount: "100000000ucre,-10uusd",
			},
			"rpc error: code = InvalidArgument desc = invalid desired amount: invalid decimal coin expression: -10uusd",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			cacheCtx, _ := s.Ctx.CacheContext()
			resp, err := s.querier.SimulateAddLiquidity(sdk.WrapSDKContext(cacheCtx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQuerySimulateRemoveLiquidity() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QuerySimulateRemoveLiquidityRequest
		expectedErr string
		postRun     func(resp *types.QuerySimulateRemoveLiquidityResponse)
	}{
		{
			"happy case",
			&types.QuerySimulateRemoveLiquidityRequest{
				PositionId: 1,
				Liquidity:  "10000000",
			},
			"",
			func(resp *types.QuerySimulateRemoveLiquidityResponse) {
				s.AssertEqual(utils.ParseCoins("631369ucre,1214070uusd"), resp.Amount)
			},
		},
		{
			"position not found",
			&types.QuerySimulateRemoveLiquidityRequest{
				PositionId: 0,
				Liquidity:  "10000000",
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
		{
			"position not found 2",
			&types.QuerySimulateRemoveLiquidityRequest{
				PositionId: 5,
				Liquidity:  "10000000",
			},
			"rpc error: code = NotFound desc = position not found",
			nil,
		},
		{
			"invalid liquidity",
			&types.QuerySimulateRemoveLiquidityRequest{
				PositionId: 2,
				Liquidity:  "10000000?",
			},
			"rpc error: code = InvalidArgument desc = invalid liquidity: 10000000?",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			cacheCtx, _ := s.Ctx.CacheContext()
			resp, err := s.querier.SimulateRemoveLiquidity(sdk.WrapSDKContext(cacheCtx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryCollectibleCoins() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryCollectibleCoinsRequest
		expectedErr string
		postRun     func(resp *types.QueryCollectibleCoinsResponse)
	}{
		{
			"query by owner",
			&types.QueryCollectibleCoinsRequest{
				Owner: utils.TestAddress(1).String(), // Alice
			},
			"",
			func(resp *types.QueryCollectibleCoinsResponse) {
				s.AssertEqual(utils.ParseCoins("21650uatom,21649ucre,1170389uusd"), resp.Fee)
				s.AssertEqual(utils.ParseCoins("8578uatom,8467ucre"), resp.FarmingRewards)
			},
		},
		{
			"query by position",
			&types.QueryCollectibleCoinsRequest{
				PositionId: 1,
			},
			"",
			func(resp *types.QueryCollectibleCoinsResponse) {
				s.AssertEqual(utils.ParseCoins("21649ucre,566242uusd"), resp.Fee)
				s.AssertEqual(utils.ParseCoins("8467ucre"), resp.FarmingRewards)
			},
		},
		{
			"invalid owner",
			&types.QueryCollectibleCoinsRequest{
				Owner: "invalidaddr",
			},
			"rpc error: code = InvalidArgument desc = invalid owner: decoding bech32 failed: invalid separator index -1",
			nil,
		},
		{
			"position not found",
			&types.QueryCollectibleCoinsRequest{
				PositionId: 5,
			},
			"rpc error: code = InvalidArgument desc = position not found: not found",
			nil,
		},
		{
			"owner has no positions",
			&types.QueryCollectibleCoinsRequest{
				Owner: utils.TestAddress(4).String(),
			},
			"",
			func(resp *types.QueryCollectibleCoinsResponse) {
				s.Require().True(resp.Fee.IsZero())
				s.Require().True(resp.FarmingRewards.IsZero())
			},
		},
		{
			"both specified",
			&types.QueryCollectibleCoinsRequest{
				Owner:      utils.TestAddress(1).String(),
				PositionId: 2,
			},
			"rpc error: code = InvalidArgument desc = owner and position id must not be specified at the same time",
			nil,
		},
		{
			"neither specified",
			&types.QueryCollectibleCoinsRequest{},
			"rpc error: code = InvalidArgument desc = owner or position id must be specified",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.CollectibleCoins(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryAllTickInfos() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllTickInfosRequest
		expectedErr string
		postRun     func(resp *types.QueryAllTickInfosResponse)
	}{
		{
			"happy case",
			&types.QueryAllTickInfosRequest{
				PoolId: 1,
			},
			"",
			func(resp *types.QueryAllTickInfosResponse) {
				s.Require().Len(resp.TickInfos, 4)
				s.Require().EqualValues(80000, resp.TickInfos[0].Tick)
				s.Require().EqualValues(89000, resp.TickInfos[1].Tick)
				s.Require().EqualValues(90100, resp.TickInfos[2].Tick)
				s.Require().EqualValues(91000, resp.TickInfos[3].Tick)
			},
		},
		{
			"pool not found",
			&types.QueryAllTickInfosRequest{
				PoolId: 3,
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"query with lower and upper tick",
			&types.QueryAllTickInfosRequest{
				PoolId:    2,
				LowerTick: "35000",
				UpperTick: "45000",
			},
			"",
			func(resp *types.QueryAllTickInfosResponse) {
				s.Require().Len(resp.TickInfos, 2)
				s.Require().EqualValues(39000, resp.TickInfos[0].Tick)
				s.Require().EqualValues(41000, resp.TickInfos[1].Tick)
			},
		},
		{
			"query with lower tick",
			&types.QueryAllTickInfosRequest{
				PoolId:    2,
				LowerTick: "35000",
			},
			"",
			func(resp *types.QueryAllTickInfosResponse) {
				s.Require().Len(resp.TickInfos, 3)
				s.Require().EqualValues(39000, resp.TickInfos[0].Tick)
				s.Require().EqualValues(41000, resp.TickInfos[1].Tick)
				s.Require().EqualValues(50000, resp.TickInfos[2].Tick)
			},
		},
		{
			"query with upper tick",
			&types.QueryAllTickInfosRequest{
				PoolId:    2,
				UpperTick: "45000",
			},
			"",
			func(resp *types.QueryAllTickInfosResponse) {
				s.Require().Len(resp.TickInfos, 3)
				s.Require().EqualValues(30000, resp.TickInfos[0].Tick)
				s.Require().EqualValues(39000, resp.TickInfos[1].Tick)
				s.Require().EqualValues(41000, resp.TickInfos[2].Tick)
			},
		},
		{
			"invalid lower tick",
			&types.QueryAllTickInfosRequest{
				PoolId:    2,
				LowerTick: "invalid",
			},
			"rpc error: code = InvalidArgument desc = invalid lower tick: strconv.ParseInt: parsing \"invalid\": invalid syntax",
			nil,
		},
		{
			"invalid upper tick",
			&types.QueryAllTickInfosRequest{
				PoolId:    2,
				UpperTick: "invalid",
			},
			"rpc error: code = InvalidArgument desc = invalid upper tick: strconv.ParseInt: parsing \"invalid\": invalid syntax",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllTickInfos(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryTickInfo() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryTickInfoRequest
		expectedErr string
		postRun     func(resp *types.QueryTickInfoResponse)
	}{
		{
			"happy case",
			&types.QueryTickInfoRequest{
				PoolId: 1,
				Tick:   89000,
			},
			"",
			func(resp *types.QueryTickInfoResponse) {
				s.Require().EqualValues(89000, resp.TickInfo.Tick)
			},
		},
		{
			"pool not found",
			&types.QueryTickInfoRequest{
				PoolId: 3,
			},
			"rpc error: code = NotFound desc = pool not found",
			nil,
		},
		{
			"tick info not found",
			&types.QueryTickInfoRequest{
				PoolId: 2,
				Tick:   12345,
			},
			"rpc error: code = NotFound desc = tick info not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.TickInfo(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryAllFarmingPlans() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryAllFarmingPlansRequest
		expectedErr string
		postRun     func(resp *types.QueryAllFarmingPlansResponse)
	}{
		{
			"happy case",
			&types.QueryAllFarmingPlansRequest{},
			"",
			func(resp *types.QueryAllFarmingPlansResponse) {
				s.Require().Len(resp.FarmingPlans, 4)
				s.Require().EqualValues(1, resp.FarmingPlans[0].Id)
				s.Require().EqualValues(2, resp.FarmingPlans[1].Id)
				s.Require().EqualValues(3, resp.FarmingPlans[2].Id)
				s.Require().EqualValues(4, resp.FarmingPlans[3].Id)
			},
		},
		{
			"query with is_private",
			&types.QueryAllFarmingPlansRequest{
				IsPrivate: "true",
			},
			"",
			func(resp *types.QueryAllFarmingPlansResponse) {
				s.Require().Len(resp.FarmingPlans, 2)
				s.Require().EqualValues(1, resp.FarmingPlans[0].Id)
				s.Require().EqualValues(4, resp.FarmingPlans[1].Id)
			},
		},
		{
			"invalid is_private",
			&types.QueryAllFarmingPlansRequest{
				IsPrivate: "invalid",
			},
			"rpc error: code = InvalidArgument desc = invalid is_private: strconv.ParseBool: parsing \"invalid\": invalid syntax",
			nil,
		},
		{
			"query with is_terminated",
			&types.QueryAllFarmingPlansRequest{
				IsTerminated: "false",
			},
			"",
			func(resp *types.QueryAllFarmingPlansResponse) {
				s.Require().Len(resp.FarmingPlans, 2)
				s.Require().EqualValues(1, resp.FarmingPlans[0].Id)
				s.Require().EqualValues(2, resp.FarmingPlans[1].Id)
			},
		},
		{
			"invalid is_terminated",
			&types.QueryAllFarmingPlansRequest{
				IsTerminated: "invalid",
			},
			"rpc error: code = InvalidArgument desc = invalid is_terminated: strconv.ParseBool: parsing \"invalid\": invalid syntax",
			nil,
		},
		{
			"query with both is_private and is_terminated",
			&types.QueryAllFarmingPlansRequest{
				IsPrivate:    "false",
				IsTerminated: "true",
			},
			"",
			func(resp *types.QueryAllFarmingPlansResponse) {
				s.Require().Len(resp.FarmingPlans, 1)
				s.Require().EqualValues(3, resp.FarmingPlans[0].Id)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.AllFarmingPlans(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryFarmingPlan() {
	s.SetupSampleScenario()

	for _, tc := range []struct {
		name        string
		req         *types.QueryFarmingPlanRequest
		expectedErr string
		postRun     func(resp *types.QueryFarmingPlanResponse)
	}{
		{
			"happy case",
			&types.QueryFarmingPlanRequest{
				PlanId: 1,
			},
			"",
			func(resp *types.QueryFarmingPlanResponse) {
				s.Require().EqualValues(1, resp.FarmingPlan.Id)
			},
		},
		{
			"plan not found",
			&types.QueryFarmingPlanRequest{
				PlanId: 0,
			},
			"rpc error: code = NotFound desc = farming plan not found",
			nil,
		},
		{
			"plan not found 2",
			&types.QueryFarmingPlanRequest{
				PlanId: 5,
			},
			"rpc error: code = NotFound desc = farming plan not found",
			nil,
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.FarmingPlan(sdk.WrapSDKContext(s.Ctx), tc.req)
			if tc.expectedErr == "" {
				s.Require().NoError(err)
				tc.postRun(resp)
			} else {
				s.Require().EqualError(err, tc.expectedErr)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryOrderBookEdgecase() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("0.000000000002410188"))
	market.OrderQuantityLimits.Min = sdk.NewInt(1)
	market.OrderQuoteLimits.Min = sdk.NewInt(1)
	s.App.ExchangeKeeper.SetMarket(s.Ctx, market)

	lpAddr := s.FundedAccount(1, enoughCoins)
	// Make last price.
	s.PlaceLimitOrder(market.Id, lpAddr, true, utils.ParseDec("0.0000000000024102"), sdk.NewInt(100000000000000000), time.Hour)
	s.PlaceLimitOrder(market.Id, lpAddr, false, utils.ParseDec("0.0000000000024102"), sdk.NewInt(100000000000000000), time.Hour)

	s.AddLiquidityByLiquidity(
		lpAddr, pool.Id, types.MinPrice, types.MaxPrice,
		sdk.NewInt(160843141868))

	querier := exchangekeeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.OrderBook(sdk.WrapSDKContext(s.Ctx), &exchangetypes.QueryOrderBookRequest{
		MarketId: market.Id,
	})
	s.Require().NoError(err)
	expected := []exchangetypes.OrderBookPriceLevel{
		{P: utils.ParseDec("0.000000000002415000"), Q: sdk.NewInt(103269500340991)},
		{P: utils.ParseDec("0.000000000002420000"), Q: sdk.NewInt(106977667069587)},
		{P: utils.ParseDec("0.000000000002425000"), Q: sdk.NewInt(106646637524773)},
		{P: utils.ParseDec("0.000000000002430000"), Q: sdk.NewInt(106317311684281)},
	}
	s.Require().GreaterOrEqual(len(resp.OrderBooks[0].Sells), len(expected))
	for i, level := range expected {
		s.AssertEqual(level.P, resp.OrderBooks[0].Sells[i].P)
		s.AssertEqual(level.Q, resp.OrderBooks[0].Sells[i].Q)
	}
	expected = []exchangetypes.OrderBookPriceLevel{
		{P: utils.ParseDec("0.000000000002410000"), Q: sdk.NewInt(4040912342627)},
		{P: utils.ParseDec("0.000000000002405000"), Q: sdk.NewInt(107644886847379)},
		{P: utils.ParseDec("0.000000000002400000"), Q: sdk.NewInt(107981102158327)},
		{P: utils.ParseDec("0.000000000002395000"), Q: sdk.NewInt(108319071332206)},
	}
	s.Require().GreaterOrEqual(len(resp.OrderBooks[0].Buys), len(expected))
	for i, level := range expected {
		s.AssertEqual(level.P, resp.OrderBooks[0].Buys[i].P)
		s.AssertEqual(level.Q, resp.OrderBooks[0].Buys[i].Q)
	}
}

func (s *KeeperTestSuite) TestQueryOrderBookEdgecase2() {
	market, pool := s.CreateMarketAndPool("ucre", "uusd", utils.ParseDec("1"))

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, lpAddr, utils.ParseDec("1"))

	s.AddLiquidityByLiquidity(
		lpAddr, pool.Id, types.MinPrice, types.MaxPrice, sdk.NewInt(10))

	querier := exchangekeeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.OrderBook(sdk.WrapSDKContext(s.Ctx), &exchangetypes.QueryOrderBookRequest{
		MarketId: market.Id,
	})
	s.Require().NoError(err)
	// Due to too low liquidity, order book is not displayed.
	s.Require().Empty(resp.OrderBooks)
}

func (s *KeeperTestSuite) TestQueryOrderBookEdgecase3() {
	market := s.CreateMarket("ucre", "uusd")
	market.OrderQuantityLimits.Min = sdk.NewInt(1)
	market.OrderQuoteLimits.Min = sdk.NewInt(1)
	s.App.ExchangeKeeper.SetMarket(s.Ctx, market)

	lpAddr := s.FundedAccount(1, enoughCoins)
	s.MakeLastPrice(market.Id, lpAddr, utils.ParseDec("5"))

	// pool price != last price
	pool := s.CreatePool(market.Id, utils.ParseDec("1000"))

	s.AddLiquidityByLiquidity(
		lpAddr, pool.Id, types.MinPrice, types.MaxPrice, sdk.NewInt(1000))

	s.PlaceLimitOrder(market.Id, lpAddr, true, utils.ParseDec("4.99"), sdk.NewInt(3_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, lpAddr, true, utils.ParseDec("4.98"), sdk.NewInt(2_000000), time.Hour)
	s.PlaceLimitOrder(market.Id, lpAddr, true, utils.ParseDec("4.97"), sdk.NewInt(1_000000), time.Hour)

	querier := exchangekeeper.Querier{Keeper: s.App.ExchangeKeeper}
	resp, err := querier.OrderBook(sdk.WrapSDKContext(s.Ctx), &exchangetypes.QueryOrderBookRequest{
		MarketId: market.Id,
	})
	s.Require().NoError(err)
	ob := resp.OrderBooks[0]
	s.Require().Empty(ob.Sells)
	expected := []exchangetypes.OrderBookPriceLevel{
		{P: utils.ParseDec("939.5"), Q: sdk.NewInt(1)},
		{P: utils.ParseDec("4.99"), Q: sdk.NewInt(3_000000)},
		{P: utils.ParseDec("4.98"), Q: sdk.NewInt(2_000000)},
		{P: utils.ParseDec("4.97"), Q: sdk.NewInt(1_000000)},
	}
	s.Require().GreaterOrEqual(len(ob.Buys), len(expected))
	for i, level := range expected {
		s.AssertEqual(level.P, ob.Buys[i].P)
		s.AssertEqual(level.Q, ob.Buys[i].Q)
	}

	ob = resp.OrderBooks[1]
	s.Require().Empty(ob.Sells)
	expected = []exchangetypes.OrderBookPriceLevel{
		{P: utils.ParseDec("939.5"), Q: sdk.NewInt(1)},
		{P: utils.ParseDec("4.9"), Q: sdk.NewInt(6_000000)},
	}
	s.Require().GreaterOrEqual(len(ob.Buys), len(expected))
	for i, level := range expected {
		s.AssertEqual(level.P, ob.Buys[i].P)
		s.AssertEqual(level.Q, ob.Buys[i].Q)
	}

	ob = resp.OrderBooks[2]
	s.Require().Empty(ob.Sells)
	expected = []exchangetypes.OrderBookPriceLevel{
		{P: utils.ParseDec("939"), Q: sdk.NewInt(1)},
		{P: utils.ParseDec("4"), Q: sdk.NewInt(6_000000)},
	}
	s.Require().GreaterOrEqual(len(ob.Buys), len(expected))
	for i, level := range expected {
		s.AssertEqual(level.P, ob.Buys[i].P)
		s.AssertEqual(level.Q, ob.Buys[i].Q)
	}
}
