package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestValidateAddPublicPlanProposal() {
	for _, tc := range []struct {
		name        string
		addRequest  []*types.AddRequestProposal
		expectedErr error
	}{
		{
			"happy case",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
				sdk.ZeroDec(),
			)},
			nil,
		},
		{
			"request case #1",
			nil,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"name case #1",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				`OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
				sdk.ZeroDec(),
			)},
			sdkerrors.Wrapf(types.ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", types.MaxNameLength),
		},
		{
			"staking coin weights case #1",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
				sdk.ZeroDec(),
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty"),
		},
		{
			"staking coin weights case #2",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.DecCoin{
						Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
						Amount: sdk.MustNewDecFromStr("0.1"),
					},
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
				sdk.ZeroDec(),
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1"),
		},
		{
			"start time & end time case #1",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				types.ParseTime("2021-08-13T00:00:00Z"),
				types.ParseTime("2021-08-06T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 100_000_000)),
				sdk.ZeroDec(),
			)},
			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
				"end time %s must be greater than start time %s",
				types.ParseTime("2021-08-06T00:00:00Z"), types.ParseTime("2021-08-13T00:00:00Z")),
		},
		{
			"epoch amount & epoch ratio case #1",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(sdk.NewInt64Coin(denom3, 1)),
				sdk.NewDec(1),
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only one of epoch amount or epoch ratio must be provided"),
		},
		{
			"epoch amount & epoch ratio case #2",
			[]*types.AddRequestProposal{types.NewAddRequestProposal(
				"testPlan",
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)),
				),
				types.ParseTime("2021-08-01T00:00:00Z"),
				types.ParseTime("2021-08-30T00:00:00Z"),
				sdk.NewCoins(),
				sdk.ZeroDec(),
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only one of epoch amount or epoch ratio must be provided"),
		},
	} {
		suite.Run(tc.name, func() {
			proposal := &types.PublicPlanProposal{
				Title:                  "testTitle",
				Description:            "testDescription",
				AddRequestProposals:    tc.addRequest,
				UpdateRequestProposals: nil,
				DeleteRequestProposals: nil,
			}

			err := proposal.ValidateBasic()
			if err == nil {
				err := keeper.HandlePublicPlanProposal(suite.ctx, suite.keeper, proposal)
				suite.Require().NoError(err)

				_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
				suite.Require().Equal(true, found)
			} else {
				suite.EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestValidateUpdatePublicPlanProposal() {
	// create a ratio public plan
	addRequests := []*types.AddRequestProposal{
		types.NewAddRequestProposal(
			"testPlan",
			suite.addrs[0].String(),
			suite.addrs[0].String(),
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
				sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
			),
			types.ParseTime("2021-08-01T00:00:00Z"),
			types.ParseTime("2021-08-30T00:00:00Z"),
			nil,
			sdk.NewDecWithPrec(10, 2), // 10%
		),
	}

	err := keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", addRequests, nil, nil),
	)
	suite.Require().NoError(err)

	plan, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)

	for _, tc := range []struct {
		name          string
		updateRequest []*types.UpdateRequestProposal
		expectedErr   error
	}{
		{
			"happy case #1 - decrease epoch ratio to 5%",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				plan.GetId(),
				plan.GetName(),
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				plan.GetStakingCoinWeights(),
				plan.GetStartTime(),
				plan.GetEndTime(),
				nil,
				sdk.NewDecWithPrec(5, 2),
			)},
			nil,
		},
		{
			"request case #1",
			nil,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"plan id case #1",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				uint64(0),
				plan.GetName(),
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				plan.GetStakingCoinWeights(),
				plan.GetStartTime(),
				plan.GetEndTime(),
				nil,
				plan.(*types.RatioPlan).EpochRatio,
			)},
			sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", uint64(0)),
		},
		{
			"name case #1",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				plan.GetId(),
				`OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`, // max length of name
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				plan.GetStakingCoinWeights(),
				plan.GetStartTime(),
				plan.GetEndTime(),
				nil,
				plan.(*types.RatioPlan).EpochRatio,
			)},
			sdkerrors.Wrapf(types.ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", types.MaxNameLength),
		},
		{
			"staking coin weights case #1",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				plan.GetId(),
				plan.GetName(),
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				sdk.NewDecCoins(),
				plan.GetStartTime(),
				plan.GetEndTime(),
				nil,
				plan.(*types.RatioPlan).EpochRatio,
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty"),
		},
		{
			"staking coin weights case #2",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				plan.GetId(),
				plan.GetName(),
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				sdk.NewDecCoins(
					sdk.DecCoin{
						Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
						Amount: sdk.MustNewDecFromStr("0.1"),
					},
				),
				plan.GetStartTime(),
				plan.GetEndTime(),
				nil,
				plan.(*types.RatioPlan).EpochRatio,
			)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1"),
		},
		{
			"start time & end time case #1",
			[]*types.UpdateRequestProposal{types.NewUpdateRequestProposal(
				plan.GetId(),
				plan.GetName(),
				plan.GetFarmingPoolAddress().String(),
				plan.GetTerminationAddress().String(),
				plan.GetStakingCoinWeights(),
				types.ParseTime("2021-08-13T00:00:00Z"),
				types.ParseTime("2021-08-06T00:00:00Z"),
				nil,
				plan.(*types.RatioPlan).EpochRatio,
			)},
			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
				"end time %s must be greater than start time %s",
				types.ParseTime("2021-08-06T00:00:00Z"), types.ParseTime("2021-08-13T00:00:00Z")),
		},
		{
			"epoch amount & epoch ratio case #1",
			[]*types.UpdateRequestProposal{
				types.NewUpdateRequestProposal(
					plan.GetId(),
					plan.GetName(),
					plan.GetFarmingPoolAddress().String(),
					plan.GetTerminationAddress().String(),
					plan.GetStakingCoinWeights(),
					plan.GetStartTime(),
					plan.GetEndTime(),
					sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)),
					plan.(*types.RatioPlan).EpochRatio,
				)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only one of epoch amount or epoch ratio must be provided"),
		},
		{
			"epoch amount & epoch ratio case #2",
			[]*types.UpdateRequestProposal{
				types.NewUpdateRequestProposal(
					plan.GetId(),
					plan.GetName(),
					plan.GetFarmingPoolAddress().String(),
					plan.GetTerminationAddress().String(),
					plan.GetStakingCoinWeights(),
					plan.GetStartTime(),
					plan.GetEndTime(),
					sdk.NewCoins(sdk.NewInt64Coin("stake", 0)),
					sdk.ZeroDec(),
				)},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only one of epoch amount or epoch ratio must be provided"),
		},
	} {
		suite.Run(tc.name, func() {
			proposal := &types.PublicPlanProposal{
				Title:                  "testTitle",
				Description:            "testDescription",
				AddRequestProposals:    nil,
				UpdateRequestProposals: tc.updateRequest,
				DeleteRequestProposals: nil,
			}

			err := proposal.ValidateBasic()
			if err == nil {
				err := keeper.HandlePublicPlanProposal(suite.ctx, suite.keeper, proposal)
				suite.Require().NoError(err)

				_, found := suite.keeper.GetPlan(suite.ctx, tc.updateRequest[0].GetPlanId())
				suite.Require().Equal(true, found)
			} else {
				suite.EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestValidateDeletePublicPlanProposal() {
	// create a ratio public plan
	addRequests := []*types.AddRequestProposal{types.NewAddRequestProposal(
		"testPlan",
		suite.addrs[0].String(),
		suite.addrs[0].String(),
		sdk.NewDecCoins(
			sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
			sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
		),
		types.ParseTime("2021-08-01T00:00:00Z"),
		types.ParseTime("2021-08-30T00:00:00Z"),
		nil,
		sdk.NewDecWithPrec(10, 2), // 10%
	)}

	err := keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", addRequests, nil, nil),
	)
	suite.Require().NoError(err)

	// should exist
	_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)

	// delete the proposal
	deleteRequests := []*types.DeleteRequestProposal{types.NewDeleteRequestProposal(uint64(1))}

	err = keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", nil, nil, deleteRequests),
	)
	suite.Require().NoError(err)

	// shouldn't exist
	_, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(false, found)
}

func (suite *KeeperTestSuite) TestUpdatePlanType() {
	// create a ratio public plan
	addRequests := []*types.AddRequestProposal{
		types.NewAddRequestProposal(
			"testPlan",
			suite.addrs[0].String(),
			suite.addrs[0].String(),
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)), // 30%
				sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1)), // 70%
			),
			types.ParseTime("2021-08-01T00:00:00Z"),
			types.ParseTime("2021-08-30T00:00:00Z"),
			nil,
			sdk.NewDecWithPrec(10, 2), // 10%
		),
	}

	err := keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", addRequests, nil, nil),
	)
	suite.Require().NoError(err)

	plan, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)
	suite.Require().Equal(plan.(*types.RatioPlan).EpochRatio, sdk.NewDecWithPrec(10, 2))

	// update the ratio plan to fixed amount plan type
	updateRequests := []*types.UpdateRequestProposal{
		types.NewUpdateRequestProposal(
			plan.GetId(),
			plan.GetName(),
			plan.GetFarmingPoolAddress().String(),
			plan.GetTerminationAddress().String(),
			plan.GetStakingCoinWeights(),
			plan.GetStartTime(),
			plan.GetEndTime(),
			sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)),
			sdk.ZeroDec(),
		)}

	err = keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", nil, updateRequests, nil),
	)
	suite.Require().NoError(err)

	plan, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)
	suite.Require().Equal(plan.(*types.FixedAmountPlan).EpochAmount, sdk.NewCoins(sdk.NewInt64Coin("stake", 100_000)))

	// update the fixed amount plan back to ratio plan
	updateRequests = []*types.UpdateRequestProposal{
		types.NewUpdateRequestProposal(
			plan.GetId(),
			plan.GetName(),
			plan.GetFarmingPoolAddress().String(),
			plan.GetTerminationAddress().String(),
			plan.GetStakingCoinWeights(),
			plan.GetStartTime(),
			plan.GetEndTime(),
			nil,
			sdk.NewDecWithPrec(7, 2), // 7%
		)}

	err = keeper.HandlePublicPlanProposal(
		suite.ctx,
		suite.keeper,
		types.NewPublicPlanProposal("testTitle", "testDescription", nil, updateRequests, nil),
	)
	suite.Require().NoError(err)

	plan, found = suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)
	suite.Require().Equal(plan.(*types.RatioPlan).EpochRatio, sdk.NewDecWithPrec(7, 2))
}
