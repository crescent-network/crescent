package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

//func (suite *KeeperTestSuite) TestGRPCLiquidValidators() {
//	liquidValidators := []types.LiquidValidator{
//		{
//			Name:               "liquidValidator1",
//			Rate:               sdk.NewDecWithPrec(5, 2),
//			SourceAddress:      suite.sourceAddrs[0].String(),
//			DestinationAddress: suite.destinationAddrs[0].String(),
//			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//		},
//		{
//			Name:               "liquidValidator2",
//			Rate:               sdk.NewDecWithPrec(5, 2),
//			SourceAddress:      suite.sourceAddrs[0].String(),
//			DestinationAddress: suite.destinationAddrs[1].String(),
//			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//		},
//		{
//			Name:               "liquidValidator3",
//			Rate:               sdk.NewDecWithPrec(5, 2),
//			SourceAddress:      suite.sourceAddrs[1].String(),
//			DestinationAddress: suite.destinationAddrs[0].String(),
//			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//		},
//		{
//			Name:               "liquidValidator4",
//			Rate:               sdk.NewDecWithPrec(5, 2),
//			SourceAddress:      suite.sourceAddrs[1].String(),
//			DestinationAddress: suite.destinationAddrs[1].String(),
//			StartTime:          types.MustParseRFC3339("0000-01-01T00:00:00Z"),
//			EndTime:            types.MustParseRFC3339("9999-12-31T00:00:00Z"),
//		},
//	}
//
//	params := suite.keeper.GetParams(suite.ctx)
//	params.LiquidValidators = liquidValidators
//	suite.keeper.SetParams(suite.ctx, params)
//
//	balance := suite.app.BankKeeper.GetAllBalances(suite.ctx, suite.sourceAddrs[0])
//	expectedCoins, _ := sdk.NewDecCoinsFromCoins(balance...).MulDec(sdk.NewDecWithPrec(5, 2)).TruncateDecimal()
//
//	suite.ctx = suite.ctx.WithBlockTime(types.MustParseRFC3339("2021-08-31T00:00:00Z"))
//	err := suite.keeper.CollectLiquidValidators(suite.ctx)
//	suite.Require().NoError(err)
//
//	for _, tc := range []struct {
//		name      string
//		req       *types.QueryLiquidValidatorsRequest
//		expectErr bool
//		postRun   func(response *types.QueryLiquidValidatorsResponse)
//	}{
//		{
//			"nil request",
//			nil,
//			true,
//			nil,
//		},
//		{
//			"query all",
//			&types.QueryLiquidValidatorsRequest{},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 4)
//			},
//		},
//		{
//			"query by not existing name",
//			&types.QueryLiquidValidatorsRequest{Name: "notfound"},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 0)
//			},
//		},
//		{
//			"query by name",
//			&types.QueryLiquidValidatorsRequest{Name: "liquidValidator1"},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 1)
//				suite.Require().Equal("liquidValidator1", resp.LiquidValidators[0].LiquidValidator.Name)
//			},
//		},
//		{
//			"invalid source addr",
//			&types.QueryLiquidValidatorsRequest{SourceAddress: "invalid"},
//			true,
//			nil,
//		},
//		{
//			"query by source addr",
//			&types.QueryLiquidValidatorsRequest{SourceAddress: suite.sourceAddrs[0].String()},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 2)
//				for _, b := range resp.LiquidValidators {
//					suite.Require().Equal(suite.sourceAddrs[0].String(), b.LiquidValidator.SourceAddress)
//				}
//			},
//		},
//		{
//			"invalid destination addr",
//			&types.QueryLiquidValidatorsRequest{DestinationAddress: "invalid"},
//			true,
//			nil,
//		},
//		{
//			"query by destination addr",
//			&types.QueryLiquidValidatorsRequest{DestinationAddress: suite.destinationAddrs[0].String()},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 2)
//				for _, b := range resp.LiquidValidators {
//					suite.Require().Equal(suite.destinationAddrs[0].String(), b.LiquidValidator.DestinationAddress)
//				}
//			},
//		},
//		{
//			"query with multiple filters",
//			&types.QueryLiquidValidatorsRequest{
//				SourceAddress:      suite.sourceAddrs[0].String(),
//				DestinationAddress: suite.destinationAddrs[1].String(),
//			},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 1)
//				suite.Require().Equal(suite.sourceAddrs[0].String(), resp.LiquidValidators[0].LiquidValidator.SourceAddress)
//				suite.Require().Equal(suite.destinationAddrs[1].String(), resp.LiquidValidators[0].LiquidValidator.DestinationAddress)
//			},
//		},
//		{
//			"correct total collected coins",
//			&types.QueryLiquidValidatorsRequest{Name: "liquidValidator1"},
//			false,
//			func(resp *types.QueryLiquidValidatorsResponse) {
//				suite.Require().Len(resp.LiquidValidators, 1)
//				suite.Require().True(coinsEq(expectedCoins, resp.LiquidValidators[0].TotalCollectedCoins))
//			},
//		},
//	} {
//		suite.Run(tc.name, func() {
//			resp, err := suite.querier.LiquidValidators(sdk.WrapSDKContext(suite.ctx), tc.req)
//			if tc.expectErr {
//				suite.Require().Error(err)
//			} else {
//				suite.Require().NoError(err)
//				tc.postRun(resp)
//			}
//		})
//	}
//}
//
//func (suite *KeeperTestSuite) TestGRPCAddresses() {
//	for _, tc := range []struct {
//		name         string
//		req          *types.QueryAddressesRequest
//		expectedAddr string
//		expectErr    bool
//	}{
//		{
//			"nil request",
//			nil,
//			"",
//			true,
//		},
//		{
//			"empty request",
//			&types.QueryAddressesRequest{},
//			"",
//			true,
//		},
//		{
//			"default module name and address type",
//			&types.QueryAddressesRequest{Name: "testSourceAddr"},
//			"cosmos1hg0v9u92ztzecpmml26206wwtghggx0flpwn5d4qc3r6dvuanxeqs4mnk5",
//			false,
//		},
//		{
//			"invalid address type",
//			&types.QueryAddressesRequest{Name: "testSourceAddr", Type: 2},
//			"",
//			true,
//		},
//	} {
//		suite.Run(tc.name, func() {
//			resp, err := suite.querier.Addresses(sdk.WrapSDKContext(suite.ctx), tc.req)
//			if tc.expectErr {
//				suite.Require().Error(err)
//			} else {
//				suite.Require().NoError(err)
//				suite.Require().Equal(resp.Address, tc.expectedAddr)
//			}
//		})
//	}
//}
