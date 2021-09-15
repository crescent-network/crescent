package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/x/farming/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	plans := []types.PlanI{
		types.NewFixedAmountPlan(
			types.NewBasePlan(
				1,
				"name1",
				types.PlanTypePrivate,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1))),
				mustParseRFC3339("2021-07-30T00:00:00Z"),
				mustParseRFC3339("2021-08-30T00:00:00Z"),
			),
			sdk.NewCoins(sdk.NewInt64Coin(denom3, 1_000_000))),
		types.NewRatioPlan(
			types.NewBasePlan(
				2,
				"name2",
				types.PlanTypePublic,
				suite.addrs[0].String(),
				suite.addrs[0].String(),
				sdk.NewDecCoins(
					sdk.NewDecCoinFromDec(denom1, sdk.NewDecWithPrec(3, 1)),
					sdk.NewDecCoinFromDec(denom2, sdk.NewDecWithPrec(7, 1))),
				mustParseRFC3339("2021-07-30T00:00:00Z"),
				mustParseRFC3339("2021-08-30T00:00:00Z"),
			),
			sdk.MustNewDecFromStr("0.01")),
	}
	//for _, plan := range plans {
	//	suite.keeper.SetPlan(suite.ctx, plan)
	//}
	suite.keeper.SetPlan(suite.ctx, plans[1])
	suite.keeper.SetPlan(suite.ctx, plans[0])

	suite.Stake(suite.addrs[1], sdk.NewCoins(
		sdk.NewInt64Coin(denom1, 1_000_000),
		sdk.NewInt64Coin(denom2, 1_000_000)))
	suite.keeper.ProcessQueuedCoins(suite.ctx)

	suite.ctx = suite.ctx.WithBlockTime(mustParseRFC3339("2021-07-31T00:00:00Z"))

	// Advance 2 epochs
	err := suite.keeper.AdvanceEpoch(suite.ctx)
	suite.Require().NoError(err)
	err = suite.keeper.AdvanceEpoch(suite.ctx)
	suite.Require().NoError(err)

	var genState *types.GenesisState
	suite.Require().NotPanics(func() {
		genState = suite.keeper.ExportGenesis(suite.ctx)
	})

	err = types.ValidateGenesis(*genState)
	suite.Require().NoError(err)

	suite.Require().NotPanics(func() {
		suite.keeper.InitGenesis(suite.ctx, *genState)
	})
	suite.Require().Equal(genState, suite.keeper.ExportGenesis(suite.ctx))
}
