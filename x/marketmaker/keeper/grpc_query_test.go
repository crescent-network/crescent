package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	resp, err := suite.querier.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	suite.Require().Equal(suite.keeper.GetParams(suite.ctx), resp.Params)
}

func (suite *KeeperTestSuite) TestGRPCMarketMakers() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2, 3, 4, 5, 6})
	suite.NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{2, 3, 4, 5, 6, 7})
	suite.NoError(err)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description",
		[]types.MarketMakerHandle{
			{Address: mmAddr.String(), PairId: 3},
			{Address: mmAddr2.String(), PairId: 3}},
		nil, nil, nil)
	suite.handleProposal(proposal)

	for _, tc := range []struct {
		name      string
		req       *types.QueryMarketMakersRequest
		expectErr bool
		postRun   func(response *types.QueryMarketMakersResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryMarketMakersRequest{},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 12)
			},
		},
		{
			"query all with page limit",
			&types.QueryMarketMakersRequest{Pagination: &query.PageRequest{
				Limit: 10,
			}},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 10)
			},
		},
		{
			"query all with page offset",
			&types.QueryMarketMakersRequest{Pagination: &query.PageRequest{
				Offset: 10,
				Limit:  10,
			}},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 2)
			},
		},
		{
			"query with eligible true",
			&types.QueryMarketMakersRequest{Eligible: "true"},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 2)
				for _, mm := range resp.Marketmakers {
					suite.Require().True(mm.Eligible)
				}
			},
		},
		{
			"query with eligible false",
			&types.QueryMarketMakersRequest{Eligible: "false"},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 10)
				for _, mm := range resp.Marketmakers {
					suite.Require().False(mm.Eligible)
				}
			},
		},
		{
			"query with pair id",
			&types.QueryMarketMakersRequest{PairId: 3},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 2)
				for _, mm := range resp.Marketmakers {
					suite.Require().True(mm.Eligible)
				}
			},
		},
		{
			"query with addr",
			&types.QueryMarketMakersRequest{Address: mmAddr.String()},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 6)
				for _, mm := range resp.Marketmakers {
					suite.Require().Equal(mmAddr.String(), mm.Address)
				}
			},
		},
		{
			"query with addr and pair",
			&types.QueryMarketMakersRequest{Address: mmAddr.String(), PairId: 1},
			false,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 1)
				for _, mm := range resp.Marketmakers {
					suite.Require().Equal(mmAddr.String(), mm.Address)
					suite.Require().Equal(uint64(1), mm.PairId)
				}
			},
		},
		{
			"query with invalid addr",
			&types.QueryMarketMakersRequest{Address: "invalidaddr"},
			true,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 0)
			},
		},
		{
			"query with invalid eligible",
			&types.QueryMarketMakersRequest{Eligible: "invalidbool"},
			true,
			func(resp *types.QueryMarketMakersResponse) {
				suite.Require().Len(resp.Marketmakers, 0)
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.MarketMakers(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCIncentive() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[5].String()
	k.SetParams(ctx, params)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2, 3, 4, 5, 6})
	suite.NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{2, 3, 4, 5, 6, 7})
	suite.NoError(err)

	// distribute incentive
	incentiveAmount := sdk.NewInt(500000000)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))
	proposal := types.NewMarketMakerProposal("title", "description",
		[]types.MarketMakerHandle{
			{
				Address: mmAddr.String(),
				PairId:  1,
			},
		}, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	for _, tc := range []struct {
		name      string
		req       *types.QueryIncentiveRequest
		expectErr bool
		postRun   func(response *types.QueryIncentiveResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query empty",
			&types.QueryIncentiveRequest{},
			true,
			nil,
		},
		{
			"query with valid address",
			&types.QueryIncentiveRequest{Address: mmAddr.String()},
			false,
			func(resp *types.QueryIncentiveResponse) {
				suite.Require().Equal(resp.Incentive.Address, mmAddr.String())
				suite.Require().Equal(resp.Incentive.Claimable.String(), incentiveCoins.String())

			},
		},
		{
			"query with invalid address",
			&types.QueryIncentiveRequest{Address: "invalidaddr"},
			true,
			nil,
		},
		{
			"query with not exist address",
			&types.QueryIncentiveRequest{Address: mmAddr2.String()},
			true,
			nil,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Incentive(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}
