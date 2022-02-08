package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	squadtypes "github.com/cosmosquad-labs/squad/types"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/types"
)

func (s *KeeperTestSuite) TestRebalancingCase1() {
	_, valOpers := s.CreateValidators([]int64{1000000, 1000000, 1000000, 1000000, 1000000})
	s.ctx = s.ctx.WithBlockHeight(100).WithBlockTime(squadtypes.MustParseRFC3339("2022-03-01T00:00:00Z"))
	params := s.keeper.GetParams(s.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(50000)
	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	newShares, bTokenMintAmt, err := s.keeper.LiquidStaking(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().NoError(err)
	s.Require().Equal(newShares, sdk.MustNewDecFromStr("49998.0"))
	s.Require().Equal(bTokenMintAmt, stakingAmt)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	proxyAccDel1, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)

	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(16666))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(16666))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(16666))

	for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
		fmt.Println(v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper))
	}
	fmt.Println("-----------")

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)

	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(12501))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(12499))

	for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
		fmt.Println(v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper))
	}
	fmt.Println("-----------")

	reds := s.app.StakingKeeper.GetRedelegations(s.ctx, types.LiquidStakingProxyAcc, 20)
	s.Require().Len(reds, 3)

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// update whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[4].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)
	proxyAccDel5, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().True(found)

	for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
		fmt.Println(v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper))
	}
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(9999))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(10002))
	s.Require().EqualValues(proxyAccDel5.Shares.TruncateInt(), sdk.NewInt(9999))

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// remove whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[3].String(), TargetWeight: sdk.NewInt(1)},
	}

	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().True(found)
	proxyAccDel5, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().False(found)

	for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
		fmt.Println(v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper))
	}
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel3.Shares.TruncateInt(), sdk.NewInt(12499))
	s.Require().EqualValues(proxyAccDel4.Shares.TruncateInt(), sdk.NewInt(12501))

	// advance block time and height for complete redelegations
	s.completeRedelegationUnbonding()

	// remove whitelist validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
	}

	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().False(found)
	proxyAccDel4, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[3])
	s.Require().False(found)
	proxyAccDel5, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[4])
	s.Require().False(found)

	for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
		fmt.Println(v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper))
	}
	s.Require().EqualValues(proxyAccDel1.Shares.TruncateInt(), sdk.NewInt(25000))
	s.Require().EqualValues(proxyAccDel2.Shares.TruncateInt(), sdk.NewInt(24998))
	// TODO: add more edge cases
}

func (s *KeeperTestSuite) TestWithdrawRewardsAndReStaking() {
	_, valOpers := s.CreateValidators([]int64{1000000, 1000000, 1000000})
	params := s.keeper.GetParams(s.ctx)

	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(10)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(10)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := sdk.NewInt(100000000)
	s.liquidStaking(s.delAddrs[0], stakingAmt)

	// no rewards
	totalRewards, totalDelShares, totalLiquidTokens := s.keeper.CheckTotalRewards(s.ctx, types.LiquidStakingProxyAcc)
	s.EqualValues(totalRewards, sdk.ZeroDec())
	s.EqualValues(totalDelShares, stakingAmt.ToDec(), totalLiquidTokens)

	// allocate rewards
	s.advanceHeight(100, false)
	totalRewards, totalDelShares, totalLiquidTokens = s.keeper.CheckTotalRewards(s.ctx, types.LiquidStakingProxyAcc)
	s.NotEqualValues(totalRewards, sdk.ZeroDec())
	s.NotEqualValues(totalLiquidTokens, sdk.ZeroDec())

	// withdraw rewards and re-staking
	whitelistedValMap := types.GetWhitelistedValMap(params.WhitelistedValidators)
	s.keeper.WithdrawRewardsAndReStaking(s.ctx, whitelistedValMap)
	totalRewardsAfter, totalDelSharesAfter, totalLiquidTokensAfter := s.keeper.CheckTotalRewards(s.ctx, types.LiquidStakingProxyAcc)
	s.EqualValues(totalRewardsAfter, sdk.ZeroDec())
	s.EqualValues(totalDelSharesAfter, totalRewards.TruncateDec().Add(totalDelShares), totalLiquidTokensAfter)
}
