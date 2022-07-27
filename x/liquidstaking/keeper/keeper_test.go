package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	chain "github.com/crescent-network/crescent/v2/app"
	utils "github.com/crescent-network/crescent/v2/types"
	"github.com/crescent-network/crescent/v2/x/farming"
	farmingtypes "github.com/crescent-network/crescent/v2/x/farming/types"
	liquiditytypes "github.com/crescent-network/crescent/v2/x/liquidity/types"
	"github.com/crescent-network/crescent/v2/x/liquidstaking"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
	"github.com/crescent-network/crescent/v2/x/mint"
)

var (
	BlockTime = 10 * time.Second
)

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

func (s *KeeperTestSuite) TearDownTest() {
	// invariant check
	crisis.EndBlocker(s.ctx, s.app.CrisisKeeper)
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

func (s *KeeperTestSuite) liquidStaking(liquidStaker sdk.AccAddress, stakingAmt sdk.Int) error {
	ctx, writeCache := s.ctx.CacheContext()
	params := s.keeper.GetParams(ctx)
	btokenBalanceBefore := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	newShares, bTokenMintAmt, err := s.keeper.LiquidStake(ctx, types.LiquidStakingProxyAcc, liquidStaker, sdk.NewCoin(sdk.DefaultBondDenom, stakingAmt))
	if err != nil {
		return err
	}
	btokenBalanceAfter := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	s.Require().NoError(err)
	s.NotEqualValues(newShares, sdk.ZeroDec())
	s.Require().EqualValues(bTokenMintAmt, btokenBalanceAfter.Sub(btokenBalanceBefore))
	writeCache()
	return nil
}

func (s *KeeperTestSuite) liquidUnstaking(liquidStaker sdk.AccAddress, ubdBTokenAmt sdk.Int, ubdComplete bool) error {
	ctx := s.ctx
	params := s.keeper.GetParams(ctx)
	balanceBefore := s.app.BankKeeper.GetBalance(ctx, liquidStaker, sdk.DefaultBondDenom).Amount
	ubdTime, unbondingAmt, _, unbondedAmt, err := s.liquidUnstakingWithResult(liquidStaker, sdk.NewCoin(params.LiquidBondDenom, ubdBTokenAmt))
	if err != nil {
		return err
	}
	if ubdComplete {
		alv := s.keeper.GetActiveLiquidValidators(ctx, params.WhitelistedValsMap())
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 200).WithBlockTime(ubdTime.Add(1))
		s.app.StakingKeeper.BlockValidatorUpdates(ctx) // EndBlock of staking keeper, mature UBD
		balanceCompleteUBD := s.app.BankKeeper.GetBalance(ctx, liquidStaker, sdk.DefaultBondDenom)
		for _, v := range alv {
			_, found := s.app.StakingKeeper.GetUnbondingDelegation(ctx, liquidStaker, v.GetOperator())
			s.Require().False(found)
		}
		s.Require().EqualValues(balanceCompleteUBD.Amount, balanceBefore.Add(unbondingAmt).Add(unbondedAmt))
	}
	return nil
}

func (s *KeeperTestSuite) liquidUnstakingWithResult(liquidStaker sdk.AccAddress, unstakingBtoken sdk.Coin) (time.Time, sdk.Int, []stakingtypes.UnbondingDelegation, sdk.Int, error) {
	ctx, writeCache := s.ctx.CacheContext()
	params := s.keeper.GetParams(ctx)
	alv := s.keeper.GetActiveLiquidValidators(ctx, params.WhitelistedValsMap())
	balanceBefore := s.app.BankKeeper.GetBalance(ctx, liquidStaker, sdk.DefaultBondDenom).Amount
	btokenBalanceBefore := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	ubdTime, unbondingAmt, ubds, unbondedAmt, err := s.keeper.LiquidUnstake(ctx, types.LiquidStakingProxyAcc, liquidStaker, unstakingBtoken)
	if err != nil {
		return ubdTime, unbondingAmt, ubds, unbondedAmt, err
	}
	balanceAfter := s.app.BankKeeper.GetBalance(ctx, liquidStaker, sdk.DefaultBondDenom).Amount
	btokenBalanceAfter := s.app.BankKeeper.GetBalance(ctx, liquidStaker, params.LiquidBondDenom).Amount
	s.Require().EqualValues(unstakingBtoken.Amount, btokenBalanceBefore.Sub(btokenBalanceAfter))
	if unbondedAmt.IsPositive() {
		s.Require().EqualValues(unbondedAmt, balanceAfter.Sub(balanceBefore))
	}
	for _, v := range alv {
		_, found := s.app.StakingKeeper.GetUnbondingDelegation(ctx, liquidStaker, v.GetOperator())
		s.Require().True(found)
	}
	writeCache()
	return ubdTime, unbondingAmt, ubds, unbondedAmt, err
}

