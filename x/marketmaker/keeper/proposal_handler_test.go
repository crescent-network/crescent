package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v4/x/marketmaker/types"
)

func (suite *KeeperTestSuite) TestMarketMakerProposal() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	params := k.GetParams(ctx)

	balanceBeforeModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	balanceBeforeMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.Require().NoError(err)

	balanceAfterModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	balanceAfterMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(balanceBeforeModuleAcc.Add(params.DepositAmount...), balanceAfterModuleAcc)
	suite.EqualValues(balanceBeforeMM.Sub(params.DepositAmount), balanceAfterMM)

	mm, found := k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.False(mm.Eligible)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
	suite.handleProposal(proposal)

	mm, found = k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.True(mm.Eligible)

	// fail include market maker already eligible
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrInvalidInclusion)

	// refunded deposit amount
	balanceAfterModuleAcc2 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	balanceAfterMM2 := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(sdk.NewCoins(), balanceAfterModuleAcc2)
	suite.EqualValues(balanceBeforeMM, balanceAfterMM2)

	// fail include not existed market maker
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 5}}, nil, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrNotExistMarketMaker)

	// fail reject market maker already eligible
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrInvalidRejection)

	// fail reject market maker not exist
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 20}}, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrNotExistMarketMaker)

	// exclude market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil)
	suite.handleProposal(proposal)

	// not refunded when exclusion
	balanceAfterModuleAcc3 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	balanceAfterMM3 := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(sdk.NewCoins(), balanceAfterModuleAcc3)
	suite.EqualValues(balanceBeforeMM, balanceAfterMM3)

	mm, found = k.GetMarketMaker(ctx, mmAddr, 1)
	suite.False(found)

	// fail exclude not existed market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 5}}, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrNotExistMarketMaker)

	// apply market maker
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{2})
	suite.Require().NoError(err)

	// fail exclude not eligible market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrInvalidExclusion)

	// reject market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil)
	suite.handleProposal(proposal)

	// refunded when rejection
	balanceAfterModuleAcc4 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	balanceAfterMM4 := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(sdk.NewCoins(), balanceAfterModuleAcc4)
	suite.EqualValues(balanceBeforeMM, balanceAfterMM4)

	// fail invalid market maker address
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: "invalidaddr", PairId: 1}}, nil, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().Error(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().Error(err)
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: "invalidaddr", PairId: 1}}, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().Error(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().Error(err)

	// fail empty market maker proposal
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)

	// fail due to duplicated market maker
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}, {Address: mmAddr.String(), PairId: 2}}, nil, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().Error(err)

	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil)
	err = proposal.ValidateBasic()
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestMarketMakerProposalDistribution() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[5].String()
	k.SetParams(ctx, params)

	balanceInitMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2})
	suite.Require().NoError(err)

	err = k.ClaimIncentives(ctx, mmAddr)
	suite.ErrorIs(err, types.ErrEmptyClaimableIncentive)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
	suite.handleProposal(proposal)

	// no incentive yet
	_, found := k.GetIncentive(ctx, mmAddr)
	suite.False(found)

	incentiveAmount := sdk.NewInt(500000000)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))

	balanceBeforeMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceBeforeBudget := suite.app.BankKeeper.GetAllBalances(ctx, params.IncentiveBudgetAcc())
	balanceBeforeReserveAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)

	// submit empty distribution proposal
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  sdk.Coins{},
			},
		})
	err = proposal.ValidateBasic()
	suite.Require().Error(err)

	// submit incentive distribution proposal
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	balanceAfterMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceAfterBudget := suite.app.BankKeeper.GetAllBalances(ctx, params.IncentiveBudgetAcc())
	balanceAfterReserveAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)

	suite.Require().EqualValues(balanceAfterMM, balanceBeforeMM)
	suite.Require().EqualValues(balanceAfterBudget, balanceBeforeBudget.Sub(incentiveCoins))
	suite.Require().EqualValues(balanceAfterReserveAcc, balanceBeforeReserveAcc.Add(incentiveCoins...))

	incentive, found := k.GetIncentive(ctx, mmAddr)
	suite.True(found)
	suite.Equal(mmAddr.String(), incentive.Address)
	suite.Equal(incentiveCoins, incentive.Claimable)

	// submit incentive distribution proposal again with multiple pairs and not eligible market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
			{
				Address: mmAddr.String(),
				PairId:  2,
				Amount:  incentiveCoins,
			},
		})
	err = proposal.ValidateBasic()
	suite.Require().NoError(err)
	err = suite.govHandler(suite.ctx, proposal)
	suite.Require().ErrorIs(err, types.ErrNotEligibleMarketMaker)

	// include market maker
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil, nil)
	suite.handleProposal(proposal)

	// submit incentive distribution proposal again with multiple pairs with eligible market makers
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
			{
				Address: mmAddr.String(),
				PairId:  2,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	balanceAfterMM = suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceAfterBudget = suite.app.BankKeeper.GetAllBalances(ctx, params.IncentiveBudgetAcc())
	balanceAfterReserveAcc = suite.app.BankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)

	suite.Require().EqualValues(balanceAfterMM, balanceInitMM)
	suite.Require().EqualValues(balanceAfterBudget, balanceBeforeBudget.Sub(incentiveCoins).Sub(incentiveCoins).Sub(incentiveCoins))
	suite.Require().EqualValues(balanceAfterReserveAcc, balanceBeforeReserveAcc.Add(incentiveCoins...).Add(incentiveCoins...).Add(incentiveCoins...))

	incentive, found = k.GetIncentive(ctx, mmAddr)
	suite.True(found)
	suite.Equal(mmAddr.String(), incentive.Address)
	suite.Equal(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount.MulRaw(3))), incentive.Claimable)

	// claim incentives
	err = k.ClaimIncentives(ctx, mmAddr)
	suite.NoError(err)
	balanceAfterMM = suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceAfterReserveAcc = suite.app.BankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)
	suite.Equal(balanceAfterMM, balanceInitMM.Add(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount.MulRaw(3))))
	suite.Equal(balanceBeforeReserveAcc, balanceAfterReserveAcc)

	// claimed all incentives, no object
	_, found = k.GetIncentive(ctx, mmAddr)
	suite.False(found)

	// claim after exclusion
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  2,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil)
	suite.handleProposal(proposal)
	err = k.ClaimIncentives(ctx, mmAddr)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestMarketMakerProposalHugeCase() {
	ctx := suite.ctx
	k := suite.keeper

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[29].String()
	k.SetParams(ctx, params)

	incentiveAmount := sdk.NewInt(100)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))

	//balanceInitMM := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)

	// apply market maker
	for i := 0; i < 25; i++ {
		err := k.ApplyMarketMaker(ctx, suite.addrs[i], []uint64{1, 2, 3, 4, 5, 6, 7})
		suite.Require().NoError(err)
	}

	proposal := types.NewMarketMakerProposal("title", "description", nil, nil, nil, nil)

	// include, distribute market maker
	for i := 0; i < 20; i++ {
		for pairId := 1; pairId <= 7; pairId++ {
			proposal.Inclusions = append(proposal.Inclusions, types.MarketMakerHandle{
				Address: suite.addrs[i].String(),
				PairId:  uint64(pairId),
			})

			proposal.Distributions = append(proposal.Distributions, types.IncentiveDistribution{
				Address: suite.addrs[i].String(),
				PairId:  uint64(pairId),
				Amount:  incentiveCoins,
			})
		}
	}

	// reject market maker
	for i := 20; i < 25; i++ {
		for pairId := 1; pairId <= 7; pairId++ {
			proposal.Rejections = append(proposal.Rejections, types.MarketMakerHandle{
				Address: suite.addrs[i].String(),
				PairId:  uint64(pairId),
			})
		}
	}
	suite.handleProposal(proposal)

	// assert incentive amount
	for i := 0; i < 20; i++ {
		incentive, found := k.GetIncentive(ctx, suite.addrs[i])
		suite.True(found)
		suite.Equal(suite.addrs[i].String(), incentive.Address)
		suite.Equal(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount.MulRaw(7))), incentive.Claimable)
	}

	// assert no incentive
	for i := 20; i < 25; i++ {
		_, found := k.GetIncentive(ctx, suite.addrs[i])
		suite.False(found)
	}
}

