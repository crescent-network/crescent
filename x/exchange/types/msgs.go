package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreateSpotMarket)(nil)
	_ sdk.Msg = (*MsgPlaceSpotLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceSpotMarketOrder)(nil)
	_ sdk.Msg = (*MsgCancelSpotOrder)(nil)
	_ sdk.Msg = (*MsgSwapExactIn)(nil)
)

// Message types for the module
const (
	TypeMsgCreateSpotMarket     = "create_spot_market"
	TypeMsgPlaceSpotLimitOrder  = "place_spot_limit_order"
	TypeMsgPlaceSpotMarketOrder = "place_spot_market_order"
	TypeMsgCancelSpotOrder      = "cancel_spot_order"
	TypeMsgSwapExactIn          = "swap_exact_in"
)

func NewMsgCreateSpotMarket(
	senderAddr sdk.AccAddress, baseDenom, quoteDenom string) *MsgCreateSpotMarket {
	return &MsgCreateSpotMarket{
		Sender:     senderAddr.String(),
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
}

func (msg MsgCreateSpotMarket) Route() string { return RouterKey }
func (msg MsgCreateSpotMarket) Type() string  { return TypeMsgCreateSpotMarket }

func (msg MsgCreateSpotMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateSpotMarket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateSpotMarket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgPlaceSpotLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, quantity sdk.Int) *MsgPlaceSpotLimitOrder {
	return &MsgPlaceSpotLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: quantity,
	}
}

func (msg MsgPlaceSpotLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceSpotLimitOrder) Type() string  { return TypeMsgPlaceSpotLimitOrder }

func (msg MsgPlaceSpotLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceSpotLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceSpotLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgPlaceSpotMarketOrder(
	senderAddr sdk.AccAddress, marketId uint64,
	isBuy bool, quantity sdk.Int) *MsgPlaceSpotMarketOrder {
	return &MsgPlaceSpotMarketOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Quantity: quantity,
	}
}

func (msg MsgPlaceSpotMarketOrder) Route() string { return RouterKey }
func (msg MsgPlaceSpotMarketOrder) Type() string  { return TypeMsgPlaceSpotMarketOrder }

func (msg MsgPlaceSpotMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceSpotMarketOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceSpotMarketOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgCancelSpotOrder(senderAddr sdk.AccAddress, orderId uint64) *MsgCancelSpotOrder {
	return &MsgCancelSpotOrder{
		Sender:  senderAddr.String(),
		OrderId: orderId,
	}
}

func (msg MsgCancelSpotOrder) Route() string { return RouterKey }
func (msg MsgCancelSpotOrder) Type() string  { return TypeMsgCancelSpotOrder }

func (msg MsgCancelSpotOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelSpotOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelSpotOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgSwapExactIn(senderAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.Coin) *MsgSwapExactIn {
	return &MsgSwapExactIn{
		Sender:    senderAddr.String(),
		Routes:    routes,
		Input:     input,
		MinOutput: minOutput,
	}
}

func (msg MsgSwapExactIn) Route() string { return RouterKey }
func (msg MsgSwapExactIn) Type() string  { return TypeMsgSwapExactIn }

func (msg MsgSwapExactIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapExactIn) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgSwapExactIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}
