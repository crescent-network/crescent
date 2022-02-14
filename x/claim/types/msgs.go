package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgClaim)(nil)
)

// Message types for the claim module
const (
	TypeMsgClaim = "claim"
)

// NewMsgClaim creates a new MsgClaim.
func NewMsgClaim(recipient sdk.AccAddress, actionType ActionType) *MsgClaim {
	return &MsgClaim{
		Recipient:  recipient.String(),
		ActionType: actionType,
	}
}

func (msg MsgClaim) Route() string { return RouterKey }

func (msg MsgClaim) Type() string { return TypeMsgClaim }

func (msg MsgClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient address: %v", err)
	}
	switch msg.ActionType {
	case ActionTypeDeposit, ActionTypeSwap, ActionTypeFarming:
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid action type: %s", msg.ActionType.String())
	}
	return nil
}

func (msg MsgClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgClaim) GetRecipient() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		panic(err)
	}
	return addr
}
