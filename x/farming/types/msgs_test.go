package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/crescent-network/crescent/v2/x/farming/types"
)

func TestMsgCreateFixedAmountPlan(t *testing.T) {
	name := "test"
	creatorAddr := sdk.AccAddress(crypto.AddressHash([]byte("creatorPoolAddr")))
	stakingCoinWeights := sdk.NewDecCoins(sdk.DecCoin{Denom: "farmingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")})
	startTime, _ := time.Parse(time.RFC3339, "2021-11-01T22:08:41+00:00") // needs to be deterministic for test
	endTime := startTime.AddDate(1, 0, 0)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgCreateFixedAmountPlan
	}{
		{
			"", // empty means no error expected
			types.NewMsgCreateFixedAmountPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, endTime, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"invalid creator address \"\": empty address string is not allowed: invalid address",
			types.NewMsgCreateFixedAmountPlan(
				name, sdk.AccAddress{}, stakingCoinWeights,
				startTime, endTime, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"end time 2020-11-01T22:08:41Z must be greater than start time 2021-11-01T22:08:41Z: invalid plan end time",
			types.NewMsgCreateFixedAmountPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, startTime.AddDate(-1, 0, 0), sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"staking coin weights must not be empty: invalid staking coin weights",
			types.NewMsgCreateFixedAmountPlan(
				name, creatorAddr, sdk.NewDecCoins(),
				startTime, endTime, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"epoch amount must not be empty: invalid request",
			types.NewMsgCreateFixedAmountPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, endTime, sdk.Coins{},
			),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgCreateFixedAmountPlan{}, tc.msg)
		require.Equal(t, types.TypeMsgCreateFixedAmountPlan, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetCreator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgCreateRatioPlan(t *testing.T) {
	name := "test"
	creatorAddr := sdk.AccAddress(crypto.AddressHash([]byte("creatorAddr")))
	stakingCoinWeights := sdk.NewDecCoins(sdk.DecCoin{Denom: "farmingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")})
	startTime, _ := time.Parse(time.RFC3339, "2021-11-01T22:08:41+00:00") // needs to be deterministic for test
	endTime := startTime.AddDate(1, 0, 0)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgCreateRatioPlan
	}{
		{
			"", // empty means no error expected
			types.NewMsgCreateRatioPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, endTime, sdk.NewDec(1),
			),
		},
		{
			"invalid creator address \"\": empty address string is not allowed: invalid address",
			types.NewMsgCreateRatioPlan(
				name, sdk.AccAddress{}, stakingCoinWeights,
				startTime, endTime, sdk.NewDec(1),
			),
		},
		{
			"end time 2020-11-01T22:08:41Z must be greater than start time 2021-11-01T22:08:41Z: invalid plan end time",
			types.NewMsgCreateRatioPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, startTime.AddDate(-1, 0, 0), sdk.NewDec(1),
			),
		},
		{
			"staking coin weights must not be empty: invalid staking coin weights",
			types.NewMsgCreateRatioPlan(
				name, creatorAddr, sdk.NewDecCoins(),
				startTime, endTime, sdk.NewDec(1),
			),
		},
		{
			"epoch ratio must be positive: -1.000000000000000000: invalid request",
			types.NewMsgCreateRatioPlan(
				name, creatorAddr, stakingCoinWeights,
				startTime, endTime, sdk.NewDec(-1),
			),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgCreateRatioPlan{}, tc.msg)
		require.Equal(t, types.TypeMsgCreateRatioPlan, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetCreator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgStake(t *testing.T) {
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoins := sdk.NewCoins(sdk.NewCoin("farmingCoinDenom", sdk.NewInt(1)))

	testCases := []struct {
		expectedErr string
		msg         *types.MsgStake
	}{
		{
			"", // empty means no error expected
			types.NewMsgStake(farmingPoolAddr, stakingCoins),
		},
		{
			"invalid farmer address \"\": empty address string is not allowed: invalid address",
			types.NewMsgStake(sdk.AccAddress{}, stakingCoins),
		},
		{
			"staking coins must not be zero: invalid request",
			types.NewMsgStake(farmingPoolAddr, sdk.NewCoins(sdk.NewCoin("farmingCoinDenom", sdk.NewInt(0)))),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgStake{}, tc.msg)
		require.Equal(t, types.TypeMsgStake, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetFarmer(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgUnstake(t *testing.T) {
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoins := sdk.NewCoins(sdk.NewCoin("farmingCoinDenom", sdk.NewInt(1)))

	testCases := []struct {
		expectedErr string
		msg         *types.MsgUnstake
	}{
		{
			"", // empty means no error expected
			types.NewMsgUnstake(farmingPoolAddr, stakingCoins),
		},
		{
			"invalid farmer address \"\": empty address string is not allowed: invalid address",
			types.NewMsgUnstake(sdk.AccAddress{}, stakingCoins),
		},
		{
			"unstaking coins must not be zero: invalid request",
			types.NewMsgUnstake(farmingPoolAddr, sdk.NewCoins(sdk.NewInt64Coin("farmingCoinDenom", 0))),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgUnstake{}, tc.msg)
		require.Equal(t, types.TypeMsgUnstake, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetFarmer(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgHarvest(t *testing.T) {
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoinDenoms := []string{"uatom", "uiris", "ukava"}

	testCases := []struct {
		expectedErr string
		msg         *types.MsgHarvest
	}{
		{
			"", // empty means no error expected
			types.NewMsgHarvest(farmingPoolAddr, stakingCoinDenoms),
		},
		{
			"invalid farmer address \"\": empty address string is not allowed: invalid address",
			types.NewMsgHarvest(sdk.AccAddress{}, stakingCoinDenoms),
		},
		{
			"staking coin denoms must be provided at least one: invalid request",
			types.NewMsgHarvest(farmingPoolAddr, []string{}),
		},
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgHarvest{}, tc.msg)
		require.Equal(t, types.TypeMsgHarvest, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetFarmer(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}
