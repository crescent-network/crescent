package types_test

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	chain "github.com/crescent-network/crescent/v4/app"
	utils "github.com/crescent-network/crescent/v4/types"
	"github.com/crescent-network/crescent/v4/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/v4/x/liquidstaking/types"
	"github.com/crescent-network/crescent/v4/x/mint"
)

var (
	whitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			TargetWeight:     sdk.NewInt(10),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(1),
		},
		{
			ValidatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			TargetWeight:     sdk.NewInt(-1),
		},
		{
			ValidatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			TargetWeight:     sdk.NewInt(0),
		},
	}
)

func TestBTokenToNativeTokenWithFee(t *testing.T) {
	testCases := []struct {
		bTokenAmount            sdk.Int
		bTokenTotalSupplyAmount sdk.Int
		netAmount               sdk.Dec
		feeRate                 sdk.Dec
		expectedOutput          sdk.Dec
	}{
		// reward added case
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(5100000000),
			feeRate:                 sdk.MustNewDecFromStr("0.0"),
			expectedOutput:          sdk.MustNewDecFromStr("102000000.0"),
		},
		// reward added case with fee
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(5100000000),
			feeRate:                 sdk.MustNewDecFromStr("0.005"),
			expectedOutput:          sdk.MustNewDecFromStr("101490000.0"),
		},
		// slashed case
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(4000000000),
			feeRate:                 sdk.MustNewDecFromStr("0.0"),
			expectedOutput:          sdk.MustNewDecFromStr("80000000.0"),
		},
		// slashed case with fee
		{
			bTokenAmount:            sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(4000000000),
			feeRate:                 sdk.MustNewDecFromStr("0.001"),
			expectedOutput:          sdk.MustNewDecFromStr("79920000.0"),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, sdk.Int{}, tc.bTokenAmount)
		require.IsType(t, sdk.Int{}, tc.bTokenTotalSupplyAmount)
		require.IsType(t, sdk.Dec{}, tc.netAmount)
		require.IsType(t, sdk.Dec{}, tc.feeRate)
		require.IsType(t, sdk.Dec{}, tc.expectedOutput)

		output := types.BTokenToNativeToken(tc.bTokenAmount, tc.bTokenTotalSupplyAmount, tc.netAmount)
		if tc.feeRate.IsPositive() {
			output = types.DeductFeeRate(output, tc.feeRate)
		}
		require.EqualValues(t, tc.expectedOutput, output)
	}
}

func TestNativeToBTokenTo(t *testing.T) {
	testCases := []struct {
		nativeTokenAmount       sdk.Int
		bTokenTotalSupplyAmount sdk.Int
		netAmount               sdk.Dec
		expectedOutput          sdk.Int
	}{
		{
			nativeTokenAmount:       sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(5000000000),
			expectedOutput:          sdk.NewInt(100000000),
		},
		{
			nativeTokenAmount:       sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(4000000000),
			expectedOutput:          sdk.NewInt(125000000),
		},
		{
			nativeTokenAmount:       sdk.NewInt(100000000),
			bTokenTotalSupplyAmount: sdk.NewInt(5000000000),
			netAmount:               sdk.NewDec(55000000000),
			expectedOutput:          sdk.NewInt(9090909),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, sdk.Int{}, tc.nativeTokenAmount)
		require.IsType(t, sdk.Int{}, tc.bTokenTotalSupplyAmount)
		require.IsType(t, sdk.Dec{}, tc.netAmount)
		require.IsType(t, sdk.Int{}, tc.expectedOutput)

		output := types.NativeTokenToBToken(tc.nativeTokenAmount, tc.bTokenTotalSupplyAmount, tc.netAmount)
		require.EqualValues(t, tc.expectedOutput, output)
	}
}

func TestActiveCondition(t *testing.T) {
	testCases := []struct {
		validator      stakingtypes.Validator
		whitelisted    bool
		tombstoned     bool
		expectedOutput bool
	}{
		// active case 1
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: true,
		},
		// active case 2
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          true,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: true,
		},
		// inactive case 1 (not whitelisted)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    false,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 2 (invalid tokens, delShares)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.Int{},
				DelegatorShares: sdk.Dec{},
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 3 (zero tokens)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Bonded,
				Tokens:          sdk.NewInt(0),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 4 (invalid status)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Unspecified,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     false,
			expectedOutput: false,
		},
		// inactive case 5 (tombstoned)
		{
			validator: stakingtypes.Validator{
				OperatorAddress: whitelistedValidators[0].ValidatorAddress,
				Jailed:          false,
				Status:          stakingtypes.Unbonding,
				Tokens:          sdk.NewInt(100000000),
				DelegatorShares: sdk.NewDec(100000000),
			},
			whitelisted:    true,
			tombstoned:     true,
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		require.IsType(t, stakingtypes.Validator{}, tc.validator)
		output := types.ActiveCondition(tc.validator, tc.whitelisted, tc.tombstoned)
		require.EqualValues(t, tc.expectedOutput, output)
	}
}