func (s *KeeperTestSuite) RequireNetAmountStateZero() {
	nas := s.keeper.GetNetAmountState(s.ctx)
	s.Require().EqualValues(nas.MintRate, sdk.ZeroDec())
	s.Require().EqualValues(nas.BtokenTotalSupply, sdk.ZeroInt())
	s.Require().EqualValues(nas.NetAmount, sdk.ZeroDec())
	s.Require().EqualValues(nas.TotalDelShares, sdk.ZeroDec())
	s.Require().EqualValues(nas.TotalLiquidTokens, sdk.ZeroInt())
	s.Require().EqualValues(nas.TotalRemainingRewards, sdk.ZeroDec())
	s.Require().EqualValues(nas.TotalUnbondingBalance, sdk.ZeroDec())
	s.Require().EqualValues(nas.ProxyAccBalance, sdk.ZeroInt())

}

// advance block time and height for complete redelegations and unbondings
func (s *KeeperTestSuite) completeRedelegationUnbonding() {
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100).WithBlockTime(s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime))
	s.app.EndBlocker(s.ctx, abci.RequestEndBlock{})
	reds := s.app.StakingKeeper.GetRedelegations(s.ctx, types.LiquidStakingProxyAcc, 100)
	s.Require().Len(reds, 0)
	ubds := s.app.StakingKeeper.GetUnbondingDelegations(s.ctx, types.LiquidStakingProxyAcc, 100)
	s.Require().Len(ubds, 0)
}

func (s *KeeperTestSuite) redelegationsErrorCount(redelegations []types.Redelegation) int {
	errCnt := 0
	for _, red := range redelegations {
		if red.Error != nil {
			errCnt++
		}
	}
	return errCnt
}

func (s *KeeperTestSuite) printRedelegationsLiquidTokens() {
	redsIng := s.app.StakingKeeper.GetRedelegations(s.ctx, types.LiquidStakingProxyAcc, 50)
	if len(redsIng) != 0 {
		fmt.Println("[Redelegations]")
		for i, red := range redsIng {
			fmt.Println("\tRedelegation #", i+1)
			fmt.Println("\t\tDelegatorAddress: ", red.DelegatorAddress)
			fmt.Println("\t\tValidatorSrcAddress : ", red.ValidatorSrcAddress)
			fmt.Println("\t\tValidatorDstAddress: ", red.ValidatorDstAddress)
			fmt.Println("\t\tEntries: ")
			for _, e := range red.Entries {
				fmt.Println("\t\t\tCreationHeight: ", e.CreationHeight)
				fmt.Println("\t\t\tCompletionTime: ", e.CompletionTime)
				fmt.Println("\t\t\tInitialBalance: ", e.InitialBalance)
				fmt.Println("\t\t\tSharesDst: ", e.SharesDst)
			}
		}
		fmt.Println("")
	}
	liquidVals := s.keeper.GetAllLiquidValidators(s.ctx)
	if len(liquidVals) != 0 {
		fmt.Println("[LiquidValidators]")
		for _, v := range s.keeper.GetAllLiquidValidators(s.ctx) {
			fmt.Printf("   OperatorAddress %s; LiquidTokens: %s\n", v.OperatorAddress, v.GetLiquidTokens(s.ctx, s.app.StakingKeeper, false))
		}
	}
}

