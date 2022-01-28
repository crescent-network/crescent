package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	farmingtypes "github.com/cosmosquad-labs/squad/x/farming/types"
	liquiditytypes "github.com/cosmosquad-labs/squad/x/liquidity/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/cosmosquad-labs/squad/app"
	"github.com/cosmosquad-labs/squad/x/liquidstaking/keeper"
)

var (
	//initialBalances = sdk.NewCoins(
	//	sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	//	sdk.NewInt64Coin(denom1, 1_000_000_000),
	//	sdk.NewInt64Coin(denom2, 1_000_000_000),
	//	sdk.NewInt64Coin(denom3, 1_000_000_000))
	//smallBalances = mustParseCoinsNormalized("1denom1,2denom2,3denom3,1000000000stake")
	PKs = simapp.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	app        *simapp.SquadApp
	ctx        sdk.Context
	keeper     keeper.Keeper
	querier    keeper.Querier
	govHandler govtypes.Handler
	addrs      []sdk.AccAddress
	delAddrs   []sdk.AccAddress
	valAddrs   []sdk.ValAddress
	//sourceAddrs           []sdk.AccAddress
	//destinationAddrs      []sdk.AccAddress
	//whitelistedValidators []liquiditytypes.WhitelistedValidator
}

//func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
//	return proposal.NewParameterChangeProposal("title", "description", changes)
//}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = simapp.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.govHandler = params.NewParamChangeProposalHandler(suite.app.ParamsKeeper)
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.MaxEntries = 200
	stakingParams.MaxValidators = 30
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	suite.keeper = suite.app.LiquidStakingKeeper
	suite.querier = keeper.Querier{Keeper: suite.keeper}
	suite.addrs = simapp.AddTestAddrs(suite.app, suite.ctx, 10, sdk.NewInt(1_000_000_000))
	suite.delAddrs = simapp.AddTestAddrs(suite.app, suite.ctx, 10, sdk.NewInt(1_000_000_000))
	suite.valAddrs = simapp.ConvertAddrsToValAddrs(suite.delAddrs)
	//dAddr1 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "destinationAddr1")
	//dAddr2 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "destinationAddr2")
	//dAddr3 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "destinationAddr3")
	//dAddr4 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "destinationAddr4")
	//dAddr5 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "destinationAddr5")
	//dAddr6 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, "farming", "GravityDEXFarmingWhitelistedValidator")
	//sAddr1 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "sourceAddr1")
	//sAddr2 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "sourceAddr2")
	//sAddr3 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "sourceAddr3")
	//sAddr4 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "sourceAddr4")
	//sAddr5 := liquiditytypes.DeriveAddress(liquiditytypes.AddressType32Bytes, liquiditytypes.ModuleName, "sourceAddr5")
	//sAddr6 := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, authtypes.FeeCollectorName).GetAddress()
	//suite.destinationAddrs = []sdk.AccAddress{dAddr1, dAddr2, dAddr3, dAddr4, dAddr5, dAddr6}
	//suite.sourceAddrs = []sdk.AccAddress{sAddr1, sAddr2, sAddr3, sAddr4, sAddr5, sAddr6}
	//for _, addr := range append(suite.addrs, suite.sourceAddrs[:3]...) {
	//	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
	//	suite.Require().NoError(err)
	//}
	//err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, suite.sourceAddrs[3], smallBalances)
	//suite.Require().NoError(err)

	//suite.whitelistedValidators = []liquiditytypes.WhitelistedValidator{
	//	{
	//		ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
	//		Weight:           sdk.MustNewDecFromStr("0.5"),
	//	},
	//}
}

//func coinsEq(exp, got sdk.Coins) (bool, string, string, string) {
//	return exp.IsEqual(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
//}
//
//func mustParseCoinsNormalized(coinStr string) (coins sdk.Coins) {
//	coins, err := sdk.ParseCoinsNormalized(coinStr)
//	if err != nil {
//		panic(err)
//	}
//	return coins
//}

func (suite *KeeperTestSuite) CreateValidators(powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})
	num := len(powers)
	addrs := simapp.AddTestAddrsIncremental(suite.app, suite.ctx, num, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(num)

	for i, power := range powers {
		val, err := stakingtypes.NewValidator(valAddrs[i], pks[i], stakingtypes.Description{})
		suite.Require().NoError(err)
		suite.app.StakingKeeper.SetValidator(suite.ctx, val)
		err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, val)
		suite.Require().NoError(err)
		suite.app.StakingKeeper.SetNewValidatorByPowerIndex(suite.ctx, val)
		suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, val.GetOperator())
		newShares, err := suite.app.StakingKeeper.Delegate(suite.ctx, addrs[i], sdk.NewInt(power), stakingtypes.Unbonded, val, true)
		suite.Require().NoError(err)
		suite.Require().Equal(newShares.TruncateInt(), sdk.NewInt(power))
	}

	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	return addrs, valAddrs
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

func (s *KeeperTestSuite) AdvanceEpoch() {
	err := s.app.FarmingKeeper.AdvanceEpoch(s.ctx)
	s.Require().NoError(err)
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
