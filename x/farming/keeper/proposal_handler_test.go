package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/keeper"
	"github.com/tendermint/farming/x/farming/types"

	_ "github.com/stretchr/testify/suite"
)

func (suite *KeeperTestSuite) TestAddPublicPlanProposal() {
	addrs := app.AddTestAddrs(suite.app, suite.ctx, 2, sdk.NewInt(100_000_000))
	farmerAddr := addrs[0]
	name := "test"
	terminationAddr := sdk.AccAddress("terminationAddr")
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// case1
	req := &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case1 := []*types.AddRequestProposal{req}

	// case2
	req = &types.AddRequestProposal{
		Name: `OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case2 := []*types.AddRequestProposal{req}

	// case3
	req = &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case3 := []*types.AddRequestProposal{req}

	// case4
	req = &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: sdk.NewDecCoins(
			sdk.DecCoin{
				Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
				Amount: sdk.MustNewDecFromStr("0.1"),
			},
		),
		StartTime:   mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:     mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount: sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:  sdk.ZeroDec(),
	}
	case4 := []*types.AddRequestProposal{req}

	// case5
	req = &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-13T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-06T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case5 := []*types.AddRequestProposal{req}

	// case6
	req = &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 1)),
		EpochRatio:         sdk.NewDec(1),
	}
	case6 := []*types.AddRequestProposal{req}

	// case7
	req = &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(),
		EpochRatio:         sdk.ZeroDec(),
	}
	case7 := []*types.AddRequestProposal{req}

	for _, tc := range []struct {
		name        string
		addRequest  []*types.AddRequestProposal
		expectedErr error
	}{
		{
			"happy case",
			case1,
			nil,
		},
		{
			"request case #1",
			nil,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"request case #2",
			[]*types.AddRequestProposal{},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"name case #1",
			case2,
			sdkerrors.Wrapf(types.ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", types.MaxNameLength),
		},
		{
			"staking coin weights case #1",
			case3,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty"),
		},
		{
			"staking coin weights case #2",
			case4,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1"),
		},
		{
			"start time & end time case #1",
			case5,
			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
				"end time %s must be greater than start time %s",
				mustParseRFC3339("2021-08-06T00:00:00Z"), mustParseRFC3339("2021-08-13T00:00:00Z")),
		},
		{
			"epoch amount & epoch ratio case #1",
			case6,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio should be provided"),
		},
		{
			"epoch amount & epoch ratio case #2",
			case7,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio must not be zero"),
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
				// TODO: need to check each field same as expected
				suite.Require().Equal(true, found)
			} else {
				suite.EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdatePublicPlanProposal() {
	addrs := app.AddTestAddrs(suite.app, suite.ctx, 2, sdk.NewInt(100_000_000))
	farmerAddr := addrs[0]
	name := "test"
	terminationAddr := sdk.AccAddress("terminationAddr")
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// add request proposal
	addReq := &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	addRequests := []*types.AddRequestProposal{addReq}

	proposal := &types.PublicPlanProposal{
		Title:                  "testTitle",
		Description:            "testDescription",
		AddRequestProposals:    addRequests,
		UpdateRequestProposals: nil,
		DeleteRequestProposals: nil,
	}

	err := proposal.ValidateBasic()
	suite.Require().NoError(err)

	err = keeper.HandlePublicPlanProposal(suite.ctx, suite.keeper, proposal)
	suite.Require().NoError(err)

	_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)

	// case1
	startTime := mustParseRFC3339("2021-08-06T00:00:00Z")
	endTime := mustParseRFC3339("2021-08-13T00:00:00Z")

	req := &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case1 := []*types.UpdateRequestProposal{req}

	// case2
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(0),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case2 := []*types.UpdateRequestProposal{req}

	// case3
	req = &types.UpdateRequestProposal{
		PlanId: uint64(1),
		Name: `OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM
		OVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERMOVERMAXLENGTHOVERMAXLENGTHOVERMAXLENGTHOVERM`,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case3 := []*types.UpdateRequestProposal{req}

	// case4
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: sdk.NewDecCoins(),
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case4 := []*types.UpdateRequestProposal{req}

	// case5
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: sdk.NewDecCoins(
			sdk.DecCoin{
				Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
				Amount: sdk.MustNewDecFromStr("0.1"),
			},
		),
		StartTime:   &startTime,
		EndTime:     &endTime,
		EpochAmount: sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:  sdk.ZeroDec(),
	}
	case5 := []*types.UpdateRequestProposal{req}

	// case6
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &endTime,
		EndTime:            &startTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
		EpochRatio:         sdk.ZeroDec(),
	}
	case6 := []*types.UpdateRequestProposal{req}

	// case7
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 1)),
		EpochRatio:         sdk.NewDec(1),
	}
	case7 := []*types.UpdateRequestProposal{req}

	// case8
	req = &types.UpdateRequestProposal{
		PlanId:             uint64(1),
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          &startTime,
		EndTime:            &endTime,
		EpochAmount:        sdk.NewCoins(),
		EpochRatio:         sdk.ZeroDec(),
	}
	case8 := []*types.UpdateRequestProposal{req}

	for _, tc := range []struct {
		name          string
		updateRequest []*types.UpdateRequestProposal
		expectedErr   error
	}{
		{
			"happy case",
			case1,
			nil,
		},
		{
			"request case #1",
			nil,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"request case #2",
			[]*types.UpdateRequestProposal{},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proposal request must not be empty"),
		},
		{
			"plan id case #1",
			case2,
			sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid plan id: %d", uint64(0)),
		},
		{
			"name case #1",
			case3,
			sdkerrors.Wrapf(types.ErrInvalidPlanNameLength, "plan name cannot be longer than max length of %d", types.MaxNameLength),
		},
		{
			"staking coin weights case #1",
			case4,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin weights must not be empty"),
		},
		{
			"staking coin weights case #2",
			case5,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total weight must be 1"),
		},
		{
			"start time & end time case #1",
			case6,
			sdkerrors.Wrapf(types.ErrInvalidPlanEndTime,
				"end time %s must be greater than start time %s",
				mustParseRFC3339("2021-08-06T00:00:00Z"), mustParseRFC3339("2021-08-13T00:00:00Z")),
		},
		{
			"epoch amount & epoch ratio case #1",
			case7,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio should be provided"),
		},
		{
			"epoch amount & epoch ratio case #2",
			case8,
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "either epoch amount or epoch ratio must not be zero"),
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

				_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
				suite.Require().Equal(true, found)
			} else {
				suite.EqualError(err, tc.expectedErr.Error())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeletePublicPlanProposal() {
	addrs := app.AddTestAddrs(suite.app, suite.ctx, 2, sdk.NewInt(100_000_000))
	farmerAddr := addrs[0]
	name := "test"
	terminationAddr := sdk.AccAddress("terminationAddr")
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{
			Denom:  "poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4",
			Amount: sdk.MustNewDecFromStr("1.0"),
		},
	)

	// add request proposal
	addReq := &types.AddRequestProposal{
		Name:               name,
		FarmingPoolAddress: farmerAddr.String(),
		TerminationAddress: terminationAddr.String(),
		StakingCoinWeights: coinWeights,
		StartTime:          mustParseRFC3339("2021-08-06T00:00:00Z"),
		EndTime:            mustParseRFC3339("2021-08-13T00:00:00Z"),
		EpochAmount:        sdk.NewCoins(sdk.NewInt64Coin("uatom", 100_000_000)),
		EpochRatio:         sdk.ZeroDec(),
	}
	addRequests := []*types.AddRequestProposal{addReq}

	proposal := &types.PublicPlanProposal{
		Title:                  "testTitle",
		Description:            "testDescription",
		AddRequestProposals:    addRequests,
		UpdateRequestProposals: nil,
		DeleteRequestProposals: nil,
	}

	err := proposal.ValidateBasic()
	suite.Require().NoError(err)

	err = keeper.HandlePublicPlanProposal(suite.ctx, suite.keeper, proposal)
	suite.Require().NoError(err)

	_, found := suite.keeper.GetPlan(suite.ctx, uint64(1))
	suite.Require().Equal(true, found)

	// delete the proposal
	req := &types.DeleteRequestProposal{
		PlanId: uint64(1),
	}
	proposals := []*types.DeleteRequestProposal{req}

	err = suite.keeper.DeletePublicPlanProposal(suite.ctx, proposals)
	suite.Require().NoError(err)
}