func (s *KeeperTestSuite) advanceHeight(height int, withBeginBlock bool) {
	feeCollector := s.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	for i := 0; i < height; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1).WithBlockTime(s.ctx.BlockTime().Add(BlockTime))
		mint.BeginBlocker(s.ctx, s.app.MintKeeper)
		feeCollectorBalance := s.app.BankKeeper.GetAllBalances(s.ctx, feeCollector)
		rewardsToBeDistributed := feeCollectorBalance.AmountOf(sdk.DefaultBondDenom)

		// mimic distribution.BeginBlock (AllocateTokens, get rewards from feeCollector, AllocateTokensToValidator, add remaining to feePool)
		err := s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, authtypes.FeeCollectorName, distrtypes.ModuleName, feeCollectorBalance)
		s.Require().NoError(err)
		totalRewards := sdk.ZeroDec()
		totalPower := int64(0)
		s.app.StakingKeeper.IterateBondedValidatorsByPower(s.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
			consPower := validator.GetConsensusPower(s.app.StakingKeeper.PowerReduction(s.ctx))
			totalPower = totalPower + consPower
			return false
		})
		if totalPower != 0 {
			s.app.StakingKeeper.IterateBondedValidatorsByPower(s.ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
				consPower := validator.GetConsensusPower(s.app.StakingKeeper.PowerReduction(s.ctx))
				powerFraction := sdk.NewDec(consPower).QuoTruncate(sdk.NewDec(totalPower))
				reward := rewardsToBeDistributed.ToDec().MulTruncate(powerFraction)
				s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, validator, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: reward}})
				totalRewards = totalRewards.Add(reward)
				return false
			})
		}
		remaining := rewardsToBeDistributed.ToDec().Sub(totalRewards)
		s.Require().False(remaining.GT(sdk.NewDec(1)))
		feePool := s.app.DistrKeeper.GetFeePool(s.ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: remaining}}...)
		s.app.DistrKeeper.SetFeePool(s.ctx, feePool)
		if withBeginBlock {
			// liquid validator set update, rebalancing, withdraw rewards, re-stake
			liquidstaking.BeginBlocker(s.ctx, s.app.LiquidStakingKeeper)
		}
		staking.EndBlocker(s.ctx, *s.app.StakingKeeper)
	}
}

// doubleSign, tombstone, slash, jail
func (s *KeeperTestSuite) doubleSign(valOper sdk.ValAddress, consAddr sdk.ConsAddress) {
	liquidValidator, found := s.keeper.GetLiquidValidator(s.ctx, valOper)
	s.Require().True(found)
	val, found := s.app.StakingKeeper.GetValidator(s.ctx, valOper)
	s.Require().True(found)
	tokens := val.Tokens
	liquidTokens := liquidValidator.GetLiquidTokens(s.ctx, s.app.StakingKeeper, false)

	// check sign info
	info, found := s.app.SlashingKeeper.GetValidatorSigningInfo(s.ctx, consAddr)
	s.Require().True(found)
	s.Require().Equal(info.Address, consAddr.String())

	// make evidence
	evidence := &evidencetypes.Equivocation{
		//Height: 0,
		//Time:   time.Unix(0, 0),
		Height:           s.ctx.BlockHeight(),
		Time:             s.ctx.BlockTime(),
		Power:            s.app.StakingKeeper.TokensToConsensusPower(s.ctx, tokens),
		ConsensusAddress: consAddr.String(),
	}

	// Double sign
	s.app.EvidenceKeeper.HandleEquivocationEvidence(s.ctx, evidence)
	// HandleEquivocationEvidence call below functions
	//s.app.SlashingKeeper.Slash()
	//s.app.SlashingKeeper.Jail(s.ctx, consAddr)
	//s.app.SlashingKeeper.JailUntil(s.ctx, consAddr, evidencetypes.DoubleSignJailEndTime)
	//s.app.SlashingKeeper.Tombstone(s.ctx, consAddr)

	// should be jailed and tombstoned
	s.Require().True(s.app.StakingKeeper.Validator(s.ctx, liquidValidator.GetOperator()).IsJailed())
	s.Require().True(s.app.SlashingKeeper.IsTombstoned(s.ctx, consAddr))

	// check tombstoned on sign info
	info, found = s.app.SlashingKeeper.GetValidatorSigningInfo(s.ctx, consAddr)
	s.Require().True(found)
	s.Require().True(info.Tombstoned)
	val, _ = s.app.StakingKeeper.GetValidator(s.ctx, valOper)
	s.Require().True(s.keeper.IsTombstoned(s.ctx, val))
	liquidTokensSlashed := liquidValidator.GetLiquidTokens(s.ctx, s.app.StakingKeeper, false)
	tokensSlashed := val.Tokens
	s.Require().True(tokensSlashed.LT(tokens))
	s.Require().True(liquidTokensSlashed.LT(liquidTokens))

	s.app.StakingKeeper.BlockValidatorUpdates(s.ctx)
	val, _ = s.app.StakingKeeper.GetValidator(s.ctx, valOper)
	// set unbonding status, no more rewards before return Bonded
	s.Require().Equal(val.Status, stakingtypes.Unbonding)
}