func (suite *KeeperTestSuite) TestMarketMakerProposalAfterResetIncentivePair() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[29].String()
	k.SetParams(ctx, params)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2, 3})
	suite.Require().NoError(err)

	// reset incentive pairs after applied
	suite.ResetIncentivePairs()

	mm, found := k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.False(mm.Eligible)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
	suite.handleProposal(proposal)

	mm, found = k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.True(mm.Eligible)

	// distribute market maker incentive
	incentiveAmount := sdk.NewInt(500000000)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))

	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, nil,
		[]types.IncentiveDistribution{
			{
				Address: mmAddr.String(),
				PairId:  1,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	incentive, found := k.GetIncentive(ctx, mmAddr)
	suite.True(found)
	suite.Equal(mmAddr.String(), incentive.Address)

	// claim incentives
	err = k.ClaimIncentives(ctx, mmAddr)
	suite.NoError(err)

	// exclude market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil)
	suite.handleProposal(proposal)
	_, found = k.GetMarketMaker(ctx, mmAddr, 1)
	suite.False(found)

	// reject market maker
	proposal = types.NewMarketMakerProposal("title", "description", nil, nil, []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil)
	suite.handleProposal(proposal)
}

func (suite *KeeperTestSuite) TestRefundDepositWhenAmountChanged() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	params := k.GetParams(ctx)
	params.DepositAmount = types.DefaultDepositAmount
	k.SetParams(ctx, params)

	balanceBeforeModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.Require().NoError(err)

	balanceAfterModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.EqualValues(balanceBeforeModuleAcc.Add(params.DepositAmount...), balanceAfterModuleAcc)

	// change deposit amount
	params.DepositAmount = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(500000000)))
	k.SetParams(ctx, params)

	// apply market maker
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{2})
	suite.Require().NoError(err)

	balanceAfterModuleAcc = suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.EqualValues(balanceBeforeModuleAcc.Add(params.DepositAmount...).
		Add(types.DefaultDepositAmount...), balanceAfterModuleAcc)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
	suite.handleProposal(proposal)

	// refunded initial deposit amount
	balanceAfterModuleAcc2 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.EqualValues(params.DepositAmount, balanceAfterModuleAcc2)

	// include market maker
	proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil, nil)
	suite.handleProposal(proposal)

	// refunded changed deposit amount
	balanceAfterModuleAcc3 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.EqualValues(sdk.NewCoins(), balanceAfterModuleAcc3)
}

