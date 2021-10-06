package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

func TestTotalEpochRatio(t *testing.T) {
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
					types.NewBasePlan(1, name2, 1, farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights, startTime, endTime),
					sdk.NewDec(1),
				),
			},
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "total epoch ratio must be lower than 1"),
		},
	}

	for _, tc := range testCases {
		err := types.ValidateTotalEpochRatio(tc.plans)
		if tc.expectedErr == nil {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Equal(t, tc.expectedErr.Error(), err.Error())
		}
	}
}

func TestPrivatePlanFarmingPoolAddress(t *testing.T) {
	testAcc1 := types.PrivatePlanFarmingPoolAddress("test1", 55)
	require.Equal(t, testAcc1, sdk.AccAddress(address.Module(types.ModuleName, []byte("PrivatePlan|55|test1"))))
	require.Equal(t, "cosmos1wce0qjwacezxz42ghqwp6aqvxjt7mu80jywhh09zv2fdv8s4595qk7tzqc", testAcc1.String())

	testAcc2 := types.PrivatePlanFarmingPoolAddress("test2", 1)
	require.Equal(t, testAcc2, sdk.AccAddress(address.Module(types.ModuleName, []byte("PrivatePlan|1|test2"))))
	require.Equal(t, "cosmos172yhzhxwgwul3s8m6qpgw2ww3auedq4k3dt224543d0sd44fgx4spcjthr", testAcc2.String())
}

// TODO: needs to cover more cases
// https://github.com/tendermint/farming/issues/90
func TestUnpackPlan(t *testing.T) {
	plan := []types.PlanI{
		types.NewRatioPlan(
			types.NewBasePlan(
				1,
				"testPlan1",
				types.PlanTypePrivate,
				types.PrivatePlanFarmingPoolAddress("farmingPoolAddr1", 1).String(),
				sdk.AccAddress("terminationAddr1").String(),
				sdk.NewDecCoins(sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")}),
				types.ParseTime("2021-08-03T00:00:00Z"),
				types.ParseTime("2021-08-07T00:00:00Z"),
			),
			sdk.NewDec(1),
		),
	}

	any, err := types.PackPlan(plan[0])
	require.NoError(t, err)

	marshaled, err := any.Marshal()
	require.NoError(t, err)

	any.Value = []byte{}
	err = any.Unmarshal(marshaled)
	require.NoError(t, err)

	reMarshal, err := any.Marshal()
	require.NoError(t, err)
	require.Equal(t, marshaled, reMarshal)

	planRecord := types.PlanRecord{
		Plan:             *any,
		FarmingPoolCoins: sdk.NewCoins(),
	}

	_, err = types.UnpackPlan(&planRecord.Plan)
	require.NoError(t, err)
}
