package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crescent-network/crescent/x/liquidstaking/types"
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
	types.Rebalancing(moduleAcc, lvs, sdk.NewDec(10000))
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
	types.Rebalancing(moduleAcc, lvs, sdk.NewDec(10000))
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
	types.Rebalancing(moduleAcc, lvs, sdk.NewDec(1*1000000))
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
	types.Rebalancing(moduleAcc, lvs, sdk.NewDec(10000))
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
	types.Rebalancing(moduleAcc, lvs, sdk.NewDec(10000))
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
