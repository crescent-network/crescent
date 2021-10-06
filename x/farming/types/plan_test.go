package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tendermint/farming/x/farming/types"
)

func TestPlanI(t *testing.T) {
	bp := types.NewBasePlan(
		1,
		"sample plan",
		types.PlanTypePublic,
		sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
		sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
		sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
		types.ParseTime("0001-01-01T00:00:00Z"),
		types.ParseTime("9999-12-31T00:00:00Z"),
	)
	plan := types.NewFixedAmountPlan(bp, sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)))
	lastDistributionTime := types.ParseTime("2021-11-01T00:00:00Z")

	require.Equal(t, bp, plan.GetBasePlan())

	for _, tc := range []struct {
		name           string
		get            func() interface{}
		set            func(types.PlanI, interface{}) error
		oldVal, newVal interface{}
		equal          func(interface{}, interface{}) bool
	}{
		{
			"Id",
			func() interface{} {
				return plan.GetId()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetId(val.(uint64))
			},
			uint64(1), uint64(2),
			nil,
		},
		{
			"Name",
			func() interface{} {
				return plan.GetName()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetName(val.(string))
			},
			"sample plan", "new plan",
			nil,
		},
		{
			"Type",
			func() interface{} {
				return plan.GetType()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetType(val.(types.PlanType))
			},
			types.PlanTypePublic, types.PlanTypePrivate,
			nil,
		},
		{
			"FarmingPoolAddress",
			func() interface{} {
				return plan.GetFarmingPoolAddress()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetFarmingPoolAddress(val.(sdk.AccAddress))
			},
			sdk.AccAddress(crypto.AddressHash([]byte("address1"))),
			sdk.AccAddress(crypto.AddressHash([]byte("address3"))),
			func(a, b interface{}) bool {
				return a.(sdk.AccAddress).Equals(b.(sdk.AccAddress))
			},
		},
		{
			"TerminationAddress",
			func() interface{} {
				return plan.GetTerminationAddress()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetTerminationAddress(val.(sdk.AccAddress))
			},
			sdk.AccAddress(crypto.AddressHash([]byte("address2"))),
			sdk.AccAddress(crypto.AddressHash([]byte("address4"))),
			func(a, b interface{}) bool {
				return a.(sdk.AccAddress).Equals(b.(sdk.AccAddress))
			},
		},
		{
			"StakingCoinWeights",
			func() interface{} {
				return plan.GetStakingCoinWeights()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetStakingCoinWeights(val.(sdk.DecCoins))
			},
			sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
				sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(5, 1)),
			),
			func(a, b interface{}) bool {
				return a.(sdk.DecCoins).IsEqual(b.(sdk.DecCoins))
			},
		},
		{
			"StartTime",
			func() interface{} {
				return plan.GetStartTime()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetStartTime(val.(time.Time))
			},
			types.ParseTime("0001-01-01T00:00:00Z"),
			types.ParseTime("2021-10-01T00:00:00Z"),
			nil,
		},
		{
			"EndTime",
			func() interface{} {
				return plan.GetEndTime()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetEndTime(val.(time.Time))
			},
			types.ParseTime("9999-12-31T00:00:00Z"),
			types.ParseTime("2021-12-31T00:00:00Z"),
			nil,
		},
		{
			"Terminated",
			func() interface{} {
				return plan.GetTerminated()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetTerminated(val.(bool))
			},
			false, true,
			nil,
		},
		{
			"LastDistributionTime",
			func() interface{} {
				return plan.GetLastDistributionTime()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetLastDistributionTime(val.(*time.Time))
			},
			(*time.Time)(nil), &lastDistributionTime,
			func(a, b interface{}) bool {
				at := a.(*time.Time)
				bt := b.(*time.Time)
				if at == nil && bt == nil {
					return true
				} else if (at == nil) != (bt == nil) {
					return false
				}
				return (*at).Equal(*bt)
			},
		},
		{
			"DistributedCoins",
			func() interface{} {
				return plan.GetDistributedCoins()
			},
			func(plan types.PlanI, val interface{}) error {
				return plan.SetDistributedCoins(val.(sdk.Coins))
			},
			sdk.NewCoins(),
			sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
			func(a, b interface{}) bool {
				return a.(sdk.Coins).IsEqual(b.(sdk.Coins))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			val := tc.get()
			if tc.equal != nil {
				require.True(t, tc.equal(tc.oldVal, val))
			} else {
				require.Equal(t, tc.oldVal, val)
			}
			err := tc.set(plan, tc.newVal)
			require.NoError(t, err)
			val = tc.get()
			if tc.equal != nil {
				require.True(t, tc.equal(tc.newVal, val))
			} else {
				require.Equal(t, tc.newVal, val)
			}
		})
	}
}

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
