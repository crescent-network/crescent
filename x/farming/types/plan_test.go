package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/farming/x/farming/types"
)

func TestGetPoolInformation(t *testing.T) {
	commonTerminationAcc := sdk.AccAddress([]byte("terminationAddr"))
	commonStartTime := time.Now().UTC()
	commonEndTime := commonStartTime.AddDate(1, 0, 0)
	commonEpochDays := uint32(1)
	commonCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)

	testCases := []struct {
		planId          uint64
		planType        types.PlanType
		farmingPoolAddr string
		rewardPoolAddr  string
		terminationAddr string
		reserveAddr     string
		coinWeights     sdk.DecCoins
	}{
		{
			planId:          uint64(1),
			planType:        types.PlanTypePublic,
			farmingPoolAddr: sdk.AccAddress([]byte("farmingPoolAddr1")).String(),
			rewardPoolAddr:  "cosmos1yqurgw7xa94psk95ctje76ferlddg8vykflaln6xsgarj5w6jkrsuvh9dj",
			reserveAddr:     "cosmos18f2zl0q0gpexruasqzav2vfwdthl4779gtmdxgqdpdl03sq9eygq42ff0u",
		},
	}

	for _, tc := range testCases {
		planName := types.PlanName(tc.planId, tc.planType, tc.farmingPoolAddr)
		rewardPoolAcc := types.GenerateRewardPoolAcc(planName)
		stakingReserveAcc := types.GenerateStakingReserveAcc(planName)
		basePlan := types.NewBasePlan(tc.planId, tc.planType, tc.farmingPoolAddr, commonTerminationAcc.String(), commonCoinWeights, commonStartTime, commonEndTime, commonEpochDays)
		require.Equal(t, basePlan.RewardPoolAddress, rewardPoolAcc.String())
		require.Equal(t, basePlan.StakingReserveAddress, stakingReserveAcc.String())
	}
}
