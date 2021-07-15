package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/farming/x/farming/types"
)

func TestGetSetNewPlan(t *testing.T) {
	app, ctx := createTestApp(true)

	farmingPoolAddr := sdk.AccAddress([]byte("farmingPoolAddr"))
	terminationAddr := sdk.AccAddress([]byte("terminationAddr"))
	farmerAddr := sdk.AccAddress([]byte("farmer"))
	stakingCoins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000000)))
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)
	basePlan := types.NewBasePlan(1, 1, farmingPoolAddr.String(), terminationAddr.String(), coinWeights, startTime, endTime, 1)
	fixedPlan := types.NewFixedAmountPlan(basePlan, sdk.NewCoins(sdk.NewCoin("testFarmCoinDenom", sdk.NewInt(1000000))))
	app.FarmingKeeper.SetPlan(ctx, fixedPlan)

	planGet, found := app.FarmingKeeper.GetPlan(ctx, 1)
	require.True(t, found)
	require.Equal(t, fixedPlan, planGet)

	plans := app.FarmingKeeper.GetAllPlans(ctx)
	require.Len(t, plans, 1)
	require.Equal(t, fixedPlan, plans[0])

	// TODO: tmp test codes for testing functionality, need to separated
	msgStake := types.NewMsgStake(fixedPlan.Id, farmerAddr, stakingCoins)
	app.FarmingKeeper.Stake(ctx, msgStake)

	stakings := app.FarmingKeeper.GetAllStakings(ctx)
	stakingsByPlan := app.FarmingKeeper.GetStakingsByPlanID(ctx, fixedPlan.Id)
	require.Equal(t, stakings, stakingsByPlan)
	plansByFarmer := app.FarmingKeeper.GetPlansByFarmerAddrIndex(ctx, farmerAddr)

	require.Equal(t, plans, plansByFarmer)
}
