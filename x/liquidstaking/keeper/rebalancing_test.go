package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/keeper"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func (suite *KeeperTestSuite) TestRebalancing() {
	lvs := types.LiquidValidators{
		{
			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(200 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(300 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(400 * 1000000),
			Weight:          sdk.NewInt(10),
		},
	}
	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
	suite.keeper.Rebalancing(suite.ctx, moduleAcc, lvs, sdk.NewDec(10000))
}

func (suite *KeeperTestSuite) TestRebalancingWithDelisting() {
	lvs := types.LiquidValidators{
		{
			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(200 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(300 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper180d0fe0w0eqnn04mwhx8h66hnttgqw32fsr6jg",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(0 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
			Status:          2,
			LiquidTokens:    sdk.NewIntFromUint64(400 * 1000000),
			Weight:          sdk.NewInt(10),
		},
	}
	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
	suite.keeper.Rebalancing(suite.ctx, moduleAcc, lvs, sdk.NewDec(10000))
}

func (suite *KeeperTestSuite) TestRebalancingUnderThreshold() {
	lvs := types.LiquidValidators{
		{
			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(101 * 1000000),
			Weight:          sdk.NewInt(10),
		},
	}
	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
	suite.keeper.Rebalancing(suite.ctx, moduleAcc, lvs, sdk.NewDec(1*1000000))
}

func (suite *KeeperTestSuite) TestRebalancingDiffWeight() {
	lvs := types.LiquidValidators{
		{
			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(20),
		},
		{
			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(200 * 1000000),
			Weight:          sdk.NewInt(20),
		},
		{
			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(300 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(400 * 1000000),
			Weight:          sdk.NewInt(10),
		},
	}
	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
	suite.keeper.Rebalancing(suite.ctx, moduleAcc, lvs, sdk.NewDec(10000))
}

func (suite *KeeperTestSuite) TestRebalancingWithDelistingDiffWeight() {
	lvs := types.LiquidValidators{
		{
			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
			Weight:          sdk.NewInt(30),
		},
		{
			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(200 * 1000000),
			Weight:          sdk.NewInt(20),
		},
		{
			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(300 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper180d0fe0w0eqnn04mwhx8h66hnttgqw32fsr6jg",
			Status:          1,
			LiquidTokens:    sdk.NewIntFromUint64(0 * 1000000),
			Weight:          sdk.NewInt(10),
		},
		{
			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
			Status:          2,
			LiquidTokens:    sdk.NewIntFromUint64(400 * 1000000),
			Weight:          sdk.NewInt(10),
		},
	}
	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
	suite.keeper.Rebalancing(suite.ctx, moduleAcc, lvs, sdk.NewDec(10000))
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
		// TODO: add more cases
	}

	for _, tc := range testCases {
		fmt.Println("------")
		require.IsType(t, types.LiquidValidators{}, tc.activeVals)
		require.IsType(t, sdk.Int{}, tc.addStakingAmt)
		require.IsType(t, map[string]sdk.Int{}, tc.expectedMap)

		totalTargetAmt := sdk.ZeroInt()
		resMap := keeper.AddStakingTargetMap(tc.activeVals, tc.addStakingAmt)
		for k, v := range resMap {
			fmt.Println(k, v.String())
			totalTargetAmt = totalTargetAmt.Add(v)
		}
		require.Equal(t, resMap, tc.expectedMap)
		require.Equal(t, tc.addStakingAmt, totalTargetAmt)
	}
}

//
//func (suite *KeeperTestSuite) TestProcessStaking() {
//	lvs := types.LiquidValidators{
//		{
//			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//		{
//			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//		{
//			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//		{
//			OperatorAddress: "cosmosvaloper1nmfag3hmkx3qyhpmq7jx5996k8uhgh87xhcqfq",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//	}
//	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
//	suite.keeper.ProcessStaking(moduleAcc, lvs, sdk.NewInt(int64(10*1000000)), sdk.NewInt(int64(20*1000000)))
//}
//
//func (suite *KeeperTestSuite) TestProcessStaking2() {
//	lvs := types.LiquidValidators{
//		{
//			OperatorAddress: "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//		{
//			OperatorAddress: "cosmosvaloper1ld6vlyy24906u3aqp5lj54f3nsg2592nm9nj5c",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//		{
//			OperatorAddress: "cosmosvaloper18hfzxheyknesfgcrttr5dg50ffnfphtwtar9fz",
//			Status:          1,
//			LiquidTokens:    sdk.NewIntFromUint64(100 * 1000000),
//			Weight:          sdk.NewInt(10),
//		},
//	}
//	moduleAcc := sdk.AccAddress(crypto.AddressHash([]byte("rebalancing")))
//	suite.keeper.ProcessStaking(moduleAcc, lvs, sdk.NewInt(int64(20*1000000)), sdk.NewInt(int64(10*1000000)))
//}
