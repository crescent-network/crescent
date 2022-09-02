package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	"github.com/crescent-network/crescent/v3/x/marketmaker/types"
)

func (suite *KeeperTestSuite) TestApplyMarketMaker() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]
	params := k.GetParams(ctx)

	// not exist mm case
	_, found := k.GetMarketMaker(ctx, mmAddr, 1)
	suite.False(found)

	// apply market maker for the pair 1
	balanceBefore := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceBeforeModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.NoError(err)

	// validate deposit amount
	balanceAfter := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	balanceAfterModuleAcc := suite.app.BankKeeper.GetAllBalances(ctx, types.DepositReserveAcc)
	suite.EqualValues(balanceBefore.Sub(params.DepositAmount), balanceAfter)
	suite.EqualValues(balanceBeforeModuleAcc.Add(params.DepositAmount...), balanceAfterModuleAcc)

	// exist mm case
	mm, found := k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.False(mm.Eligible)
	suite.Equal(uint64(1), mm.PairId)
	suite.Equal(mmAddr.String(), mm.Address)

	// apply market maker with multiple pairs
	balanceBefore = suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{2, 3})
	suite.NoError(err)

	mm, found = k.GetMarketMaker(ctx, mmAddr, 2)
	suite.True(found)
	suite.False(mm.Eligible)
	suite.Equal(uint64(2), mm.PairId)
	suite.Equal(mmAddr.String(), mm.Address)

	mm, found = k.GetMarketMaker(ctx, mmAddr, 3)
	suite.True(found)
	suite.False(mm.Eligible)
	suite.Equal(uint64(3), mm.PairId)
	suite.Equal(mmAddr.String(), mm.Address)

	// validate deposit amount
	balanceAfter = suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(balanceBefore.Sub(params.DepositAmount).Sub(params.DepositAmount), balanceAfter)

	// already exist market maker for the pair 1
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.ErrorIs(err, types.ErrAlreadyExistMarketMaker)

	// already exist market maker for the pair 1, 2
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{1, 2})
	suite.ErrorIs(err, types.ErrAlreadyExistMarketMaker)

	// If only one of them is duplicated, all of them fail
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{3, 4})
	suite.ErrorIs(err, types.ErrAlreadyExistMarketMaker)

	balanceAfter2 := suite.app.BankKeeper.GetAllBalances(ctx, mmAddr)
	suite.EqualValues(balanceAfter, balanceAfter2)

}

func (suite *KeeperTestSuite) TestApplyMarketMakerUnregisteredPair() {
	ctx := suite.ctx
	k := suite.keeper
	mmAddr := suite.addrs[0]

	// reset IncentivePairs
	suite.ResetIncentivePairs()

	// apply market maker for the unregistered pair 1
	err := k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.ErrorIs(err, types.ErrUnregisteredPairId)

	// add IncentivePairs 1~7
	suite.SetIncentivePairs()

	// apply market maker for the registered pair 1
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{1})
	suite.NoError(err)
	mm, found := k.GetMarketMaker(ctx, mmAddr, 1)
	suite.True(found)
	suite.False(mm.Eligible)

	// apply market maker for the unregistered pair 8
	err = k.ApplyMarketMaker(ctx, mmAddr, []uint64{8})
	suite.ErrorIs(err, types.ErrUnregisteredPairId)
	_, found = k.GetMarketMaker(ctx, mmAddr, 8)
	suite.False(found)
}
