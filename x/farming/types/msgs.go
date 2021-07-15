package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreateFixedAmountPlan)(nil)
	_ sdk.Msg = (*MsgCreateRatioPlan)(nil)
	_ sdk.Msg = (*MsgStake)(nil)
	_ sdk.Msg = (*MsgUnstake)(nil)
	_ sdk.Msg = (*MsgClaim)(nil)
)

// Message types for the farming module
const (
	TypeMsgCreateFixedAmountPlan = "create_fixed_amount_plan"
	TypeMsgCreateRatioPlan       = "create_ratio_plan"
	TypeMsgStake                 = "stake"
	TypeMsgUnstake               = "unstake"
	TypeMsgClaim                 = "claim"
)

// NewMsgCreateFixedAmountPlan creates a new MsgCreateFixedAmountPlan.
func NewMsgCreateFixedAmountPlan(
	farmingPoolAddr sdk.AccAddress,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochDays uint32,
	epochAmount sdk.Coins,
) *MsgCreateFixedAmountPlan {
	return &MsgCreateFixedAmountPlan{
		FarmingPoolAddress: farmingPoolAddr.String(),
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
		EpochDays:          epochDays,
		EpochAmount:        epochAmount,
	}
}

func (msg MsgCreateFixedAmountPlan) Route() string { return RouterKey }

func (msg MsgCreateFixedAmountPlan) Type() string { return TypeMsgCreateFixedAmountPlan }

func (msg MsgCreateFixedAmountPlan) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", msg.FarmingPoolAddress, err)
	}
	if !msg.EndTime.After(msg.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", msg.EndTime, msg.StartTime)
	}
	if msg.EpochDays == 0 {
		return sdkerrors.Wrapf(ErrInvalidPlanEpochDays, "epoch days must be positive")
	}
	if msg.StakingCoinWeights.Empty() {
		return ErrEmptyStakingCoinWeights
	}
	if err := msg.StakingCoinWeights.Validate(); err != nil {
		return err
	}
	if msg.EpochAmount.Empty() {
		return ErrEmptyEpochAmount
	}
	if err := msg.EpochAmount.Validate(); err != nil {
		return err
	}
	return nil
}

func (msg MsgCreateFixedAmountPlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateFixedAmountPlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateFixedAmountPlan) GetPlanCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgCreateRatioPlan creates a new MsgCreateRatioPlan.
func NewMsgCreateRatioPlan(
	farmingPoolAddr sdk.AccAddress,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochDays uint32,
	epochRatio sdk.Dec,
) *MsgCreateRatioPlan {
	return &MsgCreateRatioPlan{
		FarmingPoolAddress: farmingPoolAddr.String(),
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
		EpochDays:          epochDays,
		EpochRatio:         epochRatio,
	}
}

func (msg MsgCreateRatioPlan) Route() string { return RouterKey }

func (msg MsgCreateRatioPlan) Type() string { return TypeMsgCreateRatioPlan }

func (msg MsgCreateRatioPlan) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farming pool address %q: %v", msg.FarmingPoolAddress, err)
	}
	if !msg.EndTime.After(msg.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", msg.EndTime, msg.StartTime)
	}
	if msg.EpochDays == 0 {
		return sdkerrors.Wrapf(ErrInvalidPlanEpochDays, "epoch days must be positive")
	}
	if msg.StakingCoinWeights.Empty() {
		return ErrEmptyStakingCoinWeights
	}
	if err := msg.StakingCoinWeights.Validate(); err != nil {
		return err
	}
	if !msg.EpochRatio.IsPositive() {
		return ErrInvalidPlanEpochRatio
	}
	return nil
}

func (msg MsgCreateRatioPlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateRatioPlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateRatioPlan) GetPlanCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FarmingPoolAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgStake creates a new MsgStake.
func NewMsgStake(
	planID uint64,
	farmer sdk.AccAddress,
	stakingCoins sdk.Coins,
) *MsgStake {
	return &MsgStake{
		PlanId:       planID,
		Farmer:       farmer.String(),
		StakingCoins: stakingCoins,
	}
}

func (msg MsgStake) Route() string { return RouterKey }

func (msg MsgStake) Type() string { return TypeMsgStake }

func (msg MsgStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address %q: %v", msg.Farmer, err)
	}
	if err := msg.StakingCoins.Validate(); err != nil {
		return err
	}
	return nil
}

func (msg MsgStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStake) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgStake) GetStaker() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgUnstake creates a new MsgUnstake.
func NewMsgUnstake(
	planID uint64,
	farmer sdk.AccAddress,
	unstakingCoins sdk.Coins,
) *MsgUnstake {
	return &MsgUnstake{
		PlanId:         planID,
		Farmer:         farmer.String(),
		UnstakingCoins: unstakingCoins,
	}
}

func (msg MsgUnstake) Route() string { return RouterKey }

func (msg MsgUnstake) Type() string { return TypeMsgUnstake }

func (msg MsgUnstake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address %q: %v", msg.Farmer, err)
	}
	if err := msg.UnstakingCoins.Validate(); err != nil {
		return err
	}
	return nil
}

func (msg MsgUnstake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgUnstake) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgUnstake) GetUnstaker() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgClaim creates a new MsgClaim.
func NewMsgClaim(
	planID uint64,
	farmer sdk.AccAddress,
) *MsgClaim {
	return &MsgClaim{
		PlanId: planID,
		Farmer: farmer.String(),
	}
}

func (msg MsgClaim) Route() string { return RouterKey }

func (msg MsgClaim) Type() string { return TypeMsgClaim }

func (msg MsgClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address %q: %v", msg.Farmer, err)
	}
	return nil
}

func (msg MsgClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgClaim) GetClaimer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}
