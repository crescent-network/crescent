package types_test

import (
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/farming/x/farming/types"

	"github.com/tendermint/tendermint/crypto"
)

const (
	DefaultEpochDays = uint32(1)
)

func TestMsgCreateFixedAmountPlan(t *testing.T) {
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	// needs to be deterministic for test
	startTime, _ := time.Parse(time.RFC3339, "2021-11-01T22:08:41+00:00")
	endTime := startTime.AddDate(1, 0, 0)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgCreateFixedAmountPlan
	}{
		{
			"", // empty means no error expected
			types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"invalid farming pool address \"\": empty address string is not allowed: invalid address",
			types.NewMsgCreateFixedAmountPlan(
				sdk.AccAddress{}, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"end time 2020-11-01 22:08:41 +0000 +0000 must be greater than start time 2021-11-01 22:08:41 +0000 +0000: invalid plan end time",
			types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				startTime.AddDate(-1, 0, 0), DefaultEpochDays, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"epoch days must be positive: invalid plan epoch days",
			types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, uint32(0), sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"staking coin weights must not be empty",
			types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr, sdk.NewDecCoins(), startTime,
				endTime, DefaultEpochDays, sdk.Coins{sdk.NewCoin("uatom", sdk.NewInt(1))},
			),
		},
		{
			"epoch amount must not be empty",
			types.NewMsgCreateFixedAmountPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.Coins{},
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
			require.Equal(t, tc.msg.GetPlanCreator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgCreateRatioPlan(t *testing.T) {
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoinWeights := sdk.NewDecCoins(
		sdk.DecCoin{Denom: "testFarmStakingCoinDenom", Amount: sdk.MustNewDecFromStr("1.0")},
	)
	// needs to be deterministic for test
	startTime, _ := time.Parse(time.RFC3339, "2021-11-01T22:08:41+00:00")
	endTime := startTime.AddDate(1, 0, 0)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgCreateRatioPlan
	}{
		{
			"", // empty means no error expected
			types.NewMsgCreateRatioPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.NewDec(1),
			),
		},
		{
			"invalid farming pool address \"\": empty address string is not allowed: invalid address",
			types.NewMsgCreateRatioPlan(
				sdk.AccAddress{}, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.NewDec(1),
			),
		},
		{
			"end time 2020-11-01 22:08:41 +0000 +0000 must be greater than start time 2021-11-01 22:08:41 +0000 +0000: invalid plan end time",
			types.NewMsgCreateRatioPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				startTime.AddDate(-1, 0, 0), DefaultEpochDays, sdk.NewDec(1),
			),
		},
		{
			"epoch days must be positive: invalid plan epoch days",
			types.NewMsgCreateRatioPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, uint32(0), sdk.NewDec(1),
			),
		},
		{
			"staking coin weights must not be empty",
			types.NewMsgCreateRatioPlan(
				farmingPoolAddr, sdk.NewDecCoins(), startTime,
				endTime, DefaultEpochDays, sdk.NewDec(1),
			),
		},
		{
			"invalid plan epoch ratio",
			types.NewMsgCreateRatioPlan(
				farmingPoolAddr, stakingCoinWeights, startTime,
				endTime, DefaultEpochDays, sdk.NewDec(-1),
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
			require.Equal(t, tc.msg.GetPlanCreator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgStake(t *testing.T) {
	planId := uint64(1)
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoins := sdk.NewCoins(
		sdk.NewCoin("testFarmStakingCoinDenom", sdk.NewInt(1)),
	)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgStake
	}{
		{
			"", // empty means no error expected
			types.NewMsgStake(planId, farmingPoolAddr, stakingCoins),
		},
		// TODO" not implemented yet
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
			require.Equal(t, tc.msg.GetStaker(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgUnstake(t *testing.T) {
	planId := uint64(1)
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))
	stakingCoins := sdk.NewCoins(
		sdk.NewCoin("testFarmStakingCoinDenom", sdk.NewInt(1)),
	)

	testCases := []struct {
		expectedErr string
		msg         *types.MsgUnstake
	}{
		{
			"", // empty means no error expected
			types.NewMsgUnstake(planId, farmingPoolAddr, stakingCoins),
		},
		// TODO" not implemented yet
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
			require.Equal(t, tc.msg.GetUnstaker(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgClaim(t *testing.T) {
	planId := uint64(1)
	farmingPoolAddr := sdk.AccAddress(crypto.AddressHash([]byte("farmingPoolAddr")))

	testCases := []struct {
		expectedErr string
		msg         *types.MsgClaim
	}{
		{
			"", // empty means no error expected
			types.NewMsgClaim(planId, farmingPoolAddr),
		},
		// TODO" not implemented yet
	}

	for _, tc := range testCases {
		require.IsType(t, &types.MsgClaim{}, tc.msg)
		require.Equal(t, types.TypeMsgClaim, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetClaimer(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}
