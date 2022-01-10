package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/crescent-network/crescent/app"
	"github.com/crescent-network/crescent/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
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

	app        *simapp.CrescentApp
	ctx        sdk.Context
	keeper     keeper.Keeper
	querier    keeper.Querier
	govHandler govtypes.Handler
	addrs      []sdk.AccAddress
	delAddrs   []sdk.AccAddress
	valAddrs   []sdk.ValAddress
	//sourceAddrs           []sdk.AccAddress
	//destinationAddrs      []sdk.AccAddress
	whitelistedValidators []types.WhitelistedValidator
}

//func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
//	return proposal.NewParameterChangeProposal("title", "description", changes)
//}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	// TODO: add hooking for stakingkeeper
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
	//dAddr1 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "destinationAddr1")
	//dAddr2 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "destinationAddr2")
	//dAddr3 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "destinationAddr3")
	//dAddr4 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "destinationAddr4")
	//dAddr5 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "destinationAddr5")
	//dAddr6 := types.DeriveAddress(types.AddressType32Bytes, "farming", "GravityDEXFarmingWhitelistedValidator")
	//sAddr1 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "sourceAddr1")
	//sAddr2 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "sourceAddr2")
	//sAddr3 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "sourceAddr3")
	//sAddr4 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "sourceAddr4")
	//sAddr5 := types.DeriveAddress(types.AddressType32Bytes, types.ModuleName, "sourceAddr5")
	//sAddr6 := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, authtypes.FeeCollectorName).GetAddress()
	//suite.destinationAddrs = []sdk.AccAddress{dAddr1, dAddr2, dAddr3, dAddr4, dAddr5, dAddr6}
	//suite.sourceAddrs = []sdk.AccAddress{sAddr1, sAddr2, sAddr3, sAddr4, sAddr5, sAddr6}
	//for _, addr := range append(suite.addrs, suite.sourceAddrs[:3]...) {
	//	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr, initialBalances)
	//	suite.Require().NoError(err)
	//}
	//err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, suite.sourceAddrs[3], smallBalances)
	//suite.Require().NoError(err)

	suite.whitelistedValidators = []types.WhitelistedValidator{
		{
			ValidatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Weight:           sdk.MustNewDecFromStr("0.5"),
		},
	}
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
	num := len(powers)
	addrs := simapp.AddTestAddrsIncremental(suite.app, suite.ctx, num, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(num)
	cdc := simapp.MakeTestEncodingConfig().Marshaler

	suite.app.StakingKeeper = stakingkeeper.NewKeeper(
		cdc,
		suite.app.GetKey(stakingtypes.StoreKey),
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.GetSubspace(stakingtypes.ModuleName),
	)

	for i, power := range powers {
		val, err := stakingtypes.NewValidator(valAddrs[i], pks[i], stakingtypes.Description{})
		suite.Require().NoError(err)
		suite.app.StakingKeeper.SetValidator(suite.ctx, val)
		suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, val)
		suite.app.StakingKeeper.SetNewValidatorByPowerIndex(suite.ctx, val)
		suite.app.DistrKeeper.Hooks().AfterValidatorCreated(suite.ctx, val.GetOperator())
		suite.app.SlashingKeeper.Hooks().AfterValidatorCreated(suite.ctx, val.GetOperator())
		newShares, err := suite.app.StakingKeeper.Delegate(suite.ctx, addrs[i], sdk.NewInt(power), stakingtypes.Unbonded, val, true)
		suite.Require().NoError(err)
		suite.Require().Equal(newShares.TruncateInt(), sdk.NewInt(power))
	}

	_ = staking.EndBlocker(suite.ctx, suite.app.StakingKeeper)
	return addrs, valAddrs
}