type KeeperTestSuite struct {
	suite.Suite

	app        *chain.App
	ctx        sdk.Context
	keeper     keeper.Keeper
	querier    keeper.Querier
	govHandler govtypes.Handler
	addrs      []sdk.AccAddress
	delAddrs   []sdk.AccAddress
	valAddrs   []sdk.ValAddress
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = chain.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.govHandler = params.NewParamChangeProposalHandler(s.app.ParamsKeeper)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.MaxEntries = 7
	stakingParams.MaxValidators = 30
	s.app.StakingKeeper.SetParams(s.ctx, stakingParams)

	s.keeper = s.app.LiquidStakingKeeper
	s.querier = keeper.Querier{Keeper: s.keeper}
	s.addrs = chain.AddTestAddrs(s.app, s.ctx, 10, sdk.NewInt(1_000_000_000))
	s.delAddrs = chain.AddTestAddrs(s.app, s.ctx, 10, sdk.NewInt(1_000_000_000))
	s.valAddrs = chain.ConvertAddrsToValAddrs(s.delAddrs)

	s.ctx = s.ctx.WithBlockHeight(100).WithBlockTime(utils.ParseTime("2022-03-01T00:00:00Z"))
	params := s.keeper.GetParams(s.ctx)
	params.UnstakeFeeRate = sdk.ZeroDec()
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)
	// call mint.BeginBlocker for init k.SetLastBlockTime(ctx, ctx.BlockTime())
	mint.BeginBlocker(s.ctx, s.app.MintKeeper)
}

func (s *KeeperTestSuite) CreateValidators(powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey) {
	s.app.BeginBlocker(s.ctx, abci.RequestBeginBlock{})
	num := len(powers)
	addrs := chain.AddTestAddrsIncremental(s.app, s.ctx, num, sdk.NewInt(1000000000))
	valAddrs := chain.ConvertAddrsToValAddrs(addrs)
	pks := chain.CreateTestPubKeys(num)

	for i, power := range powers {
		val, err := stakingtypes.NewValidator(valAddrs[i], pks[i], stakingtypes.Description{})
		s.Require().NoError(err)
		s.app.StakingKeeper.SetValidator(s.ctx, val)
		err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, val)
		s.Require().NoError(err)
		s.app.StakingKeeper.SetNewValidatorByPowerIndex(s.ctx, val)
		s.app.StakingKeeper.AfterValidatorCreated(s.ctx, val.GetOperator())
		newShares, err := s.app.StakingKeeper.Delegate(s.ctx, addrs[i], sdk.NewInt(power), stakingtypes.Unbonded, val, true)
		s.Require().NoError(err)
		s.Require().Equal(newShares.TruncateInt(), sdk.NewInt(power))
	}

	s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	return addrs, valAddrs, pks
}

