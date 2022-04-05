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
	_ sdk.Msg = (*MsgHarvest)(nil)
	_ sdk.Msg = (*MsgRemovePlan)(nil)
	_ sdk.Msg = (*MsgAdvanceEpoch)(nil)
)

// Message types for the farming module
const (
	TypeMsgCreateFixedAmountPlan = "create_fixed_amount_plan"
	TypeMsgCreateRatioPlan       = "create_ratio_plan"
	TypeMsgStake                 = "stake"
	TypeMsgUnstake               = "unstake"
	TypeMsgHarvest               = "harvest"
	TypeMsgRemovePlan            = "remove_plan"
	TypeMsgAdvanceEpoch          = "advance_epoch"
)

// NewMsgCreateFixedAmountPlan creates a new MsgCreateFixedAmountPlan.
func NewMsgCreateFixedAmountPlan(
	name string,
	creatorAcc sdk.AccAddress,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochAmount sdk.Coins,
) *MsgCreateFixedAmountPlan {
	return &MsgCreateFixedAmountPlan{
		Name:               name,
		Creator:            creatorAcc.String(),
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
		EpochAmount:        epochAmount,
	}
}

func (msg MsgCreateFixedAmountPlan) Route() string { return RouterKey }

func (msg MsgCreateFixedAmountPlan) Type() string { return TypeMsgCreateFixedAmountPlan }

func (msg MsgCreateFixedAmountPlan) ValidateBasic() error {
	if err := ValidatePlanName(msg.Name); err != nil {
		return sdkerrors.Wrap(ErrInvalidPlanName, err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address %q: %v", msg.Creator, err)
	}
	if !msg.EndTime.After(msg.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", msg.EndTime.Format(time.RFC3339), msg.StartTime.Format(time.RFC3339))
	}
	if err := ValidateStakingCoinTotalWeights(msg.StakingCoinWeights); err != nil {
		return err
	}
	if msg.EpochAmount.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "epoch amount must not be empty")
	}
	if err := ValidateEpochAmount(msg.EpochAmount); err != nil {
		return err
	}
	return nil
}

func (msg MsgCreateFixedAmountPlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateFixedAmountPlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateFixedAmountPlan) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgCreateRatioPlan creates a new MsgCreateRatioPlan.
func NewMsgCreateRatioPlan(
	name string,
	creatorAcc sdk.AccAddress,
	stakingCoinWeights sdk.DecCoins,
	startTime time.Time,
	endTime time.Time,
	epochRatio sdk.Dec,
) *MsgCreateRatioPlan {
	return &MsgCreateRatioPlan{
		Name:               name,
		Creator:            creatorAcc.String(),
		StakingCoinWeights: stakingCoinWeights,
		StartTime:          startTime,
		EndTime:            endTime,
		EpochRatio:         epochRatio,
	}
}

func (msg MsgCreateRatioPlan) Route() string { return RouterKey }

func (msg MsgCreateRatioPlan) Type() string { return TypeMsgCreateRatioPlan }

func (msg MsgCreateRatioPlan) ValidateBasic() error {
	if err := ValidatePlanName(msg.Name); err != nil {
		return sdkerrors.Wrap(ErrInvalidPlanName, err.Error())
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address %q: %v", msg.Creator, err)
	}
	if !msg.EndTime.After(msg.StartTime) {
		return sdkerrors.Wrapf(ErrInvalidPlanEndTime, "end time %s must be greater than start time %s", msg.EndTime.Format(time.RFC3339), msg.StartTime.Format(time.RFC3339))
	}
	if err := ValidateStakingCoinTotalWeights(msg.StakingCoinWeights); err != nil {
		return err
	}
	if err := ValidateEpochRatio(msg.EpochRatio); err != nil {
		return err
	}
	return nil
}

func (msg MsgCreateRatioPlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateRatioPlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateRatioPlan) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgStake creates a new MsgStake.
func NewMsgStake(
	farmer sdk.AccAddress,
	stakingCoins sdk.Coins,
) *MsgStake {
	return &MsgStake{
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
	if ok := msg.StakingCoins.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coins must not be zero")
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

func (msg MsgStake) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgUnstake creates a new MsgUnstake.
func NewMsgUnstake(
	farmer sdk.AccAddress,
	unstakingCoins sdk.Coins,
) *MsgUnstake {
	return &MsgUnstake{
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
	if ok := msg.UnstakingCoins.IsZero(); ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "unstaking coins must not be zero")
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

func (msg MsgUnstake) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgHarvest creates a new MsgHarvest.
func NewMsgHarvest(
	farmer sdk.AccAddress,
	stakingCoinDenoms []string,
) *MsgHarvest {
	return &MsgHarvest{
		Farmer:            farmer.String(),
		StakingCoinDenoms: stakingCoinDenoms,
	}
}

func (msg MsgHarvest) Route() string { return RouterKey }

func (msg MsgHarvest) Type() string { return TypeMsgHarvest }

func (msg MsgHarvest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address %q: %v", msg.Farmer, err)
	}
	if len(msg.StakingCoinDenoms) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "staking coin denoms must be provided at least one")
	}
	for _, denom := range msg.StakingCoinDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return err
		}
	}
	return nil
}

func (msg MsgHarvest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgHarvest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgHarvest) GetFarmer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgRemovePlan creates a new MsgRemovePlan.
func NewMsgRemovePlan(
	creator sdk.AccAddress,
	planId uint64,
) *MsgRemovePlan {
	return &MsgRemovePlan{
		Creator: creator.String(),
		PlanId:  planId,
	}
}

func (msg MsgRemovePlan) Route() string { return RouterKey }

func (msg MsgRemovePlan) Type() string { return TypeMsgRemovePlan }

func (msg MsgRemovePlan) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address %q: %v", msg.Creator, err)
	}
	if msg.PlanId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "plan id must not be 0")
	}
	return nil
}

func (msg MsgRemovePlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRemovePlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgRemovePlan) GetCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgAdvanceEpoch creates a new MsgAdvanceEpoch.
func NewMsgAdvanceEpoch(requesterAcc sdk.AccAddress) *MsgAdvanceEpoch {
	return &MsgAdvanceEpoch{
		Requester: requesterAcc.String(),
	}
}

func (msg MsgAdvanceEpoch) Route() string { return RouterKey }

func (msg MsgAdvanceEpoch) Type() string { return TypeMsgAdvanceEpoch }

func (msg MsgAdvanceEpoch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Requester); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid requester address %q: %v", msg.Requester, err)
	}
	return nil
}

func (msg MsgAdvanceEpoch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAdvanceEpoch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
