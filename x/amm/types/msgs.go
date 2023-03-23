package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgAddLiquidity)(nil)
	_ sdk.Msg = (*MsgRemoveLiquidity)(nil)
)

// Message types for the module
const (
	TypeMsgCreatePool      = "create_pool"
	TypeMsgAddLiquidity    = "add_liquidity"
	TypeMsgRemoveLiquidity = "remove_liquidity"
)

func NewMsgCreatePool(senderAddr sdk.AccAddress) *MsgCreatePool {
	return &MsgCreatePool{
		Sender: senderAddr.String(),
	}
}

func (msg MsgCreatePool) Route() string { return RouterKey }
func (msg MsgCreatePool) Type() string  { return TypeMsgCreatePool }

func (msg MsgCreatePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgAddLiquidity(senderAddr sdk.AccAddress) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Sender: senderAddr.String(),
	}
}

func (msg MsgAddLiquidity) Route() string { return RouterKey }
func (msg MsgAddLiquidity) Type() string  { return TypeMsgAddLiquidity }

func (msg MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgAddLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgRemoveLiquidity(senderAddr sdk.AccAddress) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Sender: senderAddr.String(),
	}
}

func (msg MsgRemoveLiquidity) Route() string { return RouterKey }
func (msg MsgRemoveLiquidity) Type() string  { return TypeMsgRemoveLiquidity }

func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgRemoveLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}