func (s *KeeperTestSuite) TestLiquidStake() {
	_, valOpers, _ := s.CreateValidators([]int64{1000000, 2000000, 3000000})
	params := s.keeper.GetParams(s.ctx)
	params.MinLiquidStakingAmount = sdk.NewInt(50000)
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	stakingAmt := params.MinLiquidStakingAmount

	// fail, no active validator
	cachedCtx, _ := s.ctx.CacheContext()
	newShares, bTokenMintAmt, err := s.keeper.LiquidStake(cachedCtx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().ErrorIs(err, types.ErrActiveLiquidValidatorsNotExists)
	s.Require().Equal(newShares, sdk.ZeroDec())
	s.Require().Equal(bTokenMintAmt, sdk.ZeroInt())

	// add active validator
	params.WhitelistedValidators = []types.WhitelistedValidator{
		{ValidatorAddress: valOpers[0].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[1].String(), TargetWeight: sdk.NewInt(1)},
		{ValidatorAddress: valOpers[2].String(), TargetWeight: sdk.NewInt(1)},
	}
	s.keeper.SetParams(s.ctx, params)
	s.keeper.UpdateLiquidValidatorSet(s.ctx)

	res := s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	s.Require().Equal(sdk.ZeroDec(), res[0].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[0].LiquidTokens)

	s.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	s.Require().Equal(sdk.ZeroDec(), res[1].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[1].LiquidTokens)

	s.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	s.Require().Equal(sdk.ZeroDec(), res[2].DelShares)
	s.Require().Equal(sdk.ZeroInt(), res[2].LiquidTokens)

	// liquid staking
	newShares, bTokenMintAmt, err = s.keeper.LiquidStake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	s.Require().NoError(err)
	s.Require().Equal(newShares, stakingAmt.ToDec())
	s.Require().Equal(bTokenMintAmt, stakingAmt)

	_, found := s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[0])
	s.Require().False(found)
	_, found = s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[1])
	s.Require().False(found)
	_, found = s.app.StakingKeeper.GetDelegation(s.ctx, s.delAddrs[0], valOpers[2])
	s.Require().False(found)

	proxyAccDel1, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found := s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	s.Require().Equal(proxyAccDel1.Shares, sdk.NewDec(16668)) // 16666 + add crumb 2 to 1st active validator
	s.Require().Equal(proxyAccDel2.Shares, sdk.NewDec(16666))
	s.Require().Equal(proxyAccDel2.Shares, sdk.NewDec(16666))
	s.Require().Equal(stakingAmt.ToDec(), proxyAccDel1.Shares.Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	liquidBondDenom := s.keeper.LiquidBondDenom(s.ctx)
	balanceBeforeUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceBeforeUBD.Amount, sdk.NewInt(999950000))
	ubdBToken := sdk.NewCoin(liquidBondDenom, sdk.NewInt(10000))
	bTokenBalance := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], liquidBondDenom)
	bTokenTotalSupply := s.app.BankKeeper.GetSupply(s.ctx, liquidBondDenom)
	s.Require().Equal(bTokenBalance, sdk.NewCoin(liquidBondDenom, sdk.NewInt(50000)))
	s.Require().Equal(bTokenBalance, bTokenTotalSupply)

	// liquid unstaking
	ubdTime, unbondingAmt, ubds, unbondedAmt, err := s.keeper.LiquidUnstake(s.ctx, types.LiquidStakingProxyAcc, s.delAddrs[0], ubdBToken)
	s.Require().NoError(err)
	s.Require().EqualValues(unbondedAmt, sdk.ZeroInt())
	s.Require().Len(ubds, 3)

	// crumb excepted on unbonding
	crumb := ubdBToken.Amount.Sub(ubdBToken.Amount.QuoRaw(3).MulRaw(3)) // 1
	s.Require().EqualValues(unbondingAmt, ubdBToken.Amount.Sub(crumb))  // 9999
	s.Require().Equal(ubds[0].DelegatorAddress, s.delAddrs[0].String())
	s.Require().Equal(ubdTime, utils.ParseTime("2022-03-22T00:00:00Z"))
	bTokenBalanceAfter := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], liquidBondDenom)
	s.Require().Equal(bTokenBalanceAfter, sdk.NewCoin(liquidBondDenom, sdk.NewInt(40000)))

	balanceBeginUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceBeginUBD.Amount, balanceBeforeUBD.Amount)

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	s.Require().Equal(stakingAmt.Sub(unbondingAmt).ToDec(), proxyAccDel1.GetShares().Add(proxyAccDel2.Shares).Add(proxyAccDel3.Shares))

	// complete unbonding
	s.ctx = s.ctx.WithBlockHeight(200).WithBlockTime(ubdTime.Add(1))
	updates := s.app.StakingKeeper.BlockValidatorUpdates(s.ctx) // EndBlock of staking keeper, mature UBD
	s.Require().Empty(updates)
	balanceCompleteUBD := s.app.BankKeeper.GetBalance(s.ctx, s.delAddrs[0], sdk.DefaultBondDenom)
	s.Require().Equal(balanceCompleteUBD.Amount, balanceBeforeUBD.Amount.Add(unbondingAmt))

	proxyAccDel1, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[0])
	s.Require().True(found)
	proxyAccDel2, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[1])
	s.Require().True(found)
	proxyAccDel3, found = s.app.StakingKeeper.GetDelegation(s.ctx, types.LiquidStakingProxyAcc, valOpers[2])
	s.Require().True(found)
	// crumb added to first valid active liquid validator
	s.Require().Equal(sdk.NewDec(13335), proxyAccDel1.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel2.Shares)
	s.Require().Equal(sdk.NewDec(13333), proxyAccDel3.Shares)

	res = s.keeper.GetAllLiquidValidatorStates(s.ctx)
	s.Require().Equal(params.WhitelistedValidators[0].ValidatorAddress, res[0].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[0].TargetWeight, res[0].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[0].Status)
	s.Require().Equal(sdk.NewDec(13335), res[0].DelShares)

	s.Require().Equal(params.WhitelistedValidators[1].ValidatorAddress, res[1].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[1].TargetWeight, res[1].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[1].Status)
	s.Require().Equal(sdk.NewDec(13333), res[1].DelShares)

	s.Require().Equal(params.WhitelistedValidators[2].ValidatorAddress, res[2].OperatorAddress)
	s.Require().Equal(params.WhitelistedValidators[2].TargetWeight, res[2].Weight)
	s.Require().Equal(types.ValidatorStatusActive, res[2].Status)
	s.Require().Equal(sdk.NewDec(13333), res[2].DelShares)

	vs := s.keeper.GetAllLiquidValidators(s.ctx)
	s.Require().Len(vs.Map(), 3)

	whitelistedValsMap := types.GetWhitelistedValsMap(params.WhitelistedValidators)
	avs := s.keeper.GetActiveLiquidValidators(s.ctx, whitelistedValsMap)
	alt, _ := avs.TotalActiveLiquidTokens(s.ctx, s.app.StakingKeeper, true)
	s.Require().EqualValues(alt, sdk.NewInt(40001))
}
