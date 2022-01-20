package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/crescent-network/crescent/x/liquidstaking/types"
)

func TestDivideByCurrentWeight(t *testing.T) {
	testCases := []struct {
		activeVals      types.LiquidValidators
		addStakingAmt   sdk.Int
		expectedOutputs []sdk.Int
		expectedCrumb   sdk.Int
	}{
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt:   sdk.NewInt(10 * 1000000),
			expectedOutputs: []sdk.Int{sdk.NewInt(4 * 1000000), sdk.NewInt(4 * 1000000), sdk.NewInt(2 * 1000000)},
			expectedCrumb:   sdk.NewInt(0),
		},
		{
			activeVals: types.LiquidValidators{
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
			addStakingAmt:   sdk.NewInt(10 * 1000000),
			expectedOutputs: []sdk.Int{sdk.NewInt(3333333), sdk.NewInt(3333333), sdk.NewInt(3333333)},
			expectedCrumb:   sdk.NewInt(1),
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt:   sdk.NewInt(10),
			expectedOutputs: []sdk.Int{sdk.NewInt(4), sdk.NewInt(3), sdk.NewInt(1)},
			expectedCrumb:   sdk.NewInt(2),
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3000001),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt:   sdk.NewInt(10000000),
			expectedOutputs: []sdk.Int{sdk.NewInt(6666666), sdk.NewInt(1333333), sdk.NewInt(2000000)},
			expectedCrumb:   sdk.NewInt(1),
		},
	}

	for _, tc := range testCases {
		fmt.Println("------")
		require.IsType(t, types.LiquidValidators{}, tc.activeVals)
		require.IsType(t, sdk.Int{}, tc.addStakingAmt)
		require.IsType(t, sdk.Int{}, tc.expectedCrumb)
		require.IsType(t, []sdk.Int{}, tc.expectedOutputs)

		totalTargetAmt := sdk.ZeroInt()
		outputs, crumb := types.DivideByCurrentWeight(tc.activeVals, tc.addStakingAmt)
		for k, v := range outputs {
			fmt.Println(k, v.String())
			totalTargetAmt = totalTargetAmt.Add(v)
		}
		require.EqualValues(t, tc.expectedOutputs, outputs)
		require.EqualValues(t, tc.addStakingAmt, totalTargetAmt.Add(crumb))
		require.Equal(t, tc.expectedCrumb.String(), crumb.String())
	}
}

func TestAddStakingTargetMap(t *testing.T) {
	testCases := []struct {
		activeVals    types.LiquidValidators
		addStakingAmt sdk.Int
		expectedMap   map[string]sdk.Int
	}{
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2 * 1000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1 * 1000000),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10 * 1000000),
			expectedMap: map[string]sdk.Int{
				"a": sdk.NewInt(3 * 1000000),
				"b": sdk.NewInt(3 * 1000000),
				"c": sdk.NewInt(4 * 1000000),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"a": sdk.NewInt(3),
				"b": sdk.NewInt(3),
				"c": sdk.NewInt(4),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(8),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(7),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(4),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"a": sdk.NewInt(3),
				"b": sdk.NewInt(2),
				"c": sdk.NewInt(5),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(5),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"b": sdk.NewInt(3),
				"c": sdk.NewInt(7),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(1),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"b": sdk.NewInt(4),
				"c": sdk.NewInt(6),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"b": sdk.NewInt(5),
				"c": sdk.NewInt(5),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10),
			expectedMap: map[string]sdk.Int{
				"b": sdk.NewInt(6),
				"c": sdk.NewInt(4),
			},
		},
		{
			activeVals: types.LiquidValidators{
				{
					OperatorAddress: "a",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(10000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "b",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(2000000),
					Weight:          sdk.NewInt(1),
				},
				{
					OperatorAddress: "c",
					Status:          types.ValidatorStatusActive,
					LiquidTokens:    sdk.NewIntFromUint64(3000001),
					Weight:          sdk.NewInt(1),
				},
			},
			addStakingAmt: sdk.NewInt(10000000),
			expectedMap: map[string]sdk.Int{
				"b": sdk.NewInt(5500001),
				"c": sdk.NewInt(4499999),
			},
		},
	}

	for _, tc := range testCases {
		fmt.Println("------")
		require.IsType(t, types.LiquidValidators{}, tc.activeVals)
		require.IsType(t, sdk.Int{}, tc.addStakingAmt)
		require.IsType(t, map[string]sdk.Int{}, tc.expectedMap)

		totalTargetAmt := sdk.ZeroInt()
		resMap := types.AddStakingTargetMap(tc.activeVals, tc.addStakingAmt)
		for k, v := range resMap {
			fmt.Println(k, v.String())
			totalTargetAmt = totalTargetAmt.Add(v)
		}
		require.Equal(t, resMap, tc.expectedMap)
		require.Equal(t, tc.addStakingAmt, totalTargetAmt)
	}
}
