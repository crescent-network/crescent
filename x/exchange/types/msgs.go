package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgCreateMarket)(nil)
	_ sdk.Msg = (*MsgPlaceLimitOrder)(nil)
	_ sdk.Msg = (*MsgPlaceMarketOrder)(nil)
	_ sdk.Msg = (*MsgCancelOrder)(nil)
	_ sdk.Msg = (*MsgSwapExactAmountIn)(nil)
)

// Message types for the module
const (
	TypeMsgCreateMarket      = "create_market"
	TypeMsgPlaceLimitOrder   = "place_limit_order"
	TypeMsgPlaceMarketOrder  = "place_market_order"
	TypeMsgCancelOrder       = "cancel_order"
	TypeMsgSwapExactAmountIn = "swap_exact_amount_in"
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
	if err := sdk.ValidateDenom(msg.BaseDenom); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid base denom: %v", err)
	}
	if err := sdk.ValidateDenom(msg.QuoteDenom); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid quote denom: %v", err)
	}
	return nil
}

func NewMsgPlaceLimitOrder(
	senderAddr sdk.AccAddress, marketId uint64, isBuy bool,
	price sdk.Dec, quantity sdk.Int, isBatch bool, lifespan time.Duration) *MsgPlaceLimitOrder {
	return &MsgPlaceLimitOrder{
		Sender:   senderAddr.String(),
		MarketId: marketId,
		IsBuy:    isBuy,
		Price:    price,
		Quantity: quantity,
		IsBatch:  isBatch,
		Lifespan: lifespan,
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
	if msg.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market is must not be 0")
	}
	if msg.Price.LT(MinPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is lower than the min price; %s < %s", msg.Price, MinPrice)
	}
	if msg.Price.GT(MaxPrice) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "price is higher than the max price; %s < %s", msg.Price, MaxPrice)
	}
	if _, valid := ValidateTickPrice(msg.Price); !valid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid tick price: %s", msg.Price)
	}
	if !msg.Quantity.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be positive: %s", msg.Quantity)
	}
	if msg.Lifespan < 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "lifespan must not be negative: %v", msg.Lifespan)
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
	if msg.MarketId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market is must not be 0")
	}
	if !msg.Quantity.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quantity must be positive: %s", msg.Quantity)
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
	if msg.OrderId == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "order id must not be 0")
	}
	return nil
}

func NewMsgSwapExactAmountIn(senderAddr sdk.AccAddress, routes []uint64, input, minOutput sdk.Coin) *MsgSwapExactAmountIn {
	return &MsgSwapExactAmountIn{
		Sender:    senderAddr.String(),
		Routes:    routes,
		Input:     input,
		MinOutput: minOutput,
	}
}

func (msg MsgSwapExactAmountIn) Route() string { return RouterKey }
func (msg MsgSwapExactAmountIn) Type() string  { return TypeMsgSwapExactAmountIn }

func (msg MsgSwapExactAmountIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapExactAmountIn) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgSwapExactAmountIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if len(msg.Routes) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "routes must not be empty")
	}
	for _, marketId := range msg.Routes {
		if marketId == 0 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "market id must not be 0")
		}
	}
	if err := msg.Input.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid input: %v", err)
	}
	if !msg.Input.IsPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "input must be positive: %s", msg.Input)
	}
	if err := msg.MinOutput.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid min output: %v", err)
	}
	return nil
}