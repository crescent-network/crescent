package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func (suite *KeeperTestSuite) TestRebalancingCase1() {
	_, valOpers := suite.CreateValidators([]int64{1000000, 1000000, 1000000, 1000000, 1000000})
	suite.ctx = suite.ctx.WithBlockHeight(100).WithBlockTime(types.MustParseRFC3339("2022-03-01T00:00:00Z"))
	params := suite.keeper.GetParams(suite.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	params.CommissionRate = sdk.ZeroDec()
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	stakingAmt := sdk.NewInt(50000)
	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	newShares, bTokenMintAmt, err := suite.keeper.LiquidStaking(suite.ctx, types.LiquidStakingProxyAcc, suite.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	suite.Require().NoError(err)
	suite.Require().Equal(newShares, sdk.MustNewDecFromStr("49998.0"))
	suite.Require().Equal(bTokenMintAmt, stakingAmt)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	proxyAccDel1, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)

	suite.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(16666))
	suite.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(16666))
	suite.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(16666))

	for _, v := range suite.keeper.GetAllLiquidValidators(suite.ctx) {
		fmt.Println(v.OperatorAddress, v.LiquidTokens, v.Status)
	}
	fmt.Println("-----------")

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), Weight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	proxyAccDel1, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	suite.Require().True(found)
	proxyAccDel2, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	suite.Require().True(found)
	proxyAccDel3, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	suite.Require().True(found)
	proxyAccDel4, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	suite.Require().True(found)

	suite.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(12499))
	suite.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(12499))
	suite.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(12501))
	suite.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(12499))

	for _, v := range suite.keeper.GetAllLiquidValidators(suite.ctx) {
		fmt.Println(v.OperatorAddress, v.LiquidTokens, v.Status)
	}
	fmt.Println("-----------")

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), Weight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), Weight: sdk.NewInt(1)},
	}
	suite.keeper.SetParams(suite.ctx, params)
	liquidstaking.EndBlocker(suite.ctx, suite.keeper)

	//proxyAccDel1, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	//suite.Require().True(found)
	//proxyAccDel2, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	//suite.Require().True(found)
	//proxyAccDel3, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	//suite.Require().True(found)
	//proxyAccDel4, found = suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	//suite.Require().True(found)
	//proxyAccDel5, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	//suite.Require().True(found)

	// [TryRedelegations] failed due to redelegation restriction
	//suite.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(10000))
	//suite.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(10000))
	//suite.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(10000))
	//suite.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(10000))
	//suite.Require().EqualValues(proxyAccDel5.Shares.TruncateInt(), sdk.NewInt(10000))
	//for _, v := range suite.keeper.GetAllLiquidValidators(suite.ctx) {
	//	fmt.Println(v.OperatorAddress, v.LiquidTokens, v.Status)
	//}
	//fmt.Println("-----------")
}
