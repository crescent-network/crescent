package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/farming/types"
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
				return plan.IsTerminated()
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

func TestBasePlanValidate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		malleate    func(*types.BasePlan)
		expectedErr string
	}{
		{
			"happy case",
			func(plan *types.BasePlan) {},
			"",
		},
		{
			"invalid plan type",
			func(plan *types.BasePlan) {
				plan.Type = 3
			},
			"unknown plan type: 3: invalid plan type",
		},
		{
			"invalid farming pool addr",
			func(plan *types.BasePlan) {
				plan.FarmingPoolAddress = "invalid"
			},
			"invalid farming pool address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid termination addr",
			func(plan *types.BasePlan) {
				plan.TerminationAddress = "invalid"
			},
			"invalid termination address \"invalid\": decoding bech32 failed: invalid bech32 string length 7: invalid address",
		},
		{
			"invalid plan name",
			func(plan *types.BasePlan) {
				plan.Name = "a|b|c"
			},
			"plan name cannot contain |: invalid plan name",
		},
		{
			"too long plan name",
			func(plan *types.BasePlan) {
				plan.Name = strings.Repeat("a", 256)
			},
			"plan name cannot be longer than max length of 140: invalid plan name",
		},
		{
			"invalid staking coin weights - empty weights",
			func(plan *types.BasePlan) {
				plan.StakingCoinWeights = sdk.DecCoins{}
			},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid denom",
			func(plan *types.BasePlan) {
				plan.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "!", Amount: sdk.NewDec(1)},
				}
			},
			"invalid staking coin weights: invalid denom: !: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid amount",
			func(plan *types.BasePlan) {
				plan.StakingCoinWeights = sdk.DecCoins{
					sdk.DecCoin{Denom: "stake1", Amount: sdk.NewDec(-1)},
				}
			},
			"invalid staking coin weights: coin -1.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights #1",
			func(plan *types.BasePlan) {
				plan.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(7, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid staking coin weights - invalid sum of weights #2",
			func(plan *types.BasePlan) {
				plan.StakingCoinWeights = sdk.NewDecCoins(
					sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(7, 1)),
					sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(4, 1)),
				)
			},
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid start/end time",
			func(plan *types.BasePlan) {
				plan.StartTime = types.ParseTime("2021-10-01T00:00:00Z")
				plan.EndTime = types.ParseTime("2021-09-30T00:00:00Z")
			},
			"end time 2021-09-30 00:00:00 +0000 UTC must be greater than start time 2021-10-01 00:00:00 +0000 UTC: invalid plan end time",
		},
		{
			"valid distributed coins",
			func(plan *types.BasePlan) {
				plan.DistributedCoins = sdk.NewCoins()
			},
			"",
		},
		{
			"invalid distributed coins - invalid amount",
			func(plan *types.BasePlan) {
				plan.DistributedCoins = sdk.Coins{sdk.Coin{Denom: "reward1", Amount: sdk.ZeroInt()}}
			},
			"invalid distributed coins: coin 0reward1 amount is not positive: invalid coins",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
			tc.malleate(bp)
			err := bp.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestIsPlanActiveAt(t *testing.T) {
	plan := types.NewFixedAmountPlan(
		types.NewBasePlan(
			1,
			"sample plan",
			types.PlanTypePublic,
			sdk.AccAddress(crypto.AddressHash([]byte("address1"))).String(),
			sdk.AccAddress(crypto.AddressHash([]byte("address2"))).String(),
			sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
			types.ParseTime("2021-10-10T00:00:00Z"),
			types.ParseTime("2021-10-15T00:00:00Z"),
		),
		sdk.NewCoins(sdk.NewInt64Coin("reward1", 10000000)),
	)

	for _, tc := range []struct {
		timeStr string
		active  bool
	}{
		{"2021-09-01T00:00:00Z", false},
		{"2021-10-09T23:59:59Z", false},
		{"2021-10-10T00:00:00Z", true},
		{"2021-10-13T12:00:00Z", true},
		{"2021-10-14T23:59:59Z", true},
		{"2021-10-15T00:00:00Z", false},
		{"2021-11-01T00:00:00Z", false},
	} {
		require.Equal(t, tc.active, types.IsPlanActiveAt(plan, types.ParseTime(tc.timeStr)))
	}
}

func TestValidateStakingCoinTotalWeights(t *testing.T) {
	for _, tc := range []struct {
		name               string
		stakingCoinWeights sdk.DecCoins
		expectedErr        string
	}{
		{
			"nil case",
			nil,
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"empty case",
			sdk.DecCoins{},
			"staking coin weights must not be empty: invalid staking coin weights",
		},
		{
			"valid case 1",
			sdk.NewDecCoins(sdk.NewInt64DecCoin("stake1", 1)),
			"",
		},
		{
			"valid case 2",
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
				sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(5, 1)),
			),
			"",
		},
		{
			"invalid case 1",
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(3, 1)),
				sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
			),
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid case 2",
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("stake1", sdk.NewDecWithPrec(5, 1)),
				sdk.NewDecCoinFromDec("stake2", sdk.NewDecWithPrec(6, 1)),
			),
			"total weight must be 1: invalid staking coin weights",
		},
		{
			"invalid case 3",
			sdk.DecCoins{
				sdk.DecCoin{Denom: "stake1", Amount: sdk.NewDec(-1)},
			},
			"invalid staking coin weights: coin -1.000000000000000000stake1 amount is not positive: invalid staking coin weights",
		},
	} {
		err := types.ValidateStakingCoinTotalWeights(tc.stakingCoinWeights)
		if tc.expectedErr == "" {
			require.Nil(t, err)
		} else {
			require.Equal(t, tc.expectedErr, err.Error())
		}
	}
}

