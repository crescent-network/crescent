package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v5/x/budget/types"
)

func (suite *KeeperTestSuite) TestGRPCParams() {
	resp, err := suite.querier.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.keeper.GetParams(suite.ctx), resp.Params)
}

func (suite *KeeperTestSuite) TestGRPCBudgets() {
	budgets := []types.Budget{
		{
			Name:               "budget1",
			Rate:               sdk.NewDecWithPrec(5, 2),
			SourceAddress:      suite.sourceAddrs[0].String(),
			DestinationAddress: suite.destinationAddrs[0].String(),
			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
		},
		{
			Name:               "budget2",
			Rate:               sdk.NewDecWithPrec(5, 2),
			SourceAddress:      suite.sourceAddrs[0].String(),
			DestinationAddress: suite.destinationAddrs[1].String(),
			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
		},
		{
			Name:               "budget3",
			Rate:               sdk.NewDecWithPrec(5, 2),
			SourceAddress:      suite.sourceAddrs[1].String(),
			DestinationAddress: suite.destinationAddrs[0].String(),
			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
		},
		{
			Name:               "budget4",
			Rate:               sdk.NewDecWithPrec(5, 2),
			SourceAddress:      suite.sourceAddrs[1].String(),
			DestinationAddress: suite.destinationAddrs[1].String(),
			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
		},
	}

	params := suite.keeper.GetParams(suite.ctx)
	params.Budgets = budgets
	suite.keeper.SetParams(suite.ctx, params)

	balance := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.sourceAddrs[0])
	expectedCoins, _ := sdk.NewDecCoinsFromCoins(balance...).MulDec(sdk.NewDecWithPrec(5, 2)).TruncateDecimal()

	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-31T00:00:00Z"))
	err := suite.keeper.CollectBudgets(suite.ctx)
	suite.Require().NoError(err)

	for _, tc := range []struct {
		name      string
		req       *types.QueryBudgetsRequest
		expectErr bool
		postRun   func(response *types.QueryBudgetsResponse)
	}{
		{
			"nil request",
			nil,
			true,
			nil,
		},
		{
			"query all",
			&types.QueryBudgetsRequest{},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 4)
			},
		},
		{
			"query by not existing name",
			&types.QueryBudgetsRequest{Name: "notfound"},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 0)
			},
		},
		{
			"query by name",
			&types.QueryBudgetsRequest{Name: "budget1"},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 1)
				suite.Require().Equal("budget1", resp.Budgets[0].Budget.Name)
			},
		},
		{
			"invalid source addr",
			&types.QueryBudgetsRequest{SourceAddress: "invalid"},
			true,
			nil,
		},
		{
			"query by source addr",
			&types.QueryBudgetsRequest{SourceAddress: suite.sourceAddrs[0].String()},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 2)
				for _, b := range resp.Budgets {
					suite.Require().Equal(suite.sourceAddrs[0].String(), b.Budget.SourceAddress)
				}
			},
		},
		{
			"invalid destination addr",
			&types.QueryBudgetsRequest{DestinationAddress: "invalid"},
			true,
			nil,
		},
		{
			"query by destination addr",
			&types.QueryBudgetsRequest{DestinationAddress: suite.destinationAddrs[0].String()},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 2)
				for _, b := range resp.Budgets {
					suite.Require().Equal(suite.destinationAddrs[0].String(), b.Budget.DestinationAddress)
				}
			},
		},
		{
			"query with multiple filters",
			&types.QueryBudgetsRequest{
				SourceAddress:      suite.sourceAddrs[0].String(),
				DestinationAddress: suite.destinationAddrs[1].String(),
			},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 1)
				suite.Require().Equal(suite.sourceAddrs[0].String(), resp.Budgets[0].Budget.SourceAddress)
				suite.Require().Equal(suite.destinationAddrs[1].String(), resp.Budgets[0].Budget.DestinationAddress)
			},
		},
		{
			"correct total collected coins",
			&types.QueryBudgetsRequest{Name: "budget1"},
			false,
			func(resp *types.QueryBudgetsResponse) {
				suite.Require().Len(resp.Budgets, 1)
				suite.Require().True(coinsEq(expectedCoins, resp.Budgets[0].TotalCollectedCoins))
			},
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Budgets(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.postRun(resp)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCAddresses() {
	for _, tc := range []struct {
		name         string
		req          *types.QueryAddressesRequest
		expectedAddr string
		expectErr    bool
	}{
		{
			"nil request",
			nil,
			"",
			true,
		},
		{
			"empty request",
			&types.QueryAddressesRequest{},
			"",
			true,
		},
		{
			"default module name and address type",
			&types.QueryAddressesRequest{Name: "testSourceAddr"},
			"cosmos1hg0v9u92ztzecpmml26206wwtghggx0flpwn5d4qc3r6dvuanxeqs4mnk5",
			false,
		},
		{
			"invalid address type",
			&types.QueryAddressesRequest{Name: "testSourceAddr", Type: 2},
			"",
			true,
		},
	} {
		suite.Run(tc.name, func() {
			resp, err := suite.querier.Addresses(sdk.WrapSDKContext(suite.ctx), tc.req)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(resp.Address, tc.expectedAddr)
			}
		})
	}
}
