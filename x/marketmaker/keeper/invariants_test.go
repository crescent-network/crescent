package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v4/x/marketmaker/keeper"
	"github.com/crescent-network/crescent/v4/x/marketmaker/types"
)

func (suite *KeeperTestSuite) TestDepositReservedAmountInvariant() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]
	params := k.GetParams(ctx)

	// This is normal state, must not be broken.
	_, broken := keeper.DepositReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2})
	suite.Require().NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{3})
	suite.Require().NoError(err)

	_, found := k.GetDeposit(ctx, mmAddr, 1)
	suite.True(found)

	_, found = k.GetDeposit(ctx, mmAddr, 2)
	suite.True(found)

	_, found = k.GetDeposit(ctx, mmAddr2, 3)
	suite.True(found)

	balanceReserveAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.Require().EqualValues(params.DepositAmount.Add(params.DepositAmount...).Add(params.DepositAmount...), balanceReserveAcc)

	// This is normal state, must not be broken.
	_, broken = keeper.DepositReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	mmPair2, found := k.GetMarketMaker(ctx, mmAddr, 2)
	suite.True(found)

	// manipulate eligible of the market maker to break invariant
	mmPair2.Eligible = true
	k.SetMarketMaker(ctx, mmPair2)

	// broken deposit reserved count invariant
	_, broken = keeper.DepositReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// recovery
	mmPair2.Eligible = false
	k.SetMarketMaker(ctx, mmPair2)

	// Send coins from deposit reserve acc to break deposit amount invariant
	err = suite.app.BankKeeper.SendCoins(
		ctx, types.DepositReserveAcc, suite.addrs[3], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
	suite.Require().NoError(err)

	// broken deposit reserved amount invariant
	_, broken = keeper.DepositReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// receive coins on deposit reserve acc to recover invariant
	err = suite.app.BankKeeper.SendCoins(
		ctx, suite.addrs[3], types.DepositReserveAcc, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
	suite.Require().NoError(err)

	_, broken = keeper.DepositReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)
}

func (suite *KeeperTestSuite) TestIncentiveReservedAmountInvariant() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]

	// set incentive budget
	params := k.GetParams(ctx)
	params.IncentiveBudgetAddress = suite.addrs[5].String()
	k.SetParams(ctx, params)

	// This is normal state, must not be broken.
	_, broken := keeper.IncentiveReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2})
	suite.Require().NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{3})
	suite.Require().NoError(err)

	// include market maker
	proposal := types.NewMarketMakerProposal("title", "description", []types.MarketMakerHandle{
		{Address: mmAddr.String(), PairId: 1},
		{Address: mmAddr.String(), PairId: 2},
		{Address: mmAddr2.String(), PairId: 3},
	}, nil, nil, nil)
	suite.handleProposal(proposal)

	incentiveAmount := sdk.NewInt(500000000)
	incentiveCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, incentiveAmount))

	// submit incentive distribution proposal
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
			{
				Address: mmAddr2.String(),
				PairId:  3,
				Amount:  incentiveCoins,
			},
		})
	suite.handleProposal(proposal)

	balanceReserveAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.ClaimableIncentiveReserveAcc)
	suite.Require().EqualValues(incentiveCoins.Add(incentiveCoins...).Add(incentiveCoins...), balanceReserveAcc)

	// This is normal state, must not be broken.
	_, broken = keeper.IncentiveReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)

	incentive, found := k.GetIncentive(ctx, mmAddr)
	suite.True(found)
	suite.Require().EqualValues(incentiveCoins.Add(incentiveCoins...), incentive.Claimable)

	incentive2, found := k.GetIncentive(ctx, mmAddr2)
	suite.True(found)
	suite.Require().EqualValues(incentiveCoins, incentive2.Claimable)

	// manipulate claimable amount of the market maker to break invariant
	incentive2.Claimable = incentive2.Claimable.Add(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)))
	k.SetIncentive(ctx, incentive2)

	// broken incentive reserved invariant
	_, broken = keeper.IncentiveReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// recovery
	incentive2.Claimable = incentiveCoins
	k.SetIncentive(ctx, incentive2)

	// Send coins from incentive reserve acc to break deposit amount invariant
	err = suite.app.BankKeeper.SendCoins(
		ctx, types.ClaimableIncentiveReserveAcc, suite.addrs[3], sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
	suite.Require().NoError(err)

	// broken incentive reserved amount invariant
	_, broken = keeper.IncentiveReservedAmountInvariant(k)(ctx)
	suite.Require().True(broken)

	// receive coins on incentive reserve acc to recover invariant
	err = suite.app.BankKeeper.SendCoins(
		ctx, suite.addrs[3], types.ClaimableIncentiveReserveAcc, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2)))
	suite.Require().NoError(err)

	_, broken = keeper.IncentiveReservedAmountInvariant(k)(ctx)
	suite.Require().False(broken)
}

func (suite *KeeperTestSuite) TestDepositRecordsInvariant() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	mmAddr2 := suite.addrs[1]

	// This is normal state, must not be broken.
	_, broken := keeper.DepositRecordsInvariant(k)(ctx)
	suite.Require().False(broken)

	// apply market maker
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2})
	suite.Require().NoError(err)
	err = k.ApplyMarketMaker(ctx, mmAddr2, []uint64{3})
	suite.Require().NoError(err)

	_, found := k.GetDeposit(ctx, mmAddr, 1)
	suite.True(found)

	_, found = k.GetDeposit(ctx, mmAddr, 2)
	suite.True(found)

	_, found = k.GetDeposit(ctx, mmAddr2, 3)
	suite.True(found)

	mmPair2, found := k.GetMarketMaker(ctx, mmAddr, 2)
	suite.True(found)

	// manipulate eligible of the market maker to break invariant
	mmPair2.Eligible = true
	k.SetMarketMaker(ctx, mmPair2)

	// broken deposit record invariant
	_, broken = keeper.DepositRecordsInvariant(k)(ctx)
	suite.Require().True(broken)

	// recovery
	mmPair2.Eligible = false
	k.SetMarketMaker(ctx, mmPair2)

	_, broken = keeper.DepositRecordsInvariant(k)(ctx)
	suite.Require().False(broken)

	// manipulate force deleting deposit
	k.DeleteDeposit(ctx, mmAddr, 2)

	// broken deposit record invariant
	_, broken = keeper.DepositRecordsInvariant(k)(ctx)
	suite.Require().True(broken)
}
