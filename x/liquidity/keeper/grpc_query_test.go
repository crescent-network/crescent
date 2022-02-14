package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	squad "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidity"
	"github.com/cosmosquad-labs/squad/x/liquidity/types"

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
	s.createPool(creator, 1, squad.ParseCoins("1000000denom1,1000000denom2"), true)
	s.createPool(creator, 2, squad.ParseCoins("5000000denom1,5000000denom3"), true)
	s.createPool(creator, 3, squad.ParseCoins("3000000denom2,3000000denom3"), true)
	pair4 := s.createPool(creator, 4, squad.ParseCoins("3000000denom3,3000000denom4"), true)
	pair4.Disabled = true
	s.keeper.SetPool(s.ctx, pair4)

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
			"query all with pair id",
			&types.QueryPoolsRequest{
				PairId: 1,
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
		{
			"query all with disabled",
			&types.QueryPoolsRequest{
				Disabled: "false",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 3)
			},
		},
		{
			"query all with disabled",
			&types.QueryPoolsRequest{
				Disabled: "true",
			},
			false,
			func(resp *types.QueryPoolsResponse) {
				s.Require().Len(resp.Pools, 1)
			},
		},
		{
			"query all with both pair id and disabled",
			&types.QueryPoolsRequest{
				PairId:   4,
				Disabled: "true",
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
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("1000000denom1,1000000denom2"), true)

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
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(squad.ParseCoins("1000000denom1,1000000denom2"), resp.Pool.Balances)
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

func (s *KeeperTestSuite) TestGRPCPoolByReserveAddress() {
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("2000000denom1,2000000denom2"), true)

	for _, tc := range []struct {
		name      string
		req       *types.QueryPoolByReserveAddressRequest
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
			&types.QueryPoolByReserveAddressRequest{},
			true,
			nil,
		},
		{
			"query specific pool with the reserve account",
			&types.QueryPoolByReserveAddressRequest{
				ReserveAddress: pool.ReserveAddress,
			},
			false,
			func(resp *types.QueryPoolResponse) {
				s.Require().Equal(pool.Id, resp.Pool.Id)
				s.Require().Equal(pool.PairId, resp.Pool.PairId)
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(squad.ParseCoins("2000000denom1,2000000denom2"), resp.Pool.Balances)
				s.Require().Equal(pool.LastDepositRequestId, resp.Pool.LastDepositRequestId)
				s.Require().Equal(pool.LastWithdrawRequestId, resp.Pool.LastWithdrawRequestId)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.PoolByReserveAddress(sdk.WrapSDKContext(s.ctx), tc.req)
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
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("5000000denom1,5000000denom2"), true)

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
				s.Require().Equal(pool.ReserveAddress, resp.Pool.ReserveAddress)
				s.Require().Equal(pool.PoolCoinDenom, resp.Pool.PoolCoinDenom)
				s.Require().Equal(squad.ParseCoins("5000000denom1,5000000denom2"), resp.Pool.Balances)
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
			"query all with a single denom",
			&types.QueryPairsRequest{
				Denoms: []string{"denom1"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 2)
			},
		},
		{
			"query all with a single denom",
			&types.QueryPairsRequest{
				Denoms: []string{"denom3"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 3)
			},
		},
		{
			"query all with two denoms",
			&types.QueryPairsRequest{
				Denoms: []string{"denom3", "denom4"},
			},
			false,
			func(resp *types.QueryPairsResponse) {
				s.Require().Len(resp.Pairs, 1)
			},
		},
		{
			"query all with more than two denoms",
			&types.QueryPairsRequest{
				Denoms: []string{"denom1", "denom3", "denom4"},
			},
			true,
			nil,
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
	pair := s.createPair(creator, "denom1", "denom2", true)

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
				s.Require().Equal(pair.Id, resp.Pair.Id)
				s.Require().Equal(pair.BaseCoinDenom, resp.Pair.BaseCoinDenom)
				s.Require().Equal(pair.QuoteCoinDenom, resp.Pair.QuoteCoinDenom)
				s.Require().Equal(pair.EscrowAddress, resp.Pair.EscrowAddress)
				s.Require().Equal(pair.LastOrderId, resp.Pair.LastOrderId)
				s.Require().Equal(pair.LastPrice, resp.Pair.LastPrice)
				s.Require().Equal(pair.CurrentBatchId, resp.Pair.CurrentBatchId)
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
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("5000000denom1,5000000denom2"), true)

	depositor := s.addr(1)
	s.deposit(depositor, pool.Id, squad.ParseCoins("250000denom1,250000denom2"), true)
	s.deposit(depositor, pool.Id, squad.ParseCoins("250000denom1,250000denom2"), true)
	s.deposit(depositor, pool.Id, squad.ParseCoins("250000denom1,250000denom2"), true)
	s.deposit(depositor, pool.Id, squad.ParseCoins("250000denom1,250000denom2"), true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryDepositRequestsRequest
		expectErr bool
		postRun   func(*types.QueryDepositRequestsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryDepositRequestsRequest{},
			true,
			nil,
		},
		{
			"query all deposit requests",
			&types.QueryDepositRequestsRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryDepositRequestsResponse) {
				s.Require().Len(resp.DepositRequests, 4)
			},
		},
		{
			"query all deposit requests",
			&types.QueryDepositRequestsRequest{
				PoolId: 2,
			},
			false,
			func(resp *types.QueryDepositRequestsResponse) {
				s.Require().Len(resp.DepositRequests, 0)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.DepositRequests(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCDepositRequest() {
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("5000000denom1,5000000denom2"), true)

	depositor := s.addr(1)
	req := s.deposit(depositor, pool.Id, squad.ParseCoins("250000denom1,250000denom2"), true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryDepositRequestRequest
		expectErr bool
		postRun   func(*types.QueryDepositRequestResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryDepositRequestRequest{},
			true,
			nil,
		},
		{
			"query the deposit request with just pool id",
			&types.QueryDepositRequestRequest{
				PoolId: 1,
			},
			true,
			nil,
		},
		{
			"query the deposit request with pool id",
			&types.QueryDepositRequestRequest{
				PoolId: 1,
				Id:     1,
			},
			false,
			func(resp *types.QueryDepositRequestResponse) {
				s.Require().Equal(req.Id, resp.DepositRequest.Id)
				s.Require().Equal(req.PoolId, resp.DepositRequest.PoolId)
				s.Require().Equal(req.MsgHeight, resp.DepositRequest.MsgHeight)
				s.Require().Equal(req.Depositor, resp.DepositRequest.Depositor)
				s.Require().Equal(req.DepositCoins, resp.DepositRequest.DepositCoins)
				s.Require().NotEqual(req.AcceptedCoins, resp.DepositRequest.AcceptedCoins)
				s.Require().NotEqual(req.MintedPoolCoin, resp.DepositRequest.MintedPoolCoin)
				s.Require().NotEqual(req.Status, resp.DepositRequest.Status)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.DepositRequest(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCWithdrawRequests() {
	params := s.keeper.GetParams(s.ctx)

	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("5000000denom1,5000000denom2"), true)
	poolCoinBalance := s.app.BankKeeper.GetBalance(s.ctx, creator, pool.PoolCoinDenom)
	s.Require().Equal(params.InitialPoolCoinSupply, poolCoinBalance.Amount)

	s.withdraw(creator, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 1000))
	s.withdraw(creator, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 2500))
	s.withdraw(creator, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 6000))
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryWithdrawRequestsRequest
		expectErr bool
		postRun   func(*types.QueryWithdrawRequestsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryWithdrawRequestsRequest{},
			true,
			nil,
		},
		{
			"query all withdraw requests",
			&types.QueryWithdrawRequestsRequest{
				PoolId: 1,
			},
			false,
			func(resp *types.QueryWithdrawRequestsResponse) {
				s.Require().Len(resp.WithdrawRequests, 3)
			},
		},
		{
			"query all withdraw requests",
			&types.QueryWithdrawRequestsRequest{
				PoolId: 2,
			},
			false,
			func(resp *types.QueryWithdrawRequestsResponse) {
				s.Require().Len(resp.WithdrawRequests, 0)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.WithdrawRequests(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCWithdrawRequest() {
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)
	pool := s.createPool(creator, pair.Id, squad.ParseCoins("5000000denom1,5000000denom2"), true)

	req := s.withdraw(creator, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 50000))
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryWithdrawRequestRequest
		expectErr bool
		postRun   func(*types.QueryWithdrawRequestResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryWithdrawRequestRequest{},
			true,
			nil,
		},
		{
			"query the withdraw request with only pool id",
			&types.QueryWithdrawRequestRequest{
				PoolId: 1,
			},
			true,
			nil,
		},
		{
			"query the withdraw request with only pool id",
			&types.QueryWithdrawRequestRequest{
				PoolId: 1,
				Id:     1,
			},
			false,
			func(resp *types.QueryWithdrawRequestResponse) {
				s.Require().Equal(req.Id, resp.WithdrawRequest.Id)
				s.Require().Equal(req.PoolId, resp.WithdrawRequest.PoolId)
				s.Require().Equal(req.MsgHeight, resp.WithdrawRequest.MsgHeight)
				s.Require().Equal(req.Withdrawer, resp.WithdrawRequest.Withdrawer)
				s.Require().Equal(req.PoolCoin, resp.WithdrawRequest.PoolCoin)
				s.Require().Equal(req.WithdrawnCoins, resp.WithdrawRequest.WithdrawnCoins)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.WithdrawRequest(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCOrders() {
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)

	s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), sdk.NewInt(5000000), 10*time.Second, true)
	s.sellLimitOrder(s.addr(2), pair.Id, squad.ParseDec("1.0"), newInt(10000), time.Hour, true)
	s.sellLimitOrder(s.addr(2), pair.Id, squad.ParseDec("1.0"), newInt(700000), time.Hour, true)
	s.buyLimitOrder(s.addr(2), pair.Id, squad.ParseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryOrdersRequest
		expectErr bool
		postRun   func(*types.QueryOrdersResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryOrdersRequest{},
			true,
			nil,
		},
		{
			"query all orders",
			&types.QueryOrdersRequest{
				PairId: 1,
			},
			false,
			func(resp *types.QueryOrdersResponse) {
				s.Require().Len(resp.Orders, 5)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Orders(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
func (s *KeeperTestSuite) TestGRPCOrder() {
	creator := s.addr(0)
	pair := s.createPair(creator, "denom1", "denom2", true)

	req := s.buyLimitOrder(s.addr(1), pair.Id, squad.ParseDec("1.0"), sdk.NewInt(1000000), 10*time.Second, true)
	liquidity.EndBlocker(s.ctx, s.keeper)

	for _, tc := range []struct {
		name      string
		req       *types.QueryOrderRequest
		expectErr bool
		postRun   func(*types.QueryOrderResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"invalid request",
			&types.QueryOrderRequest{},
			true,
			nil,
		},
		{
			"query the order",
			&types.QueryOrderRequest{
				PairId: 1,
				Id:     1,
			},
			false,
			func(resp *types.QueryOrderResponse) {
				s.Require().Equal(req.Id, resp.Order.Id)
				s.Require().Equal(req.PairId, resp.Order.PairId)
				s.Require().Equal(req.MsgHeight, resp.Order.MsgHeight)
				s.Require().Equal(req.Orderer, resp.Order.Orderer)
				s.Require().Equal(req.Direction, resp.Order.Direction)
				s.Require().Equal(req.OfferCoin, resp.Order.OfferCoin)
				s.Require().Equal(req.RemainingOfferCoin, resp.Order.RemainingOfferCoin)
				s.Require().Equal(req.ReceivedCoin, resp.Order.ReceivedCoin)
				s.Require().Equal(req.Price, resp.Order.Price)
				s.Require().Equal(req.Amount, resp.Order.Amount)
				s.Require().Equal(req.OpenAmount, resp.Order.OpenAmount)
				s.Require().Equal(req.BatchId, resp.Order.BatchId)
				s.Require().Equal(req.ExpireAt, resp.Order.ExpireAt)
				s.Require().NotEqual(req.Status, resp.Order.Status)
			},
		},
	} {
		s.Run(tc.name, func() {
			resp, err := s.querier.Order(sdk.WrapSDKContext(s.ctx), tc.req)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