func TestTotalEpochRatio(t *testing.T) {
	name1 := "testPlan1"
	name2 := "testPlan2"
	farmingPoolAddr1 := sdk.AccAddress("farmingPoolAddr1")
	farmingPoolAddr2 := sdk.AccAddress("farmingPoolAddr2")
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
			sdkerrors.Wrap(types.ErrInvalidTotalEpochRatio, "total epoch ratio must be lower than 1"),
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(
						1, "plan1", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
				types.NewRatioPlan(
					types.NewBasePlan(
						2, "plan2", types.PlanTypePublic,
						farmingPoolAddr2.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
			},
			nil,
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(
						1, "plan1", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
				types.NewRatioPlan(
					types.NewBasePlan(
						2, "plan2", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2023-01-01T00:00:00Z"), types.ParseTime("2024-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
			},
			nil,
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(
						1, "plan1", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
				types.NewRatioPlan(
					types.NewBasePlan(
						2, "plan2", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-12-31T00:00:00Z"), types.ParseTime("2024-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(7, 1), // 0.7
				),
			},
			sdkerrors.Wrap(types.ErrInvalidTotalEpochRatio, "total epoch ratio must be lower than 1"),
		},
		{
			[]types.PlanI{
				types.NewRatioPlan(
					types.NewBasePlan(
						1, "plan1", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(5, 1), // 0.5
				),
				types.NewRatioPlan(
					types.NewBasePlan(
						2, "plan2", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-01-01T00:00:00Z"), types.ParseTime("2022-07-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(5, 1), // 0.5
				),
				types.NewRatioPlan(
					types.NewBasePlan(
						3, "plan3", types.PlanTypePublic,
						farmingPoolAddr1.String(), terminationAddr1.String(), stakingCoinWeights,
						types.ParseTime("2022-07-01T00:00:00Z"), types.ParseTime("2023-01-01T00:00:00Z"),
					),
					sdk.NewDecWithPrec(5, 1), // 0.5
				),
			},
			nil,
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

func TestPrivatePlanFarmingPoolAcc(t *testing.T) {
	testAcc1 := types.PrivatePlanFarmingPoolAcc("test1", 55)
	require.Equal(t, testAcc1, sdk.AccAddress(address.Module(types.ModuleName, []byte("PrivatePlan|55|test1"))))
	require.Equal(t, "cosmos1wce0qjwacezxz42ghqwp6aqvxjt7mu80jywhh09zv2fdv8s4595qk7tzqc", testAcc1.String())

	testAcc2 := types.PrivatePlanFarmingPoolAcc("test2", 1)
	require.Equal(t, testAcc2, sdk.AccAddress(address.Module(types.ModuleName, []byte("PrivatePlan|1|test2"))))
	require.Equal(t, "cosmos172yhzhxwgwul3s8m6qpgw2ww3auedq4k3dt224543d0sd44fgx4spcjthr", testAcc2.String())
}

func TestUnpackPlan(t *testing.T) {
	plan := []types.PlanI{
		types.NewRatioPlan(
			types.NewBasePlan(
				1,
				"testPlan1",
				types.PlanTypePrivate,
				types.PrivatePlanFarmingPoolAcc("farmingPoolAddr1", 1).String(),
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

	var any2 codectypes.Any
	err = any2.Unmarshal(marshaled)
	require.NoError(t, err)

	reMarshal, err := any2.Marshal()
	require.NoError(t, err)
	require.Equal(t, marshaled, reMarshal)

	planRecord := types.PlanRecord{
		Plan:             any2,
		FarmingPoolCoins: sdk.NewCoins(),
	}

	_, err = types.UnpackPlan(&planRecord.Plan)
	require.NoError(t, err)
}

func TestUnpackPlanJSON(t *testing.T) {
	plan := types.NewRatioPlan(
		types.NewBasePlan(
			1,
			"testPlan1",
			types.PlanTypePrivate,
			types.PrivatePlanFarmingPoolAcc("farmingPoolAddr1", 1).String(),
			sdk.AccAddress("terminationAddr1").String(),
			sdk.NewDecCoins(sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")}),
			types.ParseTime("2021-08-03T00:00:00Z"),
			types.ParseTime("2021-08-07T00:00:00Z"),
		),
		sdk.NewDec(1),
	)

	any, err := types.PackPlan(plan)
	require.NoError(t, err)

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	bz := cdc.MustMarshalJSON(any)

	var any2 codectypes.Any
	err = cdc.UnmarshalJSON(bz, &any2)
	require.NoError(t, err)

	plan2, err := types.UnpackPlan(&any2)
	require.NoError(t, err)

	require.Equal(t, uint64(1), plan2.GetId())
}

func TestValidatePlanName(t *testing.T) {
	for _, tc := range []struct {
		name      string
		expectErr bool
	}{
		{"", true},
		{"valid", false},
		{"valid!", false},
		{" extra space", true},
		{"contains\nline break", true},
		{"Plan #1", false},
		{"contains\x00null", true},
		{"It's valid", false},
		{"With|AccNameSplitter", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidatePlanName(tc.name)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStakingReserveAcc(t *testing.T) {
	testCases := []struct {
		stakingCoinDenom string
		expectedAcc      string
	}{
		{
			"uatom",
			"cosmos1qxs9gxctmd637l7ckpc99kw6ax6thgxx5kshpgzc8kup675xp9dsank7up",
		},
		{
			"stake",
			"cosmos1jn5vt4c3xg38ud89xjl8aumlf3akgdpllmt986w5tj9lureh65dsvk5z3t",
		},
	}

	for _, tc := range testCases {
		acc := types.StakingReserveAcc(tc.stakingCoinDenom)
		require.Equal(t, tc.expectedAcc, acc.String())
	}
}
