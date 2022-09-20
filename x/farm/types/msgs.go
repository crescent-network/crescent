package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreatePrivatePlan)(nil)
	_ sdk.Msg = (*MsgFarm)(nil)
	_ sdk.Msg = (*MsgUnfarm)(nil)
	_ sdk.Msg = (*MsgHarvest)(nil)
)

// Message types for the module
const (
	TypeMsgCreatePrivatePlan = "create_private_plan"
	TypeMsgFarm              = "farm"
	TypeMsgUnfarm            = "unfarm"
	TypeMsgHarvest           = "harvest"
)

// NewMsgCreatePrivatePlan returns a new MsgCreatePrivatePlan.
func NewMsgCreatePrivatePlan(
	creatorAddr sdk.AccAddress, description string, rewardAllocations []RewardAllocation,
	startTime, endTime time.Time) *MsgCreatePrivatePlan {
	return &MsgCreatePrivatePlan{
		Creator:           creatorAddr.String(),
		Description:       description,
		RewardAllocations: rewardAllocations,
		StartTime:         startTime,
		EndTime:           endTime,
	}
}

func (msg MsgCreatePrivatePlan) Route() string { return RouterKey }

func (msg MsgCreatePrivatePlan) Type() string { return TypeMsgCreatePrivatePlan }

func (msg MsgCreatePrivatePlan) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}
	// Create a dummy plan with valid fields and utilize Validate() method
	// for user-provided data.
	validAddr := RewardsPoolAddress // Chose random valid address
	dummyPlan := NewPlan(
		1, msg.Description, validAddr, validAddr,
		msg.RewardAllocations, msg.StartTime, msg.EndTime, true)
	if err := dummyPlan.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (msg MsgCreatePrivatePlan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePrivatePlan) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePrivatePlan) GetCreatorAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgFarm returns a new MsgFarm.
func NewMsgFarm(farmerAddr sdk.AccAddress, coin sdk.Coin) *MsgFarm {
	return &MsgFarm{
		Farmer: farmerAddr.String(),
		Coin:   coin,
	}
}

func (msg MsgFarm) Route() string { return RouterKey }

func (msg MsgFarm) Type() string { return TypeMsgFarm }

func (msg MsgFarm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if err := msg.Coin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid coin: %v", err)
	}
	if !msg.Coin.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "non-positive coin: %s", msg.Coin)
	}
	return nil
}

func (msg MsgFarm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgFarm) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgFarm) GetFarmerAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgUnfarm returns a new MsgUnfarm.
func NewMsgUnfarm(farmerAddr sdk.AccAddress, coin sdk.Coin) *MsgUnfarm {
	return &MsgUnfarm{
		Farmer: farmerAddr.String(),
		Coin:   coin,
	}
}

func (msg MsgUnfarm) Route() string { return RouterKey }

func (msg MsgUnfarm) Type() string { return TypeMsgUnfarm }

func (msg MsgUnfarm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if err := msg.Coin.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid coin: %v", err)
	}
	if !msg.Coin.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "non-positive coin: %s", msg.Coin)
	}
	return nil
}

func (msg MsgUnfarm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgUnfarm) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgUnfarm) GetFarmerAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgHarvest returns a new MsgHarvest.
func NewMsgHarvest(farmerAddr sdk.AccAddress, denom string) *MsgHarvest {
	return &MsgHarvest{
		Farmer: farmerAddr.String(),
		Denom:  denom,
	}
}

func (msg MsgHarvest) Route() string { return RouterKey }

func (msg MsgHarvest) Type() string { return TypeMsgHarvest }

func (msg MsgHarvest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Farmer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid farmer address: %v", err)
	}
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom: %v", err)
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

func (msg MsgHarvest) GetFarmerAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Farmer)
	if err != nil {
		panic(err)
	}
	return addr
}