func (suite *KeeperTestSuite) TestRefundDepositWhenAmountZero() {
	ctx := suite.ctx
	k := suite.keeper
	params := k.GetParams(ctx)
	test := func(depositAmount sdk.Coins, mmAddr sdk.AccAddress) {
		params.DepositAmount = depositAmount
		k.SetParams(ctx, params)

		balanceBeforeModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		balanceBeforeMMAddr := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)

		// apply market maker
		err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
		suite.Require().NoError(err)

		balanceAfterModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		balanceAfterMMAddr := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
		suite.EqualValues(balanceBeforeModuleAcc, balanceAfterModuleAcc)
		suite.EqualValues(balanceBeforeMMAddr, balanceAfterMMAddr)

		deposit, found := k.GetDeposit(ctx, mmAddr, 1)
		suite.EqualValues(sdk.Coins(nil), deposit.Amount)
		suite.True(found)

		// apply market maker
		err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{2})
		suite.Require().NoError(err)

		balanceAfterModuleAcc = suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		balanceAfterMMAddr = suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
		suite.EqualValues(balanceBeforeModuleAcc, balanceAfterModuleAcc)
		suite.EqualValues(balanceBeforeMMAddr, balanceAfterMMAddr)

		// include market maker
		proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 1}}, nil, nil, nil)
		suite.handleProposal(proposal)

		_, found = k.GetDeposit(ctx, mmAddr, 1)
		suite.False(found)

		// refunded initial deposit amount
		balanceAfterModuleAcc2 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		suite.EqualValues(sdk.Coins{}, balanceAfterModuleAcc2)

		// include market maker
		proposal = types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{{Address: mmAddr.String(), PairId: 2}}, nil, nil, nil)
		suite.handleProposal(proposal)

		_, found = k.GetDeposit(ctx, mmAddr, 2)
		suite.False(found)

		// refunded changed deposit amount
		balanceAfterModuleAcc3 := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
		suite.EqualValues(sdk.NewCoins(), balanceAfterModuleAcc3)
	}

	test(sdk.Coins{}, suite.addrs[0])
	test(sdk.Coins(nil), suite.addrs[1])
	test(nil, suite.addrs[2])
}
