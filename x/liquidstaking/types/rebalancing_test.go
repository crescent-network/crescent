package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/crescent-network/crescent/v2/x/liquidstaking/types"
)

var (
	liquidValidators = []types.LiquidValidator{
		{
			OperatorAddress: "cosmosvaloper15kdfwczhpmccprekhlzrvkhzw92940l3w37qqj",
		},
		{
			OperatorAddress: "cosmosvaloper1x73gyvh74ahs2rt9cqrpjkkk74nczwfpnskv3rczmsf0m6aj5dksqr58m3",
		},
		{
			OperatorAddress: "cosmosvaloper10ngyx42lfpylpllm4k3g7fz4gufnt3ptyhm5pn",
		},
		{
			OperatorAddress: "cosmosvaloper10fcwju2n8vvffkp8judj3skqpvnphasxjar5yx",
		},
	}
)

func TestDivideByWeight(t *testing.T) {
	testCases := []struct {
		whitelistedVals  []types.WhitelistedValidator
		addStakingAmt    sdk.Int
		currentDelShares []sdk.Int
		expectedOutputs  []sdk.Int
		expectedCrumb    sdk.Int
	}{
		{
			whitelistedVals: []types.WhitelistedValidator{
				{
					ValidatorAddress: liquidValidators[0].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
				{
					ValidatorAddress: liquidValidators[1].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
				{
					ValidatorAddress: liquidValidators[2].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
			},
			addStakingAmt:    sdk.NewInt(10 * 1000000),
			currentDelShares: []sdk.Int{sdk.NewInt(2000000), sdk.NewInt(2000000), sdk.NewInt(1000000)},
			expectedOutputs:  []sdk.Int{sdk.NewInt(3333333), sdk.NewInt(3333333), sdk.NewInt(3333333)},
			expectedCrumb:    sdk.NewInt(1),
		},
		{
			whitelistedVals: []types.WhitelistedValidator{
				{
					ValidatorAddress: liquidValidators[0].OperatorAddress,
					TargetWeight:     sdk.NewInt(2),
				},
				{
					ValidatorAddress: liquidValidators[1].OperatorAddress,
					TargetWeight:     sdk.NewInt(2),
				},
				{
					ValidatorAddress: liquidValidators[2].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
			},
			addStakingAmt:    sdk.NewInt(10 * 1000000),
			currentDelShares: []sdk.Int{sdk.NewInt(1000000), sdk.NewInt(1000000), sdk.NewInt(1000000)},
			expectedOutputs:  []sdk.Int{sdk.NewInt(4000000), sdk.NewInt(4000000), sdk.NewInt(2000000)},
			expectedCrumb:    sdk.NewInt(0),
		},
		{
			whitelistedVals: []types.WhitelistedValidator{
				{
					ValidatorAddress: liquidValidators[0].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
				{
					ValidatorAddress: liquidValidators[1].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
				{
					ValidatorAddress: liquidValidators[2].OperatorAddress,
					TargetWeight:     sdk.NewInt(1),
				},
			},
			addStakingAmt:    sdk.NewInt(10),
			currentDelShares: []sdk.Int{sdk.NewInt(3), sdk.NewInt(2), sdk.NewInt(1)},
			expectedOutputs:  []sdk.Int{sdk.NewInt(3), sdk.NewInt(3), sdk.NewInt(3)},
			expectedCrumb:    sdk.NewInt(1),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, []types.WhitelistedValidator{}, tc.whitelistedVals)
		require.IsType(t, sdk.Int{}, tc.addStakingAmt)
		require.IsType(t, sdk.Int{}, tc.expectedCrumb)
		require.IsType(t, []sdk.Int{}, tc.expectedOutputs)

		totalTargetAmt := sdk.ZeroInt()
		valsMap := types.GetWhitelistedValsMap(tc.whitelistedVals)
		var activeVals types.ActiveLiquidValidators
		for _, v := range tc.whitelistedVals {
			activeVals = append(activeVals, types.LiquidValidator{
				OperatorAddress: v.ValidatorAddress,
			})
		}
		outputs, crumb := types.DivideByWeight(activeVals, tc.addStakingAmt, valsMap)
		for _, v := range outputs {
			totalTargetAmt = totalTargetAmt.Add(v)
		}
		require.EqualValues(t, tc.expectedOutputs, outputs)
		require.EqualValues(t, tc.addStakingAmt, totalTargetAmt.Add(crumb))
		require.Equal(t, tc.expectedCrumb.String(), crumb.String())
	}
}

func TestMinMaxGap(t *testing.T) {
	testCases := []struct {
		name                     string
		liquidVals               types.LiquidValidators
		targetMap                map[string]sdk.Int
		liquidTokenMap           map[string]sdk.Int
		expectedMinGapVal        types.LiquidValidator
		expectedMaxGapVal        types.LiquidValidator
		expectedAmountNeeded     sdk.Int
		expectedLastRedelegation bool
	}{
		{
			name:       "zero case",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[1].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[2].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[1].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[2].OperatorAddress: sdk.ZeroInt(),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			expectedMinGapVal:        types.LiquidValidator{},
			expectedMaxGapVal:        types.LiquidValidator{},
			expectedAmountNeeded:     sdk.ZeroInt(),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 1-1",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			expectedMinGapVal:        liquidValidators[3],
			expectedMaxGapVal:        liquidValidators[0],
			expectedAmountNeeded:     sdk.NewInt(33333334),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 1-2",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334 - 33333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.NewInt(0 + 33333334),
			},
			expectedMinGapVal:        liquidValidators[3],
			expectedMaxGapVal:        liquidValidators[1],
			expectedAmountNeeded:     sdk.NewInt(33333333),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 1-3",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334 - 33333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333 - 33333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.NewInt(33333334 + 33333333),
			},
			expectedMinGapVal:        liquidValidators[3],
			expectedMaxGapVal:        liquidValidators[2],
			expectedAmountNeeded:     sdk.NewInt(33333333),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 1-4",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334 - 33333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333 - 33333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333 - 33333333),
				liquidValidators[3].OperatorAddress: sdk.NewInt(33333334 + 33333333 + 33333333),
			},
			expectedMinGapVal:        types.LiquidValidator{},
			expectedMaxGapVal:        types.LiquidValidator{},
			expectedAmountNeeded:     sdk.ZeroInt(),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 2-1",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000),
			},
			expectedMinGapVal:        liquidValidators[0],
			expectedMaxGapVal:        liquidValidators[3],
			expectedAmountNeeded:     sdk.NewInt(33333334),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 2-2",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000 + 33333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000 - 33333334),
			},
			expectedMinGapVal:        liquidValidators[1],
			expectedMaxGapVal:        liquidValidators[3],
			expectedAmountNeeded:     sdk.NewInt(33333333),
			expectedLastRedelegation: false,
		},
		{
			name:       "rebalancing case 2-3, last redelegation",
			liquidVals: liquidValidators,
			targetMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(133333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(133333333),
				liquidValidators[3].OperatorAddress: sdk.ZeroInt(),
			},
			liquidTokenMap: map[string]sdk.Int{
				liquidValidators[0].OperatorAddress: sdk.NewInt(100000000 + 33333334),
				liquidValidators[1].OperatorAddress: sdk.NewInt(100000000 + 33333333),
				liquidValidators[2].OperatorAddress: sdk.NewInt(100000000),
				liquidValidators[3].OperatorAddress: sdk.NewInt(100000000 - 33333334 - 33333333),
			},
			expectedMinGapVal:        liquidValidators[2],
			expectedMaxGapVal:        liquidValidators[3],
			expectedAmountNeeded:     sdk.NewInt(33333333),
			expectedLastRedelegation: true,
		},
	}

	for _, tc := range testCases {
		minGapVal, maxGapVal, amountNeeded, last := tc.liquidVals.MinMaxGap(tc.targetMap, tc.liquidTokenMap)
		require.EqualValues(t, minGapVal, tc.expectedMinGapVal)
		require.EqualValues(t, maxGapVal, tc.expectedMaxGapVal)
		require.EqualValues(t, amountNeeded, tc.expectedAmountNeeded)
		require.EqualValues(t, last, tc.expectedLastRedelegation)
	}
}

func TestDivideByCurrentWeight(t *testing.T) {
	testCases := []struct {
		liquidValidators []types.LiquidValidatorState
		addStakingAmt    sdk.Dec
		expectedOutputs  []sdk.Dec
		expectedCrumb    sdk.Dec
	}{
		{
			liquidValidators: []types.LiquidValidatorState{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
				},
			},
			addStakingAmt:   sdk.NewDec(10 * 1000000),
			expectedOutputs: []sdk.Dec{sdk.NewDec(4 * 1000000), sdk.NewDec(4 * 1000000), sdk.NewDec(2 * 1000000)},
			expectedCrumb:   sdk.NewDec(0),
		},
		{
			liquidValidators: []types.LiquidValidatorState{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
					Weight:          sdk.NewInt(2),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
					Weight:          sdk.NewInt(2),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt:   sdk.NewDec(10 * 1000000),
			expectedOutputs: []sdk.Dec{sdk.MustNewDecFromStr("3333333.000000000000000000"), sdk.MustNewDecFromStr("3333333.000000000000000000"), sdk.MustNewDecFromStr("3333333.000000000000000000")},
			expectedCrumb:   sdk.MustNewDecFromStr("1.000000000000000000"),
		},
		{
			liquidValidators: []types.LiquidValidatorState{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1),
				},
			},
			addStakingAmt:   sdk.NewDec(10),
			expectedOutputs: []sdk.Dec{sdk.MustNewDecFromStr("4.000000000000000000"), sdk.MustNewDecFromStr("3.000000000000000000"), sdk.MustNewDecFromStr("1.000000000000000000")},
			expectedCrumb:   sdk.MustNewDecFromStr("2.000000000000000000"),
		},
		{
			liquidValidators: []types.LiquidValidatorState{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10000000),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2000000),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3000001),
				},
			},
			addStakingAmt:   sdk.NewDec(10000000),
			expectedOutputs: []sdk.Dec{sdk.MustNewDecFromStr("6666666.000000000000000000"), sdk.MustNewDecFromStr("1333333.000000000000000000"), sdk.MustNewDecFromStr("2000000.000000000000000000")},
			expectedCrumb:   sdk.MustNewDecFromStr("1.000000000000000000"),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, []types.LiquidValidatorState{}, tc.liquidValidators)
		require.IsType(t, sdk.Dec{}, tc.addStakingAmt)
		require.IsType(t, sdk.Dec{}, tc.expectedCrumb)
		require.IsType(t, []sdk.Dec{}, tc.expectedOutputs)

		totalTargetAmt := sdk.ZeroDec()
		totalLiquidTokens := sdk.ZeroInt()
		liquidTokenMap := map[string]sdk.Int{}
		var lvs types.LiquidValidators
		for _, v := range tc.liquidValidators {
			totalLiquidTokens = totalLiquidTokens.Add(v.LiquidTokens)
			liquidTokenMap[v.OperatorAddress] = v.LiquidTokens
			lvs = append(lvs, types.LiquidValidator{
				OperatorAddress: v.OperatorAddress})
		}
		outputs, crumb := types.DivideByCurrentWeight(lvs, tc.addStakingAmt, totalLiquidTokens, liquidTokenMap)
		for _, v := range outputs {
			totalTargetAmt = totalTargetAmt.Add(v)
		}
		require.EqualValues(t, tc.expectedOutputs, outputs)
		require.EqualValues(t, tc.addStakingAmt, totalTargetAmt.Add(crumb))
		require.Equal(t, tc.expectedCrumb.String(), crumb.String())
	}
}