func (s *KeeperTestSuite) createContinuousVestingAccount(from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins, startTime, endTime time.Time) vestingtypes.ContinuousVestingAccount {
	baseAccount := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, to)
	_, ok := baseAccount.(*authtypes.BaseAccount)
	s.Require().True(ok)
	baseVestingAccount := vestingtypes.NewBaseVestingAccount(baseAccount.(*authtypes.BaseAccount), amt, endTime.Unix())
	cVestingAcc := vestingtypes.NewContinuousVestingAccountRaw(baseVestingAccount, startTime.Unix())
	s.app.AccountKeeper.SetAccount(s.ctx, cVestingAcc)
	err := s.app.BankKeeper.SendCoins(s.ctx, from, to, amt)
	s.Require().NoError(err)
	return *cVestingAcc
}

func (s *KeeperTestSuite) fundAddr(addr sdk.AccAddress, amt sdk.Coins) {
	err := s.app.BankKeeper.MintCoins(s.ctx, liquiditytypes.ModuleName, amt)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, liquiditytypes.ModuleName, addr, amt)
	s.Require().NoError(err)
}

// liquidity module keeper utils for liquid staking combine test

func (s *KeeperTestSuite) createPair(creator sdk.AccAddress, baseCoinDenom, quoteCoinDenom string, fund bool) liquiditytypes.Pair {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, params.PairCreationFee)
	}
	pair, err := s.app.LiquidityKeeper.CreatePair(s.ctx, liquiditytypes.NewMsgCreatePair(creator, baseCoinDenom, quoteCoinDenom))
	s.Require().NoError(err)
	return pair
}

func (s *KeeperTestSuite) createPool(creator sdk.AccAddress, pairId uint64, depositCoins sdk.Coins, fund bool) liquiditytypes.Pool {
	params := s.app.LiquidityKeeper.GetParams(s.ctx)
	if fund {
		s.fundAddr(creator, depositCoins.Add(params.PoolCreationFee...))
	}
	pool, err := s.app.LiquidityKeeper.CreatePool(s.ctx, liquiditytypes.NewMsgCreatePool(creator, pairId, depositCoins))
	s.Require().NoError(err)
	return pool
}

// farming module keeper utils for liquid staking combine test

func (s *KeeperTestSuite) advanceEpochDays() {
	currentEpochDays := s.app.FarmingKeeper.GetCurrentEpochDays(s.ctx)
	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Duration(currentEpochDays) * farmingtypes.Day))
	farming.EndBlocker(s.ctx, s.app.FarmingKeeper)
}

func (s *KeeperTestSuite) CreateFixedAmountPlan(farmingPoolAcc sdk.AccAddress, stakingCoinWeightsMap map[string]string, epochAmountMap map[string]int64) {
	stakingCoinWeights := sdk.NewDecCoins()
	for denom, weight := range stakingCoinWeightsMap {
		stakingCoinWeights = stakingCoinWeights.Add(sdk.NewDecCoinFromDec(denom, sdk.MustNewDecFromStr(weight)))
	}

	epochAmount := sdk.NewCoins()
	for denom, amount := range epochAmountMap {
		epochAmount = epochAmount.Add(sdk.NewInt64Coin(denom, amount))
	}

	msg := farmingtypes.NewMsgCreateFixedAmountPlan(
		fmt.Sprintf("plan%d", s.app.FarmingKeeper.GetGlobalPlanId(s.ctx)+1),
		farmingPoolAcc,
		stakingCoinWeights,
		farmingtypes.ParseTime("0001-01-01T00:00:00Z"),
		farmingtypes.ParseTime("9999-12-31T00:00:00Z"),
		epochAmount,
	)
	_, err := s.app.FarmingKeeper.CreateFixedAmountPlan(s.ctx, msg, farmingPoolAcc, farmingPoolAcc, farmingtypes.PlanTypePublic)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) Stake(farmerAcc sdk.AccAddress, amt sdk.Coins) {
	err := s.app.FarmingKeeper.Stake(s.ctx, farmerAcc, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) Unstake(farmerAcc sdk.AccAddress, amt sdk.Coins) {
	err := s.app.FarmingKeeper.Unstake(s.ctx, farmerAcc, amt)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) assertVotingPower(addr sdk.AccAddress, stakingVotingPower, liquidStakingVotingPower, validatorVotingPower sdk.Int) {
	vp := s.keeper.GetVotingPower(s.ctx, addr)
	s.Require().Equal(stakingVotingPower, vp.StakingVotingPower)
	s.Require().Equal(liquidStakingVotingPower, vp.LiquidStakingVotingPower)
	s.Require().Equal(validatorVotingPower, vp.ValidatorVotingPower)
}
