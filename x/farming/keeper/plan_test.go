package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/farming/app"
	"github.com/tendermint/farming/x/farming/types"
)

func TestGetSetNewPlan(t *testing.T) {
	simapp, ctx := createTestApp(true)

	farmingPoolAddr := sdk.AccAddress([]byte("farmingPoolAddr"))
	terminationAddr := sdk.AccAddress([]byte("terminationAddr"))
	stakingCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	coinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)

	addrs := app.AddTestAddrs(simapp, ctx, 2, sdk.NewInt(2000000))
	farmerAddr := addrs[0]

	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)
	basePlan := types.NewBasePlan(1, 1, farmingPoolAddr.String(), terminationAddr.String(), coinWeights, startTime, endTime)
	fixedPlan := types.NewFixedAmountPlan(basePlan, sdk.NewCoins(sdk.NewCoin("testFarmCoinDenom", sdk.NewInt(1000000))))
	simapp.FarmingKeeper.SetPlan(ctx, fixedPlan)

	planGet, found := simapp.FarmingKeeper.GetPlan(ctx, 1)
	require.True(t, found)
	require.Equal(t, fixedPlan, planGet)

	plans := simapp.FarmingKeeper.GetAllPlans(ctx)
	require.Len(t, plans, 1)
	require.Equal(t, fixedPlan, plans[0])

	// TODO: tmp test codes for testing functionality, need to separated
	err := simapp.FarmingKeeper.Stake(ctx, farmerAddr, stakingCoins)
	require.NoError(t, err)

	stakings := simapp.FarmingKeeper.GetAllStakings(ctx)
	fmt.Println(stakings)
	stakingByFarmer, found := simapp.FarmingKeeper.GetStakingByFarmer(ctx, farmerAddr)
	stakingsByDenom := simapp.FarmingKeeper.GetStakingsByStakingCoinDenom(ctx, sdk.DefaultBondDenom)

	require.True(t, found)
	require.Equal(t, stakings[0], stakingByFarmer)
	require.Equal(t, stakings, stakingsByDenom)

	simapp.FarmingKeeper.SetReward(ctx, sdk.DefaultBondDenom, farmerAddr, stakingCoins)

	//rewards := simapp.FarmingKeeper.GetAllRewards(ctx)
	//rewardsByFarmer := simapp.FarmingKeeper.GetRewardsByFarmer(ctx, farmerAddr)
	//rewardsByDenom := simapp.FarmingKeeper.GetRewardsByStakingCoinDenom(ctx, sdk.DefaultBondDenom)
	//
	//require.Equal(t, rewards, rewardsByFarmer)
	//require.Equal(t, rewards, rewardsByDenom)
}
