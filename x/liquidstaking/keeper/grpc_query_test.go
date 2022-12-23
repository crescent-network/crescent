package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/liquidstaking/types"
)

func (s *KeeperTestSuite) TestGRPCParams() {
	resp, err := s.querier.Params(sdk.WrapSDKContext(s.ctx), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.keeper.GetParams(s.ctx), resp.Params)
}

func (s *KeeperTestSuite) TestGRPCQueries() {
	vals, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	params.MinLiquidStakingAmount = sdk.NewInt(50000)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	// Test LiquidValidators grpc query
	res := s.keeper.GetAllLiquidValidatorStates(s.ctx)
	resp, err := s.querier.LiquidValidators(sdk.WrapSDKContext(s.ctx), &types.QueryLiquidValidatorsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(resp.LiquidValidators, res)

	resp, err = s.querier.LiquidValidators(sdk.WrapSDKContext(s.ctx), nil)
	s.Require().Nil(resp)
	s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))

	// Test States grpc query
	respStates, err := s.querier.States(sdk.WrapSDKContext(s.ctx), &types.QueryStatesRequest{})
	resNetAmountState := s.keeper.GetNetAmountState(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(respStates.NetAmountState, resNetAmountState)

	respStates, err = s.querier.States(sdk.WrapSDKContext(s.ctx), nil)
	s.Require().Nil(respStates)
	s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))

	// Test VotingPower grpc query
	respVotingPower, err := s.querier.VotingPower(sdk.WrapSDKContext(s.ctx), &types.QueryVotingPowerRequest{Voter: vals[0].String()})
	resVotingPower := s.keeper.GetVotingPower(s.ctx, vals[0])
	s.Require().NoError(err)
	s.Require().Equal(respVotingPower.VotingPower, resVotingPower)

	respVotingPower, err = s.querier.VotingPower(sdk.WrapSDKContext(s.ctx), nil)
	s.Require().Nil(respVotingPower)
	s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))

	respVotingPower, err = s.querier.VotingPower(sdk.WrapSDKContext(s.ctx), &types.QueryVotingPowerRequest{Voter: "invalidaddr"})
	s.Require().Nil(respVotingPower)
	s.Require().EqualError(err, "decoding bech32 failed: invalid separator index -1")
}
