package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/x/liquidity/types"

	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)

	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCPools() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	s.createPair(creator, "denom1", "denom3", true)
	s.createPair(creator, "denom2", "denom3", true)
	s.createPair(creator, "denom3", "denom4", true)
	s.createPool(creator, 1, parseCoins("1000000denom1,1000000denom2"), true)
	s.createPool(creator, 2, parseCoins("5000000denom1,5000000denom3"), true)
	s.createPool(creator, 3, parseCoins("3000000denom2,3000000denom3"), true)
	s.createPool(creator, 4, parseCoins("3000000denom3,3000000denom4"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolsRequest
		expectErr bool
		postRun   func(*types.QueryPoolsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryPoolsRequest{},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 4)
			},
		},
		{
			"query all with query string XDenom",
			&types.QueryPoolsRequest{
				PairId: 1,
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
		{
			"query all with query string YDenom",
			&types.QueryPoolsRequest{
				Disabled: "false",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
		{
			"query all with query string XDenom and YDenom",
			&types.QueryPoolsRequest{
				PairId:   1,
				Disabled: "false",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pools(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPool() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, 1, parseCoins("1000000denom1,1000000denom2"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolRequest
		expectErr bool
		postRun   func(*types.QueryPoolResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryPoolRequest{},
			true,
			nil,
		},
		{
			"query all pool with pool id",
			&types.QueryPoolRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryPoolResponse) {
				s.Require().Equal(pool.Id, resp.Pool.Id)
				s.Require().Equal(pool.PairId, resp.Pool.PairId)
				s.Require().Equal(parseCoins("1000000denom1,1000000denom2"), resp.Pool.Balances)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.LastDepositRequestId, resp.Pool.LastDepositRequestId)
				s.Require().Equal(pool.LastWithdrawRequestId, resp.Pool.LastWithdrawRequestId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pool(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPoolByReserveAcc() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, 1, parseCoins("1000000denom1,1000000denom2"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolByReserveAccRequest
		expectErr bool
		postRun   func(*types.QueryPoolResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryPoolByReserveAccRequest{},
			true,
			nil,
		},
		{
			"query specific pool with the reserve account",
			&types.QueryPoolByReserveAccRequest{
				ReserveAcc: pool.ReserveAddress,
			},
			false,
			func(resp *types.QueryPoolResponse) {
				s.Require().Equal(pool.Id, resp.Pool.Id)
				s.Require().Equal(pool.PairId, resp.Pool.PairId)
				s.Require().Equal(parseCoins("1000000denom1,1000000denom2"), resp.Pool.Balances)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.LastDepositRequestId, resp.Pool.LastDepositRequestId)
				s.Require().Equal(pool.LastWithdrawRequestId, resp.Pool.LastWithdrawRequestId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PoolByReserveAcc(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPoolByPoolCoinDenom() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, 1, parseCoins("5000000denom1,5000000denom2"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolByPoolCoinDenomRequest
		expectErr bool
		postRun   func(*types.QueryPoolResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryPoolByPoolCoinDenomRequest{},
			true,
			nil,
		},
		{
			"query specific pool with the pool coin denom",
			&types.QueryPoolByPoolCoinDenomRequest{
				PoolCoinDenom: pool.PoolCoinDenom,
			},
			false,
			func(resp *types.QueryPoolResponse) {
				s.Require().Equal(pool.Id, resp.Pool.Id)
				s.Require().Equal(pool.PairId, resp.Pool.PairId)
				s.Require().Equal(parseCoins("1000000denom1,1000000denom2"), resp.Pool.Balances)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.LastDepositRequestId, resp.Pool.LastDepositRequestId)
				s.Require().Equal(pool.LastWithdrawRequestId, resp.Pool.LastWithdrawRequestId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PoolByPoolCoinDenom(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPairs() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	s.createPair(creator, "denom1", "denom3", true)
	s.createPair(creator, "denom2", "denom3", true)
	s.createPair(creator, "denom3", "denom4", true)

	s.Require().Len(s.keeper.GetAllPairs(s.ctx), 4)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPairsRequest
		expectErr bool
		postRun   func(*types.QueryPairsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryPairsRequest{},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 4)
			},
		},
		{
			"query all with query string XDenom",
			&types.QueryPairsRequest{
				Denoms: []string{"denom1"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 2)
			},
		},
		{
			"query all with query string YDenom",
			&types.QueryPairsRequest{
				Denoms: []string{"denom2"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 1)
			},
		},
		{
			"query all with query strings XDenom and YDenom",
			&types.QueryPairsRequest{
				Denoms: []string{"denom1", "denom2"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 1)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pairs(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCPair() {
	creator := s.addr(0)
	s.createPair(creator, "denom1", "denom2", true)
	s.createPair(creator, "denom1", "denom3", true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPairRequest
		expectErr bool
		postRun   func(*types.QueryPairResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryPairRequest{},
			true,
			nil,
		},
		{
			"query all pool with pair id",
			&types.QueryPairRequest{
				PairId: 1,
			},
			false,
			func(resp *types.QueryPairResponse) {
				// s.Require().Equal(pool.PairId, resp.Pair.Id)
				// s.Require().Equal(pool.XCoinDenom, resp.Pair.XCoinDenom)
				// s.Require().Equal(pool.YCoinDenom, resp.Pair.YCoinDenom)
				// s.Require().Equal(types.PairEscrowAddr(resp.Pair.Id).String(), resp.Pair.EscrowAddress)
				// s.Require().Equal(uint64(0), resp.Pair.LastSwapRequestId)
				// s.Require().Equal(uint64(0), resp.Pair.LastCancelSwapRequestId)
				// s.Require().Equal(uint64(1), resp.Pair.CurrentBatchId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Pair(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCDepositRequests() {
	// k, ctx := s.keeper, s.ctx

	// params := k.GetParams(ctx)

	// // Create a normal pool
	// creator := s.addr(0)
	// xCoin, yCoin := parseCoin("1000000denom1"), parseCoin("1000000denom2")
	// s.createPool(creator, xCoin, yCoin, true)

	// pool, found := k.GetPool(ctx, 1)
	// s.Require().True(found)
	// s.Require().Equal(params.InitialPoolCoinSupply, s.getBalance(creator, pool.PoolCoinDenom).Amount)

	// // The creator withdraws pool coin
	// poolCoin := s.getBalance(creator, pool.PoolCoinDenom)
	// s.withdrawBatch(creator, pool.Id, poolCoin)
	// s.nextBlock()
}

func (s *KeeperTestSuite) TestGRPCDepositRequest() {

}

func (s *KeeperTestSuite) TestGRPCWithdrawRequests() {
	// Create Pool
	// Withdraw pool coin
}

func (s *KeeperTestSuite) TestGRPCWithdrawRequest() {

}

func (s *KeeperTestSuite) TestGRPCSwapRequests() {

}
func (s *KeeperTestSuite) TestGRPCSwapRequest() {

}
