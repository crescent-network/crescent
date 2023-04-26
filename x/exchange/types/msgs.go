package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreateMarket)(nil)
	_ sdk.Msg = (*MsgPlaceLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceMarketOrder)(nil)
	_ sdk.Msg = (*MsgCancelOrder)(nil)
	_ sdk.Msg = (*MsgSwapExactIn)(nil)
)

// Message types for the module
const (
	TypeMsgCreateMarket     = "create_market"
	TypeMsgPlaceLimitOrder  = "place_limit_order"
	TypeMsgPlaceMarketOrder = "place_market_order"
	TypeMsgCancelOrder      = "cancel_order"
	TypeMsgSwapExactIn      = "swap_exact_in"
)

func NewMsgCreateMarket(
	senderAddr sdk.AccAddress, baseDenom, quoteDenom string) *MsgCreateMarket {
	return &MsgCreateMarket{
		Sender:     senderAddr.String(),
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
}

func (msg MsgCreateMarket) Route() string { return RouterKey }
func (msg MsgCreateMarket) Type() string  { return TypeMsgCreateMarket }

func (msg MsgCreateMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateMarket) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreateMarket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgPlaceLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64,
	isBuy bool, price sdk.Dec, quantity sdk.Int) *MsgPlaceLimitOrder {
	return &MsgPlaceLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: quantity,
	}
}

func (msg MsgPlaceLimitOrder) Route() string { return RouterKey }
func (msg MsgPlaceLimitOrder) Type() string  { return TypeMsgPlaceLimitOrder }

func (msg MsgPlaceLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceLimitOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgPlaceMarketOrder(
	senderAddr sdk.AccAddress, marketId uint64,
	isBuy bool, quantity sdk.Int) *MsgPlaceMarketOrder {
	return &MsgPlaceMarketOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Quantity: quantity,
	}
}

func (msg MsgPlaceMarketOrder) Route() string { return RouterKey }
func (msg MsgPlaceMarketOrder) Type() string  { return TypeMsgPlaceMarketOrder }

func (msg MsgPlaceMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgPlaceMarketOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgPlaceMarketOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	return nil
}

func NewMsgCancelOrder(senderAddr sdk.AccAddress, orderId uint64) *MsgCancelOrder {
	return &MsgCancelOrder{
		Sender:  senderAddr.String(),
		OrderId: orderId,
	}
}

func (msg MsgCancelOrder) Route() string { return RouterKey }
func (msg MsgCancelOrder) Type() string  { return TypeMsgCancelOrder }

func (msg MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCancelOrder) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCancelOrder) ValidateBasic() error {
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
