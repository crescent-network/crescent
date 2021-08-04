package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

func TestGetPoolInformation(t *testing.T) {
	commonTerminationAcc := sdk.AccAddress("terminationAddr")
	commonStartTime := time.Now().UTC()
	commonEndTime := commonStartTime.AddDate(1, 0, 0)
	commonCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)

	testCases := []struct {
		planId          uint64
		name            string
		planType        types.PlanType
		farmingPoolAddr string
		rewardPoolAddr  string
		terminationAddr string
		reserveAddr     string
		coinWeights     sdk.DecCoins
	}{
		{
			planId:          uint64(1),
			name:            "",
			planType:        types.PlanTypePublic,
			farmingPoolAddr: sdk.AccAddress("farmingPoolAddr1").String(),
			rewardPoolAddr:  "cosmos1yqurgw7xa94psk95ctje76ferlddg8vykflaln6xsgarj5w6jkrsuvh9dj",
			reserveAddr:     "cosmos18f2zl0q0gpexruasqzav2vfwdthl4779gtmdxgqdpdl03sq9eygq42ff0u",
		},
	}

	for _, tc := range testCases {
		uniqueKey := types.PlanUniqueKey(tc.planId, tc.planType, tc.farmingPoolAddr)
		rewardPoolAcc := types.GenerateRewardPoolAcc(uniqueKey)
		basePlan := types.NewBasePlan(tc.planId, tc.name, tc.planType, tc.farmingPoolAddr, commonTerminationAcc.String(), commonCoinWeights, commonStartTime, commonEndTime)
		require.Equal(t, basePlan.RewardPoolAddress, rewardPoolAcc.String())
	}
}

func TestRatioPlans(t *testing.T) {
	name1 := "testPlan1"
	name2 := "testPlan2"
	farmingPoolAddr1 := sdk.AccAddress("farmingPoolAddr1")
	terminationAddr1 := sdk.AccAddress("terminationAddr1")
	stakingCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	startTime := time.Now().UTC()
	endTime := startTime.AddDate(1, 0, 0)

	testCases := []struct {
		plans       []types.PlanI
		expectedErr error
	}{
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(1, name1, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
			},
			nil,
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(1, name1, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
				types.NewRatioPlan(
					types.NewBasePlan(1, name1, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
			},
			sdkerrors.Wrap(types.ErrDuplicatePlanName, name1),
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(1, name1, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
				types.NewRatioPlan(
					types.NewBasePlan(1, name2, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
			},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total epoch ratio must be lower than 1"),
		},
	}

	for _, tc := range testCases {
		err := types.ValidateRatioPlans(tc.plans)
		if tc.expectedErr == nil {
			require.NoError(t, err)
		} else {
			require.Equal(t, tc.expectedErr.Error(), err.Error())
		}
	}
}
